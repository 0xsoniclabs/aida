package state

import (
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestInMemoryDb_SelfDestruct6780OnlyDeletesContractsCreatedInSameTransaction(t *testing.T) {
	a := common.Address{1}
	b := common.Address{2}

	db := MakeInMemoryStateDB(nil, 12)
	db.CreateContract(a)

	if want, got := false, db.HasSelfDestructed(a); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", a, want, got)
	}
	if want, got := false, db.HasSelfDestructed(b); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", b, want, got)
	}

	db.SelfDestruct6780(a) // < this should work

	if want, got := true, db.HasSelfDestructed(a); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", a, want, got)
	}
	if want, got := false, db.HasSelfDestructed(b); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", b, want, got)
	}

	db.SelfDestruct6780(b) // < this should be ignored

	if want, got := true, db.HasSelfDestructed(a); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", a, want, got)
	}
	if want, got := false, db.HasSelfDestructed(b); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", b, want, got)
	}
}

func TestInMemoryStateDB__GetLogs_ReturnEmptyLogsWithNilSnapshot(t *testing.T) {
	sdb := &inMemoryStateDB{state: nil}
	logs := sdb.GetLogs(common.Hash{}, 0, common.Hash{}, 0)
	assert.Empty(t, logs)
}

func TestInMemoryStateDB__GetLogs_AddsLogsWithCorrectTimestamp(t *testing.T) {
	txHash := common.Hash{0x1, 0x2, 0x3}
	blkNumber := uint64(10)
	blkHash := common.Hash{0x4, 0x5, 0x6}
	blkTimestamp := uint64(11)
	sdb := &inMemoryStateDB{state: &snapshot{
		parent: &snapshot{
			logs: []*types.Log{{Index: 1, BlockTimestamp: blkTimestamp}},
		},
		logs: []*types.Log{{Index: 0}},
	}}
	logs := sdb.GetLogs(txHash, blkNumber, blkHash, blkTimestamp)
	assert.Len(t, logs, 1) // No logs added yet
	assert.Equal(t, blkTimestamp, logs[0].BlockTimestamp)
	assert.Equal(t, uint(1), logs[0].Index)
}

// Package level function tests
func TestStateMakeEmptyGethInMemoryStateDB(t *testing.T) {
	t.Run("with variant", func(t *testing.T) {
		db, err := MakeEmptyGethInMemoryStateDB("testVariant")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
	t.Run("without variant", func(t *testing.T) {
		db, err := MakeEmptyGethInMemoryStateDB("")
		assert.NoError(t, err)
		assert.NotNil(t, db)
	})
}

// inMemoryStateDB struct method tests
func TestInMemoryStateDB_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	key := common.Hash{0x4, 0x5, 0x6}
	value := common.Hash{0x7, 0x8, 0x9}
	mem.SetTransientState(addr, key, value)

	s := slot{addr: addr, key: key}
	assert.Equal(t, 1, len(mem.state.transientStorage))
	assert.Equal(t, value, mem.state.transientStorage[s])
}

func TestInMemoryStateDB_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	key := common.Hash{0x4, 0x5, 0x6}
	value := common.Hash{0x7, 0x8, 0x9}
	mem.state.transientStorage[slot{addr: addr, key: key}] = value

	retrievedValue := mem.GetTransientState(addr, key)
	assert.Equal(t, value, retrievedValue)
}

func TestInMemoryStateDB_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     46051751,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	mem.CreateAccount(addr)

	_, exists := mem.state.createdAccounts[addr]
	assert.True(t, exists)
}

func TestInMemoryStateDB_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	mem.CreateContract(addr)

	_, exists := mem.state.createdContracts[addr]
	assert.True(t, exists)
}

func TestInMemoryStateDB_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	mockWs.EXPECT().Get(addr).Return(mockAcc).Times(2)
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(100)).Times(2)
	value := mem.SubBalance(addr, uint256.NewInt(10), tracing.BalanceChangeUnspecified)
	assert.Equal(t, uint256.NewInt(100), &value)
}

