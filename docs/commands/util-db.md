# Aida Database Utility (util-db)

## Overview

`util-db` is the Aida Database Utility — the toolkit for managing AidaDb lifecycle. While the other
Aida tools (`aida-vm-sdb`, `aida-rpc`, etc.) consume AidaDb for testing, `util-db` is responsible
for **creating, maintaining, and inspecting** the database itself.

### What is AidaDb?

AidaDb is the central filesystem-based database containing [substates](../Terminology.md),
update-sets, deleted accounts, state hashes, and metadata. It's built from real-world blockchain
data and serves as the primary data source for all offline testing and replay tools.

### Command Categories

| Category | Commands | Purpose |
|----------|----------|---------|
| **Build** | `generate`, `scrape`, `update` | Create or extend AidaDb content |
| **Transform** | `clone`, `merge`, `compact` | Reshape, combine, or optimize databases |
| **Inspect** | `info`, `validate`, `metadata` | Query and verify database integrity |
| **Prepare** | `priming` | Fast-forward a StateDB to a target block height |

## Build

To build the `util-db` application, run:
```shell
make util-db
```
The executable will be located at `build/util-db`.

## Usage

```shell
./build/util-db command [command options] [arguments...]
```

### Commands

| Command | Description |
| :--- | :--- |
| `clone` | Clone can create aida-db copy or subset |
| `compact` | Compact target db |
| `merge` | Merge source databases into aida-db |
| `info` | Prints information about AidaDb |
| `validate` | Validates AidaDb using md5 DbHash |
| `metadata` | Does action with AidaDb metadata |
| `generate` | Generates precompute substate data |
| `update` | Download aida-db patches |
| `scrape` | Stores state hashes into TargetDb for given range |
| `priming` | Performs priming of the specified database |

## Clone Command

Creates clone of aida-db for desired block range.
```shell
./build/util-db clone [subcommand] [options] <args>
```

### Subcommands

*   `db`: clone db creates aida-db subset
*   `patch`: patch is used to create aida-db patch
*   `custom`: clone custom creates a copy of aida-db components from specified range

## Merge Command

Creates target aida-db by merging source databases from arguments: `<db1> [<db2> <db3> ...]`
```shell
./build/util-db merge [options] <db1> [<db2> ...]
```

## Validate Command

Validates aida-db.
```shell
./build/util-db validate [options]
```

## Info Command

Prints information about AidaDb.
```shell
./build/util-db info [subcommand] [args] <blockNumLast>
```

### Subcommands

*   `all`: List of all records in AidaDb
*   `del-acc`: Prints info about given deleted account in AidaDb

## Generate Command

Generates precompute substate data. This is a command group with the following subcommands:

### Subcommands

#### `db-hash`

Generates new db-hash. Note that this will overwrite the current AidaDb hash.
```shell
./build/util-db generate db-hash --aida-db /path/to/aida_db
```

#### `deleted-accounts`

Executes full state transitions and records suicided and resurrected accounts.
```shell
./build/util-db generate deleted-accounts --aida-db /path/to/aida_db <blockNumFirst> <blockNumLast>
```

#### `ethereum-genesis`

Extracts WorldState from genesis JSON into first updateset.
```shell
./build/util-db generate ethereum-genesis --chainid <id> --update-db /path/to/update_db <genesis.json>
```

## Update Command

Updates aida-db by downloading patches from aida-db generation server.
```shell
./build/util-db update [options]
```

## Compact Command

Performs a full LevelDB compaction on the specified target database. This process optimizes the
database storage structure, potentially reducing disk usage and improving read performance by
merging SSTables and removing obsolete data.
```shell
./build/util-db compact [options]
```

## Metadata Command

Does action with AidaDb metadata.
```shell
./build/util-db metadata [subcommand]
```

### Subcommands

*   `print`: Print metadata
*   `generate`: Generate metadata
*   `insert`: Insert metadata
*   `remove`: Remove metadata

## Scrape Command

Stores state hashes into TargetDb for given range.
```shell
./build/util-db scrape [options] <blockNumFirst> <blockNumLast>
```

## Priming Command

Performs priming of the specified database.
```shell
./build/util-db priming [options] <blockNum>
```

## Examples

### Cloning a DB Subset

To create a smaller, standalone DB containing only the state for blocks 1000 to 2000:
```shell
./build/util-db clone db --aida-db /path/to/source_aida_db --target-db /path/to/new_db 1000 2000
```

### Merging Databases

To merge two different StateDBs into a single Aida DB:
```shell
./build/util-db merge --aida-db /path/to/merged_aida_db /path/to/db_part1 /path/to/db_part2
```

### Validating DB Integrity

To run a full validation check on an existing database:
```shell
./build/util-db validate --aida-db /path/to/aida_db --validate
```
