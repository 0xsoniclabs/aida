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

package generate

import (
	"fmt"
	"github.com/0xsoniclabs/aida/cmd/util-db/dbutils"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var DbHashCommand = cli.Command{
	Action: generateDbHashAction,
	Name:   "generate-db-hash",
	Usage:  "Generates new db-hash. Note that this will overwrite the current AidaDb hash.",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

// generateDbHashAction calculates the dbHash for given AidaDb and saves it.
func generateDbHashAction(ctx *cli.Context) error {
	log := logger.NewLogger("INFO", "DbHashGenerateCMD")

	cfg, err := utils.NewConfig(ctx, utils.NoArgs)

	aidaDb, err := db.NewDefaultBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open db; %v", err)
	}

	defer dbutils.MustCloseDB(aidaDb)

	md := utils.NewAidaDbMetadata(aidaDb, "INFO")

	log.Noticef("Starting DbHash generation for %v; this may take several hours...", cfg.AidaDb)
	hash, err := dbutils.GenerateDbHash(aidaDb, "INFO")
	if err != nil {
		return err
	}

	err = md.SetDbHash(hash)
	if err != nil {
		return fmt.Errorf("cannot set db-hash; %v", err)
	}

	return nil
}
