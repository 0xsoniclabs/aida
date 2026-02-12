# Aida Archive Block Processing (aida-vm-adb)

## Overview

`aida-vm-adb` is a tool for simulating block processing using real-world historical data
(mainnet/testnet) stored in an ArchiveDB. It allows for high-fidelity replay of historical states to
validate execution and performance.

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
Executes transactions from block `<blockNumFirst>` to `<blockNumLast>` using the historic data in
the provided archive. Each transaction loads the historic state of its block and executes the
transaction on it in read-only mode.

### Options

```
    --cpu-profile       records a CPU profile for the replay to be inspected using `pprof`
    --chainid           sets the chain-id (useful if recording from testnet)
    --aida-db           set [aida-db](../Terminology.md) directory (substate, updateset, deleted accounts)
    --db-src            sets the directory contains source state DB data
    --validate-tx       validate the effects of each transaction
    --shadow-db         use this flag when using an existing [ShadowDb](../Terminology.md)
    --vm-impl           select between `geth` and `lfvm`
    --workers           number of worker threads that execute in parallel
    --substate-db       sets directory containing substate database
    --log               level of the logging of the app action ("critical", "error", "warning", "notice", "info", "debug")
```
