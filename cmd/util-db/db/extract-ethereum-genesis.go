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
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/urfave/cli/v2"
)

var ExtractEthereumGenesisCommand = cli.Command{
	Action: extractEthereumGenesis,
	Name:   "extract-ethereum-genesis",
	Usage:  "Extracts WorldState from json into first updateset",
	Flags: []cli.Flag{
		&utils.ChainIDFlag,
		&utils.UpdateDbFlag,
		&logger.LogLevelFlag,
	},
	Description: `
Extracts WorldState from ethereum genesis.json into first updateset.`}

func extractEthereumGenesis(ctx *cli.Context) error {
	// process arguments and flags
	if ctx.Args().Len() != 1 {
		return fmt.Errorf("ethereum-update command requires exactly 1 arguments")
	}
	cfg, argErr := utils.NewConfig(ctx, utils.NoArgs)
	if argErr != nil {
		return argErr
	}
	log := logger.NewLogger(cfg.LogLevel, "Ethereum Update")

	log.Notice("Load Ethereum initial world state")
	ws, err := utildb.LoadEthereumGenesisWorldState(ctx.Args().Get(0))
	if err != nil {
		return err
	}

	udb, err := db.NewDefaultUpdateDB(cfg.UpdateDb)
	if err != nil {
		return err
	}
	defer udb.Close()

	log.Noticef("PutUpdateSet(0, %v, []common.Address{})", ws)

	return udb.PutUpdateSet(&updateset.UpdateSet{WorldState: ws, Block: 0}, make([]substatetypes.Address, 0))
}
