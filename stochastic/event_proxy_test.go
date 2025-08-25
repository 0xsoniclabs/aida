package stochastic

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TODO test all proxy calls

func TestEventProxy_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().CreateAccount(addr)
	proxy.CreateAccount(addr)
}

func TestEventProxy_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	amount := uint256.NewInt(100)
	reason := tracing.BalanceChangeUnspecified
	base.EXPECT().SubBalance(addr, amount, reason)
	proxy.SubBalance(addr, amount, reason)
}

func TestEventProxy_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	amount := uint256.NewInt(100)
	reason := tracing.BalanceChangeUnspecified
	base.EXPECT().AddBalance(addr, amount, reason)
	proxy.AddBalance(addr, amount, reason)
}

func TestEventProxy_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedBalance := uint256.NewInt(100)
	base.EXPECT().GetBalance(addr).Return(expectedBalance)
	balance := proxy.GetBalance(addr)
	assert.Equal(t, expectedBalance, balance)
}

func TestEventProxy_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedNonce := uint64(42)
	base.EXPECT().GetNonce(addr).Return(expectedNonce)
	nonce := proxy.GetNonce(addr)
	assert.Equal(t, expectedNonce, nonce)
}

func TestEventProxy_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	nonce := uint64(42)
	reason := tracing.NonceChangeUnspecified
	base.EXPECT().SetNonce(addr, nonce, reason)
	proxy.SetNonce(addr, nonce, reason)
}

func TestEventProxy_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedHash := common.HexToHash("0x5678")
	base.EXPECT().GetCodeHash(addr).Return(expectedHash)
	hash := proxy.GetCodeHash(addr)
	assert.Equal(t, expectedHash, hash)
}

func TestEventProxy_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedCode := []byte{0x01, 0x02, 0x03}
	base.EXPECT().GetCode(addr).Return(expectedCode)
	code := proxy.GetCode(addr)
	assert.Equal(t, expectedCode, code)
}

func TestEventProxy_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	code := []byte{0x01, 0x02, 0x03}
	base.EXPECT().SetCode(addr, code)
	proxy.SetCode(addr, code)
}

func TestEventProxy_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedSize := 3
	base.EXPECT().GetCodeSize(addr).Return(expectedSize)
	size := proxy.GetCodeSize(addr)
	assert.Equal(t, expectedSize, size)
}

func TestEventProxy_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	amount := uint64(100)
	base.EXPECT().AddRefund(amount)
	proxy.AddRefund(amount)
}

func TestEventProxy_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	amount := uint64(50)
	base.EXPECT().SubRefund(amount)
	proxy.SubRefund(amount)
}

func TestEventProxy_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedRefund := uint64(100)
	base.EXPECT().GetRefund().Return(expectedRefund)
	refund := proxy.GetRefund()
	assert.Equal(t, expectedRefund, refund)
}

func TestEventProxy_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	base.EXPECT().GetCommittedState(addr, key).Return(key)
	out := proxy.GetCommittedState(addr, key)
	assert.Equal(t, key, out)
}

func TestEventProxy_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	base.EXPECT().GetState(addr, key).Return(expectedValue)
	value := proxy.GetState(addr, key)
	assert.Equal(t, expectedValue, value)
}

func TestEventProxy_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	base.EXPECT().SetState(addr, key, value)
	proxy.SetState(addr, key, value)
}

func TestEventProxy_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	base.EXPECT().SetTransientState(addr, key, value)
	proxy.SetTransientState(addr, key, value)
}

func TestEventProxy_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	base.EXPECT().GetState(addr, key).Return(expectedValue)
	value := proxy.GetTransientState(addr, key)
	assert.Equal(t, expectedValue, value)
}

func TestEventProxy_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().SelfDestruct(addr)
	out := proxy.SelfDestruct(addr)
	assert.Equal(t, uint256.Int{0x0, 0x0, 0x0, 0x0}, out)
}

func TestEventProxy_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().HasSelfDestructed(addr).Return(true)
	result := proxy.HasSelfDestructed(addr)
	assert.True(t, result)
}

func TestEventProxy_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().Exist(addr).Return(true)
	result := proxy.Exist(addr)
	assert.True(t, result)
}

func TestEventProxy_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().Empty(addr).Return(false)
	result := proxy.Empty(addr)
	assert.False(t, result)
}

func TestEventProxy_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	rule := params.Rules{}
	sender := common.HexToAddress("0x1234")
	coinbase := common.HexToAddress("0x5678")
	dest := common.HexToAddress("0xabcd")
	base.EXPECT().Prepare(rule, sender, coinbase, &dest, nil, nil)
	proxy.Prepare(rule, sender, coinbase, &dest, nil, nil)
}

func TestEventProxy_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().AddAddressToAccessList(addr)
	proxy.AddAddressToAccessList(addr)
}

func TestEventProxy_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().AddressInAccessList(addr).Return(true)
	result := proxy.AddressInAccessList(addr)
	assert.True(t, result)
}

func TestEventProxy_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	base.EXPECT().SlotInAccessList(addr, slot).Return(true, false)
	res1, res2 := proxy.SlotInAccessList(addr, slot)
	assert.True(t, res1)
	assert.False(t, res2)
}

func TestEventProxy_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	base.EXPECT().AddSlotToAccessList(addr, slot)
	proxy.AddSlotToAccessList(addr, slot)
}

func TestEventProxy_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	snapshotID := 1
	base.EXPECT().RevertToSnapshot(snapshotID)
	proxy.RevertToSnapshot(snapshotID)
}

func TestEventProxy_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().Snapshot().Return(1)
	snapshotID := proxy.Snapshot()
	assert.Equal(t, 1, snapshotID)
}

