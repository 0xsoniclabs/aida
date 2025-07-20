package proxy

import (
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/0xsoniclabs/aida/tracer/operation"
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
	"testing"
)

func TestProxy_NewRecorderProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	mockDb := state.NewMockStateDB(ctrl)
	mockCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}

	proxy := NewRecorderProxy(mockDb, mockCtx)
	assert.NotNil(t, proxy)
	assert.Equal(t, mockDb, proxy.db)
	assert.Equal(t, mockCtx, proxy.ctx)
}
func TestRecorderProxy_write(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	mockCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	mockOper := operation.NewMockOperation(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: mockCtx,
	}
	mockOper.EXPECT().GetId().Return(uint8(1))
	mockOper.EXPECT().Write(gomock.Any()).Return(nil)
	proxy.write(mockOper)
}
func TestRecorderProxy_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	mockDb.EXPECT().CreateAccount(addr).Times(1)

	proxy.CreateAccount(addr)
}
func TestRecorderProxy_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	subAmount := uint256.NewInt(100)
	reason := tracing.BalanceChangeUnspecified
	expectedOutput := uint256.NewInt(1)
	mockDb.EXPECT().SubBalance(addr, subAmount, reason).Return(*expectedOutput).Times(1)

	output := proxy.SubBalance(addr, subAmount, reason)
	assert.Equal(t, *expectedOutput, output)
}
func TestRecorderProxy_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	addAmount := uint256.NewInt(100)
	reason := tracing.BalanceChangeUnspecified
	expectedOutput := uint256.NewInt(1)
	mockDb.EXPECT().AddBalance(addr, addAmount, reason).Return(*expectedOutput).Times(1)

	output := proxy.AddBalance(addr, addAmount, reason)
	assert.Equal(t, *expectedOutput, output)
}
func TestRecorderProxy_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedBalance := uint256.NewInt(100)
	mockDb.EXPECT().GetBalance(addr).Return(expectedBalance).Times(1)

	balance := proxy.GetBalance(addr)
	assert.Equal(t, expectedBalance, balance)
}
func TestRecorderProxy_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedNonce := uint64(42)
	mockDb.EXPECT().GetNonce(addr).Return(expectedNonce).Times(1)

	nonce := proxy.GetNonce(addr)
	assert.Equal(t, expectedNonce, nonce)
}
func TestRecorderProxy_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	newNonce := uint64(100)
	reason := tracing.NonceChangeUnspecified
	mockDb.EXPECT().SetNonce(addr, newNonce, reason).Times(1)

	proxy.SetNonce(addr, newNonce, reason)
}
func TestRecorderProxy_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedHash := common.Hash{0x12}
	mockDb.EXPECT().GetCodeHash(addr).Return(expectedHash).Times(1)

	hash := proxy.GetCodeHash(addr)
	assert.Equal(t, expectedHash, hash)
}
func TestRecorderProxy_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedCode := []byte{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetCode(addr).Return(expectedCode).Times(1)

	code := proxy.GetCode(addr)
	assert.Equal(t, expectedCode, code)
}
func TestRecorderProxy_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	newCode := []byte{0x01, 0x02, 0x03}
	mockDb.EXPECT().SetCode(addr, newCode).Times(1)

	proxy.SetCode(addr, newCode)
}
func TestRecorderProxy_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedSize := 3
	mockDb.EXPECT().GetCodeSize(addr).Return(expectedSize).Times(1)

	size := proxy.GetCodeSize(addr)
	assert.Equal(t, expectedSize, size)
}
func TestRecorderProxy_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	refundAmount := uint64(100)
	mockDb.EXPECT().AddRefund(refundAmount).Times(1)

	proxy.AddRefund(refundAmount)
}
func TestRecorderProxy_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	refundAmount := uint64(50)
	mockDb.EXPECT().SubRefund(refundAmount).Times(1)

	proxy.SubRefund(refundAmount)
}
func TestRecorderProxy_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedRefund := uint64(200)
	mockDb.EXPECT().GetRefund().Return(expectedRefund).Times(1)

	refund := proxy.GetRefund()
	assert.Equal(t, expectedRefund, refund)
}
func TestRecorderProxy_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	keyHash := common.Hash{0x12}
	expectedState := common.Hash{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetCommittedState(addr, keyHash).Return(expectedState).Times(1)

	stateData := proxy.GetCommittedState(addr, keyHash)
	assert.Equal(t, expectedState, stateData)
}
func TestRecorderProxy_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	keyHash := common.Hash{0x12}
	expectedState := common.Hash{0x01, 0x02, 0x03}
	mockDb.EXPECT().GetState(addr, keyHash).Return(expectedState).Times(1)

	stateData := proxy.GetState(addr, keyHash)
	assert.Equal(t, expectedState, stateData)
}
func TestRecorderProxy_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	keyHash := common.Hash{0x12}
	value := common.Hash{0x13}
	expectedHash := common.Hash{0x14}
	mockDb.EXPECT().SetState(addr, keyHash, value).Return(expectedHash).Times(1)

	h := proxy.SetState(addr, keyHash, value)
	assert.Equal(t, expectedHash, h)
}
func TestRecorderProxy_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	key := common.Hash{0x12}
	value := common.Hash{0x13}
	mockDb.EXPECT().SetTransientState(addr, key, value).Times(1)

	proxy.SetTransientState(addr, key, value)
}
func TestRecorderProxy_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	key := common.Hash{0x12}
	expectedValue := common.Hash{0x13}
	mockDb.EXPECT().GetTransientState(addr, key).Return(expectedValue).Times(1)

	value := proxy.GetTransientState(addr, key)
	assert.Equal(t, expectedValue, value)
}
func TestRecorderProxy_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedBalance := uint256.NewInt(100)
	mockDb.EXPECT().SelfDestruct(addr).Return(*expectedBalance).Times(1)

	balance := proxy.SelfDestruct(addr)
	assert.Equal(t, *expectedBalance, balance)
}
func TestRecorderProxy_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	mockDb.EXPECT().HasSelfDestructed(addr).Return(true).Times(1)

	exists := proxy.HasSelfDestructed(addr)
	assert.True(t, exists)
}
func TestRecorderProxy_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	mockDb.EXPECT().Exist(addr).Return(true).Times(1)

	exists := proxy.Exist(addr)
	assert.True(t, exists)
}
func TestRecorderProxy_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}
	addr := common.Address{0x11}
	mockDb.EXPECT().Empty(addr).Return(true).Times(1)

	isEmpty := proxy.Empty(addr)
	assert.True(t, isEmpty)
}

