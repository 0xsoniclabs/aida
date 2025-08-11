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

package compact

import (
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/urfave/cli/v2"
)

// Command compactAction given database
var Command = cli.Command{
	Action: compactAction,
	Name:   "compactAction",
	Usage:  "compactAction target db",
	Flags: []cli.Flag{
		&utils.TargetDbFlag,
	},
	Description: `
Compacts target database.
`,
}

// compactAction compacts database
func compactAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.NoArgs)
	if err != nil {
		return err
	}

	log := logger.NewLogger(cfg.LogLevel, "aida-db-compact")

	targetDb, err := leveldb.New(cfg.TargetDb, 1024, 100, "profiling", false)
	if err != nil {
		return fmt.Errorf("cannot open db; %v", err)
	}

	log.Notice("Starting compaction")

	err = targetDb.Compact(nil, nil)
	if err != nil {
		return err
	}

	log.Notice("Compaction finished")

	return nil
}
