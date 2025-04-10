// Copyright 2024 Fantom Foundation
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package tracker

import (
	"fmt"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
)

const substateProgressTrackerReportFormat = "Track: block %d, memory %d, disk %d, interval_blk_rate %.2f, interval_tx_rate %.2f, interval_gas_rate %.2f, overall_blk_rate %.2f, overall_tx_rate %.2f, overall_gas_rate %.2f"

// MakeBlockProgressTracker creates a blockProgressTracker that depends on the
// PostBlock event and is only useful as part of a sequential evaluation.
func MakeBlockProgressTracker(cfg *utils.Config, reportFrequency int) executor.Extension[txcontext.TxContext] {
	if !cfg.TrackProgress {
		return extension.NilExtension[txcontext.TxContext]{}
	}

	if reportFrequency == 0 {
		reportFrequency = ProgressTrackerDefaultReportFrequency
	}

	return makeBlockProgressTracker(cfg, reportFrequency, logger.NewLogger(cfg.LogLevel, "ProgressTracker"))
}

func makeBlockProgressTracker(cfg *utils.Config, reportFrequency int, log logger.Logger) *blockProgressTracker {
	return &blockProgressTracker{
		progressTracker:   newProgressTracker[txcontext.TxContext](cfg, reportFrequency, log),
		lastReportedBlock: int(cfg.First) - (int(cfg.First) % reportFrequency),
	}
}

// blockProgressTracker logs progress every XXX blocks depending on reportFrequency.
// Default is 100_000 blocks. This is mainly used for gathering information about process.
type blockProgressTracker struct {
	*progressTracker[txcontext.TxContext]
	overallInfo       substateProcessInfo
	lastIntervalInfo  substateProcessInfo
	lastReportedBlock int
}

type substateProcessInfo struct {
	numTransactions uint64
	gas             uint64
}

// PostTransaction increments number of transactions and saves gas used in last substate.
func (t *blockProgressTracker) PostTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.overallInfo.numTransactions++
	if ctx.ExecutionResult != nil {
		t.overallInfo.gas += ctx.ExecutionResult.GetGasUsed()
	}

	return nil
}

// PostBlock registers the completed block and may trigger the logging of an update.
func (t *blockProgressTracker) PostBlock(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	boundary := state.Block - (state.Block % t.reportFrequency)

	if state.Block-t.lastReportedBlock < t.reportFrequency {
		return nil
	}

	now := time.Now()
	overall := now.Sub(t.startOfRun)
	interval := now.Sub(t.startOfLastInterval)

	// quickly get a snapshot of the current overall progress
	t.lock.Lock()
	info := t.overallInfo
	t.lock.Unlock()

	disk, err := utils.GetDirectorySize(ctx.StateDbPath)
	if err != nil {
		return fmt.Errorf("cannot size of state-db (%v); %v", ctx.StateDbPath, err)
	}
	m := ctx.State.GetMemoryUsage()

	memory := uint64(0)
	if m != nil {
		memory = m.UsedBytes
	}

	intervalBlkRate := float64(t.reportFrequency) / interval.Seconds()
	intervalTxRate := float64(info.numTransactions-t.lastIntervalInfo.numTransactions) / interval.Seconds()
	intervalGasRate := float64(info.gas-t.lastIntervalInfo.gas) / interval.Seconds()
	t.lastIntervalInfo = info

	overallBlkRate := float64(state.Block-int(t.cfg.First)) / overall.Seconds()
	overallTxRate := float64(info.numTransactions) / overall.Seconds()
	overallGasRate := float64(info.gas) / overall.Seconds()

	t.log.Noticef(
		substateProgressTrackerReportFormat,
		boundary, memory, disk,
		intervalBlkRate, intervalTxRate, intervalGasRate,
		overallBlkRate, overallTxRate, overallGasRate,
	)

	t.lastReportedBlock = boundary
	t.startOfLastInterval = now

	return nil
}
