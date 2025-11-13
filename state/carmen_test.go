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

package state

import (
	"bytes"
	"errors"
	"testing"

	"github.com/0xsoniclabs/carmen/go/carmen"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestCarmenState_MakeCarmenStateDBInvalid tests db initialization with invalid Variant
func TestCarmenState_MakeCarmenStateDBInvalid(t *testing.T) {
	csDB, err := MakeDefaultCarmenStateDB("", "invalid-Variant", 5, "")
	if errors.Is(err, carmen.UnsupportedConfiguration) {
		t.Skip("unsupported configuration")
	}

	if err == nil {
		err = csDB.Close()
		if err != nil {
			t.Fatalf("failed to close carmen state DB: %v", err)
		}

		t.Fatalf("failed to throw error while creating carmen state DB")
	}
}

// TestCarmenState_InitCloseCarmenDB test closing db immediately after initialization
func TestCarmenState_InitCloseCarmenDB(t *testing.T) {
	for _, tc := range GetAllCarmenConfigurations() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeDefaultCarmenStateDB(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			err = csDB.Close()
			if err != nil {
				t.Fatalf("failed to close carmen state DB: %v", err)
			}
		})
	}
}

// TestCarmenState_AccountLifecycle tests account operations - create, check if it exists, if it's empty, suicide and suicide confirmation
func TestCarmenState_AccountLifecycle(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}
			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			csDB.CreateAccount(addr)

			if !csDB.Exist(addr) {
				t.Fatal("failed to create carmen state DB account")
			}

			if !csDB.Empty(addr) {
				t.Fatal("failed to create carmen state DB account; should be empty")
			}

			csDB.SelfDestruct(addr)
			if !csDB.HasSelfDestructed(addr) {
				t.Fatal("failed to suicide carmen state DB account;")
			}
		})
	}
}

// TestCarmenState_AccountBalanceOperations tests balance operations - add, subtract and check if the value is correct
func TestCarmenState_AccountBalanceOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}
			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			csDB.CreateAccount(addr)

			// get randomized balance
			additionBase := GetRandom(t, 1, 5_000_000)
			addition := uint256.NewInt(additionBase)

			csDB.AddBalance(addr, addition, 0)

			if csDB.GetBalance(addr).Cmp(addition) != 0 {
				t.Fatal("failed to add balance to carmen state DB account")
			}

			subtraction := uint256.NewInt(GetRandom(t, 1, int(additionBase)))
			expectedResult := uint256.NewInt(0).Sub(addition, subtraction)

			csDB.SubBalance(addr, subtraction, 0)

			if csDB.GetBalance(addr).Cmp(expectedResult) != 0 {
				t.Fatal("failed to subtract balance to carmen state DB account")
			}
		})
	}
}

// TestCarmenState_NonceOperations tests account nonce updating
func TestCarmenState_NonceOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			csDB.CreateAccount(addr)

			// get randomized nonce
			newNonce := GetRandom(t, 1, 5_000_000)

			csDB.SetNonce(addr, newNonce, tracing.NonceChangeUnspecified)

			if csDB.GetNonce(addr) != newNonce {
				t.Fatal("failed to update account nonce")
			}
		})
	}
}

// TestCarmenState_CodeOperations tests account code updating
func TestCarmenState_CodeOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			csDB.CreateAccount(addr)

			// generate new randomized code
			code := MakeRandomByteSlice(t, 2048)

			if csDB.GetCodeSize(addr) != 0 {
				t.Fatal("failed to update account code; wrong initial size")
			}

			csDB.SetCode(addr, code, tracing.CodeChangeUnspecified)

			if !bytes.Equal(csDB.GetCode(addr), code) {
				t.Fatal("failed to update account code; wrong value")
			}

			if csDB.GetCodeSize(addr) != len(code) {
				t.Fatal("failed to update account code; wrong size")
			}
		})
	}
}

// TestCarmenState_StateOperations tests account state update
func TestCarmenState_StateOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			csDB.CreateAccount(addr)

			// generate state key and value
			key := common.BytesToHash(MakeRandomByteSlice(t, 32))
			value := common.BytesToHash(MakeRandomByteSlice(t, 32))

			csDB.SetState(addr, key, value)

			state := csDB.GetState(addr, key)
			if state != value {
				t.Fatal("failed to update account state")
			}

			state, committed := csDB.GetStateAndCommittedState(addr, key)
			if state != value {
				t.Fatal("failed to update account state")
			}
			if committed != (common.Hash{}) {
				t.Fatal("failed to update account state")
			}
		})
	}
}

