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

package updateset

import (
	"errors"
	"fmt"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

var UpdateSetStatsCommand = cli.Command{
	Action:    reportUpdateSetStats,
	Name:      "stats",
	Usage:     "print number of accounts and storage keys in update-set",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.UpdateDbFlag,
	},
	Description: `
The stats command requires one arguments: <blockNumLast> -- the last block of update-set.`,
}

// reportUpdateSetStats reports number of accounts and storage keys in an update-set
func reportUpdateSetStats(ctx *cli.Context) error {
	var (
		err error
	)
	cfg, argErr := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if argErr != nil {
		return argErr
	}
	// initialize updateDB
	udb, err := db.NewDefaultUpdateDB(cfg.UpdateDb)
	if err != nil {
		return err
	}
	defer func(udb db.UpdateDB) {
		err = errors.Join(err, udb.Close())
	}(udb)

	iter := udb.NewUpdateSetIterator(cfg.First, cfg.Last)
	defer iter.Release()

	for iter.Next() {
		update := iter.Value()
		state := update.WorldState
		fmt.Printf("%v,%v,", update.Block, len(state))
		storage := 0
		for account := range state {
			storage = storage + len(state[account].Storage)
		}
		fmt.Printf("%v\n", storage)
	}

	if iter.Error() != nil {
		err = errors.Join(err, fmt.Errorf("failed to iterate update-set: %v", iter.Error()))
	}
	return err
}
