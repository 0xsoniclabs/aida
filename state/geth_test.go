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
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const N = 1000

// fillDb creates a new DB in the given directory, fills it with some data and returns the root hash.
// If any error occurs, the test fails. The caller is responsible for removing the directory after use.
func fillDb(t *testing.T, directory string) (common.Hash, error) {
	db, err := MakeGethStateDB(directory, "", common.Hash{}, false, nil)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}

	if err := db.BeginBlock(0); err != nil {
		t.Fatalf("BeginBlock failed: %v", err)
	}
	if err := db.BeginTransaction(0); err != nil {
		t.Fatalf("BeginTransaction failed: %v", err)
	}
	for i := 0; i < N; i++ {
		address := common.Address{byte(i), byte(i >> 8)}
		db.CreateAccount(address)
		db.SetNonce(address, 12, tracing.NonceChangeUnspecified)
		key := common.Hash{byte(i >> 8), byte(i)}
		value := common.Hash{byte(15)}
		db.SetState(address, key, value)
	}
	if err := db.EndTransaction(); err != nil {
		t.Fatalf("EndTransaction failed: %v", err)
	}
	if err := db.EndBlock(); err != nil {
		t.Fatalf("EndBlock failed: %v", err)
	}
	hash, err := db.GetHash()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("Failed to close DB: %v", err)
	}
	return hash, nil
}

// TestGethDbFilling creates a new DB in a temporary directory and fills it with some data.
// The temporary directory is removed at the end of the test. If any error occurs, the test fails.
func TestGethDbFilling(t *testing.T) {
	dir := t.TempDir()
	if _, err := fillDb(t, dir); err != nil {
		t.Errorf("Unable to fill DB: %v", err)
	}
}

// TestGethDbReloadData creates a new DB in a temporary directory, fills it with some data,
// closes it, re-opens it and checks that the data is still there. The temporary directory is removed
// at the end of the test. If any error occurs, the test fails.
func TestGethDbReloadData(t *testing.T) {
	dir := t.TempDir()
	hash, err := fillDb(t, dir)
	if err != nil {
		t.Errorf("Unable to fill DB: %v", err)
	}

	// Re-open the data base.
	db, err := MakeGethStateDB(dir, "", hash, false, nil)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	for i := 0; i < N; i++ {
		address := common.Address{byte(i), byte(i >> 8)}
		if got := db.GetNonce(address); got != 12 {
			t.Fatalf("Nonce of %v is not 12: %v", address, got)
		}
		key := common.Hash{byte(i >> 8), byte(i)}
		value := common.Hash{byte(15)}
		if got := db.GetState(address, key); got != value {
			t.Fatalf("Value of %v/%v is not %v: %v", address, key, value, got)
		}
	}
	if err = db.Close(); err != nil {
		t.Fatalf("Failed to close DB: %v", err)
	}
}

// TestGethDb_CreateAccountIsProtected checks that calling CreateAccount multiple times for the same address does not panic.
// The geth wrapper checks the existence of the account before creating it, so that the geth implementation does not panic.
func TestGethDb_CreateAccountIsProtected(t *testing.T) {
	dir := t.TempDir()
	db, err := MakeGethStateDB(dir, "", common.Hash{}, false, nil)
	require.NoError(t, err)
	addr := common.Address{0x22}
	// First create the account
	db.CreateAccount(addr)
	// Account must exist in the db
	require.True(t, db.Exist(addr))
	// Then recall it - it must not panic
	db.CreateAccount(addr)
	// Account must exist in the db
	require.True(t, db.Exist(addr))
}

// TestGethDb_CreateAccountIsProtected checks that calling CreateAccount multiple times for the same address does not panic.
// The geth wrapper checks the non-existence of the account before creating it, so that the geth implementation does not panic.
func TestGethDb_CreateContractIsProtected(t *testing.T) {
	dir := t.TempDir()
	db, err := MakeGethStateDB(dir, "", common.Hash{}, false, nil)
	require.NoError(t, err)
	addr := common.Address{0x22}
	// First create the account
	db.CreateAccount(addr)
	// Account must exist in the db
	require.True(t, db.Exist(addr))
	// Then recall it - it must not panic
	db.CreateContract(addr)
	// Account must exist in the db
	require.True(t, db.Exist(addr))
}

// TestGethDb_CreateContractDoesNotCreateAccount checks that calling CreateContract for an address that does not exist does not create the account.
// The geth wrapper checks the existence of the account before creating it, so that the geth implementation does not create it.
func TestGethDb_CreateContractDoesNotCreateAccount(t *testing.T) {
	dir := t.TempDir()
	db, err := MakeGethStateDB(dir, "", common.Hash{}, false, nil)
	require.NoError(t, err)
	addr := common.Address{0x22}
	// First try to create the contract
	db.CreateContract(addr)
	// Account must not exist in the db
	require.False(t, db.Exist(addr))
}

