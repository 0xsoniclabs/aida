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

package metadata

import (
	"errors"

	"github.com/0xsoniclabs/aida/cmd/util-db/dbutils"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var generateCommand = cli.Command{
	Action: generateAction,
	Name:   "generate",
	Usage:  "Generates new metadata for given chain-id",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&utils.ChainIDFlag,
	},
}

func generateAction(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.NoArgs)
	if argErr != nil {
		return argErr
	}

	base, err := db.NewDefaultBaseDB(cfg.AidaDb)
	if err != nil {
		return err
	}

	defer base.Close()
	sdb := db.MakeDefaultSubstateDBFromBaseDB(base)
	fb, lb, ok := utils.FindBlockRangeInSubstate(sdb)
	if !ok {
		return errors.New("cannot find block range in substate")
	}

	md := utils.NewAidaDbMetadata(base, "INFO")
	md.FirstBlock = fb
	md.LastBlock = lb
	if err = md.SetFreshMetadata(cfg.ChainID); err != nil {
		return err
	}

	return dbutils.PrintMetadata(cfg.AidaDb)

}
