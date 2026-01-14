# Aida Stochastic-Test Manager (aida-stochastic-sdb)

## Overview
`aida-stochastic-sdb` is the Stochastic Test Manager. It facilitates property-based testing by generating, recording, and replaying randomized (stochastic) sequences of StateDB operations to uncover edge cases.

## Build
To build the `aida-stochastic-sdb` application, run:
```shell
make aida-stochastic-sdb
```
The executable will be located at `build/aida-stochastic-sdb`.

## Usage
```shell
./build/aida-stochastic-sdb command [command options] [arguments...]
```

### Commands

| Command | Description |
| :--- | :--- |
| `generate` | Generate uniform stats file |
| `record` | Record Markovian stats while processing blocks |
| `replay` | Simulates StateDB operations using a Markovian Process |
| `visualize` | Produces a graphical view of the stats |

## Generate Command
Produces a stats file with uniform parameters for stochastic testing.
```shell
./build/aida-stochastic-sdb generate [options]
```

### Options
```
    --output, -o          output path
    --block-length        defines the number of transactions per block 
    --sync-period         defines the number of blocks per sync-period 
    --transaction-length  Determines indirectly the length of a transaction 
    --num-contracts       Number of contracts to create 
    --num-keys            Number of keys to generate 
    --num-values          Number of values to generate 
    --snapshot-depth      Depth of snapshot history 
```

## Record Command
Record Markovian stats while processing blocks.
```shell
./build/aida-stochastic-sdb record --aida-db /path/to/aida_db [options] <blockNumFirst> <blockNumLast>
```

### Options
```
    --sync-period         defines the number of blocks per sync-period 
    --output, -o          write the minimized trace to the given path
    --chainid             ChainID for replayer
    --aida-db             set [aida-db](Terminology) directory (substate, updateset, deleted accounts)
    --cache               Cache limit for StateDb or Priming 
    --substate-encoding   select encoding when reading [Substate](Terminology) from disk: rlp (default) or protobuf 
```

## Replay Command
Simulates StateDB operations using a Markovian Process.
```shell
./build/aida-stochastic-sdb replay --aida-db /path/to/aida_db [options] <simulation-length> <stats-file>
```

### Options
```
    --balance-range         sets the balance range of the stochastic simulation 
    --carmen-schema         select the DB schema used by Carmen's current state DB 
    --continue-on-failure   continue execute after validation failure detected
    --debug-from            sets the first block to print trace debug 
    --nonce-range           sets nonce range for stochastic simulation 
    --random-seed           Set random seed 
    --db-impl               select state DB implementation 
    --db-variant            select a state DB variant
    --db-logging            sets path to file for db-logging output
    --trace-file            set storage trace's output directory 
    --trace-debug           enable debug output for tracing
    --trace                 enable tracing
    --db-shadow-impl        select state DB implementation to shadow the prime DB implementation
    --db-shadow-variant     select a state DB variant to shadow the prime DB implementation
    --validate-state-hash   enables state hash validation
```

## Visualize Command
Produces a graphical view of the stats for the Markovian process.
```shell
./build/aida-stochastic-sdb visualize [options] <stats-file>
```

### Options
```
    --port, -v            enable visualization on `PORT` 
```