// TestCarmenState_TrxBlockSyncPeriodOperations tests creation of randomized sync-periods with blocks and transactions
func TestCarmenState_TrxBlockSyncPeriodOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeDefaultCarmenStateDB(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = csDB.Close()
				if err != nil {
					t.Fatalf("failed to close carmen state DB: %v", err)
				}
			}(csDB)

			blockNumber := 1
			trxNumber := 1
			for i := 0; i < 10; i++ {
				csDB.BeginSyncPeriod(uint64(i))

				for j := 0; j < 100; j++ {
					err = csDB.BeginBlock(uint64(blockNumber))
					if err != nil {
						t.Fatalf("cannot begin block; %v", err)
					}
					blockNumber++

					for k := 0; k < 100; k++ {
						err = csDB.BeginTransaction(uint32(trxNumber))
						if err != nil {
							t.Fatalf("cannot begin transaction; %v", err)
						}
						trxNumber++
						err = csDB.EndTransaction()
						if err != nil {
							t.Fatalf("cannot end transaction; %v", err)
						}
					}

					err = csDB.EndBlock()
					if err != nil {
						t.Fatalf("cannot end block; %v", err)
					}
				}

				assert.NotPanics(t, csDB.EndSyncPeriod)
			}
		})
	}
}

// TestCarmenState_RefundOperations tests adding and subtracting refund value
func TestCarmenState_RefundOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}
			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			refundValue := GetRandom(t, 40_000_000, 50_000_000)
			csDB.AddRefund(refundValue)

			if csDB.GetRefund() != refundValue {
				t.Fatal("failed to add refund")
			}

			reducedRefund := refundValue - uint64(30000000)

			csDB.SubRefund(uint64(30000000))

			if csDB.GetRefund() != reducedRefund {
				t.Fatal("failed to subtract refund")
			}
		})
	}
}

// TestCarmenState_AccessListOperations tests operations with creating, updating a checking AccessList
func TestCarmenState_AccessListOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}
			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			// prepare content of access list
			sender := common.BytesToAddress(MakeRandomByteSlice(t, 40))
			dest := common.BytesToAddress(MakeRandomByteSlice(t, 40))
			precompiles := []common.Address{
				common.BytesToAddress(MakeRandomByteSlice(t, 40)),
				common.BytesToAddress(MakeRandomByteSlice(t, 40)),
				common.BytesToAddress(MakeRandomByteSlice(t, 40)),
			}
			txAccesses := types.AccessList{
				types.AccessTuple{
					Address: common.BytesToAddress(MakeRandomByteSlice(t, 40)),
					StorageKeys: []common.Hash{
						common.BytesToHash(MakeRandomByteSlice(t, 32)),
						common.BytesToHash(MakeRandomByteSlice(t, 32)),
					},
				},
				types.AccessTuple{
					Address: common.BytesToAddress(MakeRandomByteSlice(t, 40)),
					StorageKeys: []common.Hash{
						common.BytesToHash(MakeRandomByteSlice(t, 32)),
						common.BytesToHash(MakeRandomByteSlice(t, 32)),
						common.BytesToHash(MakeRandomByteSlice(t, 32)),
						common.BytesToHash(MakeRandomByteSlice(t, 32)),
					},
				},
			}

			// create access list
			csDB.Prepare(params.Rules{}, sender, common.Address{}, &dest, precompiles, txAccesses)

			// add some more data after the creation for good measure
			newAddr := common.BytesToAddress(MakeRandomByteSlice(t, 40))
			newSlot := common.BytesToHash(MakeRandomByteSlice(t, 32))
			csDB.AddAddressToAccessList(newAddr)
			csDB.AddSlotToAccessList(newAddr, newSlot)

			// check content of access list
			if !csDB.AddressInAccessList(sender) {
				t.Fatal("failed to add sender address to access list")
			}

			if !csDB.AddressInAccessList(dest) {
				t.Fatal("failed to add destination address to access list")
			}

			if !csDB.AddressInAccessList(newAddr) {
				t.Fatal("failed to add new address to access list after it was already created")
			}

			for _, addr := range precompiles {
				if !csDB.AddressInAccessList(addr) {
					t.Fatal("failed to add precompile address to access list")
				}
			}

			for _, txAccess := range txAccesses {
				if !csDB.AddressInAccessList(txAccess.Address) {
					t.Fatal("failed to add transaction access address to access list")
				}

				for _, storageKey := range txAccess.StorageKeys {
					addrOK, slotOK := csDB.SlotInAccessList(txAccess.Address, storageKey)
					if !addrOK || !slotOK {
						t.Fatal("failed to add transaction access address to access list")
					}
				}
			}

			addrOK, slotOK := csDB.SlotInAccessList(newAddr, newSlot)
			if !addrOK || !slotOK {
				t.Fatal("failed to add new slot to access list after it was already created")
			}
		})
	}
}

