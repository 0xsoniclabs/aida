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

package trace

import (
	"errors"
	"fmt"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension/logger"
	"github.com/0xsoniclabs/aida/executor/extension/primer"
	"github.com/0xsoniclabs/aida/executor/extension/profiler"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/executor/extension/validator"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/0xsoniclabs/aida/tracer/operation"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

func ReplaySubstate(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	aidaDb, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer func(aidaDb db.BaseDB) {
		err = errors.Join(err, aidaDb.Close())
	}(aidaDb)

	substateIterator := executor.OpenSubstateProvider(cfg, ctx, aidaDb)
	defer substateIterator.Close()

	operationProvider, err := executor.OpenOperations(cfg)
	if err != nil {
		return err
	}

	defer substateIterator.Close()

	rCtx := context.NewReplay()

	processor := makeSubstateProcessor(cfg, rCtx, operationProvider)

	var extra = []executor.Extension[txcontext.TxContext]{
		profiler.MakeReplayProfiler[txcontext.TxContext](cfg, rCtx),
	}

	return replaySubstate(cfg, substateIterator, processor, nil, extra)
}

func makeSubstateProcessor(cfg *utils.Config, rCtx *context.Replay, operationProvider executor.Provider[[]operation.Operation]) *substateProcessor {
	return &substateProcessor{
		operationProcessor: operationProcessor{cfg, rCtx},
		operationProvider:  operationProvider,
	}
}

type substateProcessor struct {
	operationProcessor
	operationProvider executor.Provider[[]operation.Operation]
}

func (p substateProcessor) Process(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	return p.operationProvider.Run(state.Block, state.Block, func(t executor.TransactionInfo[[]operation.Operation]) error {
		return p.runTransaction(uint64(state.Block), t.Data, ctx.State)
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
			From:  int(cfg.First),
			To:    int(cfg.Last) + 1,
			State: stateDb,
		},
		processor,
		extensionList,
		nil,
	)
}
