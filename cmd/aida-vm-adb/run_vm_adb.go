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
	"github.com/0xsoniclabs/aida/executor/extension/profiler"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/executor/extension/validator"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// RunVmAdb performs block processing on an ArchiveDb
func RunVmAdb(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	cfg.SetStateDbSrcReadOnly()
	cfg.StateValidationMode = utils.SubsetCheck

	// executing archive blocks always calls ArchiveDb with block -1
	// this condition prevents an incorrect call for block that does not exist (block number -1 in this case)
	// there is nothing before block 0 so running this app on this block does nothing
	if cfg.First == 0 {
		cfg.First = 1
	}

	aidaDb, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer aidaDb.Close()

	substateIterator, err := executor.OpenSubstateProvider(cfg, ctx, aidaDb)
	if err != nil {
		return fmt.Errorf("cannot open substate provider; %w", err)
	}
	defer substateIterator.Close()

	processor, err := executor.MakeArchiveDbTxProcessor(cfg)
	if err != nil {
		return err
	}

	return run(cfg, substateIterator, nil, processor, nil)
}

func run(
	cfg *utils.Config,
	provider executor.Provider[txcontext.TxContext],
	stateDb state.StateDB,
	processor executor.Processor[txcontext.TxContext],
	extra []executor.Extension[txcontext.TxContext],
) error {
	extensionList := []executor.Extension[txcontext.TxContext]{
		profiler.MakeCpuProfiler[txcontext.TxContext](cfg),
		statedb.MakeArchivePrepper[txcontext.TxContext](),
		statedb.NewParentBlockHashProcessor(cfg),
		logger.MakeProgressLogger[txcontext.TxContext](cfg, 0),
		logger.MakeErrorLogger[txcontext.TxContext](cfg),
		validator.MakeArchiveDbValidator(cfg, validator.ValidateTxTarget{WorldState: true, Receipt: true}),
	}

	if stateDb == nil {
		extensionList = append(
			extensionList,
			statedb.MakeStateDbManager[txcontext.TxContext](cfg, ""),
			statedb.MakeArchiveBlockChecker[txcontext.TxContext](cfg),
			logger.MakeDbLogger[txcontext.TxContext](cfg),
		)
	}

	extensionList = append(extensionList, extra...)
	return executor.NewExecutor(provider, cfg.LogLevel).Run(
		executor.Params{
			From:                   int(cfg.First),
			To:                     int(cfg.Last) + 1,
			State:                  stateDb,
			NumWorkers:             cfg.Workers,
			ParallelismGranularity: executor.BlockLevel,
		},
		processor,
		extensionList,
		nil,
	)
}
