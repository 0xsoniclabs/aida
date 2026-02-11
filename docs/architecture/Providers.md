# Providers

Providers are the data sources that feed transactions into Aida's [executor pipeline](extensions/README.md). Each provider implements a generic interface that streams transaction data over a block range to a consumer function.

## Provider Interface

```go
type Provider[T any] interface {
    Run(from int, to int, consumer Consumer[T]) error
    Close()
}

type Consumer[T any] func(TransactionInfo[T]) error

type TransactionInfo[T any] struct {
    Block       int
    Transaction int
    Data        T
}
```

`Run` iterates over transactions in the block range `[from, to)` and calls the consumer for each one. `Close` releases resources.

## Implementations

| Provider | Source | Data Type | Description |
|----------|--------|-----------|-------------|
| SubstateProvider | `executor/substate_provider.go` | `txcontext.TxContext` | Replays recorded mainnet transactions |
| EthTestProvider | `executor/eth_test_provider.go` | `txcontext.TxContext` | Runs Ethereum reference state tests |
| RpcRequestProvider | `executor/rpc_request_provider.go` | `*rpc.RequestAndResults` | Replays recorded RPC requests |
| NormaTxProvider | `executor/norma_tx_provider.go` | `txcontext.TxContext` | Generates synthetic transactions |

### SubstateProvider

Opens a SubstateDB and iterates over all recorded substates in the block range `[from, to)`. Each substate is wrapped in a `txcontext.TxContext` via `substatecontext.NewTxContext`. Uses parallel decoders for performance.

This is the primary provider for mainnet transaction replay workflows (`aida-vm`, `aida-vm-adb`, etc.).

### EthTestProvider

Loads Ethereum reference state tests and splits them into individual test cases via `TestCaseSplitter`. Each test case is assigned to a synthetic block number (`2 + i`). The `from` and `to` parameters are explicitly ignored (the `Run` signature accepts them but discards both values) — all loaded tests are always processed regardless of the requested block range.

Used by the `aida-vm-ethtest` workflow for Ethereum compatibility testing.

### RpcRequestProvider

Reads recorded RPC request/response pairs from files. Supports both a single file and a directory of files. Requests of type `getLogs` are skipped.

Used by `aida-rpc` for RPC replay and comparison against recorded responses.

### NormaTxProvider

Generates synthetic transactions using the Norma framework. On startup it:

1. Creates a treasure account with initial funds
2. Deploys contracts (ERC-20, counter, store, uniswap)
3. Generates transactions until the block range is exhausted

Internally uses a `fakeRpcClient` to bridge Norma's RPC-based transaction generation interface to Aida's consumer-based pipeline.

Primarily used for synthetic load testing and benchmarking without requiring real mainnet data.

## See Also

- [Processors](Processors.md) — how transactions are executed
- [Extensions](extensions/README.md) — hooks around transaction processing