func TestEventProxy_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	log := &types.Log{
		Address: common.HexToAddress("0x1234"),
		Topics:  []common.Hash{common.HexToHash("0x5678")},
		Data:    []byte{0x01, 0x02, 0x03},
	}
	base.EXPECT().AddLog(log)
	proxy.AddLog(log)
}

func TestEventProxy_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	hash := common.Hash{0x12}
	blk := uint64(2)
	blkHash := common.Hash{2}
	blkTimestamp := uint64(13)
	base.EXPECT().GetLogs(hash, blk, blkHash, blkTimestamp)
	proxy.GetLogs(hash, blk, blkHash, blkTimestamp)
}

func TestEventProxy_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expected := utils.PointCache{}
	base.EXPECT().PointCache().Return(&expected)
	out := proxy.PointCache()
	assert.Equal(t, &expected, out)
}

func TestEventProxy_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedWitness := &stateless.Witness{}
	base.EXPECT().Witness().Return(expectedWitness)
	out := proxy.Witness()
	assert.Equal(t, expectedWitness, out)
}

func TestEventProxy_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	preimage := common.HexToHash("0x1234")
	data := []byte{0x01, 0x02, 0x03}
	base.EXPECT().AddPreimage(preimage, data)
	proxy.AddPreimage(preimage, data)
}

func TestEventProxy_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedEvents := &gethstate.AccessEvents{}
	base.EXPECT().AccessEvents().Return(expectedEvents)
	events := proxy.AccessEvents()
	assert.Equal(t, expectedEvents, events)
}

func TestEventProxy_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	hash := common.HexToHash("0x1234")
	ti := 1
	base.EXPECT().SetTxContext(hash, ti).Return()
	proxy.SetTxContext(hash, ti)
}

func TestEventProxy_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().Finalise(true).Return()
	proxy.Finalise(true)
}

func TestEventProxy_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedRoot := common.HexToHash("0x1234")
	base.EXPECT().IntermediateRoot(true).Return(expectedRoot)
	root := proxy.IntermediateRoot(true)
	assert.Equal(t, expectedRoot, root)
}

func TestEventProxy_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	block := uint64(1)
	expectedRoot := common.HexToHash("0x1234")
	base.EXPECT().Commit(block, true).Return(expectedRoot, nil)
	root, err := proxy.Commit(block, true)
	assert.Equal(t, expectedRoot, root)
	assert.NoError(t, err)
}

func TestEventProxy_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedHash := common.HexToHash("0x1234")
	base.EXPECT().GetHash().Return(expectedHash, nil)
	hash, err := proxy.GetHash()
	assert.Equal(t, expectedHash, hash)
	assert.NoError(t, err)
}

func TestEventProxy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedError := errors.New("test error")
	base.EXPECT().Error().Return(expectedError)
	err := proxy.Error()
	assert.Equal(t, err, expectedError)
}

func TestEventProxy_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedSubstate := txcontext.NewWorldState(nil)
	base.EXPECT().GetSubstatePostAlloc().Return(expectedSubstate)
	substate := proxy.GetSubstatePostAlloc()
	assert.Equal(t, expectedSubstate, substate)
}

func TestEventProxy_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	substate := txcontext.NewWorldState(nil)
	base.EXPECT().PrepareSubstate(substate, uint64(1))
	proxy.PrepareSubstate(substate, uint64(1))
}

func TestEventProxy_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().BeginTransaction(uint32(32)).Return(nil)
	err := proxy.BeginTransaction(uint32(32))
	assert.NoError(t, err)
}

func TestEventProxy_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().EndTransaction().Return(nil)
	err := proxy.EndTransaction()
	assert.NoError(t, err)
}

func TestEventProxy_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	blockNumber := uint64(1)
	base.EXPECT().BeginBlock(blockNumber).Return(nil)
	err := proxy.BeginBlock(blockNumber)
	assert.NoError(t, err)
}

func TestEventProxy_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().EndBlock().Return(nil)
	err := proxy.EndBlock()
	assert.NoError(t, err)
}

func TestEventProxy_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	blockNumber := uint64(1)
	base.EXPECT().BeginSyncPeriod(blockNumber).Return()
	proxy.BeginSyncPeriod(blockNumber)
}

func TestEventProxy_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().EndSyncPeriod().Return()
	proxy.EndSyncPeriod()
}

func TestEventProxy_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().Close().Return(nil)
	err := proxy.Close()
	assert.NoError(t, err)
}

func TestEventProxy_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	assert.Panics(t, func() {
		_, _ = proxy.StartBulkLoad(uint64(1))
	})
}

func TestEventProxy_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedUsage := &state.MemoryUsage{}
	base.EXPECT().GetMemoryUsage().Return(expectedUsage)
	usage := proxy.GetMemoryUsage()
	assert.Equal(t, expectedUsage, usage)
}

func TestEventProxy_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	blockNumber := uint64(10)
	expectedState := state.NewMockNonCommittableStateDB(ctrl)
	base.EXPECT().GetArchiveState(blockNumber).Return(expectedState, nil)
	st, err := proxy.GetArchiveState(blockNumber)
	assert.NoError(t, err)
	assert.Equal(t, expectedState, st)
}

func TestEventProxy_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	expectedHeight := uint64(100)
	base.EXPECT().GetArchiveBlockHeight().Return(expectedHeight, false, nil)
	height, _, err := proxy.GetArchiveBlockHeight()
	assert.NoError(t, err)
	assert.Equal(t, expectedHeight, height)
}

func TestEventProxy_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	base.EXPECT().GetShadowDB().Return(base)
	shadowDB := proxy.GetShadowDB()
	assert.Equal(t, base, shadowDB)
}
