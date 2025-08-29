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

package proxy

import (
	"bytes"
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/carmen/go/carmen"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func makeTestShadowDBWithCarmenTestContext(t *testing.T, ctc state.CarmenStateTestCase) state.StateDB {
	csDB, err := state.MakeDefaultCarmenStateDB(t.TempDir(), ctc.Variant, ctc.Schema, ctc.Archive)
	if errors.Is(err, carmen.UnsupportedConfiguration) {
		t.Skip("unsupported configuration")
	}

	if err != nil {
		t.Fatalf("failed to create carmen state DB: %v", err)
	}

	gsDB, err := state.MakeGethStateDB(t.TempDir(), "", common.Hash{}, false, nil)

	if err != nil {
		t.Fatalf("failed to create geth state DB: %v", err)
	}

	shadowDB := NewShadowProxy(csDB, gsDB, false)

	err = state.BeginCarmenDbTestContext(shadowDB)
	if err != nil {
		t.Fatal(err)
	}

	return shadowDB
}

func makeTestShadowDB(t *testing.T, ctc state.CarmenStateTestCase) state.StateDB {
	csDB, err := state.MakeDefaultCarmenStateDB(t.TempDir(), ctc.Variant, ctc.Schema, ctc.Archive)
	if errors.Is(err, carmen.UnsupportedConfiguration) {
		t.Skip("unsupported configuration")
	}
	if err != nil {
		t.Fatalf("failed to create carmen state DB: %v", err)
	}

	gsDB, err := state.MakeGethStateDB(t.TempDir(), "", common.Hash{}, false, nil)

	if err != nil {
		t.Fatalf("failed to create geth state DB: %v", err)
	}

	shadowDB := NewShadowProxy(csDB, gsDB, false)

	return shadowDB
}

// TestShadowState_InitCloseShadowDB test closing db immediately after initialization
func TestShadowState_InitCloseShadowDB(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDB(t, ctc)

			err := shadowDB.Close()
			if err != nil {
				t.Fatalf("failed to close shadow state DB: %v", err)
			}
		})
	}
}

// TestShadowState_AccountLifecycle tests account operations - create, check if it exists, if it's empty, suicide and suicide confirmation
func TestShadowState_AccountLifecycle(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			shadowDB.CreateAccount(addr)

			if !shadowDB.Exist(addr) {
				t.Fatal("failed to create carmen state DB account")
			}

			if !shadowDB.Empty(addr) {
				t.Fatal("failed to create carmen state DB account; should be empty")
			}

			shadowDB.SelfDestruct(addr)
			if !shadowDB.HasSelfDestructed(addr) {
				t.Fatal("failed to suicide carmen state DB account;")
			}
		})
	}
}

// TestShadowState_AccountBalanceOperations tests balance operations - add, subtract and check if the value is correct
func TestShadowState_AccountBalanceOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			shadowDB.CreateAccount(addr)

			// get randomized balance
			additionBase := state.GetRandom(t, 1, 5_000_000)
			addition := uint256.NewInt(additionBase)

			shadowDB.AddBalance(addr, addition, 0)

			if shadowDB.GetBalance(addr).Cmp(addition) != 0 {
				t.Fatal("failed to add balance to carmen state DB account")
			}

			subtraction := uint256.NewInt(state.GetRandom(t, 1, int(additionBase)))
			expectedResult := uint256.NewInt(0).Sub(addition, subtraction)

			shadowDB.SubBalance(addr, subtraction, 0)

			if shadowDB.GetBalance(addr).Cmp(expectedResult) != 0 {
				t.Fatal("failed to subtract balance to carmen state DB account")
			}
		})
	}
}

// TestShadowState_NonceOperations tests account nonce updating
func TestShadowState_NonceOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			shadowDB.CreateAccount(addr)

			// get randomized nonce
			newNonce := state.GetRandom(t, 1, 5_000_000)

			shadowDB.SetNonce(addr, newNonce, tracing.NonceChangeUnspecified)

			if shadowDB.GetNonce(addr) != newNonce {
				t.Fatal("failed to update account nonce")
			}
		})
	}
}

// TestShadowState_CodeOperations tests account code updating
func TestShadowState_CodeOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			shadowDB.CreateAccount(addr)

			// generate new randomized code
			code := state.MakeRandomByteSlice(t, 2048)

			if shadowDB.GetCodeSize(addr) != 0 {
				t.Fatal("failed to update account code; wrong initial size")
			}

			shadowDB.SetCode(addr, code)

			if !bytes.Equal(shadowDB.GetCode(addr), code) {
				t.Fatal("failed to update account code; wrong value")
			}

			if shadowDB.GetCodeSize(addr) != len(code) {
				t.Fatal("failed to update account code; wrong size")
			}
		})
	}
}

// TestShadowState_StateOperations tests account state update
func TestShadowState_StateOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			shadowDB.CreateAccount(addr)

			// generate state key and value
			key := common.BytesToHash(state.MakeRandomByteSlice(t, 32))
			value := common.BytesToHash(state.MakeRandomByteSlice(t, 32))

			shadowDB.SetState(addr, key, value)

			if shadowDB.GetState(addr, key) != value {
				t.Fatal("failed to update account state")
			}
		})
	}
}

