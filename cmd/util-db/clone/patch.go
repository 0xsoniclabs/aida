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

package clone

import (
	"fmt"
	"github.com/0xsoniclabs/aida/cmd/util-db/dbutils"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// clonePatchCommand enables creation of aida-cloneDbCommand read or subset
var clonePatchCommand = cli.Command{
	Action:    clonePatchAction,
	Name:      "patch",
	Usage:     "patch is used to create aida-db patch",
	ArgsUsage: "<blockNumFirst> <blockNumLast> <EpochNumFirst> <EpochNumLast>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.TargetDbFlag,
		&utils.CompactDbFlag,
		&utils.ValidateFlag,
		&logger.LogLevelFlag,
	},
	Description: `
Creates patch of aida-db for desired block range
`,
}

// clonePatchAction creates aida-db patch
func clonePatchAction(ctx *cli.Context) error {
	// TODO refactor
	cfg, err := utils.NewConfig(ctx, utils.NoArgs)
	if err != nil {
		return err
	}

	if ctx.Args().Len() != 4 {
		return fmt.Errorf("clone patch command requires exactly 4 arguments")
	}

	cfg.First, cfg.Last, err = utils.SetBlockRange(ctx.Args().Get(0), ctx.Args().Get(1), cfg.ChainID)
	if err != nil {
		return err
	}

	var firstEpoch, lastEpoch uint64
	firstEpoch, lastEpoch, err = utils.SetBlockRange(ctx.Args().Get(2), ctx.Args().Get(3), cfg.ChainID)
	if err != nil {
		return err
	}

	aidaDb, targetDb, err := openCloningDbs(cfg.AidaDb, cfg.TargetDb)
	if err != nil {
		return err
	}

	err = createPatchClone(cfg, aidaDb, targetDb, firstEpoch, lastEpoch, false)
	if err != nil {
		return err
	}

	dbutils.MustCloseDB(aidaDb)
	dbutils.MustCloseDB(targetDb)

	return dbutils.PrintMetadata(cfg.TargetDb)
}

// createPatchClone creates aida-cloneDbCommand clonePatchCommand
func createPatchClone(cfg *utils.Config, aidaDb, targetDb db.BaseDB, firstEpoch, lastEpoch uint64, isNewOpera bool) error {
	var isFirstGenerationFromGenesis = false

	var cloneType = utils.PatchType

	// if the patch is first, we need to make some exceptions hence cloner needs to know
	if isNewOpera {
		if firstEpoch == 5577 && cfg.ChainID == utils.MainnetChainID {
			isFirstGenerationFromGenesis = true
		} else if firstEpoch == 2458 && cfg.ChainID == utils.TestnetChainID {
			isFirstGenerationFromGenesis = true
		}
	}

	err := clone(cfg, aidaDb, targetDb, cloneType, isFirstGenerationFromGenesis)
	if err != nil {
		return err
	}

	md := utils.NewAidaDbMetadata(targetDb, cfg.LogLevel)
	err = md.SetFirstEpoch(firstEpoch)
	if err != nil {
		return err
	}

	err = md.SetLastEpoch(lastEpoch)
	if err != nil {
		return err
	}

	return nil
}
