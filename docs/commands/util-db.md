# Aida Database Utility (util-db)

## Overview
`util-db` is the Aida Database Utility — the toolkit for managing AidaDb lifecycle. While the other Aida tools (`aida-vm-sdb`, `aida-rpc`, etc.) consume AidaDb for testing, `util-db` is responsible for **creating, maintaining, and inspecting** the database itself.

### What is AidaDb?
AidaDb is the central filesystem-based database containing [substates](../Terminology.md), update-sets, deleted accounts, state hashes, and metadata. It's built from real-world blockchain data and serves as the primary data source for all offline testing and replay tools.

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

### Options
```
    --aida-db                   set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --target-db                 path to the target database
    --compact                   compact target database
    --validate                  enables validation
    --log                       level of the logging of the app action
```

## Merge Command
Creates target aida-db by merging source databases from arguments: `<db1> [<db2> <db3> ...]`
```shell
./build/util-db merge [options] <db1> [<db2> ...]
```

### Options
```
    --aida-db                   set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --delete-source-d           delete source databases while merging into one database
    --compact                   compact target database
    --log                       level of the logging of the app action
```

## Validate Command
Validates aida-db.
```shell
./build/util-db validate [options]
```

### Options
```
    --aida-db                   set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --validate                  enables validation
    --log                       level of the logging of the app action
```

## Info Command
Prints information about AidaDb.
```shell
./build/util-db info [subcommand] [args] <blockNumLast>
```

### Subcommands
*   `all`: List of all records in AidaDb
*   `del-acc`: Prints info about given deleted account in AidaDb

### Options
```
    --aida-db                   set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --account                   wanted account (for 'all' subcommand)
    --detailed                  prints detailed info (for 'del-acc' subcommand)
    --log                       level of the logging of the app action
```

## Generate Command
Generates precompute substate data.
```shell
./build/util-db generate [options] <events>
```

### Options
```
    --aida-db                   set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --db                        path to the database
    --genesis                   does not stop the program when results do not match
    --keep-db                   if set, statedb is not deleted after run
    --compact                   compact target database
    --db-tmp                    sets the temporary directory where to place DB data
    --chainid                   choose chain id
    --cache                     cache limit
    --log                       level of the logging of the app action
```

## Update Command
Updates aida-db by downloading patches from aida-db generation server.
```shell
./build/util-db update [options]
```

### Options
```
    --aida-db                   set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --chainid                   choose chain id
    --db                        path to the database
    --compact                   compact target database
    --genesis                   does not stop the program when results do not match.
    --db-tmp                    sets the temporary directory where to place DB data
    --cache                     cache limit
    --datadir                   opera datadir directory
    --output                    output path
    --log                       level of the logging of the app action
```


## Compact Command
Performs a full LevelDB compaction on the specified target database. This process optimizes the database storage structure, potentially reducing disk usage and improving read performance by merging SSTables and removing obsolete data.
```shell
./build/util-db compact [options]
```

### Options
```
    --target-db                 path to the target database
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

### Options
```
    --target-db                 path to the target database
    --chainid                   choose chain id
    --client-db                 path to the client database
    --log                       level of the logging of the app action
```

## Priming Command
Performs priming of the specified database.
```shell
./build/util-db priming [options] <blockNum>
```

### Options
```
    --aida-db                   set [aida-db](Terminology) directory
    --carmen-schema             select the DB schema used by Carmen's current state DB
    --db-impl                   select state DB implementation
    --db-variant                select a state DB variant
    --db-src                    sets the directory contains source state DB data
    --db-tmp                    sets the temporary directory where to place DB data
    --archive-mode              enables archive mode
    --archive-query-rate        defines the rate of queries to archive
    --archive-max-query-age     defines the max age of queries to archive
    --archive-variant           select a archive DB variant
    --cpu-profile               enables CPU profiling
    --memory-profile            enables memory allocation profiling
    --random-seed               set random seed
    --prime-threshold           set number of accounts written to stateDB before applying pending state updates
    --prime-random              randomize order of accounts in StateDB priming
    --update-buffer-size        buffer size for holding update set in MiB
    --custom-db-name            custom db name
    --track-progress            enable progress tracking
    --log                       level of the logging of the app action
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

## Execution Flow

Most `util-db` commands operate directly on the database without the executor pipeline. The exception is the **priming** command, which uses `executor.RunUtilPrimer` with extensions:

### Priming

Uses a simplified executor flow — no Provider or Processor, only `PreRun`/`PostRun` hooks:

- **Pipeline:** `RunUtilPrimer` (no transaction iteration)
- **Parallelism:** BlockLevel, 1 worker
- **Extensions (in order):**
  1. [DbLogger](../architecture/extensions/logger.md) — logs database statistics
  2. [StateDbManager](../architecture/extensions/statedb.md) — opens/creates the target StateDB
  3. [StateDbPrimer](../architecture/extensions/primer.md) — applies update-sets from AidaDb to fast-forward the StateDB

This is the only `util-db` command that uses the extension system. All other commands interact with the database directly via LevelDB operations.

## See Also

- [Terminology](../Terminology.md) — AidaDb, Substate, UpdateSet definitions
- [Providers](../architecture/Providers.md) — how AidaDb data is consumed by other tools
- [Extensions: Primer](../architecture/extensions/primer.md) — StateDbPrimer details