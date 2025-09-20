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

package stochastic

import (
	"fmt"
	"math"
	"time"

	"github.com/0xsoniclabs/aida/executor"
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
	Usage:     "record Markovian state while processing blocks",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.CpuProfileFlag,
		&utils.SyncPeriodLengthFlag,
		&utils.OutputFlag,
		&utils.WorkersFlag,
		&utils.ChainIDFlag,
		&utils.AidaDbFlag,
		&utils.CacheFlag,
	},
	Description: `
The stochastic record command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block for recording state.`,
}

// stochasticRecordAction implements recording of state. 
func stochasticRecordAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}
	cfg.ValidateTxState = true
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
	defer sdb.Close()
	iter := sdb.NewSubstateIterator(int(cfg.First), cfg.Workers)
	defer iter.Release()
	oldBlock := uint64(math.MaxUint64) // set to an infeasible block
	var (
		start   time.Time
		sec     float64
		lastSec float64
	)
	start = time.Now()
	sec = time.Since(start).Seconds()
	lastSec = time.Since(start).Seconds()
	recState := recorder.NewState()
	curSyncPeriod := cfg.First / cfg.SyncPeriodLength
	recState.RegisterOp(operations.BeginSyncPeriodID)
	for iter.Next() {
		tx := iter.Value()
		if oldBlock != tx.Block {
			if tx.Block > cfg.Last {
				break
			}
			if oldBlock != math.MaxUint64 {
				recState.RegisterOp(operations.EndBlockID)
				newSyncPeriod := tx.Block / cfg.SyncPeriodLength
				for curSyncPeriod < newSyncPeriod {
					recState.RegisterOp(operations.EndSyncPeriodID)
					curSyncPeriod++
					recState.RegisterOp(operations.BeginSyncPeriodID)
				}
			}
			recState.RegisterOp(operations.BeginBlockID)
			oldBlock = tx.Block
		}
		recState.RegisterOp(operations.BeginTransactionID)
		var statedb state.StateDB
		statedb = state.MakeInMemoryStateDB(substatecontext.NewWorldState(tx.InputSubstate), tx.Block)
		statedb = recorder.NewStochasticProxy(statedb, &recState)
		if _, err = processor.ProcessTransaction(statedb, int(tx.Block), tx.Transaction, substatecontext.NewTxContext(tx)); err != nil {
			return err
		}
		recState.RegisterOp(operations.EndTransactionID)

		// report progress
		sec = time.Since(start).Seconds()
		if sec-lastSec >= 15 {
			fmt.Printf("stochastic record: Elapsed time: %.0f s, at block %v\n", sec, oldBlock)
			lastSec = sec
		}
	}
	// end last block
	if oldBlock != math.MaxUint64 {
		recState.RegisterOp(operations.EndBlockID)
	}
	recState.RegisterOp(operations.EndSyncPeriodID)

	sec = time.Since(start).Seconds()
	fmt.Printf("stochastic record: Total elapsed time: %.3f s, processed %v blocks\n", sec, cfg.Last-cfg.First+1)
	fmt.Printf("stochastic record: write state file ...\n")
	if cfg.Output == "" {
		cfg.Output = "./state.json"
	}
	if err = recState.Write(cfg.Output); err != nil {
		return err
	}

	return nil
}