// TestCarmenState_GetArchiveState tests retrieving an Archive state
func TestCarmenState_GetArchiveState(t *testing.T) {
	cfgs := GetCarmenStateTestCases()
	for _, tc := range cfgs {
		if tc.Archive == "none" || tc.Archive == "" {
			continue // relevant only if the Archive is enabled
		}
		t.Run(tc.String(), func(t *testing.T) {
			tempDir := t.TempDir()
			csDB, err := MakeCarmenDbTestContext(tempDir, tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}
			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}
			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			csDB.CreateAccount(addr)

			// generate state key and value
			key := common.BytesToHash(MakeRandomByteSlice(t, 32))
			value := common.BytesToHash(MakeRandomByteSlice(t, 32))

			csDB.SetState(addr, key, value)

			err = CloseCarmenDbTestContext(csDB)
			if err != nil {
				t.Fatalf("cannot close carmen test context; %v", err)
			}

			csDB, err = MakeDefaultCarmenStateDB(tempDir, tc.Variant, tc.Schema, tc.Archive)
			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = csDB.Close()
				if err != nil {
					t.Fatalf("failed to close carmen state DB: %v", err)
				}
			}(csDB)

			archive, err := csDB.GetArchiveState(1)
			if err != nil {
				t.Fatalf("failed to retrieve Archive state of carmen state DB: %v", err)
			}

			err = archive.Release()
			if err != nil {
				t.Fatal("cannot release archive; %w", err)
			}
		})
	}
}

// TestCarmenState_SetBalanceUsingBulkInsertion tests setting an accounts balance
func TestCarmenState_SetBalanceUsingBulkInsertion(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeDefaultCarmenStateDB(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			cbl, err := csDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)
			}

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			newBalance := uint256.NewInt(GetRandom(t, 1, 5_000_000))
			cbl.SetBalance(addr, newBalance)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			err = BeginCarmenDbTestContext(csDB)
			if err != nil {
				t.Fatal(err)
			}

			if csDB.GetBalance(addr).Cmp(newBalance) != 0 {
				t.Fatal("failed to update account balance")
			}
		})
	}
}

// TestCarmenState_SetNonceUsingBulkInsertion tests setting an accounts nonce
func TestCarmenState_SetNonceUsingBulkInsertion(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeDefaultCarmenStateDB(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Closing of state DB
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			cbl, err := csDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)
			}

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			newNonce := GetRandom(t, 1, 5_000_000)

			cbl.SetNonce(addr, newNonce)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			err = BeginCarmenDbTestContext(csDB)
			if err != nil {
				t.Fatal(err)
			}

			if csDB.GetNonce(addr) != newNonce {
				t.Fatal("failed to update account nonce")
			}
		})
	}
}

// TestCarmenState_SetStateUsingBulkInsertion tests setting an accounts state
func TestCarmenState_SetStateUsingBulkInsertion(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeDefaultCarmenStateDB(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			cbl, err := csDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)
			}

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			// generate state key and value
			key := common.BytesToHash(MakeRandomByteSlice(t, 32))
			value := common.BytesToHash(MakeRandomByteSlice(t, 32))

			cbl.SetState(addr, key, value)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			err = BeginCarmenDbTestContext(csDB)
			if err != nil {
				t.Fatal(err)
			}

			if csDB.GetState(addr, key) != value {
				t.Fatal("failed to update account state")
			}
		})
	}
}

