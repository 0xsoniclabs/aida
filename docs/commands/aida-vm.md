# Aida EVM Evaluation Tool (aida-vm)

## Overview

`aida-vm` is a tool for simulating block processing with experimental StateDB implementations and/or
Virtual Machines. It provides a flexible environment for benchmarking and validating different VM
and StateDB configurations.

## Build

To build the `aida-vm` application, run:
```shell
make aida-vm
```
The executable will be located at `build/aida-vm`.

## Run

```shell
./build/aida-vm --aida-db path/to/aida-db --db-impl <geth/carmen/memory/flat> --vm-impl <geth, lfvm> <blockNumFirst> <blockNumLast>
```
This command performs block processing of the specified block range (inclusive). The initial StateDB
is primed using substate from `--aida-db`. During block processing, a transaction calls a virtual
machine which issues a series of StateDB operations to a selected storage system.