func TestInMemoryStateDB_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	mockWs.EXPECT().Get(addr).Return(mockAcc).Times(2)
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(100)).Times(2)
	value := mem.AddBalance(addr, uint256.NewInt(10), tracing.BalanceChangeUnspecified)
	assert.Equal(t, uint256.NewInt(100), &value)
}

func TestInMemoryStateDB_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	mockWs.EXPECT().Get(addr).Return(mockAcc).Times(1)
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(100)).Times(1)

	balance := mem.GetBalance(addr)
	assert.Equal(t, uint256.NewInt(100), balance)
}

func TestInMemoryStateDB_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	mockWs.EXPECT().Get(addr).Return(mockAcc).Times(1)
	mockAcc.EXPECT().GetNonce().Return(uint64(42)).Times(1)

	nonce := mem.GetNonce(addr)
	assert.Equal(t, uint64(42), nonce)
}

func TestInMemoryStateDB_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	newNonce := uint64(100)

	mem.SetNonce(addr, newNonce, tracing.NonceChangeUnspecified)

	assert.Equal(t, 0, mem.state.touched[addr])
	assert.Equal(t, newNonce, mem.state.nonces[addr])
}

func TestInMemoryStateDB_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	expected := common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	mockWs.EXPECT().Has(gomock.Any()).Return(true).Times(2)
	mockWs.EXPECT().Get(gomock.Any()).Return(mockAcc).Times(1)
	mockAcc.EXPECT().GetCode().Return([]uint8{}).Times(1)
	retrievedCodeHash := mem.GetCodeHash(addr)
	assert.Equal(t, expected, retrievedCodeHash)
}

func TestInMemoryStateDB_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	expectedCode := []byte{0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00}
	mockWs.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	mockWs.EXPECT().Get(gomock.Any()).Return(mockAcc).Times(1)
	mockAcc.EXPECT().GetCode().Return(expectedCode).Times(1)

	retrievedCode := mem.GetCode(addr)
	assert.Equal(t, expectedCode, retrievedCode)
}

func TestInMemoryStateDB_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	expectedCode := []byte{0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00}
	mockWs.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	mockWs.EXPECT().Get(gomock.Any()).Return(mockAcc).Times(1)
	mockAcc.EXPECT().GetCode().Return(expectedCode).Times(1)

	retrievedCode := mem.SetCode(addr, expectedCode)
	assert.Equal(t, expectedCode, retrievedCode)
}

func TestInMemoryStateDB_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	expectedCode := []byte{0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00}
	mockWs.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	mockWs.EXPECT().Get(gomock.Any()).Return(mockAcc).Times(1)
	mockAcc.EXPECT().GetCode().Return(expectedCode).Times(1)

	codeSize := mem.GetCodeSize(addr)
	assert.Equal(t, len(expectedCode), codeSize)
}

func TestInMemoryStateDB_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	refundAmount := uint64(100)

	mem.AddRefund(refundAmount)

	assert.Equal(t, refundAmount, mem.state.refund)
}

func TestInMemoryStateDB_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	subtractAmount := uint64(50)

	mem.SubRefund(subtractAmount)

	expectedRefund := -subtractAmount
	assert.Equal(t, expectedRefund, mem.state.refund)
}

func TestInMemoryStateDB_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	refund := mem.GetRefund()
	assert.Equal(t, uint64(0), refund)
}

func TestInMemoryStateDB_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	key := common.Hash{0x4, 0x5, 0x6}
	expectedValue := common.Hash{0x7, 0x8, 0x9}

	mockWs.EXPECT().Has(addr).Return(true)
	mockWs.EXPECT().Get(addr).Return(mockAcc)
	mockAcc.EXPECT().GetStorageAt(key).Return(expectedValue)

	value := mem.GetCommittedState(addr, key)
	assert.Equal(t, expectedValue, value)
}

