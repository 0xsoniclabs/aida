# StateDB Extensions

Extensions managing StateDB lifecycle, event emission, and state preparation. Source: `executor/extension/statedb/`.

## Overview

| Extension | Purpose | Hooks |
|-----------|---------|-------|
| [StateDbManager](#statedbmanager) | Opens/creates/closes StateDB | PreRun, PostRun |
| [LiveDbBlockChecker](#livedbblockchecker) | Checks LiveDB block alignment | PreRun |
| [ArchiveBlockChecker](#archiveblockchecker) | Checks archive block alignment | PreRun |
| [BlockEventEmitter](#blockeventemitter) | Emits BeginBlock/EndBlock | PreBlock, PostBlock |
| [TransactionEventEmitter](#transactioneventemitter) | Emits BeginTransaction/EndTransaction | PreTransaction, PostTransaction |
| [TxGeneratorBlockEventEmitter](#txgeneratorblockeventemitter) | Block events for tx generator | PreTransaction, PostTransaction |
| [StateDbPrepper](#statedbprepper) | Calls PrepareSubstate on StateDB | PreTransaction |
| [StatePrepper](#stateprepper) | Feeds substate input state to DB | PreTransaction |
| [ArchivePrepper](#archiveprepper) | Gets archive from StateDB | PreBlock, PreTransaction, PostTransaction, PostBlock |
| [TemporaryArchivePrepper](#temporaryarchiveprepper) | Per-transaction temporary archive | PreTransaction, PostTransaction |
| [TemporaryStatePrepper](#temporarystateprepper) | Per-transaction fresh StateDB | PreTransaction, PostTransaction |
| [SyncPeriodEmitter](#syncperiodemitter) | BeginSyncPeriod/EndSyncPeriod | PreBlock |
| [ParentBlockHashProcessor](#parentblockhashprocessor) | Saves parent block hash (EIP-2935) | PreBlock |
| [StateDbCorrectorPreScope](#statedbcorrectorprescope) | Fixes known blockchain exceptions | PreRun, PreBlock, PreTransaction, PostBlock |
| [EthStateTestDbPrepper](#ethstatetestdbprepper) | Fresh StateDB per Ethereum test | PreBlock |
| [EthStateScopeTestEventEmitter](#ethstatescopestesteventemitter) | Block/tx events for Ethereum tests | PreTransaction, PostTransaction |

## Details

### StateDbManager

Opens, creates, or reuses a StateDB based on configuration flags (`--keep-db`, `--db-src`, `--db-impl`, `--db-variant`). Closes the database on PostRun. This is typically the first extension registered, ensuring the StateDB exists for all subsequent extensions.

### LiveDbBlockChecker

Validates that an existing LiveDB's block height is aligned with the requested processing range. Runs during PreRun and aborts early if there's a mismatch.

### ArchiveBlockChecker

Similar to LiveDbBlockChecker but validates archive state block alignment.

### BlockEventEmitter

Calls `BeginBlock()` on the StateDB at PreBlock and `EndBlock()` at PostBlock. Required for StateDB implementations that track block boundaries.

### TransactionEventEmitter

Calls `BeginTransaction()` at PreTransaction and `EndTransaction()` at PostTransaction on the StateDB.

### TxGeneratorBlockEventEmitter

Variant of BlockEventEmitter designed for the transaction generator workflow. Handles the first block specially (different initialization).

### StateDbPrepper

Calls `PrepareSubstate()` on the StateDB before each transaction, feeding it the data needed by in-memory DB implementations to process the upcoming transaction.

### StatePrepper

Related to StateDbPrepper. Feeds the substate's input state (accounts, storage) into the StateDB before transaction processing.

### ArchivePrepper

Retrieves the archive view from the StateDB at PreBlock (assigning it to `ctx.Archive`), manages archive transaction boundaries (BeginTransaction at PreTransaction, EndTransaction at PostTransaction), and releases the archive at PostBlock.

### TemporaryArchivePrepper

Creates a temporary archive snapshot per transaction and releases it afterward. Designed for RPC replay workflows where each request needs an independent archive view.

### TemporaryStatePrepper

Creates a completely fresh StateDB for each transaction. Used in off-chain or in-memory modes where transactions are processed independently.

### SyncPeriodEmitter

Manages `BeginSyncPeriod()` and `EndSyncPeriod()` calls on the StateDB at block boundaries. Sync periods are a concept from the Sonic/Opera chain architecture.

### ParentBlockHashProcessor

Saves the parent block hash into the StateDB per [EIP-2935](https://eips.ethereum.org/EIPS/eip-2935) (Prague hard fork and later). Runs at PreBlock.

### StateDbCorrectorPreScope

Applies known blockchain exception fixes to the state. Initializes the exception database at PreRun, loads and applies pre-block fixes at PreBlock (including catch-up for skipped blocks), applies per-transaction fixes at PreTransaction, and handles post-block fixes (e.g., trailing skipped transactions, miner rewards) at PostBlock. Handles edge cases where the canonical chain had irregular state transitions.

### EthStateTestDbPrepper

Creates a fresh StateDB for each Ethereum reference state test at PreBlock, ensuring test isolation.

### EthStateScopeTestEventEmitter

Emits block and transaction lifecycle events adapted for Ethereum state test execution. Runs at PreTransaction and PostTransaction.

## See Also

- [Extension System Overview](README.md)
- [Validators](validator.md) — validate state after these extensions prepare it
- [Primers](primer.md) — prime state before processing
