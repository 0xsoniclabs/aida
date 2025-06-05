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

package utils

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
)

// GenerateUpdateSet generates an update set for a block range.
func GenerateUpdateSet(first uint64, last uint64, cfg *Config, aidaDb db.BaseDB) (substate.WorldState, []substatetypes.Address, error) {
	var (
		deletedAccountDB *db.DestroyedAccountDB
		deletedAccounts  []substatetypes.Address
		err              error
	)
	sdb := db.MakeDefaultSubstateDBFromBaseDB(aidaDb)
	err = sdb.SetSubstateEncoding(cfg.SubstateEncoding)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set substate encoding. %v", err)
	}

	stateIter := sdb.NewSubstateIterator(int(first), cfg.Workers)
	update := make(substate.WorldState)
	defer stateIter.Release()

	// Todo rewrite in wrapping functions
	deletedAccountDB = db.MakeDefaultDestroyedAccountDBFromBaseDB(aidaDb)
	for stateIter.Next() {
		tx := stateIter.Value()
		// exceeded block range?
		if tx.Block > last {
			break
		}

		// if this transaction has suicided accounts, clear their states.
		destroyed, resurrected, err := deletedAccountDB.GetDestroyedAccounts(tx.Block, tx.Transaction)
		if err != nil {
			return update, deletedAccounts, fmt.Errorf("failed to get deleted account. %v", err)
		}
		// reset storagea
		if len(destroyed) > 0 {
			deletedAccounts = append(deletedAccounts, destroyed...)
		}
		if len(resurrected) > 0 {
			deletedAccounts = append(deletedAccounts, resurrected...)
			ClearAccountStorage(update, resurrected)
		}

		// merge output substate to update
		update.Merge(tx.OutputSubstate)
	}
	return update, deletedAccounts, nil
}

// GenerateWorldStateFromUpdateDB generates an initial world-state
// from pre-computed update-set
func GenerateWorldStateFromUpdateDB(cfg *Config, target uint64) (ws substate.WorldState, err error) {
	ws = make(substate.WorldState)
	block := uint64(0)
	// load pre-computed update-set from update-set db
	udb, err := db.NewDefaultUpdateDB(cfg.AidaDb)
	if err != nil {
		return nil, err
	}
	defer func(udb db.UpdateDB) {
		err = errors.Join(err, udb.Close())
	}(udb)
	updateIter := udb.NewUpdateSetIterator(block, target)
	for updateIter.Next() {
		blk := updateIter.Value()
		if blk.Block > target {
			break
		}
		block = blk.Block
		// Reset accessed storage locations of suicided accounts prior to updateset block.
		// The known accessed storage locations in the updateset range has already been
		// reset when generating the update set database.
		ClearAccountStorage(ws, blk.DeletedAccounts)
		ws.Merge(blk.WorldState)
		block++
	}
	updateIter.Release()

	// advance from the latest precomputed updateset to the target block
	update, _, err := GenerateUpdateSet(block, target, cfg, udb)
	if err != nil {
		return nil, err
	}
	ws.Merge(update)
	err = DeleteDestroyedAccountsFromWorldState(substatecontext.NewWorldState(ws), cfg, target)
	return ws, err
}

// ClearAccountStorage clears storage of all input accounts.
func ClearAccountStorage(update substate.WorldState, accounts []substatetypes.Address) {
	for _, addr := range accounts {
		if _, found := update[addr]; found {
			update[addr].Storage = make(map[substatetypes.Hash]substatetypes.Hash)
		}
	}
}

// OverwriteWorldState overwrites the StateDb with the expected state.
func OverwriteWorldState(cfg *Config, alloc txcontext.WorldState, db state.VmStateDB) error {
	if cfg.StateValidationMode != SubsetCheck {
		return nil
	}

	alloc.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		if !db.Exist(addr) {
			db.CreateAccount(addr)
		}
		accBalance := acc.GetBalance()
		balance := db.GetBalance(addr)
		if accBalance.Cmp(balance) != 0 {
			db.SubBalance(addr, balance, tracing.BalanceChangeUnspecified)
			db.AddBalance(addr, accBalance, tracing.BalanceChangeUnspecified)
		}
		if nonce := db.GetNonce(addr); nonce != acc.GetNonce() {
			db.SetNonce(addr, acc.GetNonce(), tracing.NonceChangeUnspecified)

		}
		if code := db.GetCode(addr); bytes.Compare(code, acc.GetCode()) != 0 {
			db.SetCode(addr, acc.GetCode())
		}

		acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
			if db.GetState(addr, keyHash) != valueHash {
				db.SetState(addr, keyHash, valueHash)
			}
		})

	})

	return nil
}
