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

package validate

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var Command = cli.Command{
	Action: validateAction,
	Name:   "validate",
	Usage:  "Validates AidaDb using md5 DbHash.",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.ChainIDFlag,
	},
}

// validateAction calculates the dbHash for given AidaDb and compares it to expected hash either found in metadata or online
func validateAction(ctx *cli.Context) error {
	log := logger.NewLogger("INFO", "ValidateCMD")

	cfg, err := utils.NewConfig(ctx, utils.NoArgs)
	if err != nil {
		return fmt.Errorf("cannot parse config; %v", err)
	}

	aidaDb, err := db.NewDefaultSubstateDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open db; %v", err)
	}

	defer utildb.MustCloseDB(aidaDb)
	md := utils.NewAidaDbMetadata(aidaDb, "INFO")

	if md.GetChainID() == 0 {
		log.Warning("cannot find db-hash in your aida-db metadata, this operation is needed because db-hash was not found inside your aida-db; please make sure you specified correct chain-id with flag --%v", utils.ChainIDFlag.Name)
		if err = md.SetChainID(cfg.ChainID); err != nil {
			return fmt.Errorf("cannot set chain-id inside aida-db metadata; %v", err)
		}
	}

	// validation only makes sense if user has pure AidaDb
	dbType := md.GetDbType()
	if dbType != utils.GenType {
		return fmt.Errorf("validation cannot be performed - your db type (%v) cannot be validated; aborting", dbType)
	}

	var saveHash = false
	// if db hash is not present, look for it in patches.json
	expectedHash := md.GetDbHash()
	if len(expectedHash) == 0 {
		// we want to save the hash inside metadata
		saveHash = true
		expectedHash, err = utildb.FindDbHashOnline(md.GetChainID(), log, md)
		if err != nil {
			return fmt.Errorf("validation cannot be performed - could not find expected db hash; %v", err)
		}
	}

	log.Noticef("Found DbHash for your Db: %v", hex.EncodeToString(expectedHash))

	log.Noticef("Starting DbHash calculation for %v; this may take several hours...", cfg.AidaDb)
	trueHash, err := utildb.GenerateDbHash(aidaDb, "INFO")
	if err != nil {
		return err
	}

	if !bytes.Equal(expectedHash, trueHash) {
		return fmt.Errorf("hashes are different! expected: %v; your aida-db:%v", hex.EncodeToString(expectedHash), hex.EncodeToString(trueHash))
	}

	log.Noticef("Validation successful!")

	if saveHash {
		err = md.SetDbHash(trueHash)
		if err != nil {
			return err
		}
	}

	return nil
}
