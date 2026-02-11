# Processors

Processors execute transactions against a StateDB. They sit between [Providers](Providers.md) (data source) and [Extensions](extensions/README.md) (pre/post hooks) in the executor pipeline.

## Processor Interface

```go
type Processor[T any] interface {
    Process(State[T], *Context) error
}
```

`Process` receives immutable transaction state and a mutable context (containing the StateDB) and applies the transaction.

## Implementations

| Processor | Constructor | Target DB | Use Case |
|-----------|------------|-----------|----------|
| LiveDbTxProcessor | `MakeLiveDbTxProcessor(cfg)` | Live StateDB | Standard transaction replay |
| ArchiveDbTxProcessor | `MakeArchiveDbTxProcessor(cfg)` | Archive StateDB | Archive validation |
| EthTestProcessor | `MakeEthTestProcessor(cfg)` | Live StateDB | Ethereum reference tests |

### LiveDbTxProcessor

Processes transactions against the **live** (current) StateDB via `ctx.State`. This is the default processor for most Aida workflows.

### ArchiveDbTxProcessor

Processes transactions against the **archive** StateDB via `ctx.Archive`. Used for validating that archive states produce correct results when replaying historical transactions.

### EthTestProcessor

Processes Ethereum reference test transactions. Before execution, validates:
- Blob gas parameters
- Transaction byte encoding
- Sender recovery

Before execution, validates fork-specific blob gas limits — the maximum number of blob transactions per block differs by fork: **Cancun = 6**, **Prague/Osaka = 9** (pre-Cancun forks allow 0). Processing errors are not treated as fatal — the [validators](extensions/validator.md) decide whether results are acceptable.

## TxProcessor (Shared Base)

All processors delegate to a shared `TxProcessor` created by the `MakeTxProcessor(cfg)` factory. The factory selects the backend based on `cfg.EvmImpl`: values `""`, `"opera"`, or `"ethereum"` produce an `aidaProcessor`; any other value creates a `toscaProcessor` with the specified interpreter.

### Error Handling

Each processor uses `isErrFatal()` to decide whether an error should halt execution:

- If `ContinueOnFailure` is **false** (default), all errors are fatal.
- If `ContinueOnFailure` is **true**, errors are sent to the `ErrorInput` channel (collected by [ErrorLogger](extensions/logger.md)) and execution continues.
- `MaxNumErrors` sets an upper bound — once reached, subsequent errors become fatal again. A value of `0` means unlimited.

The error counter is atomic, making it safe for parallel execution.

### Backends

The TxProcessor supports two backends:

#### aidaProcessor

Uses go-ethereum's `core.ApplyMessage` with the standard EVM implementation. This is the default backend.

#### toscaProcessor

Uses Tosca's processor with a configurable interpreter. Supported interpreters include:
- `floria`
- `geth`
- `opera`
- and others

Maps EVM revision identifiers to Tosca's revision enum (Istanbul → Osaka).

### Pseudo-Transactions

The TxProcessor also handles special pseudo-transactions that don't go through the EVM:
- **SFC** (Special Function Contract) calls
- **Genesis** state initialization
- **Lachesis-Opera transition** state changes

## See Also

- [Providers](Providers.md) — data sources feeding the processor
- [Extensions](extensions/README.md) — hooks around processing
