// Copyright 2025 Sonic Labs
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

package stochastic

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// StochasticRecordCommand data structure for the record app
var StochasticRecordCommand = cli.Command{
	Action:    stochasticRecordAction,
	Name:      "record",
	Usage:     "record Markovian stats while processing blocks",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.CpuProfileFlag,
		&utils.SyncPeriodLengthFlag,
		&utils.OutputFlag,
		&utils.WorkersFlag,
		&utils.ChainIDFlag,
		&utils.AidaDbFlag,
		&utils.CacheFlag,
		&utils.SubstateEncodingFlag,
	},
	Description: `
The stochastic record command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block for recording stats.`,
}

// stochasticRecordAction implements recording of stats.
func stochasticRecordAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}
	if cfg.SyncPeriodLength == 0 {
		return fmt.Errorf("sync-period must be greater than zero")
	}
	cfg.ValidateTxState = true
	log := logger.NewLogger(cfg.LogLevel, "StochasticRecord")
	if err := utils.StartCPUProfile(cfg); err != nil {
		return err
	}
	defer utils.StopCPUProfile(cfg)
	processor, err := executor.MakeLiveDbTxProcessor(cfg)
	if err != nil {
		return err
	}
	sdb, err := db.NewReadOnlySubstateDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer func(sdb db.SubstateDB) {
		err = errors.Join(err, sdb.Close())
	}(sdb)
	iter := sdb.NewSubstateIterator(int(cfg.First), cfg.Workers)
	defer iter.Release()
	oldBlock := uint64(math.MaxUint64) // set to an infeasible block
	var sec float64
	start := time.Now()
	lastSec := time.Since(start).Seconds()
	stats := recorder.NewStats()
	curSyncPeriod := cfg.First / cfg.SyncPeriodLength
	err = stats.CountOp(operations.BeginSyncPeriodID)
	if err != nil {
		return err
	}
	for iter.Next() {
		tx := iter.Value()
		if oldBlock != tx.Block {
			if tx.Block > cfg.Last {
				break
			}
			if oldBlock != math.MaxUint64 {
				err = stats.CountOp(operations.EndBlockID)
				if err != nil {
					return err
				}
				newSyncPeriod := tx.Block / cfg.SyncPeriodLength
				for curSyncPeriod < newSyncPeriod {
					err = stats.CountOp(operations.EndSyncPeriodID)
					if err != nil {
						return err
					}
					curSyncPeriod++
					err = stats.CountOp(operations.BeginSyncPeriodID)
					if err != nil {
						return err
					}
				}
			}
			err = stats.CountOp(operations.BeginBlockID)
			if err != nil {
				return err
			}
			oldBlock = tx.Block
		}
		err = stats.CountOp(operations.BeginTransactionID)
		if err != nil {
			return err
		}
		var statedb state.StateDB
		statedb = state.MakeInMemoryStateDB(substatecontext.NewWorldState(tx.InputSubstate), tx.Block)
		statedb = recorder.NewStochasticProxy(statedb, &stats)
		if _, err = processor.ProcessTransaction(statedb, int(tx.Block), tx.Transaction, substatecontext.NewTxContext(tx)); err != nil {
			return err
		}
		err = stats.CountOp(operations.EndTransactionID)
		if err != nil {
			return err
		}

		// report progress
		sec = time.Since(start).Seconds()
		if sec-lastSec >= 15 {
			log.Infof("Elapsed time: %.0f s, at block %v", sec, oldBlock)
			lastSec = sec
		}
	}
	// end last block
	if oldBlock != math.MaxUint64 {
		err = stats.CountOp(operations.EndBlockID)
		if err != nil {
			return err
		}
	}
	err = stats.CountOp(operations.EndSyncPeriodID)
	if err != nil {
		return err
	}

	sec = time.Since(start).Seconds()
	log.Noticef("Total elapsed time: %.3f s, processed %v blocks", sec, cfg.Last-cfg.First+1)
	log.Notice("Write stats file ...")
	if cfg.Output == "" {
		cfg.Output = "./stats.json"
	}
	if err = stats.Write(cfg.Output); err != nil {
		return err
	}

	return nil
}