func TestGethStateDB_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	gomock.InOrder(mockDb.EXPECT().Exist(addr),
		mockDb.EXPECT().CreateAccount(addr))
	g.CreateAccount(addr)
}

func TestGethStateDB_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	gomock.InOrder(mockDb.EXPECT().Exist(addr))
	g.CreateContract(addr)
}

func TestGethStateDB_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().Exist(addr).Return(true)
	exists := g.Exist(addr)
	assert.True(t, exists)
}

func TestGethStateDB_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().Empty(addr).Return(true)
	empty := g.Empty(addr)
	assert.True(t, empty)
}

func TestGethStateDB_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := uint256.NewInt(0)
	mockDb.EXPECT().SelfDestruct(addr).Return(*expected)
	value := g.SelfDestruct(addr)
	assert.Equal(t, expected, &value)
}

func TestGethStateDB_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := uint256.NewInt(0)
	mockDb.EXPECT().SelfDestruct6780(addr).Return(*expected, true)
	value, value2 := g.SelfDestruct6780(addr)
	assert.Equal(t, expected, &value)
	assert.True(t, value2)
}
func TestGethStateDB_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().HasSelfDestructed(addr).Return(true)
	result := g.HasSelfDestructed(addr)
	assert.True(t, result)
}

func TestGethStateDB_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := uint256.NewInt(100)
	mockDb.EXPECT().GetBalance(addr).Return(expected)
	result := g.GetBalance(addr)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	value := uint256.NewInt(50)
	expected := uint256.NewInt(150)
	mockDb.EXPECT().AddBalance(addr, value, gomock.Any()).Return(*expected)
	result := g.AddBalance(addr, value, 0)
	assert.Equal(t, expected, &result)
}

func TestGethStateDB_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	value := uint256.NewInt(50)
	expected := uint256.NewInt(50)
	mockDb.EXPECT().SubBalance(addr, value, gomock.Any()).Return(*expected)
	result := g.SubBalance(addr, value, 0)
	assert.Equal(t, expected, &result)
}

func TestGethStateDB_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := uint64(42)
	mockDb.EXPECT().GetNonce(addr).Return(expected)
	result := g.GetNonce(addr)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	nonce := uint64(42)
	mockDb.EXPECT().SetNonce(addr, nonce, gomock.Any())
	g.SetNonce(addr, nonce, 0)
}

func TestGethStateDB_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetStateAndCommittedState(addr, key).Return(common.Hash{}, expected)
	result := g.GetCommittedState(addr, key)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetState(addr, key).Return(expected)
	result := g.GetState(addr, key)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	expected := common.HexToHash("0xdef0")
	mockDb.EXPECT().SetState(addr, key, value).Return(expected)
	result := g.SetState(addr, key, value)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := common.HexToHash("0x5678")
	mockDb.EXPECT().GetStorageRoot(addr).Return(expected)
	result := g.GetStorageRoot(addr)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetTransientState(addr, key).Return(expected)
	result := g.GetTransientState(addr, key)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	mockDb.EXPECT().SetTransientState(addr, key, value)
	g.SetTransientState(addr, key, value)
}

func TestGethStateDB_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := []byte{0x60, 0x80, 0x60, 0x40}
	mockDb.EXPECT().GetCode(addr).Return(expected)
	result := g.GetCode(addr)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := common.HexToHash("0x5678")
	mockDb.EXPECT().GetCodeHash(addr).Return(expected)
	result := g.GetCodeHash(addr)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	expected := 256
	mockDb.EXPECT().GetCodeSize(addr).Return(expected)
	result := g.GetCodeSize(addr)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	code := []byte{0x60, 0x80, 0x60, 0x40}
	expected := []byte{0x60, 0x80}
	mockDb.EXPECT().SetCode(addr, code).Return(expected)
	result := g.SetCode(addr, code)
	assert.Equal(t, expected, result)
}

func TestGethStateDB_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	expected := 42
	mockDb.EXPECT().Snapshot().Return(expected)
	result := g.Snapshot()
	assert.Equal(t, expected, result)
}

func TestGethStateDB_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	id := 42
	mockDb.EXPECT().RevertToSnapshot(id)
	g.RevertToSnapshot(id)
}

func TestGethStateDB_Error(t *testing.T) {
	g := gethStateDB{}
	result := g.Error()
	assert.Nil(t, result)
}

func TestGethStateDB_BeginTransaction(t *testing.T) {
	g := gethStateDB{}
	err := g.BeginTransaction(1)
	assert.Nil(t, err)
}

