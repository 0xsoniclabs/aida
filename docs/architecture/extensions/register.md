# Register Extensions

Extensions that record processing progress to **filesystem-based register files** for external monitoring tools (IPC). Unlike [Trackers](tracker.md) which output to console/logs, registers write to files that external processes can poll.

Source: `executor/extension/register/`.

## Overview

| Extension | Constructor | Purpose | Hooks |
|-----------|------------|---------|-------|
| RegisterProgress | `MakeRegisterProgress(cfg)` | Records block progress to register file | PreBlock or PreTransaction |
| RegisterRequestProgress | — | Records RPC request progress to register file | PreTransaction |

## Details

### RegisterProgress

Writes the current block number to a register file at configurable intervals. External tools can monitor this file to track processing progress. Can be configured to fire on PreBlock or PreTransaction depending on the workflow.

### RegisterRequestProgress

Same concept as RegisterProgress but tracks RPC request processing progress rather than block progress.

## See Also

- [Extension System Overview](README.md)
- [Trackers](tracker.md) — in-process progress tracking
