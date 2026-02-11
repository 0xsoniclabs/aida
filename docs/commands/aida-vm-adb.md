# Aida Archive Block Processing (aida-vm-adb)

## Overview
`aida-vm-adb` is a tool for simulating block processing using real-world historical data (mainnet/testnet) stored in an ArchiveDB. It allows for high-fidelity replay of historical states to validate execution and performance.

## Build
To build the `aida-vm-adb` application, run:
```shell
make aida-vm-adb
```
The executable will be located at `build/aida-vm-adb`.

## Run
```shell
./build/aida-vm-adb --substate-db path/to/substatedb --db-src path/to/statedb/with/archive <blockNumFirst> <blockNumLast>
```
Executes transactions from block `<blockNumFirst>` to `<blockNumLast>` using the historic data in the provided archive. Each transaction loads the historic state of its block and executes the transaction on it in read-only mode.

### Options
```
    --cpu-profile       records a CPU profile for the replay to be inspected using `pprof`
    --chainid           sets the chain-id (useful if recording from testnet)
    --aida-db           set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --db-src            sets the directory contains source state DB data
    --validate-tx       validate the effects of each transaction
    --shadow-db         use this flag when using an existing [ShadowDb](Terminology)
    --vm-impl           select between `geth` and `lfvm`
    --workers           number of worker threads that execute in parallel
    --substate-db       sets directory containing substate database
    --log               level of the logging of the app action ("critical", "error", "warning", "notice", "info", "debug")
```

## Execution Flow

Uses the standard [Provider](../architecture/Providers.md) → [Processor](../architecture/Processors.md) → [Extensions](../architecture/extensions/README.md) pipeline.

- **Provider:** SubstateProvider
- **Processor:** ArchiveDbTxProcessor
- **Parallelism:** BlockLevel, configurable workers (parallel capable via `--workers`)

**Extensions (in registration order):**

1. CpuProfiler
2. ArchivePrepper
3. ParentBlockHashProcessor
4. ProgressLogger
5. ErrorLogger
6. ArchiveDbValidator (WorldState + Receipt) — uses `MakeArchiveDbValidator`, not `MakeLiveDbValidator`
7. StateDbManager *(if no external stateDb)*
8. ArchiveBlockChecker *(if no external stateDb)*
9. DbLogger *(if no external stateDb)*