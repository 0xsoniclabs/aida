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

package info

import (
	"errors"
	"fmt"

	"github.com/0xsoniclabs/aida/cmd/util-db/dbutils/dbcomponent"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var printCountCommand = cli.Command{
	Action:    printCountAction,
	Name:      "count",
	Usage:     "Count records in AidaDb.",
	ArgsUsage: "<firstBlockNum>, <lastBlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.DbComponentFlag,
		&utils.SubstateEncodingFlag,
		&logger.LogLevelFlag,
	},
}

// printCountAction prints count of given db component in given AidaDb
func printCountAction(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "AidaDb-Count")

	base, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return err
	}

	defer func() {
		err = base.Close()
		if err != nil {
			log.Warningf("Error closing aida db: %w", err)
		}
	}()

	return printCount(cfg, base, log)
}

// printCount prints count of given db component in given AidaDb
func printCount(cfg *utils.Config, base db.BaseDB, log logger.Logger) error {
	dbComponent, err := dbcomponent.ParseDbComponent(cfg.DbComponent)
	if err != nil {
		return err
	}

	log.Noticef("Inspecting database between blocks %v-%v", cfg.First, cfg.Last)

	var errResult error

	// print substate count
	if dbComponent == dbcomponent.Substate || dbComponent == dbcomponent.All {
		sdb := db.MakeDefaultSubstateDBFromBaseDB(base)
		err = sdb.SetSubstateEncoding(cfg.SubstateEncoding)
		if err != nil {
			return fmt.Errorf("cannot set substate encoding; %w", err)
		}
		count := getSubstateCount(cfg, sdb)
		log.Noticef("Found %v substates", count)
	}

	// print update count
	if dbComponent == dbcomponent.Update || dbComponent == dbcomponent.All {
		count, err := getUpdateCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print update count; %w", err)
		} else {
			log.Noticef("Found %v updates", count)
		}
	}

	// print deleted count
	if dbComponent == dbcomponent.Delete || dbComponent == dbcomponent.All {
		count, err := getDeletedCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print deleted count; %w", err)
		} else {
			log.Noticef("Found %v deleted accounts", count)
		}
	}

	// print state hash count
	if dbComponent == dbcomponent.StateHash || dbComponent == dbcomponent.All {
		count, err := getStateHashCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print state hash count; %w", err)
		} else {
			log.Noticef("Found %v state-hashes", count)
		}
	}

	// print block hash count
	if dbComponent == dbcomponent.BlockHash || dbComponent == dbcomponent.All {
		count, err := getBlockHashCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print block hash count; %w", err)
		} else {
			log.Noticef("Found %v block-hashes", count)
		}
	}

	// print exception count
	if dbComponent == dbcomponent.Exception || dbComponent == dbcomponent.All {
		count, err := getExceptionCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print exception count; %w", err)
		} else {
			log.Noticef("Found %v exceptions", count)
		}
	}

	return errResult
}
