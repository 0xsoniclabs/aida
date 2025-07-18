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

package validator

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// validateWorldState compares states of accounts in stateDB to an expected set of states.
// If fullState mode, check if expected state is contained in stateDB.
// If partialState mode, check for equality of sets.
func validateWorldState(cfg *utils.Config, db state.VmStateDB, expectedAlloc txcontext.WorldState, log logger.Logger) error {
	var err error
	switch cfg.StateValidationMode {
	case utils.SubsetCheck:
		err = doSubsetValidation(expectedAlloc, db)
	case utils.EqualityCheck:
		vmAlloc := db.GetSubstatePostAlloc()
		isEqual := expectedAlloc.Equal(vmAlloc)
		if !isEqual {
			err = fmt.Errorf("inconsistent output: alloc")
			printAllocationDiffSummary(expectedAlloc, vmAlloc, log)

			return err
		}
	}
	return err
}

// printIfDifferent compares two values of any types and reports differences if any.
func printIfDifferent[T comparable](label string, want, have T, log logger.Logger) bool {
	if want != have {
		log.Errorf("Different %s:\nwant: %v\nhave: %v\n", label, want, have)
		return true
	}
	return false
}

// printIfDifferentBytes compares two values of byte type and reports differences if any.
func printIfDifferentBytes(label string, want, have []byte, log logger.Logger) bool {
	if !bytes.Equal(want, have) {
		log.Errorf("Different %s:\nwant: %v\nhave: %v\n", label, want, have)
		return true
	}
	return false
}

// printIfDifferentUint256 compares two values of big int type and reports differences if any.
func printIfDifferentUint256(label string, want, have *uint256.Int, log logger.Logger) bool {
	if want == nil && have == nil {
		return false
	}
	if want == nil || have == nil || want.Cmp(have) != 0 {
		log.Errorf("Different %s:\nwant: %v\nhave: %v\n", label, want, have)
		return true
	}
	return false
}

// printLogDiffSummary compares two tx logs and reports differences if any.
func printLogDiffSummary(label string, want, have *types.Log, log logger.Logger) {
	printIfDifferent(fmt.Sprintf("%s.address", label), want.Address, have.Address, log)
	if !printIfDifferent(fmt.Sprintf("%s.Topics size", label), len(want.Topics), len(have.Topics), log) {
		for i := range want.Topics {
			printIfDifferent(fmt.Sprintf("%s.Topics[%d]", label, i), want.Topics[i], have.Topics[i], log)
		}
	}
	printIfDifferentBytes(fmt.Sprintf("%s.data", label), want.Data, have.Data, log)
}

// printAllocationDiffSummary compares atrributes and existence of accounts and reports differences if any.
func printAllocationDiffSummary(want, have txcontext.WorldState, log logger.Logger) {
	printIfDifferent("substate alloc size", want.Len(), have.Len(), log)

	want.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		if have.Get(addr) == nil {
			log.Errorf("\tmissing address=%v\n", addr)
		}
	})

	have.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		if want.Get(addr) == nil {
			log.Errorf("\textra address=%v\n", addr)
		}
	})

	have.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		wantAcc := want.Get(addr)
		if wantAcc != nil {
			printAccountDiffSummary(fmt.Sprintf("key=%v:", addr), wantAcc, acc, log)
		}
	})

}

// PrintAccountDiffSummary compares attributes of two accounts and reports differences if any.
func printAccountDiffSummary(label string, want, have txcontext.Account, log logger.Logger) {
	printIfDifferent(fmt.Sprintf("%s.Nonce", label), want.GetNonce(), have.GetNonce(), log)
	printIfDifferentUint256(fmt.Sprintf("%s.Balance", label), want.GetBalance(), have.GetBalance(), log)
	printIfDifferentBytes(fmt.Sprintf("%s.Code", label), want.GetCode(), have.GetCode(), log)

	printIfDifferent(fmt.Sprintf("len(%s.Storage)", label), want.GetStorageSize(), have.GetStorageSize(), log)

	want.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
		haveValueHash := have.GetStorageAt(keyHash)
		if haveValueHash != valueHash {
			log.Errorf("\t%s.Storage misses key %v val %v\n", label, keyHash, valueHash)
		}
	})

	have.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
		wantValueHash := want.GetStorageAt(keyHash)
		if wantValueHash != valueHash {
			log.Errorf("\t%s.Storage has extra key %v\n", label, keyHash)
		}
	})

	have.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
		wantValueHash := want.GetStorageAt(keyHash)
		printIfDifferent(fmt.Sprintf("%s.Storage[%v]", label, keyHash), wantValueHash, valueHash, log)
	})

}

