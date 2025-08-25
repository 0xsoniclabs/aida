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
	"fmt"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension/logger"
	"github.com/0xsoniclabs/aida/executor/extension/primer"
	"github.com/0xsoniclabs/aida/executor/extension/profiler"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/executor/extension/validator"
	aidaLogger "github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var RunReplaySubstateCmd = cli.Command{
	Action:    RunReplaySubstate,
	Name:      "replay-substate",
	Usage:     "executes storage trace using substates",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.ChainIDFlag,
		&utils.CpuProfileFlag,
		&utils.RandomizePrimingFlag,
		&utils.RandomSeedFlag,
		&utils.PrimeThresholdFlag,
		&utils.ProfileFlag,
		&utils.StateDbImplementationFlag,
		&utils.StateDbVariantFlag,
		&utils.StateDbLoggingFlag,
		&utils.ShadowDbImplementationFlag,
		&utils.ShadowDbVariantFlag,
		&utils.SyncPeriodLengthFlag,
		&utils.WorkersFlag,
		&utils.TraceFileFlag,
		&utils.TraceDirectoryFlag,
		&utils.TraceDebugFlag,
		&utils.DebugFromFlag,
		//&utils.ValidateFlag,
		//&utils.ValidateTxStateFlag,
		&utils.AidaDbFlag,
		&aidaLogger.LogLevelFlag,
	},
	Description: `
The trace replay-substate command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay storage traces.`,
}

func RunReplaySubstate(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	aidaDb, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer aidaDb.Close()

	substateIterator := executor.OpenSubstateProvider(cfg, ctx, aidaDb)
	defer substateIterator.Close()

	files, err := getTraceFiles(cfg.TraceFile, cfg.TraceDirectory)
	if err != nil {
		return err
	}

	for _, filename := range files {
		file, first, _, err := tracer.NewFileReader(filename)
		if err != nil {
			return err
		}

		if cfg.First < first {
			return fmt.Errorf("chosen first block %d is less than the first block %d in the trace file %s", cfg.First, first, filename)
		}

		provider := executor.NewTraceProvider(file)

		processor := makeSubstateProcessor(provider)
		err = replaySubstate(cfg, substateIterator, processor, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to replay substate: %w", err)
		}
	}
	return nil
}

func makeSubstateProcessor(operationProvider executor.Provider[tracer.Operation]) *substateProcessor {
	return &substateProcessor{
		traceProcessor:    traceProcessor{},
		operationProvider: operationProvider,
	}
}

type substateProcessor struct {
	traceProcessor
	operationProvider      executor.Provider[tracer.Operation]
	currentBlockOperations []tracer.Operation
}

func (p substateProcessor) Process(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	return p.operationProvider.Run(state.Block, state.Block, func(t executor.TransactionInfo[tracer.Operation]) error {
		return p.traceProcessor.Process(executor.State[tracer.Operation]{
			Block:       state.Block,
			Transaction: state.Transaction,
			Data:        t.Data,
		}, ctx)
	})
}

func replaySubstate(
	cfg *utils.Config,
	provider executor.Provider[txcontext.TxContext],
	processor executor.Processor[txcontext.TxContext],
	stateDb state.StateDB,
	extra []executor.Extension[txcontext.TxContext],
) error {
	var extensionList = []executor.Extension[txcontext.TxContext]{
		profiler.MakeCpuProfiler[txcontext.TxContext](cfg),
		logger.MakeProgressLogger[txcontext.TxContext](cfg, 0),
		profiler.MakeMemoryUsagePrinter[txcontext.TxContext](cfg),
		profiler.MakeMemoryProfiler[txcontext.TxContext](cfg),
		validator.MakeLiveDbValidator(cfg, validator.ValidateTxTarget{WorldState: true, Receipt: true}),
	}

	if stateDb == nil {
		extensionList = append(extensionList, statedb.MakeStateDbManager[txcontext.TxContext](cfg, ""))
	}

	if cfg.DbImpl == "memory" {
		extensionList = append(extensionList, statedb.MakeStateDbPrepper())
	} else {
		extensionList = append(extensionList, primer.MakeTxPrimer(cfg))
	}

	extensionList = append(extensionList, extra...)

	return executor.NewExecutor(provider, cfg.LogLevel).Run(
		executor.Params{
			From:                   int(cfg.First),
			To:                     int(cfg.Last) + 1,
			State:                  stateDb,
			ParallelismGranularity: executor.BlockLevel,
		},
		processor,
		extensionList,
		nil,
	)
}
