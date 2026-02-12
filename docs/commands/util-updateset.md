# Aida UpdateSet Utility (util-updateset)

## Overview

`util-updateset` is the UpdateSet Manager. It generates **[UpdateSets](../Terminology.md)**—collections of
state changes—used to efficiently 'prime' the world state at arbitrary block heights for fast
replaying.

## Build

To build the `util-updateset` application, run:
```shell
make util-updateset
```
The executable will be located at `build/util-updateset`.

## Usage

```shell
./build/util-updateset command [command options] [arguments...]
```

### Commands

| Command | Description |
| :--- | :--- |
| `generate` | Generate update-set from substate |
| `stats` | Print number of accounts and storage keys in update-set |

## Generate Command

Generate update-set from substate.
```shell
./build/util-updateset generate --aida-db /path/to/aida_db [options] <blockNumLast> <interval>
```
`<blockNumLast>` is last block of the inclusive range of blocks to generate update set.
`<interval>` is the block interval of writing update set to updateDB.

### Options

```
    --chainid               ChainID for replayer
    --aida-db               set [aida-db](../Terminology.md) directory (substate, updateset, deleted accounts)
    --update-buffer-size    buffer size for holding update set in MB
    --validate              enables all validations
```

## Stats Command

Print number of accounts and storage keys in update-set.
```shell
./build/util-updateset stats --update-db /path/to/update_db <blockNumLast>
```
The stats command requires one arguments: `<blockNumLast>` - the last block of update-set.

### Options

```
    --update-db             set update-set database directory
```