// TestShadowState_TrxBlockSyncPeriodOperations tests creation of randomized sync-periods with blocks and transactions
func TestShadowState_TrxBlockSyncPeriodOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDB(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := shadowDB.Close()
				if err != nil {
					t.Fatalf("failed to close shadow state DB: %v", err)
				}
			}(shadowDB)

			blockNumber := 1
			trxNumber := 1
			for i := 0; i < 10; i++ {
				shadowDB.BeginSyncPeriod(uint64(i))

				for j := 0; j < 100; j++ {
					err := shadowDB.BeginBlock(uint64(blockNumber))
					if err != nil {
						t.Fatalf("cannot begin block; %v", err)
					}
					blockNumber++

					for k := 0; k < 100; k++ {
						err = shadowDB.BeginTransaction(uint32(trxNumber))
						if err != nil {
							t.Fatalf("cannot begin transaction; %v", err)
						}
						trxNumber++
						err = shadowDB.EndTransaction()
						if err != nil {
							t.Fatalf("cannot end transaction; %v", err)
						}
					}

					err = shadowDB.EndBlock()
					if err != nil {
						t.Fatalf("cannot end block; %v", err)
					}
				}

				shadowDB.EndSyncPeriod()
			}
		})
	}
}

// TestShadowState_RefundOperations tests adding and subtracting refund value
func TestShadowState_RefundOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			refundValue := state.GetRandom(t, 40_000_000, 50_000_000)
			shadowDB.AddRefund(refundValue)

			if shadowDB.GetRefund() != refundValue {
				t.Fatal("failed to add refund")
			}

			reducedRefund := refundValue - uint64(30000000)

			shadowDB.SubRefund(uint64(30000000))

			if shadowDB.GetRefund() != reducedRefund {
				t.Fatal("failed to subtract refund")
			}
		})
	}
}

// TestShadowState_AccessListOperations tests operations with creating, updating a checking AccessList
func TestShadowState_AccessListOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			// prepare content of access list
			rules := params.Rules{}
			sender := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))
			coinbase := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))
			dest := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))
			precompiles := []common.Address{
				common.BytesToAddress(state.MakeRandomByteSlice(t, 40)),
				common.BytesToAddress(state.MakeRandomByteSlice(t, 40)),
				common.BytesToAddress(state.MakeRandomByteSlice(t, 40)),
			}
			txAccesses := types.AccessList{
				types.AccessTuple{
					Address: common.BytesToAddress(state.MakeRandomByteSlice(t, 40)),
					StorageKeys: []common.Hash{
						common.BytesToHash(state.MakeRandomByteSlice(t, 32)),
						common.BytesToHash(state.MakeRandomByteSlice(t, 32)),
					},
				},
				types.AccessTuple{
					Address: common.BytesToAddress(state.MakeRandomByteSlice(t, 40)),
					StorageKeys: []common.Hash{
						common.BytesToHash(state.MakeRandomByteSlice(t, 32)),
						common.BytesToHash(state.MakeRandomByteSlice(t, 32)),
						common.BytesToHash(state.MakeRandomByteSlice(t, 32)),
						common.BytesToHash(state.MakeRandomByteSlice(t, 32)),
					},
				},
			}

			// create access list
			shadowDB.Prepare(rules, sender, coinbase, &dest, precompiles, txAccesses)

			// add some more data after the creation for good measure
			newAddr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))
			newSlot := common.BytesToHash(state.MakeRandomByteSlice(t, 32))
			shadowDB.AddAddressToAccessList(newAddr)
			shadowDB.AddSlotToAccessList(newAddr, newSlot)

			// check content of access list
			if !shadowDB.AddressInAccessList(sender) {
				t.Fatal("failed to add sender address to access list")
			}

			if !shadowDB.AddressInAccessList(dest) {
				t.Fatal("failed to add destination address to access list")
			}

			if !shadowDB.AddressInAccessList(newAddr) {
				t.Fatal("failed to add new address to access list after it was already created")
			}

			for _, addr := range precompiles {
				if !shadowDB.AddressInAccessList(addr) {
					t.Fatal("failed to add precompile address to access list")
				}
			}

			for _, txAccess := range txAccesses {
				if !shadowDB.AddressInAccessList(txAccess.Address) {
					t.Fatal("failed to add transaction access address to access list")
				}

				for _, storageKey := range txAccess.StorageKeys {
					addrOK, slotOK := shadowDB.SlotInAccessList(txAccess.Address, storageKey)
					if !addrOK || !slotOK {
						t.Fatal("failed to add transaction access address to access list")
					}
				}
			}

			addrOK, slotOK := shadowDB.SlotInAccessList(newAddr, newSlot)
			if !addrOK || !slotOK {
				t.Fatal("failed to add new slot to access list after it was already created")
			}
		})
	}
}