func TestRecorderProxy_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	rules := params.Rules{}
	sender := common.Address{0x01}
	coinbase := common.Address{0x02}
	dest := &common.Address{0x03}
	precompiles := []common.Address{{0x04}}
	txAccesses := types.AccessList{}
	mockDb.EXPECT().Prepare(rules, sender, coinbase, dest, precompiles, txAccesses).Times(1)

	proxy.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func TestRecorderProxy_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	mockDb.EXPECT().AddAddressToAccessList(addr).Times(1)

	proxy.AddAddressToAccessList(addr)
}

func TestRecorderProxy_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	mockDb.EXPECT().AddressInAccessList(addr).Return(true).Times(1)

	ok := proxy.AddressInAccessList(addr)
	assert.True(t, ok)
}

func TestRecorderProxy_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	slot := common.Hash{0x12}
	mockDb.EXPECT().SlotInAccessList(addr, slot).Return(true, false).Times(1)

	addressOk, slotOk := proxy.SlotInAccessList(addr, slot)
	assert.True(t, addressOk)
	assert.False(t, slotOk)
}

func TestRecorderProxy_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	slot := common.Hash{0x12}
	mockDb.EXPECT().AddSlotToAccessList(addr, slot).Times(1)

	proxy.AddSlotToAccessList(addr, slot)
}

