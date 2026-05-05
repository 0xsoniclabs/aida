# Logger Extensions

Extensions for progress reporting and debug logging. Source: `executor/extension/logger/`.

## Overview

| Extension | Purpose |
|-----------|---------|
| ProgressLogger | Logs processing progress at intervals |
| ErrorLogger | Collects and writes errors to log file |
| DbLogger | Logs database statistics |
| DeltaLogger | Produces delta-debugger compatible traces |
| EthStateTestLogger | Logs Ethereum test pass/fail counts |

## Details

### ProgressLogger

Periodically logs the current block number, transaction throughput, and gas rate at configurable
intervals. Starts the timer at PreRun, logs on each PreBlock if the interval has elapsed, and prints
a final summary at PostRun.

### ErrorLogger

Starts a **dedicated goroutine** at PreRun that asynchronously reads from the `ErrorInput` channel
in the Context, making it unique among extensions in using concurrent processing. Collects all
errors and writes them to a log file at PostRun. Provides a centralized error collection point for
non-fatal errors (used by processors when `ContinueOnFailure` is enabled).

### DbLogger

Logs StateDB statistics (cache sizes, disk usage, operation counts) at PostRun.

### DeltaLogger

Produces traces compatible with delta-debugging workflows (specifically designed for
`aida-stochastic-sdb` and delta-debugger tooling). At PreRun, wraps the StateDB in a
`DeltaLoggingProxy` that records all state operations. Also wraps freshly created StateDBs at
PreTransaction. Flushes and closes the trace file at PostRun.

### EthStateTestLogger

Tracks pass/fail counts for Ethereum reference state tests. Increments counters at PostTransaction
and prints a summary at PostRun.