// TestShadowState_SetBalanceUsingBulkInsertion tests setting an accounts balance
func TestShadowState_SetBalanceUsingBulkInsertion(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDB(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			cbl, err := shadowDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)

			}

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			newBalance := uint256.NewInt(state.GetRandom(t, 1, 5_000_000))
			cbl.SetBalance(addr, newBalance)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			err = state.BeginCarmenDbTestContext(shadowDB)
			if err != nil {
				t.Fatal(err)
			}

			if shadowDB.GetBalance(addr).Cmp(newBalance) != 0 {
				t.Fatal("failed to update account balance")
			}
		})
	}
}

// TestShadowState_SetNonceUsingBulkInsertion tests setting an accounts nonce
func TestShadowState_SetNonceUsingBulkInsertion(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDB(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			cbl, err := shadowDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)

			}

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			newNonce := state.GetRandom(t, 1, 5_000_000)

			cbl.SetNonce(addr, newNonce)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			err = state.BeginCarmenDbTestContext(shadowDB)
			if err != nil {
				t.Fatal(err)
			}

			if shadowDB.GetNonce(addr) != newNonce {
				t.Fatal("failed to update account nonce")
			}
		})
	}
}

// TestShadowState_SetStateUsingBulkInsertion tests setting an accounts state
func TestShadowState_SetStateUsingBulkInsertion(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDB(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			cbl, err := shadowDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)
			}

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			// generate state key and value
			key := common.BytesToHash(state.MakeRandomByteSlice(t, 32))
			value := common.BytesToHash(state.MakeRandomByteSlice(t, 32))

			cbl.SetState(addr, key, value)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			//this is needed because new carmen API needs txCtx for db interactions
			err = shadowDB.BeginBlock(1)
			if err != nil {
				t.Fatalf("cannot begin block; %v", err)
			}
			err = shadowDB.BeginTransaction(0)
			if err != nil {
				t.Fatalf("cannot begin tx; %v", err)
			}

			if shadowDB.GetState(addr, key) != value {
				t.Fatal("failed to update account state")
			}
		})
	}
}

// TestShadowState_SetCodeUsingBulkInsertion tests setting an accounts code
func TestShadowState_SetCodeUsingBulkInsertion(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDB(t, ctc)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := state.CloseCarmenDbTestContext(shadowDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(shadowDB)

			cbl, err := shadowDB.StartBulkLoad(0)
			if err != nil {
				t.Fatal(err)

			}

			addr := common.BytesToAddress(state.MakeRandomByteSlice(t, 40))

			cbl.CreateAccount(addr)

			// generate new randomized code
			code := state.MakeRandomByteSlice(t, 2048)

			cbl.SetCode(addr, code)

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			err = state.BeginCarmenDbTestContext(shadowDB)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(shadowDB.GetCode(addr), code) {
				t.Fatal("failed to update account code")
			}
		})
	}
}

// TestShadowState_BulkloadOperations tests multiple operation in one bulkload
func TestShadowState_BulkloadOperations(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			shadowDB := makeTestShadowDBWithCarmenTestContext(t, ctc)

			// generate 100 randomized accounts
			accounts := [100]common.Address{}

			for i := 0; i < len(accounts); i++ {
				accounts[i] = common.BytesToAddress(state.MakeRandomByteSlice(t, 40))
				shadowDB.CreateAccount(accounts[i])
			}

			if err := shadowDB.EndTransaction(); err != nil {
				t.Fatalf("cannot end tx; %v", err)
			}
			if err := shadowDB.EndBlock(); err != nil {
				t.Fatalf("cannot end block; %v", err)
			}

			cbl, err := shadowDB.StartBulkLoad(7)
			if err != nil {
				t.Fatal(err)

			}

			for _, account := range accounts {
				// randomized operation
				operationType := state.GetRandom(t, 0, 4)

				switch operationType {
				case 1:
					// set balance
					newBalance := uint256.NewInt(uint64(state.GetRandom(t, 0, 5_000_000)))

					cbl.SetBalance(account, newBalance)
				case 2:
					// set code
					code := state.MakeRandomByteSlice(t, 2048)

					cbl.SetCode(account, code)
				case 3:
					// set state
					key := common.BytesToHash(state.MakeRandomByteSlice(t, 32))
					value := common.BytesToHash(state.MakeRandomByteSlice(t, 32))

					cbl.SetState(account, key, value)
				case 4:
					// set nonce
					newNonce := uint64(state.GetRandom(t, 0, 5_000_000))

					cbl.SetNonce(account, newNonce)
				default:
					// set code by default
					code := state.MakeRandomByteSlice(t, 2048)

					cbl.SetCode(account, code)
				}
			}

			err = cbl.Close()
			if err != nil {
				t.Fatalf("failed to close bulk load: %v", err)
			}

			// This is placed at the end instead of in a defer clause to
			// avoid being called in case of a panic occurring during the
			// test. This would make error diagnostic very difficult.
			if err := shadowDB.Close(); err != nil {
				t.Fatalf("failed to close shadow state DB: %v", err)
			}
		})
	}
}

