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
	"fmt"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension/logger"
	"github.com/0xsoniclabs/aida/executor/extension/primer"
	"github.com/0xsoniclabs/aida/executor/extension/profiler"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/0xsoniclabs/aida/tracer/operation"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

func ReplayTrace(ctx *cli.Context) error {
	cfg, err := config.NewConfig(ctx, config.BlockRangeArgs)
	if err != nil {
		return err
	}

	operationProvider, err := executor.OpenOperations(cfg)
	if err != nil {

	}

	defer operationProvider.Close()

	rCtx := context.NewReplay()

	processor := operationProcessor{cfg, rCtx}

	var extra = []executor.Extension[[]operation.Operation]{
		profiler.MakeReplayProfiler[[]operation.Operation](cfg, rCtx),
	}

	var aidaDb db.BaseDB
	// we need to open substate if we are priming
	if cfg.First > 0 && !cfg.SkipPriming {
		aidaDb, err = db.NewReadOnlyBaseDB(cfg.AidaDb)
		if err != nil {
			return fmt.Errorf("cannot open aida-db; %w", err)
		}
		defer aidaDb.Close()
	}

	return replay(cfg, operationProvider, processor, extra, aidaDb)
}

type operationProcessor struct {
	cfg  *config.Config
	rCtx *context.Replay
}

func (p operationProcessor) Process(state executor.State[[]operation.Operation], ctx *executor.Context) error {
	p.runTransaction(uint64(state.Block), state.Data, ctx.State)
	return nil
}

func (p operationProcessor) runTransaction(block uint64, operations []operation.Operation, stateDb state.StateDB) {
	for _, op := range operations {
		operation.Execute(op, stateDb, p.rCtx)
		if p.cfg.Debug && block >= p.cfg.DebugFrom {
			operation.Debug(&p.rCtx.Context, op)
		}
	}
}

func replay(cfg *config.Config, provider executor.Provider[[]operation.Operation], processor executor.Processor[[]operation.Operation], extra []executor.Extension[[]operation.Operation], aidaDb db.BaseDB) error {
	var extensionList = []executor.Extension[[]operation.Operation]{
		profiler.MakeCpuProfiler[[]operation.Operation](cfg),
		statedb.MakeStateDbManager[[]operation.Operation](cfg, ""),
		profiler.MakeMemoryUsagePrinter[[]operation.Operation](cfg),
		profiler.MakeMemoryProfiler[[]operation.Operation](cfg),
		logger.MakeProgressLogger[[]operation.Operation](cfg, 0),
		primer.MakeStateDbPrimer[[]operation.Operation](cfg),
	}

	extensionList = append(extensionList, extra...)

	return executor.NewExecutor(provider, cfg.LogLevel).Run(
		executor.Params{
			From: int(cfg.First),
			To:   int(cfg.Last) + 1,
		},
		processor,
		extensionList,
		aidaDb,
	)
}