func TestGethStateDB_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	mockDb.EXPECT().Finalise(true).AnyTimes()
	err := g.EndTransaction()
	assert.Nil(t, err)
}

func TestGethStateDB_BeginBlock(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	err := g.BeginBlock(1)
	assert.Nil(t, err)
}

func TestGethStateDB_EndBlock(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	err := g.EndBlock()
	assert.Nil(t, err)
}

func TestGethStateDB_BeginSyncPeriod(t *testing.T) {
	g := gethStateDB{}
	assert.NotPanics(t, func() {
		g.BeginSyncPeriod(1)
	})
}

func TestGethStateDB_EndSyncPeriod(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		stateRoot:     common.Hash{},
		evmState:      state.NewDatabase(trieDb, nil),
		triegc:        prque.New[uint64, common.Hash](nil),
		isArchiveMode: false,
	}

	assert.NotPanics(t, g.EndSyncPeriod)
}

func TestGethStateDB_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		db:        mockDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	hash, err := g.GetHash()
	assert.Nil(t, err)
	assert.Equal(t, common.Hash{}, hash)
}

func TestGethStateDB_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	sDb, _ := state.New(types.EmptyRootHash, state.NewDatabase(trieDb, nil))
	g := gethStateDB{
		db:        sDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	assert.NotPanics(t, func() {
		g.Finalise(true)
	})
}

func TestGethStateDB_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	sDb, _ := state.New(types.EmptyRootHash, state.NewDatabase(trieDb, nil))
	g := gethStateDB{
		db:        sDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	result := g.IntermediateRoot(true)
	assert.Equal(t, common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), result)
}

func TestGethStateDB_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	sDb, _ := state.New(types.EmptyRootHash, state.NewDatabase(trieDb, nil))
	g := gethStateDB{
		db:        sDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	hash, err := g.Commit(1, true)
	assert.Equal(t, common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), hash)
	assert.Nil(t, err)
}

func TestGethStateDB_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	sDb, _ := state.New(types.EmptyRootHash, state.NewDatabase(trieDb, nil))
	g := gethStateDB{
		db:        sDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	thash := common.HexToHash("0x1234")
	g.SetTxContext(thash, 1)
}

func TestGethStateDB_PrepareSubstate(t *testing.T) {
	g := gethStateDB{}
	assert.NotPanics(t, func() {
		g.PrepareSubstate(nil, 1)
	})
}

func TestGethStateDB_GetSubstatePostAlloc(t *testing.T) {
	g := gethStateDB{}
	result := g.GetSubstatePostAlloc()
	assert.Nil(t, result)
}

func TestGethStateDB_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()
	levelDb, _ := leveldb.New(dir, 0, 0, "test", false)
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	sDb, _ := state.New(types.EmptyRootHash, state.NewDatabase(trieDb, nil))
	g := gethStateDB{
		db:        sDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
		backend:   levelDb,
	}
	err := g.Close()
	assert.Nil(t, err)
}

func TestGethStateDB_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	gas := uint64(100)
	mockDb.EXPECT().AddRefund(gas)
	g.AddRefund(gas)
}

func TestGethStateDB_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	gas := uint64(50)
	mockDb.EXPECT().SubRefund(gas)
	g.SubRefund(gas)
}

func TestGethStateDB_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	expected := uint64(75)
	mockDb.EXPECT().GetRefund().Return(expected)
	result := g.GetRefund()
	assert.Equal(t, expected, result)
}

func TestGethStateDB_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	mockDb.EXPECT().Prepare(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	g.Prepare(params.Rules{}, common.Address{}, common.Address{}, nil, nil, nil)
}

func TestGethStateDB_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().AddressInAccessList(addr).Return(true)
	result := g.AddressInAccessList(addr)
	assert.True(t, result)
}

func TestGethStateDB_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	mockDb.EXPECT().SlotInAccessList(addr, slot).Return(true, true)
	addrOk, slotOk := g.SlotInAccessList(addr, slot)
	assert.True(t, addrOk)
	assert.True(t, slotOk)
}

func TestGethStateDB_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().AddAddressToAccessList(addr)
	g.AddAddressToAccessList(addr)
}

func TestGethStateDB_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	mockDb.EXPECT().AddSlotToAccessList(addr, slot)
	g.AddSlotToAccessList(addr, slot)
}

func TestGethStateDB_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	log := &types.Log{}
	mockDb.EXPECT().AddLog(log)
	g.AddLog(log)
}

func TestGethStateDB_AddPreimage(t *testing.T) {
	g := gethStateDB{}
	hash := common.HexToHash("0x1234")
	preimage := []byte{0x56, 0x78}
	assert.NotPanics(t, func() {
		g.AddPreimage(hash, preimage)
	})
}

