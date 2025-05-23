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

package main

import (
	"math"
	"time"

	"github.com/0xsoniclabs/aida/executor/extension/validator"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension/logger"
	"github.com/0xsoniclabs/aida/executor/extension/profiler"
	"github.com/0xsoniclabs/aida/executor/extension/register"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/executor/extension/tracker"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

const (
	txGeneratorDefaultProgressReportFrequency = 100
)

// RunTxGenerator performs sequential block processing on a StateDb using transaction generator
func RunTxGenerator(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.LastBlockArg)
	if err != nil {
		return err
	}

	cfg.StateValidationMode = utils.SubsetCheck
	cfg.ChainID = utils.EthTestsChainID // Use EthTests chain ID for configurable forks

	db, dbPath, err := utils.PrepareStateDB(cfg)
	if err != nil {
		return err
	}

	provider := executor.NewNormaTxProvider(cfg, db)

	processor, err := executor.MakeLiveDbTxProcessor(cfg)
	if err != nil {
		return err
	}

	return runTransactions(cfg, provider, db, dbPath, processor, nil)
}

func runTransactions(
	cfg *utils.Config,
	provider executor.Provider[txcontext.TxContext],
	stateDb state.StateDB,
	stateDbPath string,
	processor executor.Processor[txcontext.TxContext],
	extra []executor.Extension[txcontext.TxContext],
) error {

	var progressReportFrequency int = txGeneratorDefaultProgressReportFrequency
	if cfg.BlockLength > 0 {
		progressReportFrequency = int(math.Ceil(float64(50_000) / float64(cfg.BlockLength)))
	}

	// order of extensionList has to be maintained
	var extensionList = []executor.Extension[txcontext.TxContext]{
		profiler.MakeVirtualMachineStatisticsPrinter[txcontext.TxContext](cfg),
		statedb.MakeStateDbManager[txcontext.TxContext](cfg, stateDbPath),
		register.MakeRegisterProgress(cfg,
			progressReportFrequency,
			register.OnPreTransaction,
		),
		// RegisterProgress should be the as top-most as possible on the list
		// In this case, after StateDb is created.
		// Any error that happen in extension above it will not be correctly recorded.
		logger.MakeDbLogger[txcontext.TxContext](cfg),
		logger.MakeProgressLogger[txcontext.TxContext](cfg, 15*time.Second),
		logger.MakeErrorLogger[txcontext.TxContext](cfg),
		tracker.MakeBlockProgressTracker(cfg, cfg.TrackerGranularity),
		profiler.MakeMemoryUsagePrinter[txcontext.TxContext](cfg),
		profiler.MakeMemoryProfiler[txcontext.TxContext](cfg),
		validator.MakeShadowDbValidator(cfg),
		statedb.MakeTxGeneratorBlockEventEmitter[txcontext.TxContext](),
	}

	extensionList = append(extensionList, extra...)

	return executor.NewExecutor(provider, cfg.LogLevel).Run(
		executor.Params{
			From:                   int(cfg.First),
			To:                     int(cfg.Last),
			State:                  stateDb,
			ParallelismGranularity: executor.TransactionLevel,
		},
		processor,
		extensionList,
		nil,
	)
}