func TestShadowState_GetShadowDB(t *testing.T) {
	for _, ctc := range state.GetCarmenStateTestCases() {
		t.Run(ctc.String(), func(t *testing.T) {
			csDB, err := state.MakeDefaultCarmenStateDB(t.TempDir(), ctc.Variant, ctc.Schema, ctc.Archive)
			if errors.Is(err, carmen.UnsupportedConfiguration) {
				t.Skip("unsupported configuration")
			}

			if err != nil {
				t.Fatalf("failed to create carmen state DB: %v", err)
			}

			gsDB, err := state.MakeGethStateDB(t.TempDir(), "", common.Hash{}, false, nil)

			if err != nil {
				t.Fatalf("failed to create geth state DB: %v", err)
			}

			shadowDB := NewShadowProxy(csDB, gsDB, false)

			// Close DB after test ends
			defer func(shadowDB state.StateDB) {
				err := shadowDB.Close()
				if err != nil {
					t.Fatalf("failed to close shadow state DB: %v", err)
				}
			}(shadowDB)

			if shadowDB.GetShadowDB() != gsDB {
				t.Fatal("Wrong return value of GetShadowDB")
			}

		})
	}
}

func TestShadowState_GetLogs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	pdb := state.NewMockStateDB(ctrl)
	sdb := state.NewMockStateDB(ctrl)
	db := NewShadowProxy(pdb, sdb, false)
	txHash := common.HexToHash("0x1")
	blockHash := common.HexToHash("0x2")
	log1 := &types.Log{}
	block := uint64(0)
	blkTimestamp := uint64(10)

	pdb.EXPECT().GetLogs(txHash, block, blockHash, blkTimestamp).Return([]*types.Log{log1})
	sdb.EXPECT().GetLogs(txHash, block, blockHash, blkTimestamp).Return([]*types.Log{log1})

	db.GetLogs(txHash, block, blockHash, blkTimestamp)
	if err := db.Error(); err != nil {
		t.Fatalf("Failed to compare logs; %v", err)
	}
}

func TestShadowState_GetLogsExpectError_LengthDifferent(t *testing.T) {
	ctrl := gomock.NewController(t)
	pdb := state.NewMockStateDB(ctrl)
	sdb := state.NewMockStateDB(ctrl)
	db := NewShadowProxy(pdb, sdb, false)
	txHash := common.HexToHash("0x1")
	blockHash := common.HexToHash("0x2")
	log1 := &types.Log{}
	block := uint64(0)
	blkTimestamp := uint64(10)

	pdb.EXPECT().GetLogs(txHash, block, blockHash, blkTimestamp).Return(nil)
	sdb.EXPECT().GetLogs(txHash, block, blockHash, blkTimestamp).Return([]*types.Log{log1})

	db.GetLogs(txHash, block, blockHash, blkTimestamp)
	if err := db.Error(); err == nil {
		t.Fatal("Expect mismatched GetLogs lengths")
	}
}

func TestShadowState_GetLogsExpectError_BloomDifferent(t *testing.T) {
	ctrl := gomock.NewController(t)
	pdb := state.NewMockStateDB(ctrl)
	sdb := state.NewMockStateDB(ctrl)
	db := NewShadowProxy(pdb, sdb, false)
	txHash := common.HexToHash("0x1")
	blockHash := common.HexToHash("0x2")
	log1 := &types.Log{}
	log2 := &types.Log{Address: common.HexToAddress("0x3")}
	block := uint64(0)
	blkTimestamp := uint64(10)

	pdb.EXPECT().GetLogs(txHash, block, blockHash, blkTimestamp).Return([]*types.Log{log1})
	sdb.EXPECT().GetLogs(txHash, block, blockHash, blkTimestamp).Return([]*types.Log{log2})

	db.GetLogs(txHash, block, blockHash, blkTimestamp)
	if err := db.Error(); err == nil {
		t.Fatal("Expect mismatched log values")
	}
}

func TestShadowState_GetHash_SuccessWithValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	pdb := state.NewMockStateDB(ctrl)
	sdb := state.NewMockStateDB(ctrl)
	db := NewShadowProxy(pdb, sdb, true)
	expectedHash := common.HexToHash("0x1")

	pdb.EXPECT().GetHash().Return(expectedHash, nil)
	sdb.EXPECT().GetHash().Return(expectedHash, nil)

	_, err := db.GetHash()
	if err != nil {
		t.Fatalf("Failed to execute GetHash; %v", err)
	}
	if err := db.Error(); err != nil {
		t.Fatalf("Failed to execute GetHash; %v", err)
	}
}

func TestShadowState_GetHash_SuccessWithoutValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	pdb := state.NewMockStateDB(ctrl)
	sdb := state.NewMockStateDB(ctrl)
	db := NewShadowProxy(pdb, sdb, false)
	primeHash := common.HexToHash("0x1")

	// hash of shadow is not called
	pdb.EXPECT().GetHash().Return(primeHash, nil)

	_, err := db.GetHash()
	if err != nil {
		t.Fatalf("Failed to execute GetHash; %v", err)
	}
	if err := db.Error(); err != nil {
		t.Fatalf("Failed to execute GetHash; %v", err)
	}
}

