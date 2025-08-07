package scrape

import (
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var Command = cli.Command{
	Action:    scrapeAction,
	Name:      "scrape",
	Usage:     "Stores state hashes into TargetDb for given range",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.TargetDbFlag,
		&utils.ChainIDFlag,
		&utils.ClientDbFlag,
		&logger.LogLevelFlag,
	},
}

// scrapeAction stores state hashes into Target for given range
func scrapeAction(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "UtilDb-Scrape")
	log.Infof("Scraping for range %d-%d", cfg.First, cfg.Last)

	database, err := db.NewDefaultBaseDB(cfg.TargetDb)
	if err != nil {
		return fmt.Errorf("error opening stateHash leveldb %s: %v", cfg.TargetDb, err)
	}
	defer database.Close()

	err = utils.StateAndBlockHashScraper(ctx.Context, cfg.ChainID, cfg.ClientDb, database, cfg.First, cfg.Last, log)
	if err != nil {
		return err
	}

	log.Infof("Scraping finished")
	return nil
}
