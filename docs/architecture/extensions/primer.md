# Primer Extensions

Extensions that initialize StateDB with pre-existing state. Source: `executor/extension/primer/`.

## Overview

| Extension | Constructor | Purpose | Hooks |
|-----------|------------|---------|-------|
| StateDbPrimer | `MakeStateDbPrimer(cfg)` | Fast-forwards StateDB to target block | PreRun |
| EthStateTestDbPrimer | — | Primes with Ethereum test initial state | PreTransaction |
| TxPrimer | — | Primes with substate input per transaction | PreTransaction |

## Details

### StateDbPrimer

Primes the StateDB with **block-level update-sets** from AidaDb to fast-forward it to the target starting block without replaying all prior transactions. This is essential for starting replay from an arbitrary block height. Runs at PreRun. Typically used in **mutually exclusive** workflows from TxPrimer — StateDbPrimer is for block-level fast-forward, while TxPrimer handles per-transaction injection.

### EthStateTestDbPrimer

Loads the initial account allocations (balances, nonces, code, storage) from an Ethereum reference test into the StateDB before each test transaction.

### TxPrimer

Primes the StateDB with the substate's **transaction-level input state** (expected pre-state) before each transaction. Used when the StateDB doesn't already contain the required state. Typically used in workflows where transactions are processed independently (e.g., in-memory mode), as opposed to StateDbPrimer's block-level approach.

## See Also

- [Extension System Overview](README.md)
- [StateDB Extensions](statedb.md) — manages the StateDB being primed
