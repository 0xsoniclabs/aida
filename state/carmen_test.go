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

package state

import (
	"bytes"
	"errors"
	"testing"

	carmen "github.com/0xsoniclabs/carmen/go/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
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

			csDB.SetCode(addr, code)

			if bytes.Compare(csDB.GetCode(addr), code) != 0 {
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

			if csDB.GetState(addr, key) != value {
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

				csDB.EndSyncPeriod()
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

				switch {
				case operationType == 1:
					// set balance
					newBalance := uint256.NewInt(GetRandom(t, 0, 5_000_000))

					cbl.SetBalance(account, newBalance)
				case operationType == 2:
					// set code
					code := MakeRandomByteSlice(t, 2048)

					cbl.SetCode(account, code)
				case operationType == 3:
					// set state
					key := common.BytesToHash(MakeRandomByteSlice(t, 32))
					value := common.BytesToHash(MakeRandomByteSlice(t, 32))

					cbl.SetState(account, key, value)
				case operationType == 4:
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