// TestCarmenState_SetCodeUsingBulkInsertion tests setting an accounts code
func TestCarmenState_SetCodeUsingBulkInsertion(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeDefaultCarmenStateDB(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = CloseCarmenDbTestContext(csDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(csDB)

			cbl, err := csDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)
			}

			addr := common.BytesToAddress(MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			// generate new randomized code
			code := MakeRandomByteSlice(t, 2048)

			cbl.SetCode(addr, code)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			err = BeginCarmenDbTestContext(csDB)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(csDB.GetCode(addr), code) {
				t.Fatal("failed to update account code")
			}
		})
	}
}

// TestCarmenState_BulkloadOperations tests multiple operation in one bulkload
func TestCarmenState_BulkloadOperations(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeCarmenDbTestContext(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				if err = csDB.Close(); err != nil {
					t.Fatalf("cannot close db; %v", err)
				}
			}(csDB)

			// generate 100 randomized accounts
			accounts := [100]common.Address{}

			for i := 0; i < len(accounts); i++ {
				accounts[i] = common.BytesToAddress(MakeRandomByteSlice(t, 40))
				csDB.CreateAccount(accounts[i])
			}

			if err = csDB.EndTransaction(); err != nil {
				t.Fatalf("cannot end tx; %v", err)
			}
			if err = csDB.EndBlock(); err != nil {
				t.Fatalf("cannot end block; %v", err)
			}

			cbl, err := csDB.StartBulkLoad(7)
			if err != nil {
				t.Fatal(err)
			}

			for _, account := range accounts {
				// randomized operation
				operationType := GetRandom(t, 0, 4)

				switch operationType {
				case 1:
					// set balance
					newBalance := uint256.NewInt(GetRandom(t, 0, 5_000_000))

					cbl.SetBalance(account, newBalance)
				case 2:
					// set code
					code := MakeRandomByteSlice(t, 2048)

					cbl.SetCode(account, code)
				case 3:
					// set state
					key := common.BytesToHash(MakeRandomByteSlice(t, 32))
					value := common.BytesToHash(MakeRandomByteSlice(t, 32))

					cbl.SetState(account, key, value)
				case 4:
					// set nonce
					newNonce := GetRandom(t, 0, 5_000_000)

					cbl.SetNonce(account, newNonce)
				default:
					// set code by default
					code := MakeRandomByteSlice(t, 2048)

					cbl.SetCode(account, code)
				}
			}

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}
		})
	}
}

// TestCarmenState_GetShadowDB tests retrieval of shadow DB

func TestCarmenState_GetShadowDB(t *testing.T) {
	for _, tc := range GetCarmenStateTestCases() {
		t.Run(tc.String(), func(t *testing.T) {
			csDB, err := MakeDefaultCarmenStateDB(t.TempDir(), tc.Variant, tc.Schema, tc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			// Close DB after test ends
			defer func(csDB StateDB) {
				err = csDB.Close()
				if err != nil {
					t.Fatalf("failed to close carmen state DB: %v", err)
				}
			}(csDB)

			// check that shadowDB returns the DB object itself
			if csDB.GetShadowDB() != nil {
				t.Fatal("failed to retrieve shadow DB")
			}
		})
	}
}

// carmenStateDB struct method tests
func TestCarmenStateDB_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().CreateAccount(carmen.Address(addr))
	c.CreateAccount(addr)
}

func TestCarmenStateDB_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().CreateContract(carmen.Address(addr))
	c.CreateContract(addr)
}

func TestCarmenStateDB_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().Exist(carmen.Address(addr)).Return(true)
	exists := c.Exist(addr)
	assert.True(t, exists)
}

func TestCarmenStateDB_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().Empty(carmen.Address(addr)).Return(true)
	empty := c.Empty(addr)
	assert.True(t, empty)
}

func TestCarmenStateDB_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().GetBalance(carmen.Address(addr)).Return(carmen.Amount{})
	mockTxCtx.EXPECT().SelfDestruct(carmen.Address(addr))
	a := c.SelfDestruct(addr)
	assert.Equal(t, uint256.Int{}, a)
}