func TestRecorderProxy_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	snapshot := 42
	mockDb.EXPECT().RevertToSnapshot(snapshot).Times(1)

	proxy.RevertToSnapshot(snapshot)
}

func TestRecorderProxy_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedSnapshot := 123
	mockDb.EXPECT().Snapshot().Return(expectedSnapshot).Times(1)

	snapshot := proxy.Snapshot()
	assert.Equal(t, expectedSnapshot, snapshot)
}

func TestRecorderProxy_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	log := &types.Log{
		Address: common.Address{0x11},
		Topics:  []common.Hash{{0x12}},
		Data:    []byte{0x01, 0x02},
	}
	mockDb.EXPECT().AddLog(log).Times(1)

	proxy.AddLog(log)
}

func TestRecorderProxy_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	hash := common.Hash{0x01}
	block := uint64(100)
	blockHash := common.Hash{0x02}
	blkTimestamp := uint64(1234567890)
	expectedLogs := []*types.Log{{Address: common.Address{0x11}}}
	mockDb.EXPECT().GetLogs(hash, block, blockHash, blkTimestamp).Return(expectedLogs).Times(1)

	logs := proxy.GetLogs(hash, block, blockHash, blkTimestamp)
	assert.Equal(t, expectedLogs, logs)
}

func TestRecorderProxy_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedCache := &utils.PointCache{}
	mockDb.EXPECT().PointCache().Return(expectedCache).Times(1)

	cache := proxy.PointCache()
	assert.Equal(t, expectedCache, cache)
}

func TestRecorderProxy_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedWitness := &stateless.Witness{}
	mockDb.EXPECT().Witness().Return(expectedWitness).Times(1)

	witness := proxy.Witness()
	assert.Equal(t, expectedWitness, witness)
}

func TestRecorderProxy_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Hash{0x11}
	image := []byte{0x01, 0x02, 0x03}
	mockDb.EXPECT().AddPreimage(addr, image).Times(1)

	proxy.AddPreimage(addr, image)
}

func TestRecorderProxy_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedEvents := &geth.AccessEvents{}
	mockDb.EXPECT().AccessEvents().Return(expectedEvents).Times(1)

	events := proxy.AccessEvents()
	assert.Equal(t, expectedEvents, events)
}

func TestRecorderProxy_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	thash := common.Hash{0x11}
	ti := 42
	mockDb.EXPECT().SetTxContext(thash, ti).Times(1)

	proxy.SetTxContext(thash, ti)
}

func TestRecorderProxy_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	mockDb.EXPECT().Finalise(true).Times(1)

	proxy.Finalise(true)
}

func TestRecorderProxy_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedRoot := common.Hash{0x11}
	mockDb.EXPECT().IntermediateRoot(true).Return(expectedRoot).Times(1)

	root := proxy.IntermediateRoot(true)
	assert.Equal(t, expectedRoot, root)
}

func TestRecorderProxy_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	block := uint64(100)
	expectedRoot := common.Hash{0x11}
	mockDb.EXPECT().Commit(block, true).Return(expectedRoot, nil).Times(1)

	root, err := proxy.Commit(block, true)
	assert.NoError(t, err)
	assert.Equal(t, expectedRoot, root)
}

func TestRecorderProxy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	mockDb.EXPECT().Error().Return(nil).Times(1)

	err = proxy.Error()
	assert.NoError(t, err)
}

func TestRecorderProxy_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedAlloc := txcontext.NewMockWorldState(ctrl)
	mockDb.EXPECT().GetSubstatePostAlloc().Return(expectedAlloc).Times(1)

	alloc := proxy.GetSubstatePostAlloc()
	assert.Equal(t, expectedAlloc, alloc)
}

