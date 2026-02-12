# Tracker Extensions

Extensions for tracking and reporting processing progress via **console/log output**. For
filesystem-based IPC used by external monitoring tools, see [Register Extensions](register.md).

Source: `executor/extension/tracker/`.

## Overview

| Extension | Purpose |
|-----------|---------|
| BlockProgressTracker | Reports block processing progress |
| RequestProgressTracker | Reports RPC request processing progress |

## Details

### ProgressTracker (Base)

Shared base implementation providing configurable reporting granularity, rate calculation, and ETA
estimation. Both concrete trackers embed this.

### BlockProgressTracker

Tracks block processing progress. After each block (PostBlock), updates internal counters and
reports progress at the configured granularity (e.g., every N blocks or every N seconds).

### RequestProgressTracker

Same as BlockProgressTracker but tracks RPC request processing progress, firing at PostTransaction
since each transaction corresponds to one RPC request in the replay workflow.
