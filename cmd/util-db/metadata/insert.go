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

package metadata

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/utildb/metadata"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// insertCommand is a generic command for inserting any metadata key/value pair into AidaDb
var insertCommand = cli.Command{
	Action: insertAction,
	Name:   "insert",
	Usage:  "inserts key/value metadata pair into AidaDb",
	Flags: []cli.Flag{
		&config.AidaDbFlag,
	},
	Description: `
Inserts key/value pair into AidaDb according to arguments:
<key> <value>
If given key is not metadata-key, operation fails.
`,
}

// insertAction key/value pair into AidaDb
func insertAction(ctx *cli.Context) (finalErr error) {
	var (
		err error
		val uint64
	)

	aidaDbPath := ctx.String(config.AidaDbFlag.Name)

	if ctx.Args().Len() != 2 {
		return fmt.Errorf("this command requires two arguments - <keyArg> <value>")
	}

	keyArg := ctx.Args().Get(0)
	valArg := ctx.Args().Get(1)

	// open db
	base, err := db.NewDefaultBaseDB(aidaDbPath)
	if err != nil {
		return err
	}

	defer func() {
		finalErr = errors.Join(finalErr, base.Close())
	}()

	md := metadata.NewAidaDbMetadata(base, "INFO")

	switch db.MetadataPrefix + keyArg {
	case metadata.FirstBlockPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetFirstBlock(val); err != nil {
			return err
		}
	case metadata.LastBlockPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetLastBlock(val); err != nil {
			return err
		}
	case metadata.FirstEpochPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetFirstEpoch(val); err != nil {
			return err
		}
	case metadata.LastEpochPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetLastEpoch(val); err != nil {
			return err
		}
	case metadata.TypePrefix:
		num64, err := strconv.ParseUint(valArg, 10, 8)
		if err != nil {
			return err
		}
		if err = md.SetDbType(metadata.AidaDbType(uint8(num64))); err != nil {
			return err
		}
	case metadata.ChainIDPrefix:
		val, err := strconv.ParseInt(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetChainID(config.ChainID(val)); err != nil {
			return err
		}
	case metadata.TimestampPrefix:
		if err = md.SetTimestamp(); err != nil {
			return err
		}
	case metadata.DbHashPrefix:
		hash, err := hex.DecodeString(valArg)
		if err != nil {
			return fmt.Errorf("cannot decode db-hash string into []byte; %v", err)
		}
		if err = md.SetDbHash(hash); err != nil {
			return err
		}
	case db.UpdatesetIntervalKey:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetUpdatesetInterval(val); err != nil {
			return err
		}
	case db.UpdatesetSizeKey:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetUpdatesetSize(val); err != nil {
			return err
		}
	default:
		return fmt.Errorf("incorrect keyArg: %v", keyArg)
	}

	return nil
}