func TestCarmenStateDB_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().GetBalance(carmen.Address(addr)).Return(carmen.Amount{})
	mockTxCtx.EXPECT().SelfDestruct6780(carmen.Address(addr)).Return(true)
	a, value := c.SelfDestruct6780(addr)
	assert.Equal(t, uint256.Int{}, a)
	assert.Equal(t, true, value)
}

func TestCarmenStateDB_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().HasSelfDestructed(carmen.Address(addr)).Return(true)
	hasSelfDestructed := c.HasSelfDestructed(addr)
	assert.True(t, hasSelfDestructed)
}

func TestCarmenStateDB_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	expectedBalance := uint256.NewInt(1000)
	mockTxCtx.EXPECT().GetBalance(carmen.Address(addr)).Return(carmen.NewAmount(uint64(1000)))
	balance := c.GetBalance(addr)
	assert.Equal(t, expectedBalance, balance)
}

func TestCarmenStateDB_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	additionalBalance := uint256.NewInt(500)
	mockTxCtx.EXPECT().GetBalance(carmen.Address(addr)).Return(carmen.NewAmount(uint64(500)))
	mockTxCtx.EXPECT().AddBalance(carmen.Address(addr), carmen.NewAmount(uint64(500)))
	value := c.AddBalance(addr, additionalBalance, tracing.BalanceChangeUnspecified)
	assert.Equal(t, uint256.NewInt(500), &value)
}

func TestCarmenStateDB_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	subtractBalance := uint256.NewInt(500)
	mockTxCtx.EXPECT().GetBalance(carmen.Address(addr)).Return(carmen.NewAmount(uint64(500)))
	mockTxCtx.EXPECT().SubBalance(carmen.Address(addr), carmen.NewAmount(uint64(500)))
	value := c.SubBalance(addr, subtractBalance, tracing.BalanceChangeUnspecified)
	assert.Equal(t, uint256.NewInt(500), &value)
}

func TestCarmenStateDB_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	expectedNonce := uint64(42)
	mockTxCtx.EXPECT().GetNonce(carmen.Address(addr)).Return(expectedNonce)
	nonce := c.GetNonce(addr)
	assert.Equal(t, expectedNonce, nonce)
}

func TestCarmenStateDB_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	newNonce := uint64(100)
	mockTxCtx.EXPECT().SetNonce(carmen.Address(addr), newNonce)
	c.SetNonce(addr, newNonce, tracing.NonceChangeUnspecified)
}

func TestCarmenStateDB_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	expectedValue := common.BytesToHash([]byte("testValue"))
	mockTxCtx.EXPECT().GetCommittedState(carmen.Address(addr), carmen.Key(key)).Return(carmen.Value(expectedValue))
	value := c.GetCommittedState(addr, key)
	assert.Equal(t, expectedValue, value)
}

func TestCarmenStateDB_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	expectedValue := common.BytesToHash([]byte("testValue"))
	mockTxCtx.EXPECT().GetState(carmen.Address(addr), carmen.Key(key)).Return(carmen.Value(expectedValue))
	value := c.GetState(addr, key)
	assert.Equal(t, expectedValue, value)
}

func TestCarmenStateDB_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	value := common.BytesToHash([]byte("testValue"))
	mockTxCtx.EXPECT().GetState(carmen.Address(addr), carmen.Key(key)).Return(carmen.Value{})
	mockTxCtx.EXPECT().SetState(carmen.Address(addr), carmen.Key(key), carmen.Value(value))
	out := c.SetState(addr, key, value)
	assert.Equal(t, common.Hash{}, out)
}

func TestCarmenStateDB_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().HasEmptyStorage(carmen.Address(addr)).Return(false)
	root := c.GetStorageRoot(addr)
	assert.Equal(t, common.Hash{0x01}, root)
}

func TestCarmenStateDB_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	value := common.BytesToHash([]byte("testValue"))
	mockTxCtx.EXPECT().SetTransientState(carmen.Address(addr), carmen.Key(key), carmen.Value(value))
	c.SetTransientState(addr, key, value)
}

func TestCarmenStateDB_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	expectedValue := common.BytesToHash([]byte("testValue"))
	mockTxCtx.EXPECT().GetTransientState(carmen.Address(addr), carmen.Key(key)).Return(carmen.Value(expectedValue))
	value := c.GetTransientState(addr, key)
	assert.Equal(t, expectedValue, value)
}

