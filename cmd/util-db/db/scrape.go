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
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var ScrapeCommand = cli.Command{
	Action:    scrapePrepare,
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

// scrapePrepare stores state hashes into Target for given range
func scrapePrepare(ctx *cli.Context) error {
	cfg, argErr := config.NewConfig(ctx, config.BlockRangeArgs)
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
