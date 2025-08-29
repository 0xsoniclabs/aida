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

package clone

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// cloneCustomCommand enables creation of aida-db read or subset
var cloneCustomCommand = cli.Command{
	Action:    cloneCustomAction,
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
clone custom is a specialized clone tool which copies specific components in aida-db from 
 the given block range.
`,
}

// cloneCustomAction creates aida-db copy or subset
func cloneCustomAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	aidaDb, targetDb, err := openCloningDbs(cfg.AidaDb, cfg.TargetDb)
	if err != nil {
		return err
	}

	err = clone(cfg, aidaDb, targetDb, utils.CustomType)
	if err != nil {
		return err
	}

	utildb.MustCloseDB(aidaDb)
	utildb.MustCloseDB(targetDb)

	return nil
}
