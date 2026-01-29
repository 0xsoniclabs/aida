# Aida EVM Evaluation Tool (aida-vm)

## Overview
`aida-vm` is a tool for simulating block processing with experimental StateDB implementations and/or Virtual Machines. It provides a flexible environment for benchmarking and validating different VM and StateDB configurations.

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
This command performs block processing of the specified block range (inclusive). The initial StateDB is primed using substate from `--aida-db`. During block processing, a transaction calls a virtual machine which issues a series of StateDB operations to a selected storage system.

### Options
```
    --aida-db                  set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --deletion-db              sets the directory containing deleted accounts database
    --update-db                set update-set database directory
    --substate-db              data directory for substate recorder/replayer
    --carmen-schema            select the DB schema used by Carmen's current state DB
    --db-impl                  select state DB implementation
    --db-variant               select a state DB variant
    --db-src                   sets the directory contains source state DB data
    --db-tmp                   sets the temporary directory where to place state DB data; uses system default if empty
    --db-logging               enable logging of all DB operations
    --archive                  set node type to archival mode. If set, the node keep all the EVM state history; otherwise the state history will be pruned.
    --archive-variant          set the archive implementation variant for the selected DB implementation, ignored if not running in archive mode
    --shadow-db                use this flag when using an existing [ShadowDb](Terminology)
    --db-shadow-impl           select state DB implementation to shadow the prime DB implementation
    --db-shadow-variant        select a state DB variant to shadow the prime DB implementation
    --vm-impl                  select VM implementation
    --memory-breakdown         enables printing of memory usage breakdown
    --memory-profile           enables memory allocation profiling
    --profile                  enables profiling
    --cpu-profile              enables CPU profiling
    --random-seed              set random seed
    --prime-threshold          set number of accounts written to stateDB before applying pending state updates
    --prime-random             randomize order of accounts in StateDB priming
    --skip-priming             if set, DB priming should be skipped; most useful with the 'memory' DB implementation
    --update-buffer-size       buffer size for holding update set in MiB
    --chainid                  ChainID for replayer
    --continue-on-failure      continue execute after validation failure detected
    --quiet                    disable progress report
    --sync-period              defines the number of blocks per sync-period
    --keep-db                  if set, statedb is not deleted after run
    --max-transactions         limit the maximum number of processed transactions, default: unlimited
    --validate-tx              enables transaction state validation
    --validate-ws              enables end-state validation
    --validate                 enables validation
    --workers                  number of worker threads that execute in parallel
    --erigonbatchsize          batch size for the execution stage
    --log                      level of the logging of the app action ("critical", "error", "warning", "notice", "info", "debug")
```