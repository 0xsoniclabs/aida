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

package primer

import (
	"fmt"

	"github.com/0xsoniclabs/aida/config"
	"github.com/urfave/cli/v2"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension/logger"
	"github.com/0xsoniclabs/aida/executor/extension/primer"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/substate/db"
)

// RunPrimer performs sequential block processing on a StateDb
func RunPrimer(ctx *cli.Context) error {
	cfg, err := config.NewConfig(ctx, config.LastBlockArg)
	if err != nil {
		return err
	}

	aidaDb, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer aidaDb.Close()

	// set config for primming command
	cfg.StateValidationMode = config.SubsetCheck
	cfg.KeepDb = true
	// This is necessary to pass the check inside the priming exstension
	cfg.First = cfg.Last

	return runPriming(cfg, aidaDb)
}

func runPriming(
	cfg *config.Config,
	aidaDb db.BaseDB,
) error {
	var extensionList = []executor.Extension[txcontext.TxContext]{
		logger.MakeDbLogger[txcontext.TxContext](cfg),
		statedb.MakeStateDbManager[txcontext.TxContext](cfg, ""),
	}

	extensionList = append(extensionList, []executor.Extension[txcontext.TxContext]{
		primer.MakeStateDbPrimer[txcontext.TxContext](cfg),
	}...,
	)

	return executor.RunUtilPrimer(
		executor.Params{
			To:                     int(cfg.Last),
			NumWorkers:             1, // vm-sdb can run only with one worker
			ParallelismGranularity: executor.BlockLevel,
		},
		extensionList,
		aidaDb,
	)
}