func TestShadowState_GetHash_FailWithValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	pdb := state.NewMockStateDB(ctrl)
	sdb := state.NewMockStateDB(ctrl)
	db := NewShadowProxy(pdb, sdb, true)
	primeHash := common.HexToHash("0x1")
	shadowHash := common.HexToHash("0x2")

	pdb.EXPECT().GetHash().Return(primeHash, nil)
	sdb.EXPECT().GetHash().Return(shadowHash, nil)

	_, err := db.GetHash()
	if err == nil {
		t.Fatal("Expect a mistach of state hashes")
	}
	if err := db.Error(); err == nil {
		t.Fatal("Expect a mistach of state hashes")
	}
}

func TestShadowState_GetStorageRoot_CallsBothMethods_And_ReturnsPrimaryResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	pdb := state.NewMockStateDB(ctrl)
	sdb := state.NewMockStateDB(ctrl)
	db := NewShadowProxy(pdb, sdb, true)

	addr := common.Address{1}
	primaryHash := common.Hash{1}
	shadowHash := common.Hash{2}

	// both databases must be called
	pdb.EXPECT().GetStorageRoot(addr).Return(primaryHash)
	sdb.EXPECT().GetStorageRoot(addr).Return(shadowHash)

	if got, want := db.GetStorageRoot(addr), primaryHash; got != want {
		if got == shadowHash {
			t.Error("proxy returned shadow-db hash but must return primary-db hash")
		} else {
			t.Errorf("unexpected hash, got: %s, want: %s", got, want)
		}
	}
}

func TestProxy_NewShadowProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPdb := state.NewMockStateDB(ctrl)
	mockSdb := state.NewMockStateDB(ctrl)
	shadowDb := NewShadowProxy(mockPdb, mockSdb, false)
	assert.NotNil(t, shadowDb)
}

func TestShadowVmStateDb_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().CreateAccount(addr).Times(2)
	shadow.CreateAccount(addr)
}
func TestShadowVmStateDb_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().Exist(addr).Return(true).Times(2)
	exist := shadow.Exist(addr)
	assert.True(t, exist)
}
func TestShadowVmStateDb_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().Empty(addr).Return(true).Times(2)
	isEmpty := shadow.Empty(addr)
	assert.True(t, isEmpty)
}
func TestShadowVmStateDb_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().SelfDestruct(addr).Times(2)
	shadow.SelfDestruct(addr)
}
func TestShadowVmStateDb_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().HasSelfDestructed(addr).Return(true).Times(2)
	hasSelfDestructed := shadow.HasSelfDestructed(addr)
	assert.True(t, hasSelfDestructed)
}
func TestShadowVmStateDb_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	balance := uint256.NewInt(1000)
	mockDb.EXPECT().GetBalance(addr).Return(balance).Times(2)
	result := shadow.GetBalance(addr)
	assert.Equal(t, balance, result)
}
func TestShadowVmStateDb_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	value := uint256.NewInt(1000)
	reason := tracing.BalanceChangeUnspecified
	expectedBalance := uint256.NewInt(2000)
	mockDb.EXPECT().AddBalance(addr, value, reason).Return(*expectedBalance).Times(2)
	balance := shadow.AddBalance(addr, value, reason)
	assert.Equal(t, *expectedBalance, balance)
}
func TestShadowVmStateDb_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	value := uint256.NewInt(1000)
	reason := tracing.BalanceChangeUnspecified
	expectedBalance := uint256.NewInt(500)
	mockDb.EXPECT().SubBalance(addr, value, reason).Return(*expectedBalance).Times(2)
	balance := shadow.SubBalance(addr, value, reason)
	assert.Equal(t, *expectedBalance, balance)
}
func TestShadowVmStateDb_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	expectedNonce := uint64(42)
	mockDb.EXPECT().GetNonce(addr).Return(expectedNonce).Times(2)
	nonce := shadow.GetNonce(addr)
	assert.Equal(t, expectedNonce, nonce)
}
func TestShadowVmStateDb_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	newNonce := uint64(100)
	reason := tracing.NonceChangeUnspecified
	mockDb.EXPECT().SetNonce(addr, newNonce, reason).Times(2)
	shadow.SetNonce(addr, newNonce, reason)
}
func TestShadowVmStateDb_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	key := common.Hash{0x34}
	expectedValue := common.Hash{0x56}
	mockDb.EXPECT().GetCommittedState(addr, key).Return(expectedValue).Times(2)
	value := shadow.GetCommittedState(addr, key)
	assert.Equal(t, expectedValue, value)
}
func TestShadowVmStateDb_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	key := common.Hash{0x34}
	expectedValue := common.Hash{0x56}
	mockDb.EXPECT().GetState(addr, key).Return(expectedValue).Times(2)
	value := shadow.GetState(addr, key)
	assert.Equal(t, expectedValue, value)
}
func TestShadowVmStateDb_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	key := common.Hash{0x34}
	value := common.Hash{0x56}
	expectedHash := common.Hash{0x78}
	mockDb.EXPECT().SetState(addr, key, value).Return(expectedHash).Times(2)
	h := shadow.SetState(addr, key, value)
	assert.Equal(t, expectedHash, h)
}
func TestShadowVmStateDb_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	key := common.Hash{0x34}
	value := common.Hash{0x56}
	mockDb.EXPECT().SetTransientState(addr, key, value).Times(2)
	shadow.SetTransientState(addr, key, value)
}
func TestShadowVmStateDb_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	key := common.Hash{0x34}
	expectedValue := common.Hash{0x56}
	mockDb.EXPECT().GetTransientState(addr, key).Return(expectedValue).Times(2)
	value := shadow.GetTransientState(addr, key)
	assert.Equal(t, expectedValue, value)
}
func TestShadowVmStateDb_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	expectedCode := []byte{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetCode(addr).Return(expectedCode).Times(2)
	code := shadow.GetCode(addr)
	assert.Equal(t, expectedCode, code)
}
func TestShadowVmStateDb_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	expectedSize := 3
	mockDb.EXPECT().GetCodeSize(addr).Return(expectedSize).Times(2)
	size := shadow.GetCodeSize(addr)
	assert.Equal(t, expectedSize, size)
}
func TestShadowVmStateDb_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetCodeHash(addr).Return(expectedHash).Times(2)
	hash := shadow.GetCodeHash(addr)
	assert.Equal(t, expectedHash, hash)
}
func TestShadowVmStateDb_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	code := []byte{0x01, 0x02, 0x03}
	expectedHash := []byte{0x04, 0x05, 0x06}
	mockDb.EXPECT().SetCode(addr, code).Return(expectedHash).Times(2)
	hash := shadow.SetCode(addr, code)
	assert.Equal(t, expectedHash, hash)
}
func TestShadowVmStateDb_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	expectedSnapshot := 0
	mockDb.EXPECT().Snapshot().Return(1).Times(2)
	snapshot := shadow.Snapshot()
	assert.Equal(t, expectedSnapshot, snapshot)
}
func TestShadowVmStateDb_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:  mockDb,
		shadow: mockDb,
		snapshots: []snapshotPair{
			{},
		},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	snapshot := 0
	mockDb.EXPECT().RevertToSnapshot(snapshot).Times(2)
	shadow.RevertToSnapshot(snapshot)
}
func TestShadowVmStateDb_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	mockDb.EXPECT().BeginTransaction(uint32(1)).Return(nil).Times(2)
	err := shadow.BeginTransaction(uint32(1))
	assert.NoError(t, err)
}
func TestShadowVmStateDb_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	mockDb.EXPECT().EndTransaction().Return(nil).Times(2)
	err := shadow.EndTransaction()
	assert.NoError(t, err)
}
func TestShadowVmStateDb_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	mockDb.EXPECT().Finalise(true).Times(2)
	shadow.Finalise(true)
}
func TestShadowStateDb_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	blk := uint64(123)
	mockDb.EXPECT().BeginBlock(blk).Return(nil).Times(2)
	err := shadow.BeginBlock(blk)
	assert.NoError(t, err)
}