func TestCarmenStateDB_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	expectedCode := []byte{0x01, 0x02, 0x03}
	mockTxCtx.EXPECT().GetCode(carmen.Address(addr)).Return(expectedCode)
	code := c.GetCode(addr)
	assert.Equal(t, expectedCode, code)
}

func TestCarmenStateDB_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	expectedSize := 3
	mockTxCtx.EXPECT().GetCodeSize(carmen.Address(addr)).Return(expectedSize)
	size := c.GetCodeSize(addr)
	assert.Equal(t, expectedSize, size)
}

func TestCarmenStateDB_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	mockTxCtx.EXPECT().GetCodeHash(carmen.Address(addr)).Return(carmen.Hash{
		0x01, 0x02, 0x03,
	})
	hash := c.GetCodeHash(addr)
	assert.Equal(t, expectedHash, hash)
}

func TestCarmenStateDB_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	code := []byte{0x01, 0x02, 0x03}
	mockTxCtx.EXPECT().GetCode(carmen.Address(addr)).Return(code)
	mockTxCtx.EXPECT().SetCode(carmen.Address(addr), code)
	c.SetCode(addr, code, tracing.CodeChangeUnspecified)
}

func TestCarmenStateDB_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	snapshotID := 1
	mockTxCtx.EXPECT().Snapshot().Return(snapshotID)
	ss := c.Snapshot()
	assert.Equal(t, snapshotID, ss)
}

func TestCarmenStateDB_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	snapshotID := 1
	mockTxCtx.EXPECT().RevertToSnapshot(snapshotID)
	c.RevertToSnapshot(snapshotID)
}

func TestCarmenStateDB_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	mockTxCtx.EXPECT().Commit().Return(nil)
	err := c.EndTransaction()
	assert.NoError(t, err)
}

func TestCarmenStateDB_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockQueryCtx := carmen.NewMockQueryContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	mockQueryCtx.EXPECT().GetStateHash().Return(carmen.Hash{0x01, 0x02, 0x03})
	mockDb.EXPECT().QueryHeadState(gomock.Any()).Do(func(ff func(ctxt carmen.QueryContext)) {
		ff(mockQueryCtx)
	}).Return(nil)
	hash, err := c.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestCarmenStateDB_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	mockDb.EXPECT().Close().Return(nil)
	err := c.Close()
	assert.NoError(t, err)
}

func TestCarmenStateDB_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	mockTxCtx.EXPECT().AddRefund(uint64(100))
	c.AddRefund(uint64(100))
}

func TestCarmenStateDB_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	mockTxCtx.EXPECT().SubRefund(uint64(50))
	c.SubRefund(uint64(50))
}

func TestCarmenStateDB_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	expectedRefund := uint64(200)
	mockTxCtx.EXPECT().GetRefund().Return(expectedRefund)
	refund := c.GetRefund()
	assert.Equal(t, expectedRefund, refund)
}

func TestCarmenStateDB_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	sender := common.HexToAddress("0x1234")
	coinbase := common.HexToAddress("0x5678")
	dest := common.HexToAddress("0x9abc")
	mockTxCtx.EXPECT().ClearAccessList()
	mockTxCtx.EXPECT().AddAddressToAccessList(gomock.Any()).Times(3)
	c.Prepare(params.TestRules, sender, coinbase, &dest, []common.Address{sender}, nil)
}

func TestCarmenStateDB_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().IsAddressInAccessList(carmen.Address(addr)).Return(true)
	inList := c.AddressInAccessList(addr)
	assert.True(t, inList)
}

func TestCarmenStateDB_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	mockTxCtx.EXPECT().IsSlotInAccessList(carmen.Address(addr), carmen.Key(key)).Return(true, false)
	addrOk, slotOk := c.SlotInAccessList(addr, key)
	assert.True(t, addrOk)
	assert.False(t, slotOk)
}

func TestCarmenStateDB_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	mockTxCtx.EXPECT().AddAddressToAccessList(carmen.Address(addr))
	c.AddAddressToAccessList(addr)
}

