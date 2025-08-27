package metadata

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// insertCommand is a generic command for inserting any metadata key/value pair into AidaDb
var insertCommand = cli.Command{
	Action: insertAction,
	Name:   "insert",
	Usage:  "inserts key/value metadata pair into AidaDb",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
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

	aidaDbPath := ctx.String(utils.AidaDbFlag.Name)

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

	md := utils.NewAidaDbMetadata(base, "INFO")

	switch db.MetadataPrefix + keyArg {
	case utils.FirstBlockPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetFirstBlock(val); err != nil {
			return err
		}
	case utils.LastBlockPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetLastBlock(val); err != nil {
			return err
		}
	case utils.FirstEpochPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetFirstEpoch(val); err != nil {
			return err
		}
	case utils.LastEpochPrefix:
		val, err = strconv.ParseUint(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetLastEpoch(val); err != nil {
			return err
		}
	case utils.TypePrefix:
		num64, err := strconv.ParseUint(valArg, 10, 8)
		if err != nil {
			return err
		}
		if err = md.SetDbType(utils.AidaDbType(uint8(num64))); err != nil {
			return err
		}
	case utils.ChainIDPrefix:
		val, err := strconv.ParseInt(valArg, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse uint %v; %v", valArg, err)
		}
		if err = md.SetChainID(utils.ChainID(val)); err != nil {
			return err
		}
	case utils.TimestampPrefix:
		if err = md.SetTimestamp(); err != nil {
			return err
		}
	case utils.DbHashPrefix:
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