func TestShadowStateDb_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	mockDb.EXPECT().EndBlock().Return(nil).Times(2)
	err := shadow.EndBlock()
	assert.NoError(t, err)
}

func TestShadowStateDb_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	number := uint64(456)
	mockDb.EXPECT().BeginSyncPeriod(number).Times(2)
	shadow.BeginSyncPeriod(number)
}

func TestShadowStateDb_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	mockDb.EXPECT().EndSyncPeriod().Times(2)
	shadow.EndSyncPeriod()
}

func TestShadowStateDb_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetHash().Return(expectedHash, nil).Times(1)
	hash, err := shadow.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestShadowNonCommittableStateDb_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockNonCommittableStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowNonCommittableStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetHash().Return(expectedHash, nil).Times(1)
	hash, err := shadow.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestShadowNonCommittableStateDb_getHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockNonCommittableStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockOp := func(s state.NonCommittableStateDB) (common.Hash, error) {
		return common.Hash{0x01, 0x02, 0x03}, nil
	}
	shadow := &shadowNonCommittableStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	hash, err := shadow.getHash("test", mockOp)
	assert.Error(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestShadowNonCommittableStateDb_getStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockNonCommittableStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockOp := func(s state.VmStateDB) (common.Hash, error) {
		return common.Hash{0x01, 0x02, 0x03}, nil
	}
	shadow := &shadowNonCommittableStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	expectedHash := common.Hash{0x01, 0x02, 0x03}
	hash, err := shadow.getStateHash("test", mockOp)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestShadowStateDb_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	mockDb.EXPECT().Close().Return(nil).Times(2)
	err := shadow.Close()
	assert.NoError(t, err)
}

func TestShadowNonCommittableStateDb_Release(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockNonCommittableStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowNonCommittableStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	mockDb.EXPECT().Release().Return(nil).Times(2)
	err := shadow.Release()
	assert.NoError(t, err)
}

func TestShadowVmStateDb_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	amount := uint64(1000)
	mockDb.EXPECT().AddRefund(amount).Times(2)
	mockDb.EXPECT().GetRefund().Return(amount).Times(2)
	shadow.AddRefund(amount)
}

