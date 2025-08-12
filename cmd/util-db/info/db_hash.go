package info

import (
	"encoding/hex"
	"fmt"
	"github.com/0xsoniclabs/aida/cmd/util-db/flags"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var printDbHashCommand = cli.Command{
	Action: printDbHashAction,
	Name:   "db-hash",
	Usage:  "Prints db-hash (md5) of AidaDb. If this it is not present, it is generated.",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&flags.ForceFlag,
	},
}

func printDbHashAction(ctx *cli.Context) error {
	var force = ctx.Bool(flags.ForceFlag.Name)

	aidaDb, err := db.NewReadOnlyBaseDB(ctx.String(utils.AidaDbFlag.Name))
	if err != nil {
		return fmt.Errorf("cannot open db; %v", err)
	}

	defer utildb.MustCloseDB(aidaDb)

	var dbHash []byte

	log := logger.NewLogger("INFO", "AidaDb-Db-Hash")

	md := utils.NewAidaDbMetadata(aidaDb, "INFO")

	// first try to extract from db
	dbHash = md.GetDbHash()
	if len(dbHash) != 0 && !force {
		log.Infof("Db-Hash (metadata): %v", hex.EncodeToString(dbHash))
		return nil
	}

	// if not found in db, we need to iterate and create the hash
	if dbHash, err = utildb.GenerateDbHash(aidaDb, "INFO"); err != nil {
		return err
	}

	fmt.Printf("Db-Hash (calculated): %v", hex.EncodeToString(dbHash))
	return nil
}
