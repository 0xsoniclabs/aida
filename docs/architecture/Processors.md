# Processors

Processors execute transactions against a StateDB. They sit between [Providers](Providers.md) (data
source) and [Extensions](extensions/README.md) (pre/post hooks) in the executor pipeline.

## Processor Interface

```go
type Processor[T any] interface {
    Process(State[T], *Context) error
}
```

`Process` receives immutable transaction state and a mutable context (containing the StateDB) and
applies the transaction.

## Implementations

| Processor | Constructor | Target DB | Use Case |
|-----------|------------|-----------|----------|
| LiveDbTxProcessor | `MakeLiveDbTxProcessor(cfg)` | Live StateDB | Standard transaction replay |
| ArchiveDbTxProcessor | `MakeArchiveDbTxProcessor(cfg)` | Archive StateDB | Archive validation |
| EthTestProcessor | `MakeEthTestProcessor(cfg)` | Live StateDB | Ethereum reference tests |

### LiveDbTxProcessor

Processes transactions against the **live** (current) StateDB via `ctx.State`. This is the default
processor for most Aida workflows.

### ArchiveDbTxProcessor

Processes transactions against the **archive** StateDB via `ctx.Archive`. Used for validating that
archive states produce correct results when replaying historical transactions.

### EthTestProcessor

Processes [Ethereum reference state tests](https://github.com/ethereum/tests). Unlike the other
processors, EthTestProcessor **pre-filters invalid transactions** before calling the EVM — because
invalid transactions would update the sender's nonce and corrupt the expected state hash.

Pre-execution checks:
- **Blob gas limit** — must not exceed the fork's per-block maximum (Cancun = 6 blobs,
  Prague/Osaka = 9, pre-Cancun = 0)
- **Transaction encoding** — raw bytes must unmarshal as a valid transaction
- **Sender recovery** — the sender address must be recoverable from the signature

If any check fails, the result is recorded as an error but execution continues — the
[validators](extensions/validator.md) decide whether the test passed or failed.

## TxProcessor (Shared Base)

All processors delegate to a shared `TxProcessor` created by the `MakeTxProcessor(cfg)` factory. The
factory selects the backend based on `cfg.EvmImpl`: values `""`, `"opera"`, or `"ethereum"` produce
an `aidaProcessor`; any other value creates a `toscaProcessor` with the specified interpreter.

### Error Handling

Each processor uses `isErrFatal()` to decide whether an error should halt execution:

- If `ContinueOnFailure` is **false** (default), all errors are fatal.
- If `ContinueOnFailure` is **true**, errors are sent to the `ErrorInput` channel (collected by
  [ErrorLogger](extensions/logger.md)) and execution continues.
- `MaxNumErrors` sets an upper bound — once reached, subsequent errors become fatal again. A value
  of `0` means unlimited.

The error counter is atomic, making it safe for parallel execution.

### Backends

The TxProcessor supports two backends:

#### aidaProcessor

Uses go-ethereum's `core.ApplyMessage` with the standard EVM implementation. This is the default
backend.

#### toscaProcessor

Uses [Tosca's](../Terminology.md#tosca) processor with a configurable interpreter. Supported interpreters include:
- `floria` (an experimental block processor)
- `geth`
- `opera`
- and others

Maps EVM revision identifiers to Tosca's revision enum (Istanbul → Osaka).

### Pseudo-Transactions

The TxProcessor also handles special pseudo-transactions that don't go through the EVM:
- **SFC** (Special Function Contract) calls
- **Genesis** state initialization
- **Lachesis-Opera transition** state changes
