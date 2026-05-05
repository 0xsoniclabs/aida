# Aida Stochastic-Test Manager (aida-stochastic-sdb)

## Overview

`aida-stochastic-sdb` is the Stochastic Test Manager. It facilitates property-based testing by
generating, recording, and replaying randomized (stochastic) sequences of StateDB operations to
uncover edge cases.

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

## Record Command

Record Markovian stats while processing blocks.
```shell
./build/aida-stochastic-sdb record --aida-db /path/to/aida_db [options] <blockNumFirst> <blockNumLast>
```

## Replay Command

Simulates StateDB operations using a Markovian Process.
```shell
./build/aida-stochastic-sdb replay --aida-db /path/to/aida_db [options] <simulation-length> <stats-file>
```

## Visualize Command

Produces a graphical view of the stats for the Markovian process.
```shell
./build/aida-stochastic-sdb visualize [options] <stats-file>
```