func TestInMemoryStateDB_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	key := common.Hash{0x4, 0x5, 0x6}
	expectedValue := common.Hash{0x7, 0x8, 0x9}

	mockWs.EXPECT().Has(addr).Return(true).Times(1)
	mockWs.EXPECT().Get(addr).Return(mockAcc).Times(1)
	mockAcc.EXPECT().GetStorageAt(key).Return(expectedValue).Times(1)

	value := mem.GetState(addr, key)
	assert.Equal(t, expectedValue, value)
}

func TestInMemoryStateDB_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	key := common.Hash{0x4, 0x5, 0x6}
	value := common.Hash{0x7, 0x8, 0x9}

	mockWs.EXPECT().Has(addr).Return(true)
	mockWs.EXPECT().Get(addr).Return(mockAcc)
	mockAcc.EXPECT().GetStorageAt(key).Return(common.Hash{})
	hash := mem.SetState(addr, key, value)
	assert.Equal(t, common.Hash{}, hash)
}

func TestInMemoryStateDB_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	storageRoot := mem.GetStorageRoot(addr)
	assert.Equal(t, common.Hash{}, storageRoot)
}

func TestInMemoryStateDB_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	mockWs.EXPECT().Get(addr).Return(mockAcc)
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(100))
	value := mem.SelfDestruct(addr)

	assert.Equal(t, uint256.NewInt(100), &value)
}

func TestInMemoryStateDB_HasBeenCreatedInThisTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	value := mem.hasBeenCreatedInThisTransaction(addr)
	assert.False(t, value)
}

func TestInMemoryStateDB_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	mockWs.EXPECT().Get(addr).Return(mockAcc)
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(100))

	// Call SelfDestruct6780
	value, value2 := mem.SelfDestruct6780(addr)
	assert.Equal(t, uint256.NewInt(100), &value)
	assert.False(t, value2)
}

func TestInMemoryStateDB_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	hasSelfDestructed := mem.HasSelfDestructed(addr)
	assert.False(t, hasSelfDestructed)
}

func TestInMemoryStateDB_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	mockWs.EXPECT().Has(addr).Return(true)

	exists := mem.Exist(addr)
	assert.True(t, exists)
}

func TestInMemoryStateDB_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)
	addr := common.Address{0x1, 0x2, 0x3}

	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	mockWs.EXPECT().Has(addr).Return(false).Times(1)
	mockWs.EXPECT().Get(addr).Return(mockAcc).Times(2)
	mockAcc.EXPECT().GetNonce().Return(uint64(0))
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(0))

	isEmpty := mem.Empty(addr)
	assert.True(t, isEmpty)
}

func TestInMemoryStateDB_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	mem.Prepare(params.TestRules, addr, addr, &addr, []common.Address{addr}, nil)
	assert.Equal(t, 1, len(mem.state.accessed_accounts))
}

func TestInMemoryStateDB_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	value := mem.AddressInAccessList(addr)
	assert.False(t, value)

}

func TestInMemoryStateDB_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	key := common.Hash{0x4, 0x5, 0x6}

	value, value2 := mem.SlotInAccessList(addr, key)
	assert.False(t, value)
	assert.False(t, value2)
}

func TestInMemoryStateDB_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}

	mem.AddAddressToAccessList(addr)
	assert.Contains(t, mem.state.accessed_accounts, addr)
}

func TestInMemoryStateDB_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	addr := common.Address{0x1, 0x2, 0x3}
	key := common.Hash{0x4, 0x5, 0x6}

	mem.AddSlotToAccessList(addr, key)
	assert.Contains(t, mem.state.accessed_slots, slot{addr: addr, key: key})
}

func TestInMemoryStateDB_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	assert.NotPanics(t, func() {
		mem.RevertToSnapshot(0)
	})
}

func TestInMemoryStateDB_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	ss := mem.Snapshot()
	assert.Equal(t, 0, ss)
	assert.Equal(t, 1, mem.snapshot_counter)
}

