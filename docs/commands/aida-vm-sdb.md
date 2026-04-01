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

## Ethereum Test Command

Execute ethereum tests.
```shell
./build/aida-vm-sdb ethereum-test [options] <blockNumFirst> <blockNumLast>
```

## Tx Generator Command

Generates transactions for specified block range and executes them over StateDb.
```shell
./build/aida-vm-sdb tx-generator [options] <blockNumFirst> <blockNumLast>
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
