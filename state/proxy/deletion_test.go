package proxy

import (
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProxy_NewDeletionProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")
	assert.NotNil(t, proxy)
	assert.Equal(t, mockDb, proxy.db)
	assert.Equal(t, mockChan, proxy.ch)
}
func TestDeletionProxy_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	mockDb.EXPECT().CreateAccount(address).Times(1)

	proxy.CreateAccount(address)
}
func TestDeletionProxy_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	amount := uint256.NewInt(100)

	mockDb.EXPECT().SubBalance(address, amount, gomock.Any()).Return(*amount).Times(1)

	result := proxy.SubBalance(address, amount, tracing.BalanceChangeUnspecified)
	assert.Equal(t, *amount, result)
}
func TestDeletionProxy_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	amount := uint256.NewInt(100)

	mockDb.EXPECT().AddBalance(address, amount, gomock.Any()).Return(*amount).Times(1)

	result := proxy.AddBalance(address, amount, tracing.BalanceChangeUnspecified)
	assert.Equal(t, *amount, result)
}
func TestDeletionProxy_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedBalance := uint256.NewInt(100)

	mockDb.EXPECT().GetBalance(address).Return(expectedBalance).Times(1)

	balance := proxy.GetBalance(address)
	assert.Equal(t, expectedBalance, balance)
}
func TestDeletionProxy_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedNonce := uint64(42)

	mockDb.EXPECT().GetNonce(address).Return(expectedNonce).Times(1)

	nonce := proxy.GetNonce(address)
	assert.Equal(t, expectedNonce, nonce)
}
func TestDeletionProxy_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	nonce := uint64(42)

	mockDb.EXPECT().SetNonce(address, nonce, gomock.Any()).Times(1)

	proxy.SetNonce(address, nonce, tracing.NonceChangeUnspecified)
}
func TestDeletionProxy_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedHash := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")

	mockDb.EXPECT().GetCodeHash(address).Return(expectedHash).Times(1)

	hash := proxy.GetCodeHash(address)
	assert.Equal(t, expectedHash, hash)
}
func TestDeletionProxy_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedCode := []byte{0x01, 0x02, 0x03}

	mockDb.EXPECT().GetCode(address).Return(expectedCode).Times(1)

	code := proxy.GetCode(address)
	assert.Equal(t, expectedCode, code)
}
func TestDeletionProxy_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	code := []byte{0x01, 0x02, 0x03}

	mockDb.EXPECT().SetCode(address, code).Return(code).Times(1)

	result := proxy.SetCode(address, code)
	assert.Equal(t, code, result)
}
func TestDeletionProxy_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedSize := 3

	mockDb.EXPECT().GetCodeSize(address).Return(expectedSize).Times(1)

	size := proxy.GetCodeSize(address)
	assert.Equal(t, expectedSize, size)
}
func TestDeletionProxy_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	refundAmount := uint64(100)

	mockDb.EXPECT().AddRefund(refundAmount).Times(1)

	proxy.AddRefund(refundAmount)
}
func TestDeletionProxy_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	refundAmount := uint64(50)

	mockDb.EXPECT().SubRefund(refundAmount).Times(1)

	proxy.SubRefund(refundAmount)
}
func TestDeletionProxy_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedRefund := uint64(200)

	mockDb.EXPECT().GetRefund().Return(expectedRefund).Times(1)

	refund := proxy.GetRefund()
	assert.Equal(t, expectedRefund, refund)
}
func TestDeletionProxy_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")
	expectedValue := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().GetCommittedState(address, slot).Return(expectedValue).Times(1)

	value := proxy.GetCommittedState(address, slot)
	assert.Equal(t, expectedValue, value)
}
func TestDeletionProxy_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")
	expectedValue := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().GetState(address, slot).Return(expectedValue).Times(1)

	value := proxy.GetState(address, slot)
	assert.Equal(t, expectedValue, value)
}
func TestDeletionProxy_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")
	value := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().SetState(address, slot, value).Return(value).Times(1)

	result := proxy.SetState(address, slot, value)
	assert.Equal(t, value, result)
}
func TestDeletionProxy_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")
	value := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().SetTransientState(address, slot, value).Times(1)

	proxy.SetTransientState(address, slot, value)
}
func TestDeletionProxy_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")
	expectedValue := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().GetTransientState(address, slot).Return(expectedValue).Times(1)

	value := proxy.GetTransientState(address, slot)
	assert.Equal(t, expectedValue, value)
}
func TestDeletionProxy_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedRefund := uint256.NewInt(100)

	mockDb.EXPECT().SelfDestruct(address).Return(*expectedRefund).Times(1)

	refund := proxy.SelfDestruct(address)
	assert.Equal(t, *expectedRefund, refund)
}
func TestDeletionProxy_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().HasSelfDestructed(address).Return(true).Times(1)

	destructed := proxy.HasSelfDestructed(address)
	assert.True(t, destructed)
}
func TestDeletionProxy_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().Exist(address).Return(true).Times(1)

	exists := proxy.Exist(address)
	assert.True(t, exists)
}
func TestDeletionProxy_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().Empty(address).Return(true).Times(1)

	empty := proxy.Empty(address)
	assert.True(t, empty)
}
func TestDeletionProxy_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	rules := params.TestRules
	from := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	to := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")
	accessList := types.AccessList{}

	mockDb.EXPECT().Prepare(rules, from, to, nil, nil, accessList).Times(1)

	proxy.Prepare(rules, from, to, nil, nil, accessList)

}
func TestDeletionProxy_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().AddAddressToAccessList(address).Times(1)

	proxy.AddAddressToAccessList(address)
}
func TestDeletionProxy_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().AddressInAccessList(address).Return(true).Times(1)

	result := proxy.AddressInAccessList(address)
	assert.True(t, result)
}
func TestDeletionProxy_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")

	mockDb.EXPECT().SlotInAccessList(address, slot).Return(true, true).Times(1)

	inList, isReadOnly := proxy.SlotInAccessList(address, slot)
	assert.True(t, inList)
	assert.True(t, isReadOnly)
}
func TestDeletionProxy_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")

	mockDb.EXPECT().AddSlotToAccessList(address, slot).Times(1)

	proxy.AddSlotToAccessList(address, slot)
}
func TestDeletionProxy_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	snapshotId := 1

	mockDb.EXPECT().RevertToSnapshot(snapshotId).Times(1)

	proxy.RevertToSnapshot(snapshotId)
}
func TestDeletionProxy_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	mockDb.EXPECT().Snapshot().Return(1).Times(1)

	snapshotId := proxy.Snapshot()
	assert.Equal(t, 1, snapshotId)
}
func TestDeletionProxy_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	log := &types.Log{
		Address: common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		Data:    []byte{0x01, 0x02, 0x03},
	}

	mockDb.EXPECT().AddLog(log).Times(1)

	proxy.AddLog(log)
}
func TestDeletionProxy_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")
	startBlock := uint64(1)
	endBlock := uint64(10)
	expectedLogs := []*types.Log{
		{Data: []byte{0x01}},
		{Data: []byte{0x02}},
	}

	mockDb.EXPECT().GetLogs(address, startBlock, address, endBlock).Return(expectedLogs).Times(1)

	logs := proxy.GetLogs(address, startBlock, address, endBlock)
	assert.Equal(t, expectedLogs, logs)
}
func TestDeletionProxy_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedPointCache := &utils.PointCache{}
	mockDb.EXPECT().PointCache().Return(expectedPointCache).Times(1)
	pointCache := proxy.PointCache()
	assert.Equal(t, expectedPointCache, pointCache)
}
func TestDeletionProxy_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedWitness := &stateless.Witness{}
	mockDb.EXPECT().Witness().Return(expectedWitness).Times(1)
	witness := proxy.Witness()
	assert.NotNil(t, witness)
	assert.Equal(t, expectedWitness, witness)
}
func TestDeletionProxy_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	slot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")
	preimage := []byte{0x01, 0x02, 0x03}

	mockDb.EXPECT().AddPreimage(slot, preimage).Times(1)

	proxy.AddPreimage(slot, preimage)
}
func TestDeletionProxy_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedAccessEvents := &geth.AccessEvents{}
	mockDb.EXPECT().AccessEvents().Return(expectedAccessEvents).Times(1)

	events := proxy.AccessEvents()
	assert.Equal(t, expectedAccessEvents, events)
}
func TestDeletionProxy_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	txHash := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")
	txIndex := 1

	mockDb.EXPECT().SetTxContext(txHash, txIndex).Times(1)

	proxy.SetTxContext(txHash, txIndex)
}
func TestDeletionProxy_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	mockDb.EXPECT().Finalise(true).Times(1)

	proxy.Finalise(true)
}
func TestDeletionProxy_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedRoot := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")
	mockDb.EXPECT().IntermediateRoot(true).Return(expectedRoot).Times(1)

	root := proxy.IntermediateRoot(true)
	assert.Equal(t, expectedRoot, root)
}
func TestDeletionProxy_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedRoot := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")
	expectedBlock := uint64(100)

	mockDb.EXPECT().Commit(expectedBlock, true).Return(expectedRoot, nil).Times(1)

	root, err := proxy.Commit(expectedBlock, true)
	assert.NoError(t, err)
	assert.Equal(t, expectedRoot, root)
}
func TestDeletionProxy_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedHash := common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")
	mockDb.EXPECT().GetHash().Return(expectedHash, nil).Times(1)

	hash, err := proxy.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}
