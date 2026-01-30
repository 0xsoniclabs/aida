# Terminology

This document defines common terms and concepts used across the Aida testing infrastructure.

## AidaDb
The central filesystem-based database containing substates, update-sets, and deleted accounts. It serves as the primary data source for offline testing and replay.

## Substate
A minimal fragment of the world-state (accounts, balances, nonces, code, and storage key/value pairs) and the transaction message required to execute a specific transaction or block in isolation.

## UpdateSet (or Update-Set)
A set of state changes generated from a range of blocks. It is used to "fast-forward" or initialize the world state at a specific block height without replaying the entire history.

## Priming
The process of pre-loading a StateDB with the necessary state data (from a Substate or UpdateSet) before execution begins. This ensures the StateDB has the correct context (e.g., account balances, storage values) to process transactions.

## ShadowDb
A wrapper mechanism that performs operations on two StateDbs simultaneously:
1.  **Prime DB**: The main database where operations are executed.
2.  **Shadow DB**: A secondary database where the same operations are mirrored for verification purposes.
Useful for comparing different StateDB implementations or configurations.