func TestGethStateDB_AccessEvents(t *testing.T) {
	g := gethStateDB{
		accessEvents: nil,
	}
	result := g.AccessEvents()
	assert.Nil(t, result)
}

func TestGethStateDB_GetLogs(t *testing.T) {
	g := gethStateDB{}
	hash := common.HexToHash("0x1234")
	result := g.GetLogs(hash, 1, common.Hash{}, 123456)
	assert.Equal(t, []*types.Log{}, result)
}

func TestGethStateDB_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	mockDb.EXPECT().PointCache().Return(nil)
	result := g.PointCache()
	assert.Nil(t, result)
}

func TestGethStateDB_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethStateDB{
		db: mockDb,
	}
	mockDb.EXPECT().Witness().Return(nil)
	result := g.Witness()
	assert.Nil(t, result)
}

func TestGethStateDB_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockStateDB(ctrl)
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		db:        mockDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}
	result, err := g.StartBulkLoad(1)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestGethStateDB_GetArchiveState(t *testing.T) {
	g := gethStateDB{}
	result, err := g.GetArchiveState(1)
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "archive states are not (yet) supported")
}

func TestGethStateDB_GetArchiveBlockHeight(t *testing.T) {
	g := gethStateDB{}
	height, exists, err := g.GetArchiveBlockHeight()
	assert.Equal(t, uint64(0), height)
	assert.False(t, exists)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "archive states are not (yet) supported")
}

func TestGethStateDB_GetMemoryUsage(t *testing.T) {
	g := gethStateDB{}
	result := g.GetMemoryUsage()
	assert.Equal(t, uint64(0), result.UsedBytes)
}

func TestGethStateDB_TrieCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockStateDB(ctrl)
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		db:        mockDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
		block:     triesInMemory + 1,
	}
	err := g.trieCommit()
	assert.Nil(t, err)
}

func TestGethStateDB_TrieCleanCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockStateDB(ctrl)
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		db:        mockDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
		block:     triesInMemory + 1,
	}
	err := g.trieCleanCommit()
	assert.Nil(t, err)
}

func TestGethStateDB_TrieCap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockStateDB(ctrl)
	db := rawdb.NewMemoryDatabase()
	trieDb := triedb.NewDatabase(db, &triedb.Config{})
	g := gethStateDB{
		db:        mockDb,
		stateRoot: common.Hash{},
		evmState:  state.NewDatabase(trieDb, nil),
		triegc:    prque.New[uint64, common.Hash](nil),
	}

	err := g.trieCap()
	assert.Nil(t, err)
}

func TestGethStateDB_GetShadowDB(t *testing.T) {
	g := gethStateDB{}
	result := g.GetShadowDB()
	assert.Nil(t, result)
}

// gethBulkLoad struct method tests
func TestGethBulkLoad_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethBulkLoad{
		db: &gethStateDB{
			db: mockDb,
		},
	}
	addr := common.HexToAddress("0x1234")
	gomock.InOrder(mockDb.EXPECT().Exist(addr),
		mockDb.EXPECT().CreateAccount(addr))
	g.CreateAccount(addr)
}

func TestGethBulkLoad_SetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethBulkLoad{
		db: &gethStateDB{
			db: mockDb,
		},
	}
	addr := common.HexToAddress("0x1234")
	value := uint256.NewInt(100)
	mockDb.EXPECT().GetBalance(addr).Return(value)
	mockDb.EXPECT().AddBalance(addr, value, gomock.Any()).Return(*value)
	g.SetBalance(addr, value)
}

func TestGethBulkLoad_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethBulkLoad{
		db: &gethStateDB{
			db: mockDb,
		},
	}
	addr := common.HexToAddress("0x1234")
	nonce := uint64(42)
	mockDb.EXPECT().SetNonce(addr, nonce, gomock.Any())
	g.SetNonce(addr, nonce)
}

func TestGethBulkLoad_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethBulkLoad{
		db: &gethStateDB{
			db: mockDb,
		},
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	mockDb.EXPECT().SetState(addr, key, value)
	g.SetState(addr, key, value)
}

func TestGethBulkLoad_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethBulkLoad{
		db: &gethStateDB{
			db: mockDb,
		},
	}
	addr := common.HexToAddress("0x1234")
	code := []byte{0x60, 0x80, 0x60, 0x40}
	mockDb.EXPECT().SetCode(addr, code).Return(code)
	g.SetCode(addr, code)
}

func TestGethBulkLoad_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockVmStateDB(ctrl)
	g := gethBulkLoad{
		db: &gethStateDB{
			db: mockDb,
		},
	}
	err := g.Close()
	assert.Nil(t, err)
}