func TestShadowVmStateDb_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	amount := uint64(500)
	mockDb.EXPECT().SubRefund(amount).Times(2)
	mockDb.EXPECT().GetRefund().Return(uint64(1000)).Times(2)
	shadow.SubRefund(amount)
}

func TestShadowVmStateDb_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	expectedRefund := uint64(750)
	mockDb.EXPECT().GetRefund().Return(expectedRefund).Times(2)
	refund := shadow.GetRefund()
	assert.Equal(t, expectedRefund, refund)
}

func TestShadowVmStateDb_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	rules := params.Rules{}
	sender := common.Address{0x01}
	coinbase := common.Address{0x02}
	dest := common.Address{0x03}
	precompiles := []common.Address{{0x04}}
	txAccesses := types.AccessList{}
	mockDb.EXPECT().Prepare(rules, sender, coinbase, &dest, precompiles, txAccesses).Times(2)
	shadow.Prepare(rules, sender, coinbase, &dest, precompiles, txAccesses)
}

func TestShadowVmStateDb_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().AddressInAccessList(addr).Return(true).Times(2)
	result := shadow.AddressInAccessList(addr)
	assert.True(t, result)
}

func TestShadowVmStateDb_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	slot := common.Hash{0x34}
	mockDb.EXPECT().SlotInAccessList(addr, slot).Return(true, true).Times(2)
	addrOk, slotOk := shadow.SlotInAccessList(addr, slot)
	assert.True(t, addrOk)
	assert.True(t, slotOk)
}

func TestShadowVmStateDb_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().AddAddressToAccessList(addr).Times(2)
	shadow.AddAddressToAccessList(addr)
}

func TestShadowVmStateDb_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	slot := common.Hash{0x34}
	mockDb.EXPECT().AddSlotToAccessList(addr, slot).Times(2)
	shadow.AddSlotToAccessList(addr, slot)
}

func TestShadowVmStateDb_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	log := &types.Log{Address: common.Address{0x12}}
	mockDb.EXPECT().AddLog(log).Times(2)
	shadow.AddLog(log)
}

func TestShadowVmStateDb_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	hash := common.Hash{0x01}
	block := uint64(123)
	blockHash := common.Hash{0x02}
	timestamp := uint64(456)
	expectedLogs := []*types.Log{{Address: common.Address{0x12}}}
	mockDb.EXPECT().GetLogs(hash, block, blockHash, timestamp).Return(expectedLogs).Times(2)
	logs := shadow.GetLogs(hash, block, blockHash, timestamp)
	assert.Equal(t, expectedLogs, logs)
}

func TestShadowVmStateDb_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	expectedHash := common.Hash{0x34}
	mockDb.EXPECT().GetStorageRoot(addr).Return(expectedHash).Times(2)
	hash := shadow.GetStorageRoot(addr)
	assert.Equal(t, expectedHash, hash)
}

func TestShadowVmStateDb_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	mockDb.EXPECT().CreateContract(addr).Times(2)
	shadow.CreateContract(addr)
}

func TestShadowVmStateDb_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	addr := common.Address{0x12}
	expectedBalance := uint256.NewInt(1000)
	mockDb.EXPECT().SelfDestruct6780(addr).Return(*expectedBalance, true).Times(2)
	balance, destroyed := shadow.SelfDestruct6780(addr)
	assert.Equal(t, *expectedBalance, balance)
	assert.True(t, destroyed)
}

func TestShadowVmStateDb_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	mockDb.EXPECT().PointCache().Return(nil).Times(1)
	cache := shadow.PointCache()
	assert.Nil(t, cache)
}

func TestShadowVmStateDb_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	mockDb.EXPECT().Witness().Return(nil).Times(1)
	witness := shadow.Witness()
	assert.Nil(t, witness)
}

func TestShadowStateDb_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	mockDb.EXPECT().Finalise(true).Times(2)
	shadow.Finalise(true)
}

func TestShadowStateDb_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	expectedHash := common.Hash{0x12}
	mockDb.EXPECT().IntermediateRoot(true).Return(expectedHash).Times(2)
	hash := shadow.IntermediateRoot(true)
	assert.Equal(t, expectedHash, hash)
}

func TestShadowStateDb_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	block := uint64(123)
	expectedHash := common.Hash{0x34}
	mockDb.EXPECT().Commit(block, true).Return(expectedHash, nil).Times(2)
	hash, err := shadow.Commit(block, true)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestShadowVmStateDb_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	err := shadow.Error()
	assert.NoError(t, err)
}

func TestShadowVmStateDb_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	thash := common.Hash{0x12}
	ti := 42
	mockDb.EXPECT().SetTxContext(thash, ti).Times(2)
	shadow.SetTxContext(thash, ti)
}

func TestShadowStateDb_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	substate := txcontext.NewMockWorldState(ctrl)
	block := uint64(123)
	mockDb.EXPECT().PrepareSubstate(substate, block).Times(2)
	shadow.PrepareSubstate(substate, block)
}

