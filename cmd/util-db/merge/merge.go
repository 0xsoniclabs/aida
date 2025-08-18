package merge

import (
	"fmt"

	"github.com/0xsoniclabs/aida/cmd/util-db/flags"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
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
		dbs, err = utildb.OpenSourceDatabases(sourcePaths)
		if err != nil {
			return err
		}
		md, err = utils.ProcessMergeMetadata(cfg, targetDb, dbs, sourcePaths)
		if err != nil {
			return err
		}

		// todo this should not be necessary - do not close aida-db in metadata
		targetDb, err = db.NewDefaultBaseDB(cfg.AidaDb)
		if err != nil {
			return fmt.Errorf("cannot re-open db: %w", err)
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
