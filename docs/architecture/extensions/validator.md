# Validator Extensions

Extensions that validate processing results for correctness. Source:
`executor/extension/validator/`.

## Overview

| Extension | Purpose |
|-----------|---------|
| StateHashValidator | Validates state hash against AidaDb |
| ShadowDbValidator | Compares primary and shadow StateDBs (proxy pattern) |
| LiveDbValidator | Validates live StateDB output vs expected |
| RpcComparator | Compares RPC results vs recorded responses |
| EthStateTestErrorValidator | Validates error expectations |
| EthStateTestStateHashValidator | Validates post-state hash |
| EthStateTestLogHashValidator | Validates log hash |
| EthereumPreTransactionUpdater | Fixes exceptions before processing |
| EthereumPostTransactionUpdater | Fixes exceptions after processing |

## Details

### StateHashValidator

After each block, computes the state hash of the StateDB and compares it against the expected hash
stored in AidaDb. Detects state drift caused by incorrect transaction processing.

### ShadowDbValidator

Uses the **proxy pattern** — wraps the StateDB in a shadow proxy that mirrors all operations to both
a primary and shadow StateDB implementation. See [ShadowDb](../../Terminology.md#shadowdb) for details. After each
**block** (PostBlock), compares state hashes between the two and checks for accumulated errors.
Useful for validating new StateDB implementations against a known-good reference.

### LiveDbValidator

**Constructor:** `MakeLiveDbValidator(cfg)`

Validates the live StateDB's state against expected substates. At **PreTransaction**, validates (or
optionally overwrites) the input world state. At **PostTransaction**, validates the output world
state and optionally the transaction receipt. Checks account balances, nonces, code, and storage
values.

### RpcComparator

Compares the results of replayed RPC requests against the originally recorded responses. Reports
mismatches for debugging RPC compatibility.

### EthStateTestErrorValidator

For Ethereum reference tests, validates that the transaction produced the expected error (or
succeeded when expected to succeed).

### EthStateTestStateHashValidator

Validates that the post-transaction state root hash matches the expected hash from the Ethereum test
specification.

### EthStateTestLogHashValidator

Validates that the hash of transaction logs matches the expected log hash from the Ethereum test
specification.

### EthereumPreTransactionUpdater

Applies known Ethereum mainnet exception fixes to the state **before** transaction processing.
Handles irregular state transitions from historical hard forks or consensus bugs.

### EthereumPostTransactionUpdater

Applies known Ethereum mainnet exception fixes **after** transaction processing. This is a workaround
for correting any unknow changes outside a transaction.

### Shared Utilities (`utils.go`)

Core validation logic shared across all state validators:
- **SubsetCheck** (`doSubsetValidation`) — verifies expected state is contained in the StateDB
  (default mode)
- **EqualityCheck** — verifies two world states are identical, with detailed diff reporting

These are selected via `cfg.StateValidationMode` and underpin all world-state validation in the
system.
