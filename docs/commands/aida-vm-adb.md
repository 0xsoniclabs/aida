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
