# Aida Profile Manager (aida-profile)

## Overview
`aida-profile` is a tool for obtaining usage statistics and profiling data from the world-state.

## Build
To build the `aida-profile` application, run:
```shell
make aida-profile
```
The executable will be located at `build/aida-profile`.

## Usage
```shell
./build/aida-profile command [command options] [arguments...]
```

### Commands

| Command | Description |
| :--- | :--- |
| `code-size` | Reports code size and nonce of smart contracts in the specified block range |
| `storage-size` | Returns the change in storage size by transactions in the specified block range |
| `address-stats` | Computes usage statistics of addresses |
| `key-stats` | Computes usage statistics of accessed storage keys |
| `location-stats` | Computes usage statistics of accessed storage locations |

## Code Size Command
Reports code size and nonce of smart contracts in the specified block range.
```shell
./build/aida-profile code-size --substate-db /path/to/substate_db <blockNumFirst> <blockNumLast>
```

### Options
```
    --substate-db           sets directory containing substate database
    --log                   level of the logging of the app action
```

## Storage Size Command
Returns the change in storage size by transactions in the specified block range.
```shell
./build/aida-profile storage-size --substate-db /path/to/substate_db <blockNumFirst> <blockNumLast>
```

### Options
```
    --substate-db           sets directory containing substate database
    --log                   level of the logging of the app action
```

## Address Stats Command
Computes usage statistics of addresses.
```shell
./build/aida-profile address-stats --substate-db /path/to/substate_db <blockNumFirst> <blockNumLast>
```

## Key Stats Command
Computes usage statistics of accessed storage keys.
```shell
./build/aida-profile key-stats --substate-db /path/to/substate_db <blockNumFirst> <blockNumLast>
```

## Location Stats Command
Computes usage statistics of accessed storage locations.
```shell
./build/aida-profile location-stats --substate-db /path/to/substate_db <blockNumFirst> <blockNumLast>
```
