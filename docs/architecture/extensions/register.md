# Register Extensions

Extensions that record processing progress to **filesystem-based register files** for external
monitoring tools (IPC). Unlike [Trackers](tracker.md) which output to console/logs, registers write
to files that external processes can poll.

Source: `executor/extension/register/`.

## Overview

| Extension | Purpose |
|-----------|---------|
| RegisterProgress | Records block progress to register file |
| RegisterRequestProgress | Records RPC request progress to register file |

## Details

### RegisterProgress

Writes the current block number to a register file at configurable intervals. External tools can
monitor this file to track processing progress. Can be configured to fire on PreBlock or
PreTransaction depending on the workflow.

### RegisterRequestProgress

Same concept as RegisterProgress but tracks RPC request processing progress rather than block
progress.
