# Aida Block Processing Manager (aida-vm-sdb)

## Overview

`aida-vm-sdb` is the Block Processing Manager. It functions similarly to a block processor,
orchestrating the execution of transactions and blocks. It accepts transaction information in
various formats, including [Substates](../Terminology.md), JSON, and directly from a transaction feeder
(see `txcontext` code package for details). It includes tools for iterating over substates, running
Ethereum tests, and generating transactions for stress testing.

## Build

To build the `aida-vm-sdb` application, run:
```shell
make aida-vm-sdb
```
The executable will be located at `build/aida-vm-sdb`.

## Usage

```shell
./build/aida-vm-sdb command [command options] [arguments...]
```

### Commands

| Command | Description |
| :--- | :--- |
| `substate` | Iterates over substates that are executed into a StateDb |
| `ethereum-test` (ethtest) | Execute ethereum tests |
| `tx-generator` | Generates transactions for specified block range and executes them over StateDb |

## Substate Command

Iterates over substates that are executed into a StateDb.
```shell
./build/aida-vm-sdb substate --aida-db /path/to/aida_db [options] <blockNumFirst> <blockNumLast>
```

### Options

```
    --aida-db                   set [aida-db](../Terminology.md) directory (substate, updateset, deleted accounts)
    --carmen-checkpoint-interval interval for carmen checkpoint
    --carmen-checkpoint-period  period for carmen checkpoint
    --carmen-schema             select the DB schema used by Carmen's current state DB
    --db-impl                   select state DB implementation
    --db-variant                select a state DB variant
    --db-src                    sets the directory contains source state DB data
    --db-src-overwrite          Modify source db directly
    --db-logging                sets path to file for db-logging output
    --validate-state-hash       enables state hash validation
    --archive-mode              enables archive mode
    --archive-query-rate        defines the rate of queries to archive
    --archive-max-query-age     defines the max age of queries to archive
    --archive-variant           select a archive DB variant
    --shadow-db                 use this flag when using an existing [ShadowDb](../Terminology.md)
    --db-shadow-impl            select state DB implementation to shadow the prime DB implementation
    --db-shadow-variant         select a state DB variant to shadow the prime DB implementation
    --evm-impl                  select EVM implementation
    --vm-impl                   select VM implementation
    --random-seed               Set random seed
    --prime-threshold           set number of accounts written to stateDB before applying pending state updates
    --register-run              When enabled, register results/metadata to an external service.
    --overwrite-run-id          Use provided run id instead of auto-generating run id
    --prime-random              randomize order of accounts in StateDB priming
    --skip-priming              if set, DB priming should be skipped; most useful with the 'memory' DB implementation
    --update-buffer-size        buffer size for holding update set in MB
    --chainid                   ChainID for replayer
    --continue-on-failure       continue execute after validation failure detected
    --sync-period               defines the number of blocks per sync-period
    --keep-db                   if set, state-db is not deleted after run
    --custom-db-name            custom db name
    --validate-tx               enables transaction state validation
    --validate                  enables all validations
    --overwrite-pre-world-state Overwrites pre-world state
    --tracker-granularity       chooses how often will tracker report achieved block
    --substate-encoding         select encoding when reading substate from disk: rlp (default) or protobuf
```

## Ethereum Test Command

Execute ethereum tests.
```shell
./build/aida-vm-sdb ethereum-test [options] <blockNumFirst> <blockNumLast>
```

### Options

```
    --carmen-schema             select the DB schema used by Carmen's current state DB
    --db-impl                   select state DB implementation
    --db-variant                select a state DB variant
    --db-logging                sets path to file for db-logging output
    --shadow-db                 use this flag when using an existing [ShadowDb](../Terminology.md)
    --db-shadow-impl            select state DB implementation to shadow the prime DB implementation
    --db-shadow-variant         select a state DB variant to shadow the prime DB implementation
    --evm-impl                  select EVM implementation
    --vm-impl                   select VM implementation
    --random-seed               Set random seed
    --prime-threshold           set number of accounts written to stateDB before applying pending state updates
    --chainid                   ChainID for replayer
    --continue-on-failure       continue execute after validation failure detected
    --validate                  enables all validations
    --validate-state-hash       enables state hash validation
    --max-num-errors            max num errors
    --ethtest-type              ethereum test type
    --fork                      fork name
```

## Tx Generator Command

Generates transactions for specified block range and executes them over StateDb.
```shell
./build/aida-vm-sdb tx-generator [options] <blockNumFirst> <blockNumLast>
```

### Options

```
    --tx-generator-type         tx generator type
    --carmen-schema             select the DB schema used by Carmen's current state DB
    --db-impl                   select state DB implementation
    --db-variant                select a state DB variant
    --db-src                    sets the directory contains source state DB data
    --db-src-overwrite          Modify source db directly
    --db-logging                sets path to file for db-logging output
    --validate-state-hash       enables state hash validation
    --shadow-db                 use this flag when using an existing [ShadowDb](../Terminology.md)
    --db-shadow-impl            select state DB implementation to shadow the prime DB implementation
    --db-shadow-variant         select a state DB variant to shadow the prime DB implementation
    --register-run              When enabled, register results/metadata to an external service.
    --overwrite-run-id          Use provided run id instead of auto-generating run id
    --evm-impl                  select EVM implementation
    --vm-impl                   select VM implementation
    --chainid                   ChainID for replayer
    --continue-on-failure       continue execute after validation failure detected
    --keep-db                   if set, state-db is not deleted after run
    --validate                  enables all validations
    --block-length              defines the number of transactions per block
    --tracker-granularity       chooses how often will tracker report achieved block
    --fork                      fork name
```

## Examples

### Iterating Over Substates

To execute transactions sequentially using substates for blocks 1,000,000 to 1,001,000:
```shell
./build/aida-vm-sdb substate --aida-db /path/to/aida_db 1000000 1001000
```

### Generating Transactions

To generate synthetic transactions for stress testing the StateDB:
```shell
./build/aida-vm-sdb tx-generator --aida-db /path/to/test_db --block-length 100 0 1000
```

### Running Ethereum Tests

To execute standard Ethereum tests against the configured VM:
```shell
./build/aida-vm-sdb ethereum-test --vm-impl geth --ethtest-type GeneralStateTests
```
