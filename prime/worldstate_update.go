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

package prime

import (
	"errors"
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
)

// generateUpdateSet generates an update set for a block range.
func generateUpdateSet(first uint64, last uint64, cfg *utils.Config, sdb db.SubstateDB, ddb db.DestroyedAccountDB) (substate.WorldState, []types.Address, error) {
	var (
		deletedAccounts []types.Address
	)

	substateIter := sdb.NewSubstateIterator(int(first), cfg.Workers)
	update := make(substate.WorldState)
	defer substateIter.Release()

	// Todo rewrite in wrapping functions
	for substateIter.Next() {
		tx := substateIter.Value()
		// exceeded block range?
		if tx.Block > last {
			break
		}

		// if this transaction has suicided accounts, clear their states.
		destroyed, resurrected, err := ddb.GetDestroyedAccounts(tx.Block, tx.Transaction)
		if err != nil {
			return update, deletedAccounts, fmt.Errorf("failed to get deleted account. %v", err)
		}
		// reset storage of destroyed accounts
		if len(destroyed) > 0 {
			deletedAccounts = append(deletedAccounts, destroyed...)
		}
		if len(resurrected) > 0 {
			// Resurrected accounts are contained in this transaction, it should be cleared before merging.
			// This is to ensure that storage keys are consistent with the new data in the substate.
			// It is possible that the old value has additional keys which may not get replaced.

			// Because we know that resurrected account will have latest value from this transaction,
			// we can safely clear the storage here, and add it to list of accounts to be deleted before priming.
			deletedAccounts = append(deletedAccounts, resurrected...)
			ClearAccountStorage(update, resurrected)
		}

		// merge output substate to update
		update.Merge(tx.OutputSubstate)
	}
	return update, deletedAccounts, nil
}

// ClearAccountStorage clears storage of all input accounts.
func ClearAccountStorage(update substate.WorldState, accounts []types.Address) {
	for _, addr := range accounts {
		if _, found := update[addr]; found {
			update[addr].Storage = make(map[types.Hash]types.Hash)
		}
	}
}

// deleteDestroyedAccountsFromWorldState removes previously suicided accounts from
// the world state.
func deleteDestroyedAccountsFromWorldState(ws txcontext.WorldState, cfg *utils.Config, target uint64) (err error) {
	log := logger.NewLogger(cfg.LogLevel, "DelDestAcc")

	src, err := db.NewReadOnlyDestroyedAccountDB(cfg.DeletionDb)
	if err != nil {
		return err
	}
	defer func(src db.DestroyedAccountDB) {
		err = errors.Join(err, src.Close())
	}(src)
	list, err := src.GetAccountsDestroyedInRange(0, target)
	if err != nil {
		return err
	}
	for _, cur := range list {
		if ws.Has(common.Address(cur)) {
			log.Debugf("Remove %v from world state", cur)
			ws.Delete(common.Address(cur))
		}
	}
	return nil
}
