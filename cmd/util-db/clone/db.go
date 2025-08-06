package clone

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// cloneDbCommand enables creation of aida-cloneDbCommand read or subset
var cloneDbCommand = cli.Command{
	Action:    cloneDbAction,
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

// cloneDbAction creates aida-db copy or subset
func cloneDbAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	aidaDb, targetDb, err := openCloningDbs(cfg.AidaDb, cfg.TargetDb)
	if err != nil {
		return err
	}

	err = clone(cfg, aidaDb, targetDb, utils.CloneType, false)
	if err != nil {
		return err
	}

	utildb.MustCloseDB(aidaDb)
	utildb.MustCloseDB(targetDb)

	return utildb.PrintMetadata(cfg.TargetDb)
}
