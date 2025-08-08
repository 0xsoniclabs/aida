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

package merge

import (
	"fmt"
	"os"

	"github.com/0xsoniclabs/aida/cmd/util-db/dbutils"

	"github.com/0xsoniclabs/aida/cmd/util-db/flags"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// Command merges given databases into aida-db
var Command = cli.Command{
	Action: mergeAction,
	Name:   "merge",
	Usage:  "merge source databases into aida-db",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.DeleteSourceDbsFlag,
		&logger.LogLevelFlag,
		&utils.CompactDbFlag,
		&flags.SkipMetadata,
		&utils.SubstateEncodingFlag,
	},
	Description: `
Creates target aida-db by merging source databases from arguments:
<db1> [<db2> <db3> ...]
`,
}

// mergeAction two or more Dbs together
func mergeAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.OneToNArgs)
	if err != nil {
		return err
	}

	sourcePaths := make([]string, ctx.Args().Len())
	for i := 0; i < ctx.Args().Len(); i++ {
		sourcePaths[i] = ctx.Args().Get(i)
	}

	// we need a destination where to save merged aida-db
	if cfg.AidaDb == "" {
		return fmt.Errorf("you need to specify where you want aida-db to save (--aida-db)")
	}

	targetDb, err := db.NewDefaultBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open db; %v", err)
	}

	var (
		dbs []db.BaseDB
		md  *utils.AidaDbMetadata
	)

	if !cfg.SkipMetadata {
		dbs, err = openSourceDatabases(sourcePaths)
		if err != nil {
			return err
		}
		md, err = utils.ProcessMergeMetadata(cfg, targetDb, dbs, sourcePaths)
		if err != nil {
			return err
		}

		targetDb = md.Db

		for _, db := range dbs {
			dbutils.MustCloseDB(db)
		}
	}

	dbs, err = openSourceDatabases(sourcePaths)
	if err != nil {
		return err
	}

	m := dbutils.NewMerger(cfg, targetDb, dbs, sourcePaths, md)

	if err = m.Merge(); err != nil {
		return err
	}

	m.CloseSourceDbs()

	return m.FinishMerge()
}

// openSourceDatabases opens all databases required for merge
func openSourceDatabases(sourceDbPaths []string) ([]db.BaseDB, error) {
	if len(sourceDbPaths) < 1 {
		return nil, fmt.Errorf("no source database were specified\n")
	}

	var sourceDbs []db.BaseDB
	for i := 0; i < len(sourceDbPaths); i++ {
		path := sourceDbPaths[i]
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source database %s; doesn't exist\n", path)
		}
		db, err := db.NewReadOnlyBaseDB(path)
		if err != nil {
			return nil, fmt.Errorf("source database %s; error: %v", path, err)
		}
		sourceDbs = append(sourceDbs, db)
	}

	return sourceDbs, nil
}
