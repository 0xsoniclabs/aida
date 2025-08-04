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
	"github.com/0xsoniclabs/aida/executor/extension/tracker"
	"github.com/0xsoniclabs/aida/executor/extension/validator"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

func RunReplay(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
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

	var files []string
	if cfg.TraceDirectory != "" {
		entries, err := os.ReadDir(cfg.TraceDirectory)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				return fmt.Errorf("given dir %s contains subdirectories, please provide a single trace file or a directory with trace files", cfg.TraceDirectory)
			}
			files = append(files, entry.Name())
		}
	} else if cfg.TraceFile != "" {
		files = append(files, cfg.TraceFile)
	} else {
		return fmt.Errorf("no trace file (--%s) or directory (--%s) provided", utils.TraceFileFlag.Name, utils.TraceDirectoryFlag.Name)
	}

	for _, filename := range files {
		file, err := tracer.NewFileReader(filename)
		if err != nil {
			return err
		}

		provider := executor.NewTraceProvider(file)
		err = replay(cfg, provider, nil, &traceProcessor{}, nil, aidaDb)
		if err != nil {
			return err
		}
	}

	return nil
}

func replay(
	cfg *utils.Config,
	provider executor.Provider[tracer.Operation],
	stateDb state.StateDB,
	processor executor.Processor[tracer.Operation],
	extra []executor.Extension[tracer.Operation],
	aidaDb db.BaseDB,
) error {
	// order of extensionList has to be maintained
	var extensionList = []executor.Extension[tracer.Operation]{
		profiler.MakeCpuProfiler[tracer.Operation](cfg),
		profiler.MakeDiagnosticServer[tracer.Operation](cfg),
	}

	if stateDb == nil {
		extensionList = append(
			extensionList,
			statedb.MakeStateDbManager[tracer.Operation](cfg, ""),
			statedb.MakeLiveDbBlockChecker[tracer.Operation](cfg),
			validator.MakeShadowDbValidator[tracer.Operation](cfg),
			logger.MakeDbLogger[tracer.Operation](cfg),
		)
	}
	extensionList = append(extensionList, extra...)

	extensionList = append(extensionList, []executor.Extension[tracer.Operation]{
		// RegisterProgress should be the as top-most as possible on the list
		// In this case, after StateDb is created.
		// Any error that happen in extension above it will not be correctly recorded.
		profiler.MakeThreadLocker[tracer.Operation](),
		profiler.MakeVirtualMachineStatisticsPrinter[tracer.Operation](cfg),
		logger.MakeProgressLogger[tracer.Operation](cfg, 15*time.Second),
		logger.MakeErrorLogger[tracer.Operation](cfg),
		tracker.MakeBlockProgressTracker[tracer.Operation](cfg, cfg.TrackerGranularity),
		primer.MakeStateDbPrimer[tracer.Operation](cfg),
		profiler.MakeMemoryUsagePrinter[tracer.Operation](cfg),
		profiler.MakeMemoryProfiler[tracer.Operation](cfg),
		validator.MakeStateHashValidator[tracer.Operation](cfg),
		profiler.MakeOperationProfiler[tracer.Operation](cfg),
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
