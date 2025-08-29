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
	"strconv"
	"strings"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var printStateHashCommand = cli.Command{
	Action:    printStateHashAction,
	Name:      "state-hash",
	Usage:     "Prints state hash for given block number.",
	ArgsUsage: "<BlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

func printStateHashAction(ctx *cli.Context) error {
	return printHash(ctx, "state-hash")
}

var printBlockHashCommand = cli.Command{
	Action:    printBlockHashAction,
	Name:      "block-hash",
	Usage:     "Prints block hash for given block number.",
	ArgsUsage: "<BlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

func printBlockHashAction(ctx *cli.Context) error {
	return printHash(ctx, "block-hash")
}

// printTableHashCommand calculates md5 of actual data stored.
// Using []byte value from database, it decodes it and calculates md5 of the decoded objects.
var printTableHashCommand = cli.Command{
	Action: printTableHashAction,
	Name:   "print-table-hash",
	Usage:  "Calculates hash of AidaDb table data.",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.DbComponentFlag,
		&utils.SubstateEncodingFlag,
		&logger.LogLevelFlag,
	},
	Description: `
Creates hash of substates, updatesets, deletion and state-hashes using decoded objects from database rather than []byte value representation, because that is not deterministic.
`,
}

// printTableHashAction creates hash of substates, updatesets, deletion and state-hashes.
func printTableHashAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	database, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return err
	}

	log := logger.NewLogger(cfg.LogLevel, "printTableHash")
	defer func() {
		err = database.Close()
		if err != nil {
			log.Warningf("could not close database: %v", err)
		}
	}()

	log.Info("Inspecting database...")
	err = utildb.TableHash(cfg, database, log)
	if err != nil {
		return err
	}
	log.Info("Finished")

	return nil
}

var printPrefixHashCommand = cli.Command{
	Action:    printPrefixHashAction,
	Name:      "print-prefix-hash",
	Usage:     "Prints hash of data inside AidaDb for given prefix.",
	ArgsUsage: "<prefix>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

// printPrefixHashAction calculates md5 of prefix in given AidaDb
func printPrefixHashAction(ctx *cli.Context) error {
	log := logger.NewLogger("INFO", "GeneratePrefixHash")

	cfg, err := utils.NewConfig(ctx, utils.NoArgs)
	if err != nil {
		return err
	}

	database, err := db.NewReadOnlyBaseDB(cfg.AidaDb)
	if err != nil {
		return err
	}

	defer func() {
		err = database.Close()
		if err != nil {
			log.Warningf("could not close database: %v", err)
		}
	}()

	if ctx.Args().Len() == 0 || ctx.Args().Len() >= 2 {
		return fmt.Errorf("generate-prefix-hash command requires exactly 1 argument")
	}

	prefix := ctx.Args().Slice()[0]
	log.Noticef("Generating hash for prefix %v", prefix)
	_, err = utildb.GeneratePrefixHash(database, prefix, "INFO")
	return err
}

func printHash(ctx *cli.Context, hashType string) error {
	cfg, argErr := utils.NewConfig(ctx, utils.OneToNArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "AidaDb-Print-"+strings.ToUpper(hashType))

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
