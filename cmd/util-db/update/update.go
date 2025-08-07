package update

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// Command downloads aida-db and new patches
var Command = cli.Command{
	Action: updateAction,
	Name:   "update",
	Usage:  "download aida-db patches",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.ChainIDFlag,
		&logger.LogLevelFlag,
		&utils.CompactDbFlag,
		&utils.DbTmpFlag,
		&utils.UpdateTypeFlag,
		&utils.SubstateEncodingFlag,
	},
	Description: ` 
Updates aida-db by downloading patches from aida-db generation server.
`,
}

// updateAction updates aida-db by downloading patches from aida-db generation server.
func updateAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.NoArgs)
	if err != nil {
		return err
	}
	if err = utildb.Update(cfg); err != nil {
		return err
	}

	return utildb.PrintMetadata(cfg.AidaDb)
}