func TestRecorderProxy_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	substate := txcontext.NewMockWorldState(ctrl)
	block := uint64(100)
	mockDb.EXPECT().PrepareSubstate(substate, block).Times(1)

	proxy.PrepareSubstate(substate, block)
}

func TestRecorderProxy_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	number := uint32(42)
	mockDb.EXPECT().BeginTransaction(number).Times(1)

	err = proxy.BeginTransaction(number)
	assert.NoError(t, err)
}

func TestRecorderProxy_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	mockDb.EXPECT().EndTransaction().Times(1)

	err = proxy.EndTransaction()
	assert.NoError(t, err)
}

func TestRecorderProxy_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	number := uint64(100)
	mockDb.EXPECT().BeginBlock(number).Times(1)

	err = proxy.BeginBlock(number)
	assert.NoError(t, err)
}

func TestRecorderProxy_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	mockDb.EXPECT().EndBlock().Times(1)

	err = proxy.EndBlock()
	assert.NoError(t, err)
}

func TestRecorderProxy_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	number := uint64(100)
	mockDb.EXPECT().BeginSyncPeriod(number).Times(1)

	proxy.BeginSyncPeriod(number)
}

func TestRecorderProxy_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	mockDb.EXPECT().EndSyncPeriod().Times(1)

	proxy.EndSyncPeriod()
}

func TestRecorderProxy_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedHash := common.Hash{0x11}
	mockDb.EXPECT().GetHash().Return(expectedHash, nil).Times(1)

	hash, err := proxy.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestRecorderProxy_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	block := uint64(100)

	archiveState, err := proxy.GetArchiveState(block)
	assert.Error(t, err)
	assert.Nil(t, archiveState)
	assert.Contains(t, err.Error(), "archive states are not (yet) supported")
}

func TestRecorderProxy_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	height, ok, err := proxy.GetArchiveBlockHeight()
	assert.Error(t, err)
	assert.Equal(t, uint64(0), height)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "archive states are not (yet) supported")
}

func TestRecorderProxy_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	mockDb.EXPECT().Close().Return(nil).Times(1)

	err = proxy.Close()
	assert.NoError(t, err)
}

func TestRecorderProxy_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	block := uint64(100)

	assert.Panics(t, func() {
		_, _ = proxy.StartBulkLoad(block)
	})
}

func TestRecorderProxy_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	expectedUsage := &state.MemoryUsage{}
	mockDb.EXPECT().GetMemoryUsage().Return(expectedUsage).Times(1)

	usage := proxy.GetMemoryUsage()
	assert.Equal(t, expectedUsage, usage)
}

func TestRecorderProxy_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	shadowDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	mockDb.EXPECT().GetShadowDB().Return(shadowDb).Times(1)

	result := proxy.GetShadowDB()
	assert.Equal(t, shadowDb, result)
}

func TestRecorderProxy_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	mockDb.EXPECT().CreateContract(addr).Times(1)

	proxy.CreateContract(addr)
}

func TestRecorderProxy_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedBalance := uint256.NewInt(100)
	mockDb.EXPECT().SelfDestruct6780(addr).Return(*expectedBalance, true).Times(1)

	balance, ok := proxy.SelfDestruct6780(addr)
	assert.Equal(t, *expectedBalance, balance)
	assert.True(t, ok)
}

func TestRecorderProxy_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	if err != nil {
		t.Fatalf("failed to create record context: %v", err)
	}
	mockDb := state.NewMockStateDB(ctrl)
	proxy := &RecorderProxy{
		db:  mockDb,
		ctx: recordCtx,
	}

	addr := common.Address{0x11}
	expectedRoot := common.Hash{0x12}
	mockDb.EXPECT().GetStorageRoot(addr).Return(expectedRoot).Times(1)

	root := proxy.GetStorageRoot(addr)
	assert.Equal(t, expectedRoot, root)
}
