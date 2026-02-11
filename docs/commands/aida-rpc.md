# Aida RPC
## Overview
**aida-rpc** is a tool for testing RPC interfaces by replaying historic API requests recorded from mainnet against a local StateDB. It verifies data correctness and the consistency of the RPC interface implementation.

It replays **RPC requests** into the **StateDB** and compares the result with response in record. Any unmatched results are logged and if not specifically turned off with `--continue-on-failure` flag, it will shut down the replay since any inconsistency in data needs to be investigated immediately.

Substate is necessary for extracting timestamp of block in order to start EVM correctly.

[ShadowDb](ShadowDb) can be used with aida-rpc - see [ShadowDb documentation](ShadowDb) for more details

As of right now, these are the supported methods for both `eth` and `ftm` namespaces:
1. getBalance
2. getTransactionCount
3. call
4. getCode
5. getStorageAt

![API-Replay](https://user-images.githubusercontent.com/84449820/234000908-d1108a9f-0b61-448f-8fb8-9feb4cd13a83.png)

## Requirements
You need a configured Go language environment to build the CLI application.
Please check the [Go documentation](https://go.dev)
for the details of installing the language compiler on your system.

## Build
To build the `aida-rpc` application, run `make aida-rpc`.

The `aida-rpc` executable application will be created in `build/` folder.

## Run
```
./build/aida-rpc --api-recording path/to/api-recording --db-src path/to/statedb/with/archive --substate-db path/to/substate <blockNumFirst> <blockNumLast>
```
executes recorded requests into StateDB with block range between **blockNumFirst-blockNumLast** and compares its results with recorded responses. \
**Requests need to be in block range of given StateDB otherwise they will not be executed.**

### Options
```
GLOBAL:
    --rpc-recording, -r     Path to source file with recorded API data
    --vm-impl               select VM implementation 
    --chainid               ChainID for replayer
    --continue-on-failure   continue execute after validation failure detected
    --validate              enables all validations
    --register-run          When enabled, register results/metadata to an external service.
    --overwrite-run-id      Use provided run id instead of auto-generating run id
    --shadow-db             use this flag when using an existing [ShadowDb](Terminology) 
    --db-src                sets the directory contains source state DB data
    --db-logging            sets path to file for db-logging output
    --trace                 enable tracing
    --trace-file            set storage trace's output directory 
    --trace-debug           enable debug output for tracing
```

## Execution Flow

Uses the standard [Provider](../architecture/Providers.md) → [Processor](../architecture/Processors.md) → [Extensions](../architecture/extensions/README.md) pipeline.

- **Provider:** RpcRequestProvider
- **Processor:** rpcProcessor (custom processor that calls `rpc.Execute()`)
- **Parallelism:** BlockLevel, configurable workers

**Extensions (in registration order):**

1. RegisterRequestProgress (OnPreBlock)
2. CpuProfiler
3. ProgressLogger (15s)
4. ErrorLogger
5. RequestProgressTracker
6. TemporaryArchivePrepper
7. RpcComparator
8. StateDbManager *(if no external stateDb)*
9. ArchiveBlockChecker *(if no external stateDb)*
10. DbLogger *(if no external stateDb)*
