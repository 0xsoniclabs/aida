package generate

import (
	"fmt"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utildb/metadata"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var generateDbHashCommand = cli.Command{
	Action: generateDbHashAction,
	Name:   "db-hash",
	Usage:  "Generates new db-hash. Note that this will overwrite the current AidaDb hash.",
	Flags: []cli.Flag{
		&config.AidaDbFlag,
	},
}

// generateDbHashAction calculates the dbHash for given AidaDb and saves it.
func generateDbHashAction(ctx *cli.Context) error {
	log := logger.NewLogger("INFO", "DbHashGenerateCMD")

	cfg, err := config.NewConfig(ctx, config.NoArgs)
	if err != nil {
		return err
	}

	aidaDb, err := db.NewDefaultBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open db; %v", err)
	}

	defer utildb.MustCloseDB(aidaDb)

	md := metadata.NewAidaDbMetadata(aidaDb, "INFO")

	log.Noticef("Starting DbHash generation for %v; this may take several hours...", cfg.AidaDb)
	hash, err := utildb.GenerateDbHash(aidaDb, "INFO")
	if err != nil {
		return err
	}

	err = md.SetDbHash(hash)
	if err != nil {
		return fmt.Errorf("cannot set db-hash; %v", err)
	}

	return nil
}
