# Aida Storage Trace Manager (aida-sdb)

## Overview
`aida-sdb` is a tool for managing storage traces and replaying them on a StateDB or using Substates. It combines recording capabilities with replay functionality to simulate block processing in isolation.

## Build
To build the `aida-sdb` application, run:
```shell
make aida-sdb
```
The executable will be located at `build/aida-sdb`.

## Usage
```shell
./build/aida-sdb command [command options] [arguments...]
```

### Commands

| Command | Description |
| :--- | :--- |
| `record` | Captures and records StateDB operations while processing blocks |
| `replay` | Executes storage trace |
| `replay-substate` | Executes storage trace using substates |

## Record Command
Captures and records StateDB operations while processing blocks.
```shell
./build/aida-sdb record --aida-db /path/to/aida_db --trace-file /path/to/output <blockNumFirst> <blockNumLast>
```

### Options
```
    --aida-db               set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --trace-file            set storage trace's output directory
    --db-impl               select state DB implementation
    --db-variant            select a state DB variant
    --vm-impl               select VM implementation
    --chainid               ChainID for replayer
    --validate              enables all validations
    --log                   level of the logging of the app action
```

## Replay Command
Executes storage trace.
```shell
./build/aida-sdb replay --aida-db /path/to/aida-db --trace-file /path/to/trace_file <blockNumFirst> <blockNumLast>
```

### Options
```
    --aida-db               set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --trace-file            set storage trace's output directory
    --db-impl               select state DB implementation
    --db-variant            select a state DB variant
    --db-tmp                sets the temporary directory where to place DB data
    --validate              enables all validations
    --log                   level of the logging of the app action
```

## Replay Substate Command
Executes storage trace using substates.
```shell
./build/aida-sdb replay-substate --aida-db /path/to/aida-db --trace-file /path/to/trace_file <blockNumFirst> <blockNumLast>
```

### Options
```
    --aida-db               set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --trace-file            set storage trace's output directory
    --db-impl               select state DB implementation
    --db-variant            select a state DB variant
    --vm-impl               select VM implementation
    --chainid               ChainID for replayer
    --validate              enables all validations
    --log                   level of the logging of the app action
```