func TestCarmenStateDB_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	mockTxCtx.EXPECT().AddSlotToAccessList(carmen.Address(addr), carmen.Key(key))
	c.AddSlotToAccessList(addr, key)
}

func TestCarmenStateDB_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToAddress("0x1234")
	topics := []common.Hash{common.HexToHash("0x1"), common.HexToHash("0x2")}
	data := []byte{0x01, 0x02, 0x03}
	mockTxCtx.EXPECT().AddLog(gomock.Any())
	c.AddLog(&types.Log{
		Address: addr,
		Topics:  topics,
		Data:    data,
	})
}

func TestCarmenStateDB_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	addr := common.HexToHash("0x1234")
	mockTxCtx.EXPECT().GetLogs().Return([]*carmen.Log{})
	logs := c.GetLogs(addr, uint64(0), addr, uint64(0))
	assert.Equal(t, []*types.Log{}, logs)
}

func TestCarmenStateDB_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	assert.Panics(t, func() {
		c.PointCache()
	})
}

func TestCarmenStateDB_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	out := c.Witness()
	assert.Nil(t, out)
}

func TestCarmenStateDB_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	assert.NotPanics(t, func() {
		c.Finalise(false)
	})
}

func TestCarmenStateDB_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	root := c.IntermediateRoot(false)
	assert.Equal(t, common.Hash{}, root)
}

func TestCarmenStateDB_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	h, out := c.Commit(uint64(9), false)
	assert.Equal(t, common.Hash{}, h)
	assert.Nil(t, out)
}

func TestCarmenStateDB_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	assert.NotPanics(t, func() {
		c.SetTxContext(common.Hash{}, 0)
	})
}

func TestCarmenStateDB_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	assert.NotPanics(t, func() {
		c.PrepareSubstate(nil, uint64(0))
	})
}

func TestCarmenStateDB_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	out := c.GetSubstatePostAlloc()
	assert.Nil(t, out)
}

func TestCarmenStateDB_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	assert.Panics(t, func() {
		c.AddPreimage(common.Hash{}, nil)
	})
}

func TestCarmenStateDB_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	out := c.AccessEvents()
	assert.Nil(t, out)
}

func TestCarmenStateDB_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	err := c.Error()
	assert.Nil(t, err)
}

func TestCarmenStateDB_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	mockDb.EXPECT().StartBulkLoad(gomock.Any()).Return(nil, nil)
	bulkLoad, err := c.StartBulkLoad(0)
	assert.NoError(t, err)
	assert.NotNil(t, bulkLoad)
}

func TestCarmenStateDB_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockMem := NewMockproxyMemoryFootprint(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	mockDb.EXPECT().GetMemoryFootprint().Return(mockMem)
	mockMem.EXPECT().Total().Return(uint64(9))
	memoryUsage := c.GetMemoryUsage()
	assert.Equal(t, uint64(9), memoryUsage.UsedBytes)
}

func TestCarmenStateDB_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenStateDB{
		db:    mockDb,
		txCtx: mockTxCtx,
	}
	shadowDB := c.GetShadowDB()
	assert.Nil(t, shadowDB)
}

// carmenHeadState struct method tests
func TestCarmenHeadStateBeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHeadBlockContext(ctrl)
	c := &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	mockDb.EXPECT().BeginBlock(gomock.Any()).Return(mockBlkCtx, nil)
	err := c.BeginBlock(uint64(9))
	assert.NoError(t, err)
}

func TestCarmenHeadStateBeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHeadBlockContext(ctrl)
	c := &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	mockBlkCtx.EXPECT().BeginTransaction().Return(mockTxCtx, nil)
	err := c.BeginTransaction(uint32(9))
	assert.NoError(t, err)
}

func TestCarmenHeadStateEndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHeadBlockContext(ctrl)
	c := &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	mockBlkCtx.EXPECT().Commit().Return(nil)
	err := c.EndBlock()
	assert.NoError(t, err)
}

func TestCarmenHeadStateBeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHeadBlockContext(ctrl)
	c := &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	assert.NotPanics(t, func() {
		c.BeginSyncPeriod(uint64(9))
	})
}

func TestCarmenHeadStateEndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHeadBlockContext(ctrl)
	c := &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}

	assert.NotPanics(t, c.EndSyncPeriod)
}

func TestCarmenHeadStateGetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHeadBlockContext(ctrl)
	c := &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	mockDb.EXPECT().GetHistoricContext(uint64(1)).Return(nil, nil)
	archiveState, err := c.GetArchiveState(uint64(1))
	assert.NoError(t, err)
	assert.NotNil(t, archiveState)
}

func TestCarmenHeadStateGetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHeadBlockContext(ctrl)
	c := &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	mockDb.EXPECT().GetArchiveBlockHeight().Return(int64(42), nil)
	height, value, err := c.GetArchiveBlockHeight()
	assert.Equal(t, uint64(42), height)
	assert.Equal(t, false, value)
	assert.NoError(t, err)
}

// carmenHistoricState struct method tests
func TestCarmenHistoricState_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHistoricBlockContext(ctrl)
	c := &carmenHistoricState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	mockBlkCtx.EXPECT().BeginTransaction().Return(mockTxCtx, nil)
	err := c.BeginTransaction(uint32(9))
	assert.NoError(t, err)
}

func TestCarmenHistoricState_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	c := &carmenHistoricState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
	}
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetHistoricStateHash(gomock.Any()).Return(carmen.Hash(expectedHash), nil)
	hash, err := c.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestCarmenHistoricState_Release(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := carmen.NewMockDatabase(ctrl)
	mockTxCtx := carmen.NewMockTransactionContext(ctrl)
	mockBlkCtx := carmen.NewMockHistoricBlockContext(ctrl)
	c := &carmenHistoricState{
		carmenStateDB: carmenStateDB{
			db:    mockDb,
			txCtx: mockTxCtx,
		},
		blkCtx: mockBlkCtx,
	}
	mockBlkCtx.EXPECT().Close().Return(nil)
	err := c.Release()
	assert.NoError(t, err)
}

// carmenBulkLoad struct method tests
func TestCarmenBulkLoad_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBulk := carmen.NewMockBulkLoad(ctrl)
	c := &carmenBulkLoad{
		load: mockBulk,
	}
	addr := common.HexToAddress("0x1234")
	mockBulk.EXPECT().CreateAccount(carmen.Address(addr))
	c.CreateAccount(addr)
}

func TestCarmenBulkLoad_SetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBulk := carmen.NewMockBulkLoad(ctrl)
	c := &carmenBulkLoad{
		load: mockBulk,
	}
	addr := common.HexToAddress("0x1234")
	balance := uint256.NewInt(1000)
	mockBulk.EXPECT().SetBalance(carmen.Address(addr), carmen.NewAmount(uint64(1000)))
	c.SetBalance(addr, balance)
}

func TestCarmenBulkLoad_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBulk := carmen.NewMockBulkLoad(ctrl)
	c := &carmenBulkLoad{
		load: mockBulk,
	}
	addr := common.HexToAddress("0x1234")
	nonce := uint64(42)
	mockBulk.EXPECT().SetNonce(carmen.Address(addr), nonce)
	c.SetNonce(addr, nonce)
}

func TestCarmenBulkLoad_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBulk := carmen.NewMockBulkLoad(ctrl)
	c := &carmenBulkLoad{
		load: mockBulk,
	}
	addr := common.HexToAddress("0x1234")
	key := common.BytesToHash([]byte("testKey"))
	value := common.BytesToHash([]byte("testValue"))
	mockBulk.EXPECT().SetState(carmen.Address(addr), carmen.Key(key), carmen.Value(value))
	c.SetState(addr, key, value)
}

func TestCarmenBulkLoad_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBulk := carmen.NewMockBulkLoad(ctrl)
	c := &carmenBulkLoad{
		load: mockBulk,
	}
	addr := common.HexToAddress("0x1234")
	code := []byte{0x01, 0x02, 0x03}
	mockBulk.EXPECT().SetCode(carmen.Address(addr), code)
	c.SetCode(addr, code)
}

func TestCarmenBulkLoad_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBulk := carmen.NewMockBulkLoad(ctrl)
	c := &carmenBulkLoad{
		load: mockBulk,
	}
	mockBulk.EXPECT().Finalize().Return(nil)
	err := c.Close()
	assert.NoError(t, err)
}