func TestShadowVmStateDb_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	expectedWorldState := txcontext.NewMockWorldState(ctrl)
	mockDb.EXPECT().GetSubstatePostAlloc().Return(expectedWorldState).Times(2)
	worldState := shadow.GetSubstatePostAlloc()
	assert.Equal(t, expectedWorldState, worldState)
}

func TestShadowVmStateDb_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	hash := common.Hash{0x12}
	plain := []byte{0x34, 0x56}
	mockDb.EXPECT().AddPreimage(hash, plain).Times(2)
	shadow.AddPreimage(hash, plain)
}

func TestShadowVmStateDb_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowVmStateDb{
		prime:            mockDb,
		shadow:           mockDb,
		snapshots:        []snapshotPair{},
		err:              nil,
		log:              mockLogger,
		compareStateHash: false,
	}
	mockDb.EXPECT().AccessEvents().Return(nil).Times(1)
	events := shadow.AccessEvents()
	assert.Nil(t, events)
}

func TestShadowStateDb_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	block := uint64(123)
	mockDb.EXPECT().StartBulkLoad(block).Return(mockBulkLoad, nil).Times(2)
	bulkLoad, err := shadow.StartBulkLoad(block)
	assert.NoError(t, err)
	assert.NotNil(t, bulkLoad)
}

func TestShadowStateDb_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockArchiveState := state.NewMockNonCommittableStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	block := uint64(123)
	mockDb.EXPECT().GetArchiveState(block).Return(mockArchiveState, nil).Times(2)
	archiveState, err := shadow.GetArchiveState(block)
	assert.NoError(t, err)
	assert.NotNil(t, archiveState)
}

func TestShadowStateDb_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	mockDb.EXPECT().GetArchiveBlockHeight().Return(uint64(0), true, nil).Times(2)
	height, exists, err := shadow.GetArchiveBlockHeight()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), height)
	assert.True(t, exists)
}

func TestShadowStateDb_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockDb,
			shadow:           mockDb,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockDb,
		shadow: mockDb,
	}
	expectedUsage := &state.MemoryUsage{}
	mockDb.EXPECT().GetMemoryUsage().Return(expectedUsage).Times(2)
	usage := shadow.GetMemoryUsage()
	assert.Equal(t, expectedUsage.UsedBytes, usage.UsedBytes)
	assert.Contains(t, usage.Breakdown.String(), "Primary")
}

func TestShadowStateDb_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrime := state.NewMockStateDB(ctrl)
	mockShadow := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	shadow := &shadowStateDb{
		shadowVmStateDb: shadowVmStateDb{
			prime:            mockPrime,
			shadow:           mockShadow,
			snapshots:        []snapshotPair{},
			err:              nil,
			log:              mockLogger,
			compareStateHash: false,
		},
		prime:  mockPrime,
		shadow: mockShadow,
	}
	shadowDB := shadow.GetShadowDB()
	assert.Equal(t, mockShadow, shadowDB)
}

func TestShadowBulkLoad_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	shadow := &shadowBulkLoad{
		prime:  mockBulkLoad,
		shadow: mockBulkLoad,
	}
	addr := common.Address{0x12}
	mockBulkLoad.EXPECT().CreateAccount(addr).Times(2)
	shadow.CreateAccount(addr)
}

func TestShadowBulkLoad_SetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	shadow := &shadowBulkLoad{
		prime:  mockBulkLoad,
		shadow: mockBulkLoad,
	}
	addr := common.Address{0x12}
	value := uint256.NewInt(1000)
	mockBulkLoad.EXPECT().SetBalance(addr, value).Times(2)
	shadow.SetBalance(addr, value)
}

func TestShadowBulkLoad_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	shadow := &shadowBulkLoad{
		prime:  mockBulkLoad,
		shadow: mockBulkLoad,
	}
	addr := common.Address{0x12}
	value := uint64(42)
	mockBulkLoad.EXPECT().SetNonce(addr, value).Times(2)
	shadow.SetNonce(addr, value)
}

func TestShadowBulkLoad_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	shadow := &shadowBulkLoad{
		prime:  mockBulkLoad,
		shadow: mockBulkLoad,
	}
	addr := common.Address{0x12}
	key := common.Hash{0x34}
	value := common.Hash{0x56}
	mockBulkLoad.EXPECT().SetState(addr, key, value).Times(2)
	shadow.SetState(addr, key, value)
}

func TestShadowBulkLoad_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	shadow := &shadowBulkLoad{
		prime:  mockBulkLoad,
		shadow: mockBulkLoad,
	}
	addr := common.Address{0x12}
	code := []byte{0x01, 0x02, 0x03}
	mockBulkLoad.EXPECT().SetCode(addr, code).Times(2)
	shadow.SetCode(addr, code)
}

func TestShadowBulkLoad_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	shadow := &shadowBulkLoad{
		prime:  mockBulkLoad,
		shadow: mockBulkLoad,
	}
	mockBulkLoad.EXPECT().Close().Return(nil).Times(2)
	err := shadow.Close()
	assert.NoError(t, err)
}