// doSubsetValidation validates whether the given alloc is contained in the db object.
// NB: We can only check what must be in the db (but cannot check whether db stores more).
func doSubsetValidation(alloc txcontext.WorldState, db state.VmStateDB) error {
	var err string

	alloc.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		if !db.Exist(addr) {
			err += fmt.Sprintf("  Account %v does not exist\n", addr.Hex())
		}
		accBalance := acc.GetBalance()

		if balance := db.GetBalance(addr); accBalance.Cmp(balance) != 0 {
			err += fmt.Sprintf("  Failed to validate balance for account %v\n"+
				"    have %v\n"+
				"    want %v\n",
				addr.Hex(), balance, accBalance)
		}
		if nonce := db.GetNonce(addr); nonce != acc.GetNonce() {
			err += fmt.Sprintf("  Failed to validate nonce for account %v\n"+
				"    have %v\n"+
				"    want %v\n",
				addr.Hex(), nonce, acc.GetNonce())
		}
		if code := db.GetCode(addr); bytes.Compare(code, acc.GetCode()) != 0 {
			err += fmt.Sprintf("  Failed to validate code for account %v\n"+
				"    have len %v\n"+
				"    want len %v\n",
				addr.Hex(), len(code), len(acc.GetCode()))
		}

		// validate Storage
		acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
			if db.GetState(addr, keyHash) != valueHash {
				err += fmt.Sprintf("  Failed to validate storage for account %v, key %v\n"+
					"    have %v\n"+
					"    want %v\n",
					addr.Hex(), keyHash.Hex(), db.GetState(addr, keyHash).Hex(), valueHash.Hex())
			}
		})

	})

	if len(err) > 0 {
		return fmt.Errorf("%s", err)
	}
	return nil
}

// updateStateDbOnEthereumChain is used to fix exceptions in ethereum dataset inconsistencies
func updateStateDbOnEthereumChain(alloc txcontext.WorldState, db state.StateDB, overwriteAccount bool) error {
	stateModified := false
	alloc.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		if !db.Exist(addr) {
			db.CreateAccount(addr)
		}

		accBalance := acc.GetBalance()
		balance := db.GetBalance(addr)
		// balance increments covers block rewards
		// or zero balance exception for slashed accounts - dao fork
		if overwriteAccount ||
			balance.Cmp(accBalance) < 0 ||
			(slices.Contains(params.DAODrainList(), addr) && accBalance.Eq(uint256.NewInt(0))) {
			if accBalance.Cmp(balance) != 0 {
				db.SubBalance(addr, balance, tracing.BalanceChangeUnspecified)
				db.AddBalance(addr, accBalance, tracing.BalanceChangeUnspecified)
			}
		}

		if overwriteAccount {
			if nonce := db.GetNonce(addr); nonce != acc.GetNonce() {
				db.SetNonce(addr, acc.GetNonce(), tracing.NonceChangeUnspecified)
			}
			if code := db.GetCode(addr); bytes.Compare(code, acc.GetCode()) != 0 {
				db.SetCode(addr, acc.GetCode())
			}
		}

		// BeaconRootsAddress, ConsolidationQueueAddress and WithdrawalQueueAddress are a special cases, where the storage is diverging
		if overwriteAccount || addr == params.BeaconRootsAddress || addr == params.ConsolidationQueueAddress || addr == params.WithdrawalQueueAddress {
			acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
				if db.GetState(addr, keyHash) != valueHash {
					db.SetState(addr, keyHash, valueHash)
					stateModified = true
				}
			})
		}
	})

	// If state was modified by this function, the current transaction must be
	// ended and a new transaction must be started to make sure that the
	// changed state is considered committed state. Without this, refunds for
	// state modifications are off while running the main transaction.
	if stateModified {
		err := db.EndTransaction()
		if err != nil {
			return fmt.Errorf("cannot end transaction: %w", err)
		}
		err = db.BeginTransaction(utils.PseudoTx) // the transaction number is ignored
		if err != nil {
			return fmt.Errorf("cannot end transaction: %w", err)
		}
	}

	return nil
}
