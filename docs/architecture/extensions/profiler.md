# Profiler Extensions

Extensions for performance profiling and diagnostics. Source: `executor/extension/profiler/`.

## Overview

| Extension | Purpose |
|-----------|---------|
| CpuProfiler | Records CPU profile to file |
| MemoryProfiler | Records memory profile to file |
| MemoryUsagePrinter | Prints memory breakdown |
| DiagnosticServer | HTTP server for real-time diagnostics |
| ThreadLocker | Locks thread to OS thread |
| VirtualMachineStatisticsPrinter | Prints VM statistics |
| BlockRuntimeAndGasCollector | Collects block runtime and gas data |
| OperationProfiler | Profiles individual StateDB operations (instrument pattern) |

## Details

### CpuProfiler

Starts Go's `pprof` CPU profiler at PreRun and writes the profile to a file at PostRun. Activated
via `--cpu-profile` flag.

### MemoryProfiler

Writes a memory heap profile at PostRun. Activated via `--memory-profile` flag.

### MemoryUsagePrinter

Logs a breakdown of memory usage (heap, stack, GC stats) after each block. Useful for tracking
memory growth during long replay runs.

### DiagnosticServer

Starts a background HTTP server (typically on `/debug/pprof/`) at PreRun for real-time profiling and
diagnostics. Stops at PostRun.

### ThreadLocker

Calls `runtime.LockOSThread()` at PreRun to pin the executor goroutine to a single OS thread.
Ensures consistent profiling results by preventing goroutine migration. Unlocks at PostRun.

### VirtualMachineStatisticsPrinter

Prints VM-specific statistics (instruction counts, gas usage breakdowns, etc.) at PostRun. Output
depends on the VM backend in use.

### BlockRuntimeAndGasCollector

Tracks wall-clock time and gas usage per block and per transaction. Opens the profile database at
PreRun, resets context at PreBlock, starts a per-transaction timer at PreTransaction, records delta
gas per transaction at PostTransaction, finalizes block data at PostBlock, and closes the database
at PostRun.

### OperationProfiler

Uses the **instrument pattern** — wraps the StateDB in a `ProfilerProxy` at PreRun that intercepts
all StateDB method calls (GetBalance, SetStorage, etc.), recording call counts and durations via
incremental analytics. Supports configurable profiling depth (interval, block, or transaction
level). Also uses PreBlock/PostBlock/PostTransaction hooks to manage analytics aggregation and
reporting at each depth level, and PostRun to flush final results.