func TestInMemoryStateDB_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	log := &types.Log{
		Address: common.Address{0x1, 0x2, 0x3},
		Data:    []byte{0x4, 0x5, 0x6},
	}

	mem.AddLog(log)
	assert.Contains(t, mem.state.logs, log)
}

func TestInMemoryStateDB_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	assert.Panics(t, func() {
		mem.AddPreimage(common.Hash{}, nil)
	})

}

func TestInMemoryStateDB_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockEvents := state.NewAccessEvents(utils.NewPointCache(4096))
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: mockEvents,
	}
	out := mem.AccessEvents()
	assert.Equal(t, mockEvents, out)
}

func TestInMemoryStateDB_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	assert.NotPanics(t, func() {
		mem.SetTxContext(common.Hash{}, 0)
	})
}

func TestInMemoryStateDB_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	assert.NotPanics(t, func() {
		mem.Finalise(false)
	})
}

func TestInMemoryStateDB_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	assert.Panics(t, func() {
		_ = mem.IntermediateRoot(false)
	})
}

func TestInMemoryStateDB_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	a, b := mem.Commit(uint64(0), false)
	assert.Equal(t, common.Hash{}, a)
	assert.Nil(t, b)
}

func TestInMemoryStateDB_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	txHash := common.Hash{0x1, 0x2, 0x3}
	blkNumber := uint64(10)
	blkHash := common.Hash{0x4, 0x5, 0x6}
	blkTimestamp := uint64(11)

	logs := mem.GetLogs(txHash, blkNumber, blkHash, blkTimestamp)
	assert.Empty(t, logs)
}

func TestInMemoryStateDB_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	assert.Panics(t, func() {
		_ = mem.PointCache()
	})
}

func TestInMemoryStateDB_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	a := mem.Witness()
	assert.Nil(t, a)
}

func TestInMemoryStateDB_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	a := mem.Error()
	assert.Nil(t, a)
}

func TestInMemoryStateDB_getEffects(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	mem.state.touched = map[common.Address]int{
		{0x1, 0x2, 0x3}: 1,
	}
	mem.state.storage = map[slot]common.Hash{
		{addr: common.Address{0x1, 0x2, 0x3}, key: common.Hash{0x4, 0x5, 0x6}}: {0x7, 0x8, 0x9},
	}
	mockWs.EXPECT().Get(gomock.Any()).Return(mockAcc).Times(3)
	mockWs.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	mockAcc.EXPECT().GetNonce().Return(uint64(0))
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(0))
	mockAcc.EXPECT().GetCode().Return([]byte{})
	ws := mem.getEffects()
	assert.NotNil(t, ws)
}

func TestInMemoryStateDB_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	mem.state.touched = map[common.Address]int{
		{0x1, 0x2, 0x3}: 1,
	}
	mem.state.storage = map[slot]common.Hash{
		{addr: common.Address{0x1, 0x2, 0x3}, key: common.Hash{0x4, 0x5, 0x6}}: {0x7, 0x8, 0x9},
	}
	mockWs.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	mockAcc.EXPECT().GetNonce().Return(uint64(1)).Times(3)
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(100)).Times(2)
	mockAcc.EXPECT().GetCode().Return([]byte{0x06}).Times(2)
	mockWs.EXPECT().Get(gomock.Any()).Return(mockAcc).Times(4)
	mockAcc.EXPECT().ForEachStorage(gomock.Any()).Do(func(ff func(key common.Hash, value common.Hash)) {
		ff(common.Hash{0x1, 0x2, 0x3}, common.Hash{0x4, 0x5, 0x6})
	})
	mockWs.EXPECT().ForEachAccount(gomock.Any()).Do(func(ff func(addr common.Address, acc txcontext.Account)) {
		ff(common.Address{0x1, 0x2, 0x3}, mockAcc)
	})
	ws := mem.GetSubstatePostAlloc()
	assert.NotNil(t, ws)
}