func TestDeletionProxy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedError := assert.AnError
	mockDb.EXPECT().Error().Return(expectedError).Times(1)

	err := proxy.Error()
	assert.Equal(t, expectedError, err)
}
func TestDeletionProxy_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedSubstate := txcontext.NewMockWorldState(ctrl)
	mockDb.EXPECT().GetSubstatePostAlloc().Return(expectedSubstate).Times(1)

	substate := proxy.GetSubstatePostAlloc()
	assert.Equal(t, expectedSubstate, substate)
}
func TestDeletionProxy_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	worldState := txcontext.NewMockWorldState(ctrl)
	blockNumber := uint64(100)

	mockDb.EXPECT().PrepareSubstate(worldState, blockNumber).Times(1)

	proxy.PrepareSubstate(worldState, blockNumber)
}
func TestDeletionProxy_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")
	expectedBlock := uint32(100)
	mockDb.EXPECT().BeginTransaction(expectedBlock).Return(nil).Times(1)

	err := proxy.BeginTransaction(expectedBlock)
	assert.NoError(t, err)
}
func TestDeletionProxy_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	mockDb.EXPECT().EndTransaction().Return(nil).Times(1)

	err := proxy.EndTransaction()
	assert.NoError(t, err)
}
func TestDeletionProxy_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedBlockNumber := uint64(100)
	mockDb.EXPECT().BeginBlock(expectedBlockNumber).Return(nil).Times(1)

	err := proxy.BeginBlock(expectedBlockNumber)
	assert.NoError(t, err)
}
func TestDeletionProxy_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	mockDb.EXPECT().EndBlock().Return(nil).Times(1)

	err := proxy.EndBlock()
	assert.NoError(t, err)
}
func TestDeletionProxy_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedSyncPeriod := uint64(100)
	mockDb.EXPECT().BeginSyncPeriod(expectedSyncPeriod).Times(1)

	proxy.BeginSyncPeriod(expectedSyncPeriod)
}
func TestDeletionProxy_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	mockDb.EXPECT().EndSyncPeriod().Times(1)

	proxy.EndSyncPeriod()
}
func TestDeletionProxy_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedBlock := uint64(100)
	expectedState := state.NewMockNonCommittableStateDB(ctrl)

	mockDb.EXPECT().GetArchiveState(expectedBlock).Return(expectedState, nil).Times(1)

	st, err := proxy.GetArchiveState(expectedBlock)
	assert.NoError(t, err)
	assert.Equal(t, expectedState, st)
}
func TestDeletionProxy_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedHeight := uint64(100)
	mockDb.EXPECT().GetArchiveBlockHeight().Return(expectedHeight, true, nil).Times(1)

	height, exists, err := proxy.GetArchiveBlockHeight()
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, expectedHeight, height)
}
func TestDeletionProxy_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	mockDb.EXPECT().Close().Return(nil).Times(1)

	err := proxy.Close()
	assert.NoError(t, err)
}
func TestDeletionProxy_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	mockLogger := logger.NewMockLogger(ctrl)
	proxy := &DeletionProxy{
		db:  mockDb,
		ch:  mockChan,
		log: mockLogger,
	}

	expectedBlock := uint64(100)

	mockLogger.EXPECT().Fatal(gomock.Any()).Times(1)
	bulkLoad, err := proxy.StartBulkLoad(expectedBlock)
	assert.Nil(t, err)
	assert.Nil(t, bulkLoad)
}
func TestDeletionProxy_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedMemoryUsage := &state.MemoryUsage{
		UsedBytes: 1024,
		Breakdown: nil,
	}

	mockDb.EXPECT().GetMemoryUsage().Return(expectedMemoryUsage).Times(1)

	memoryUsage := proxy.GetMemoryUsage()
	assert.Equal(t, expectedMemoryUsage, memoryUsage)
}
func TestDeletionProxy_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	expectedShadowDB := state.NewMockStateDB(ctrl)

	mockDb.EXPECT().GetShadowDB().Return(expectedShadowDB).Times(1)

	shadowDB := proxy.GetShadowDB()
	assert.Equal(t, expectedShadowDB, shadowDB)
}
func TestDeletionProxy_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	mockDb.EXPECT().CreateContract(address).Return().Times(1)

	proxy.CreateContract(address)
}
func TestDeletionProxy_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedRefund := uint256.NewInt(100)

	mockDb.EXPECT().SelfDestruct6780(address).Return(*expectedRefund, true).Times(1)

	refund, affected := proxy.SelfDestruct6780(address)
	assert.True(t, affected)
	assert.Equal(t, *expectedRefund, refund)
}
func TestDeletionProxy_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedRoot := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef")

	mockDb.EXPECT().GetStorageRoot(address).Return(expectedRoot).Times(1)

	root := proxy.GetStorageRoot(address)
	assert.Equal(t, expectedRoot, root)
}
