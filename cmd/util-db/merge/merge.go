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

package merge

import (
	"fmt"

	"github.com/0xsoniclabs/aida/cmd/util-db/flags"
	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utildb/metadata"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// Command merges given databases into aida-db
var Command = cli.Command{
	Action: mergeAction,
	Name:   "merge",
	Usage:  "merge source databases into aida-db",
	Flags: []cli.Flag{
		&config.AidaDbFlag,
		&config.DeleteSourceDbsFlag,
		&logger.LogLevelFlag,
		&config.CompactDbFlag,
		&flags.SkipMetadata,
		&config.SubstateEncodingFlag,
	},
	Description: `
Creates target aida-db by merging source databases from arguments:
<db1> [<db2> <db3> ...]
`,
}

// mergeAction two or more Dbs together
func mergeAction(ctx *cli.Context) error {
	cfg, err := config.NewConfig(ctx, config.OneToNArgs)
	if err != nil {
		return err
	}

	sourcePaths := make([]string, ctx.Args().Len())
	for i := 0; i < ctx.Args().Len(); i++ {
		sourcePaths[i] = ctx.Args().Get(i)
	}

	targetDb, err := db.NewDefaultSubstateDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open db; %v", err)
	}

	var (
		dbs []db.SubstateDB
		md  *metadata.AidaDbMetadata
	)

	if !cfg.SkipMetadata {
		dbs, err = utildb.OpenSourceDatabases(sourcePaths)
		if err != nil {
			return fmt.Errorf("cannot open source databases: %w", err)
		}

		// merge metadata from all source dbs
		targetMD := metadata.NewAidaDbMetadata(targetDb, cfg.LogLevel)
		for _, db := range dbs {
			sourceMD := metadata.NewAidaDbMetadata(db, cfg.LogLevel)
			if err := targetMD.Merge(sourceMD); err != nil {
				return fmt.Errorf("cannot merge metadata: %w", err)
			}
		}
		for _, database := range dbs {
			utildb.MustCloseDB(database)
		}
	}
	dbs, err = utildb.OpenSourceDatabases(sourcePaths)
	if err != nil {
		return err
	}

	m := utildb.NewMerger(cfg, targetDb, dbs, sourcePaths, md)
	if err = m.Merge(); err != nil {
		return err
	}

	m.CloseSourceDbs()
	return m.FinishMerge()
}
