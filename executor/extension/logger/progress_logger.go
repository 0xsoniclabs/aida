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

package logger

import (
	"sync"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
)

const (
	ProgressLoggerDefaultReportFrequency = 15 * time.Second // how often will ticker trigger
	progressLoggerReportFormat           = "Elapsed time: %v; current block %d; last interval rate ~%.2f Tx/s, ~%.2f MGas/s"
	finalSummaryProgressReportFormat     = "Total elapsed time: %v; last block %d; total transaction rate ~%.2f Tx/s, ~%.2f MGas/s"
)

// MakeProgressLogger creates progress logger. It logs progress about processor depending on reportFrequency.
// If reportFrequency is 0, it is set to ProgressLoggerDefaultReportFrequency.
func MakeProgressLogger[T any](cfg *utils.Config, reportFrequency time.Duration) executor.Extension[T] {
	if cfg.NoHeartbeatLogging {
		return extension.NilExtension[T]{}
	}

	if reportFrequency <= 0 {
		reportFrequency = ProgressLoggerDefaultReportFrequency
	}

	return makeProgressLogger[T](cfg, reportFrequency, logger.NewLogger(cfg.LogLevel, "Progress-Logger"))
}

func makeProgressLogger[T any](cfg *utils.Config, reportFrequency time.Duration, logger logger.Logger) *progressLogger[T] {
	return &progressLogger[T]{
		cfg:             cfg,
		log:             logger,
		inputCh:         make(chan txProgressInfo, cfg.Workers*10),
		wg:              new(sync.WaitGroup),
		reportFrequency: reportFrequency,
	}
}

// txProgressInfo keeps information to be reported from a transaction.
type txProgressInfo struct {
	block   int
	gasUsed uint64
}

// progressLogger logs human-readable information about progress
// in "heartbeat" depending on reportFrequency.
type progressLogger[T any] struct {
	extension.NilExtension[T]
	cfg             *utils.Config
	log             logger.Logger
	inputCh         chan txProgressInfo
	wg              *sync.WaitGroup
	reportFrequency time.Duration
}

// PreRun starts the report goroutine
func (l *progressLogger[T]) PreRun(_ executor.State[T], ctx *executor.Context) error {
	l.wg.Add(1)

	// pass the value for thread safety
	path := ""
	if ctx != nil && ctx.StateDbPath != "" {
		path = ctx.StateDbPath
	}
	go l.startReport(l.reportFrequency, path)
	return nil
}

// PostRun gracefully closes the Extension and awaits the report goroutine correct closure.
func (l *progressLogger[T]) PostRun(executor.State[T], *executor.Context, error) error {
	close(l.inputCh)
	l.wg.Wait()

	return nil
}

func (l *progressLogger[T]) PostTransaction(state executor.State[T], ctx *executor.Context) error {
	var gas uint64
	if ctx.ExecutionResult != nil {
		gas = ctx.ExecutionResult.GetGasUsed()
	}
	l.inputCh <- txProgressInfo{block: state.Block, gasUsed: gas}
	return nil
}

// startReport runs in own goroutine. It accepts data from Executor from PostBock func.
// It reports current progress every time we hit the ticker with defaultReportFrequencyInSeconds.
func (l *progressLogger[T]) startReport(reportFrequency time.Duration, stateDbPath string) {
	defer l.wg.Done()

	var (
		currentBlock                 int
		totalTx, currentIntervalTx   uint64
		totalGas, currentIntervalGas uint64
	)

	start := time.Now()
	lastReport := time.Now()
	ticker := time.NewTicker(reportFrequency)

	defer func() {
		elapsed := time.Since(start)
		txRate := float64(totalTx) / elapsed.Seconds()
		gasRate := float64(totalGas) / elapsed.Seconds()

		l.log.Noticef(finalSummaryProgressReportFormat, elapsed.Round(time.Second), currentBlock, txRate, gasRate/1e6)
	}()

	var (
		in txProgressInfo
		ok bool
	)
	for {
		select {
		case in, ok = <-l.inputCh:
			if !ok {
				return
			}

			if in.block > currentBlock {
				currentBlock = in.block
			}

			currentIntervalTx++
			totalTx++

			currentIntervalGas += in.gasUsed
			totalGas += in.gasUsed

		case now := <-ticker.C:
			// skip if no data are present
			if currentIntervalTx == 0 {
				continue
			}
			elapsed := now.Sub(start)
			txRate := float64(currentIntervalTx) / now.Sub(lastReport).Seconds()
			gasRate := float64(currentIntervalGas) / now.Sub(lastReport).Seconds()

			if stateDbPath != "" {
				used, err := utils.GetDirectorySize(stateDbPath)
				if err != nil {
					l.log.Errorf("failed to get size of state-db (%v); %v", stateDbPath, err)
					continue
				}
				free, err := utils.GetFreeSpace(stateDbPath)
				if err != nil {
					l.log.Errorf("failed to get free space of state-db (%v); %v", stateDbPath, err)
					continue
				}

				GiB := float64(1 << 30)
				l.log.Infof(progressLoggerReportFormat+"; disk usage %.2f GiB, free space %.2f GiB",
					elapsed.Round(1*time.Second), currentBlock, txRate, gasRate/1e6,
					float64(used)/GiB, float64(free)/GiB,
				)
			} else {
				l.log.Infof(progressLoggerReportFormat,
					elapsed.Round(1*time.Second), currentBlock, txRate, gasRate/1e6,
				)
			}

			lastReport = now

			currentIntervalTx = 0
			currentIntervalGas = 0
		}
	}

}
