package info

import (
	"errors"
	"fmt"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utildb/dbcomponent"
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
		count := utildb.GetSubstateCount(cfg, sdb)
		log.Noticef("Found %v substates", count)
	}

	// print update count
	if dbComponent == dbcomponent.Update || dbComponent == dbcomponent.All {
		count, err := utildb.GetUpdateCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print update count; %w", err)
		} else {
			log.Noticef("Found %v updates", count)
		}
	}

	// print deleted count
	if dbComponent == dbcomponent.Delete || dbComponent == dbcomponent.All {
		count, err := utildb.GetDeletedCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print deleted count; %w", err)
		} else {
			log.Noticef("Found %v deleted accounts", count)
		}
	}

	// print state hash count
	if dbComponent == dbcomponent.StateHash || dbComponent == dbcomponent.All {
		count, err := utildb.GetStateHashCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print state hash count; %w", err)
		} else {
			log.Noticef("Found %v state-hashes", count)
		}
	}

	// print block hash count
	if dbComponent == dbcomponent.BlockHash || dbComponent == dbcomponent.All {
		count, err := utildb.GetBlockHashCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print block hash count; %w", err)
		} else {
			log.Noticef("Found %v block-hashes", count)
		}
	}

	// print exception count
	if dbComponent == dbcomponent.Exception || dbComponent == dbcomponent.All {
		count, err := utildb.GetExceptionCount(cfg, base)
		if err != nil {
			errResult = errors.Join(errResult, err)
			log.Warningf("cannot print exception count; %w", err)
		} else {
			log.Noticef("Found %v exceptions", count)
		}
	}

	return errResult
}
