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

package db

import (
	"fmt"
	"github.com/0xsoniclabs/aida/config"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// CloneCommand clones aida-db as standalone or patch database
var CloneCommand = cli.Command{
	Name:  "clone",
	Usage: `Used for creation of standalone subset of aida-db or patch`,
	Subcommands: []*cli.Command{
		&CloneDb,
		&ClonePatch,
		&CloneCustom,
	},
}

// ClonePatch enables creation of aida-db read or subset
var ClonePatch = cli.Command{
	Action:    clonePatch,
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

// CloneDb enables creation of aida-db read or subset
var CloneDb = cli.Command{
	Action:    createDbClone,
	Name:      "db",
	Usage:     "clone db creates aida-db subset",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.TargetDbFlag,
		&utils.CompactDbFlag,
		&utils.ValidateFlag,
		&logger.LogLevelFlag,
	},
	Description: `
Creates clone db is used to create subset of aida-db to have more compact database, but still fully usable for desired block range.
`,
}

// CloneDb enables creation of aida-db read or subset
var CloneCustom = cli.Command{
	Action:    createCustomClone,
	Name:      "custom",
	Usage:     "clone custom creates a copy of aida-db components from specified range",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.DbComponentFlag,
		&utils.TargetDbFlag,
		&utils.CompactDbFlag,
		&utils.ValidateFlag,
		&logger.LogLevelFlag,
	},
	Description: `
Clone custom is a specialized clone tool which copies specific components in aida-db from 
 the given block range.
`,
}

// clonePatch creates aida-db patch
func clonePatch(ctx *cli.Context) error {
	// TODO refactor
	cfg, err := config.NewConfig(ctx, config.NoArgs)
	if err != nil {
		return err
	}

	if ctx.Args().Len() != 4 {
		return fmt.Errorf("clone patch command requires exactly 4 arguments")
	}

	cfg.First, cfg.Last, err = config.SetBlockRange(ctx.Args().Get(0), ctx.Args().Get(1), cfg.ChainID)
	if err != nil {
		return err
	}

	var firstEpoch, lastEpoch uint64
	firstEpoch, lastEpoch, err = config.SetBlockRange(ctx.Args().Get(2), ctx.Args().Get(3), cfg.ChainID)
	if err != nil {
		return err
	}

	aidaDb, targetDb, err := utildb.OpenCloningDbs(cfg.AidaDb, cfg.TargetDb)
	if err != nil {
		return err
	}

	err = utildb.CreatePatchClone(cfg, aidaDb, targetDb, firstEpoch, lastEpoch, false)
	if err != nil {
		return err
	}

	utildb.MustCloseDB(aidaDb)
	utildb.MustCloseDB(targetDb)

	return utildb.PrintMetadata(cfg.TargetDb)
}

// createDbClone creates aida-db copy or subset
func createDbClone(ctx *cli.Context) error {
	cfg, err := config.NewConfig(ctx, config.BlockRangeArgs)
	if err != nil {
		return err
	}

	aidaDb, targetDb, err := utildb.OpenCloningDbs(cfg.AidaDb, cfg.TargetDb)
	if err != nil {
		return err
	}

	err = utildb.Clone(cfg, aidaDb, targetDb, utildb.CloneType, false)
	if err != nil {
		return err
	}

	utildb.MustCloseDB(aidaDb)
	utildb.MustCloseDB(targetDb)

	return utildb.PrintMetadata(cfg.TargetDb)
}

// createDbClone creates aida-db copy or subset
func createCustomClone(ctx *cli.Context) error {
	cfg, err := config.NewConfig(ctx, config.BlockRangeArgs)
	if err != nil {
		return err
	}

	aidaDb, targetDb, err := utildb.OpenCloningDbs(cfg.AidaDb, cfg.TargetDb)
	if err != nil {
		return err
	}

	err = utildb.Clone(cfg, aidaDb, targetDb, utildb.CustomType, false)
	if err != nil {
		return err
	}

	utildb.MustCloseDB(aidaDb)
	utildb.MustCloseDB(targetDb)

	return nil
}
