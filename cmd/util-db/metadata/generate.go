package metadata

import (
	"errors"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var generateCommand = cli.Command{
	Action: generateAction,
	Name:   "generate",
	Usage:  "Generates new metadata for given chain-id",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.ChainIDFlag,
	},
}

func generateAction(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.NoArgs)
	if argErr != nil {
		return argErr
	}

	base, err := db.NewDefaultBaseDB(cfg.AidaDb)
	if err != nil {
		return err
	}

	defer base.Close()
	sdb := db.MakeDefaultSubstateDBFromBaseDB(base)
	fb, lb, ok := utils.FindBlockRangeInSubstate(sdb)
	if !ok {
		return errors.New("cannot find block range in substate")
	}

	md := utils.NewAidaDbMetadata(base, "INFO")
	md.FirstBlock = fb
	md.LastBlock = lb
	if err = md.SetFreshMetadata(cfg.ChainID); err != nil {
		return err
	}

	return utildb.PrintMetadata(cfg.AidaDb)

}
