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

package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension/logger"
	"github.com/0xsoniclabs/aida/executor/extension/primer"
	"github.com/0xsoniclabs/aida/executor/extension/profiler"
	"github.com/0xsoniclabs/aida/executor/extension/register"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/executor/extension/tracker"
	"github.com/0xsoniclabs/aida/executor/extension/validator"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

const (
	substateDefaultProgressReportFrequency = 100_000
)

// RunSubstate performs sequential block processing on a StateDb
func RunSubstate(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	cfg.StateValidationMode = utils.SubsetCheck

	aidaDb, err := db.NewReadOnlySubstateDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer func(aidaDb db.BaseDB) {
		err = errors.Join(err, aidaDb.Close())
	}(aidaDb)

	substateIterator, err := executor.OpenSubstateProvider(cfg, ctx, aidaDb)
	if err != nil {
		return err
	}
	defer substateIterator.Close()

	processor, err := executor.MakeLiveDbTxProcessor(cfg)
	if err != nil {
		return err
	}

	return runSubstates(cfg, substateIterator, nil, processor, nil, aidaDb)
}

func runSubstates(cfg *utils.Config, provider executor.Provider[txcontext.TxContext], stateDb state.StateDB, processor executor.Processor[txcontext.TxContext], extra []executor.Extension[txcontext.TxContext], aidaDb db.BaseDB) error {
	// order of extensionList has to be maintained
	var extensionList = []executor.Extension[txcontext.TxContext]{
		profiler.MakeCpuProfiler[txcontext.TxContext](cfg),
		profiler.MakeDiagnosticServer[txcontext.TxContext](cfg),
	}

	if stateDb == nil {
		extensionList = append(
			extensionList,
			statedb.MakeStateDbManager[txcontext.TxContext](cfg, ""),
			statedb.MakeLiveDbBlockChecker[txcontext.TxContext](cfg),
			validator.MakeShadowDbValidator(cfg),
			logger.MakeDbLogger[txcontext.TxContext](cfg),
		)
	}

	archiveInquirer, err := statedb.MakeArchiveInquirer(cfg)
	if err != nil {
		return err
	}

	extensionList = append(extensionList, logger.MakeDeltaLogger[txcontext.TxContext](cfg))
	extensionList = append(extensionList, extra...)

	extensionList = append(extensionList, []executor.Extension[txcontext.TxContext]{
		register.MakeRegisterProgress(cfg,
			substateDefaultProgressReportFrequency,
			register.OnPreBlock,
		),
		// RegisterProgress should be the as top-most as possible on the list
		// In this case, after StateDb is created.
		// Any error that happen in extension above it will not be correctly recorded.
		profiler.MakeThreadLocker[txcontext.TxContext](),
		profiler.MakeVirtualMachineStatisticsPrinter[txcontext.TxContext](cfg),
		logger.MakeProgressLogger[txcontext.TxContext](cfg, 15*time.Second),
		logger.MakeErrorLogger[txcontext.TxContext](cfg),
		tracker.MakeBlockProgressTracker(cfg, cfg.TrackerGranularity),
		primer.MakeStateDbPrimer[txcontext.TxContext](cfg),
		profiler.MakeMemoryUsagePrinter[txcontext.TxContext](cfg),
		profiler.MakeMemoryProfiler[txcontext.TxContext](cfg),
		statedb.MakeStateDbPrepper(),
		archiveInquirer,
		validator.MakeStateHashValidator[txcontext.TxContext](cfg),
		statedb.MakeBlockEventEmitter[txcontext.TxContext](),
		statedb.NewParentBlockHashProcessor(cfg),
		statedb.MakeTransactionEventEmitter[txcontext.TxContext](),
		validator.MakeEthereumDbPreTransactionUpdater(cfg),
		statedb.MakeStateDbCorrector(cfg),
		validator.MakeLiveDbValidator(cfg, validator.ValidateTxTarget{WorldState: true, Receipt: true}),
		validator.MakeEthereumDbPostTransactionUpdater(cfg),
		profiler.MakeOperationProfiler[txcontext.TxContext](cfg),

		// block profile extension should be always last because:
		// 1) Pre-Func are called forwards so this is called last and
		// 2) Post-Func are called backwards so this is called first
		// that means the gap between time measurements will be as small as possible
		profiler.MakeBlockRuntimeAndGasCollector(cfg),
	}...,
	)

	return executor.NewExecutor(provider, cfg.LogLevel).Run(
		executor.Params{
			From:                   int(cfg.First),
			To:                     int(cfg.Last) + 1,
			NumWorkers:             1, // vm-sdb can run only with one worker
			State:                  stateDb,
			ParallelismGranularity: executor.BlockLevel,
		},
		processor,
		extensionList,
		aidaDb,
	)
}