func TestInMemoryStateDB_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	err := mem.BeginTransaction(uint32(0))
	assert.NoError(t, err)
}

func TestInMemoryStateDB_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	err := mem.EndTransaction()
	assert.NoError(t, err)
}

func TestInMemoryStateDB_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	err := mem.BeginBlock(uint64(1))
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), mem.blockNum)
}

func TestInMemoryStateDB_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	err := mem.EndBlock()
	assert.NoError(t, err)
}

func TestInMemoryStateDB_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	assert.NotPanics(t, func() {
		mem.BeginSyncPeriod(uint64(0))
	})
}

func TestInMemoryStateDB_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	err := mem.EndSyncPeriod()
	assert.NoError(t, err)
}

func TestInMemoryStateDB_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	a, b := mem.GetHash()
	assert.Equal(t, common.Hash{}, a)
	assert.Nil(t, b)

}

func TestInMemoryStateDB_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	a := mem.Close()
	assert.Nil(t, a)
}

func TestInMemoryStateDB_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}

	a := mem.GetMemoryUsage()
	assert.Equal(t, uint64(0), a.UsedBytes)
}

func TestInMemoryStateDB_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	a, b := mem.GetArchiveState(uint64(1))
	assert.Nil(t, a)
	assert.Error(t, b)
}

func TestInMemoryStateDB_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	a, b, c := mem.GetArchiveBlockHeight()
	assert.Equal(t, uint64(0), a)
	assert.Equal(t, false, b)
	assert.Error(t, c)
}

func TestInMemoryStateDB_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockWs2 := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	mem.PrepareSubstate(mockWs2, uint64(1))
	assert.Equal(t, uint64(1), mem.blockNum)
	assert.Equal(t, mockWs2, mem.ws)
}

func TestInMemoryStateDB_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	a, b := mem.StartBulkLoad(uint64(1))
	assert.NotNil(t, a)
	assert.Nil(t, b)
}

func TestInMemoryStateDB_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mem := &inMemoryStateDB{
		ws:           mockWs,
		state:        makeSnapshot(nil, 0),
		blockNum:     1,
		accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
	}
	a := mem.GetShadowDB()
	assert.Nil(t, a)
}

// gethInMemoryBulkLoad struct method tests
func TestGethInMemoryBulkLoad_CreateAccount(t *testing.T) {
	g := &gethInMemoryBulkLoad{}
	assert.NotPanics(t, func() {
		g.CreateAccount(common.Address{0x1, 0x2, 0x3})
	})
}

func TestGethInMemoryBulkLoad_SetBalance(t *testing.T) {
	g := &gethInMemoryBulkLoad{}
	assert.NotPanics(t, func() {
		g.SetBalance(common.Address{0x1, 0x2, 0x3}, uint256.NewInt(100))
	})
}

func TestGethInMemoryBulkLoad_SetNonce(t *testing.T) {
	g := &gethInMemoryBulkLoad{}
	assert.NotPanics(t, func() {
		g.SetNonce(common.Address{0x1, 0x2, 0x3}, 42)
	})
}

func TestGethInMemoryBulkLoad_SetState(t *testing.T) {
	g := &gethInMemoryBulkLoad{}
	assert.NotPanics(t, func() {
		g.SetState(common.Address{0x1, 0x2, 0x3}, common.Hash{0x4, 0x5, 0x6}, common.Hash{0x7, 0x8, 0x9})
	})
}

func TestGethInMemoryBulkLoad_SetCode(t *testing.T) {
	g := &gethInMemoryBulkLoad{}
	assert.NotPanics(t, func() {
		g.SetCode(common.Address{0x1, 0x2, 0x3}, []byte{0x60, 0x00, 0x60, 0x00})
	})
}

func TestGethInMemoryBulkLoad_Close(t *testing.T) {
	g := &gethInMemoryBulkLoad{}
	a := g.Close()
	assert.Nil(t, a)
}
