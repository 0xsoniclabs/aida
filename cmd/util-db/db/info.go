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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/aida/cmd/util-db/flags"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utildb/dbcomponent"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var InfoCommand = cli.Command{
	Name:  "info",
	Usage: "Prints information about AidaDb",
	Subcommands: []*cli.Command{
		&cmdDelAcc,
		&cmdCount,
		&cmdRange,
		&cmdPrintStateHash,
		&cmdPrintBlockHash,
		&cmdPrintException,
	},
}

var cmdCount = cli.Command{
	Action:    printCountRun,
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

var cmdRange = cli.Command{
	Action: printRangeRun,
	Name:   "range",
	Usage:  "Prints range of all types in AidaDb",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.DbComponentFlag,
		&utils.SubstateEncodingFlag,
		&logger.LogLevelFlag,
	},
}

var cmdDelAcc = cli.Command{
	Action:    printDeletedAccountInfo,
	Name:      "del-acc",
	Usage:     "Prints deletion info about an account in AidaDb.",
	ArgsUsage: "<firstBlockNum>, <lastBlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&logger.LogLevelFlag,
		&flags.Account,
	},
}

var cmdPrintStateHash = cli.Command{
	Action:    printStateHash,
	Name:      "state-hash",
	Usage:     "Prints state hash for given block number.",
	ArgsUsage: "<BlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

var cmdPrintBlockHash = cli.Command{
	Action:    printBlockHash,
	Name:      "block-hash",
	Usage:     "Prints block hash for given block number.",
	ArgsUsage: "<BlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

var cmdPrintException = cli.Command{
	Action:    printException,
	Name:      "exception",
	Usage:     "Prints exception for given block number.",
	ArgsUsage: "<BlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

// printCountRun prints count of given db component in given AidaDb
func printCountRun(ctx *cli.Context) error {
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

// printRangeRun prints range of given db component in given AidaDb
func printRangeRun(ctx *cli.Context) error {
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

	baseDb, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}

	// print substate range
	if dbComponent == dbcomponent.Substate || dbComponent == dbcomponent.All {
		sdb := db.MakeDefaultSubstateDBFromBaseDB(baseDb)
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
		udb := db.MakeDefaultUpdateDBFromBaseDB(baseDb)
		firstUsBlock, lastUsBlock, err := utildb.FindBlockRangeInUpdate(udb)
		if err != nil {
			log.Warningf("cannot find updateset range; %w", err)
		} else {
			log.Infof("Updateset block range: %v - %v", firstUsBlock, lastUsBlock)
		}
	}

	// print deleted range
	if dbComponent == dbcomponent.Delete || dbComponent == dbcomponent.All {
		ddb := db.MakeDefaultDestroyedAccountDBFromBaseDB(baseDb)
		first, last, err := utildb.FindBlockRangeInDeleted(ddb)
		if err != nil {
			log.Warningf("cannot find deleted range; %w", err)
		} else {
			log.Infof("Deleted block range: %v - %v", first, last)
		}
	}

	// print state hash range
	if dbComponent == dbcomponent.StateHash || dbComponent == dbcomponent.All {
		firstStateHashBlock, lastStateHashBlock, err := utildb.FindBlockRangeInStateHash(baseDb, log)
		if err != nil {
			log.Warningf("cannot find state hash range; %w", err)
		} else {
			log.Infof("State Hash range: %v - %v", firstStateHashBlock, lastStateHashBlock)
		}
	}

	// print block hash range
	if dbComponent == dbcomponent.BlockHash || dbComponent == dbcomponent.All {
		firstBlockHashBlock, lastBlockHashBlock, err := utildb.FindBlockRangeOfBlockHashes(baseDb, log)
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

// printDeletedAccountInfo for given deleted account in AidaDb
func printDeletedAccountInfo(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "AidaDb-Deleted-Account-Info")

	db, err := db.NewReadOnlyDestroyedAccountDB(cfg.DeletionDb)
	if err != nil {
		return err
	}

	accounts, err := db.GetAccountsDestroyedInRange(cfg.First, cfg.Last)
	if err != nil {
		return fmt.Errorf("cannot Get all destroyed accounts; %w", err)
	}

	wantedAcc := ctx.String(flags.Account.Name)

	for _, acc := range accounts {
		if strings.Compare(acc.String(), wantedAcc) == 0 {
			log.Noticef("Account %v, got deleted in %v - %v", wantedAcc, cfg.First, cfg.Last)
			return nil
		}
	}

	log.Warningf("Account %v, didn't get deleted in %v - %v", cfg.First, cfg.Last)

	return nil

}

// printTableHash creates hash of substates, updatesets, deletion and state-hashes.
func printTableHash(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	database, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return err
	}

	defer database.Close()

	log := logger.NewLogger(cfg.LogLevel, "printTableHash")
	log.Info("Inspecting database...")
	err = utildb.TableHash(cfg, database, log)
	if err != nil {
		return err
	}
	log.Info("Finished")

	return nil
}

func printStateHash(ctx *cli.Context) error {
	return printHash(ctx, "state-hash")
}

func printBlockHash(ctx *cli.Context) error {
	return printHash(ctx, "block-hash")
}

func printHash(ctx *cli.Context, hashType string) error {
	cfg, argErr := utils.NewConfig(ctx, utils.OneToNArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "AidaDb-Print-"+strings.Title(hashType))

	if ctx.Args().Len() != 1 {
		return fmt.Errorf("%s command requires exactly 1 argument", hashType)
	}
	blockNumStr := ctx.Args().Slice()[0]
	blockNumInt32, err := strconv.ParseInt(blockNumStr, 10, 32)
	if err != nil {
		return fmt.Errorf("cannot parse block number %s; %v", blockNumStr, err)
	}
	blockNum := int(blockNumInt32)

	return printHashForBlock(cfg, log, blockNum, hashType)
}

// printHashForBlock prints state or block hash for given block number in AidaDb
func printHashForBlock(cfg *utils.Config, log logger.Logger, blockNum int, hashType string) error {
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

	provider := utils.MakeHashProvider(base)
	switch hashType {
	case "state-hash":
		bytes, err := provider.GetStateRootHash(blockNum)
		if err != nil {
			return fmt.Errorf("cannot get state hash for block %v; %v", blockNum, err)
		}
		log.Noticef("State hash for block %v is %v", blockNum, bytes)
	case "block-hash":
		bytesHash, err := provider.GetBlockHash(blockNum)
		if err != nil {
			return fmt.Errorf("cannot get block hash for block %v; %v", blockNum, err)
		}
		log.Noticef("Block hash for block %v is %v", blockNum, bytesHash)
	default:
		return fmt.Errorf("unsupported hash type: %s", hashType)
	}

	return nil
}

func printException(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.OneToNArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "AidaDb-PrintException")

	if ctx.Args().Len() != 1 {
		return fmt.Errorf("printException command requires exactly 1 argument")
	}
	blockNumStr := ctx.Args().Slice()[0]
	blockNum, err := strconv.ParseUint(ctx.Args().Slice()[0], 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse block number %s; %v", blockNumStr, err)
	}

	return printExceptionForBlock(cfg, log, blockNum)
}

func printExceptionForBlock(cfg *utils.Config, log logger.Logger, blockNum uint64) error {
	exceptionDb, err := db.NewReadOnlyExceptionDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}

	exception, err := exceptionDb.GetException(blockNum)
	if err != nil {
		return fmt.Errorf("cannot get exception for block %d; %v", blockNum, err)
	}

	log.Noticef("Exception for block %v: %v", blockNum, exception)

	return nil
}

// printPrefixHash calculates md5 of prefix in given AidaDb
func printPrefixHash(ctx *cli.Context) error {
	log := logger.NewLogger("INFO", "GeneratePrefixHash")

	cfg, err := utils.NewConfig(ctx, utils.NoArgs)

	database, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return err
	}

	defer database.Close()

	if ctx.Args().Len() == 0 || ctx.Args().Len() >= 2 {
		return fmt.Errorf("generate-prefix-hash command requires exactly 1 argument")
	}

	prefix := ctx.Args().Slice()[0]
	log.Noticef("Generating hash for prefix %v", prefix)
	_, err = utildb.GeneratePrefixHash(database, prefix, "INFO")
	return err
}
