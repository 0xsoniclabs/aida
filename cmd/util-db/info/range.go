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

package info

import (
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utildb/dbcomponent"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var printRangeCommand = cli.Command{
	Action: rangeAction,
	Name:   "range",
	Usage:  "Prints range of all types in AidaDb",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.DbComponentFlag,
		&utils.SubstateEncodingFlag,
		&logger.LogLevelFlag,
	},
}

// rangeAction prints range of given db component in given AidaDb
func rangeAction(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.NoArgs)
	if argErr != nil {
		return argErr
	}
	log := logger.NewLogger(cfg.LogLevel, "AidaDb-Range")

	return printRange(cfg, log)
}

// printRange prints range of given db component in given AidaDb
func printRange(cfg *utils.Config, log logger.Logger) error {
	dbComponent, err := dbcomponent.ParseDbComponent(cfg.DbComponent)
	if err != nil {
		return err
	}

	baseDb, err := db.NewReadOnlySubstateDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}

	// print substate range
	if dbComponent == dbcomponent.Substate || dbComponent == dbcomponent.All {
		sdb, err := db.MakeDefaultSubstateDBFromBaseDB(baseDb)
		if err != nil {
			return err
		}
		err = sdb.SetSubstateEncoding(cfg.SubstateEncoding)
		if err != nil {
			return fmt.Errorf("cannot set substate encoding; %w", err)
		}

		firstBlock, lastBlock, ok := utils.FindBlockRangeInSubstate(sdb)
		if !ok {
			log.Warning("No substate found")
		} else {
			log.Infof("Substate block range: %v - %v", firstBlock, lastBlock)
		}
	}

	// print update range
	if dbComponent == dbcomponent.Update || dbComponent == dbcomponent.All {
		udb, err := db.MakeDefaultUpdateDBFromBaseDB(baseDb)
		if err != nil {
			return err
		}
		firstUsBlock, lastUsBlock, err := utildb.FindBlockRangeInUpdate(udb)
		if err != nil {
			log.Warningf("cannot find updateset range; %w", err)
		} else {
			log.Infof("Updateset block range: %v - %v", firstUsBlock, lastUsBlock)
		}
	}

	// print deleted range
	if dbComponent == dbcomponent.Delete || dbComponent == dbcomponent.All {
		ddb, err := db.MakeDefaultDestroyedAccountDBFromBaseDB(baseDb)
		if err != nil {
			return err
		}
		first, last, err := utildb.FindBlockRangeInDeleted(ddb)
		if err != nil {
			log.Warningf("cannot find deleted range; %w", err)
		} else {
			log.Infof("Deleted block range: %v - %v", first, last)
		}
	}

	// print state hash range
	if dbComponent == dbcomponent.StateHash || dbComponent == dbcomponent.All {
		shdb := db.MakeDefaultStateHashDBFromBaseDB(baseDb)
		firstStateHashBlock, lastStateHashBlock, err := utildb.FindBlockRangeInStateHash(shdb)
		if err != nil {
			log.Warningf("cannot find state hash range; %w", err)
		} else {
			log.Infof("State Hash range: %v - %v", firstStateHashBlock, lastStateHashBlock)
		}
	}

	// print block hash range
	if dbComponent == dbcomponent.BlockHash || dbComponent == dbcomponent.All {
		bhbd := db.MakeDefaultBlockHashDBFromBaseDB(baseDb)
		firstBlockHashBlock, lastBlockHashBlock, err := utildb.FindBlockRangeOfBlockHashes(bhbd)
		if err != nil {
			log.Warningf("cannot find block hash range; %w", err)
		} else {
			log.Infof("Block Hash range: %v - %v", firstBlockHashBlock, lastBlockHashBlock)
		}
	}

	// print exception range
	if dbComponent == dbcomponent.Exception || dbComponent == dbcomponent.All {
		edb := db.MakeDefaultExceptionDBFromBaseDB(baseDb)
		firstUsBlock, lastUsBlock, err := utildb.FindBlockRangeInException(edb)
		if err != nil {
			log.Warningf("cannot find exception range; %w", err)
		} else {
			log.Infof("Exception block range: %v - %v", firstUsBlock, lastUsBlock)
		}
	}

	err = baseDb.Close()
	if err != nil {
		return fmt.Errorf("cannot close aida db; %w", err)
	}
	return nil
}
