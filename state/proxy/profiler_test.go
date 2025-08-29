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
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer/operation"
	"github.com/0xsoniclabs/aida/utils/analytics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProxy_NewProfilerProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	p := NewProfilerProxy(mockDb, mockAnalytics, "info")

	assert.NotNil(t, p)
	assert.Equal(t, mockDb, p.db)
	assert.Equal(t, mockAnalytics, p.anlt)
}
func TestProfilerProxy_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().CreateContract(expectedAddr)
	p.CreateContract(expectedAddr)
}
func TestProfilerProxy_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedBalance := uint256.NewInt(11)
	mockDb.EXPECT().SubBalance(expectedAddr, expectedBalance, tracing.BalanceChangeUnspecified).Return(*expectedBalance)
	balance := p.SubBalance(expectedAddr, expectedBalance, tracing.BalanceChangeUnspecified)
	assert.Equal(t, *expectedBalance, balance)
}
func TestProfilerProxy_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedBalance := uint256.NewInt(11)
	mockDb.EXPECT().AddBalance(expectedAddr, expectedBalance, tracing.BalanceChangeUnspecified).Return(*expectedBalance)
	balance := p.AddBalance(expectedAddr, expectedBalance, tracing.BalanceChangeUnspecified)
	assert.Equal(t, *expectedBalance, balance)
}
func TestProfilerProxy_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedBalance := uint256.NewInt(11)
	mockDb.EXPECT().GetBalance(expectedAddr).Return(expectedBalance)
	balance := p.GetBalance(expectedAddr)
	assert.Equal(t, expectedBalance, balance)
}
func TestProfilerProxy_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedNonce := uint64(42)
	mockDb.EXPECT().GetNonce(expectedAddr).Return(expectedNonce)
	nonce := p.GetNonce(expectedAddr)
	assert.Equal(t, expectedNonce, nonce)
}
func TestProfilerProxy_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedNonce := uint64(42)
	mockDb.EXPECT().SetNonce(expectedAddr, expectedNonce, tracing.NonceChangeUnspecified)
	p.SetNonce(expectedAddr, expectedNonce, tracing.NonceChangeUnspecified)
}
func TestProfilerProxy_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedCodeHash := common.HexToHash("0x5678")
	mockDb.EXPECT().GetCodeHash(expectedAddr).Return(expectedCodeHash)
	codeHash := p.GetCodeHash(expectedAddr)
	assert.Equal(t, expectedCodeHash, codeHash)
}
func TestProfilerProxy_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedCode := []byte{0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00}
	mockDb.EXPECT().GetCode(expectedAddr).Return(expectedCode)
	code := p.GetCode(expectedAddr)
	assert.Equal(t, expectedCode, code)
}
func TestProfilerProxy_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedCode := []byte{0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00}
	mockDb.EXPECT().SetCode(expectedAddr, expectedCode)
	p.SetCode(expectedAddr, expectedCode)
}
func TestProfilerProxy_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedCodeSize := 8
	mockDb.EXPECT().GetCodeSize(expectedAddr).Return(expectedCodeSize)
	codeSize := p.GetCodeSize(expectedAddr)
	assert.Equal(t, expectedCodeSize, codeSize)
}
func TestProfilerProxy_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedRefund := uint64(10)
	mockDb.EXPECT().AddRefund(expectedRefund)
	p.AddRefund(expectedRefund)
}
func TestProfilerProxy_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedRefund := uint64(10)
	mockDb.EXPECT().SubRefund(expectedRefund)
	p.SubRefund(expectedRefund)
}
func TestProfilerProxy_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedRefund := uint64(10)
	mockDb.EXPECT().GetRefund().Return(expectedRefund)
	refund := p.GetRefund()
	assert.Equal(t, expectedRefund, refund)
}
func TestProfilerProxy_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedKeyHash := common.HexToHash("0x5678")
	expectedHash := common.HexToHash("0x1234")
	mockDb.EXPECT().GetCommittedState(expectedAddr, expectedKeyHash).Return(expectedHash)
	h := p.GetCommittedState(expectedAddr, expectedKeyHash)
	assert.Equal(t, expectedHash, h)
}
func TestProfilerProxy_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedKeyHash := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetState(expectedAddr, expectedKeyHash).Return(expectedValue)
	value := p.GetState(expectedAddr, expectedKeyHash)
	assert.Equal(t, expectedValue, value)
}
func TestProfilerProxy_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedKeyHash := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	mockDb.EXPECT().SetState(expectedAddr, expectedKeyHash, expectedValue)
	p.SetState(expectedAddr, expectedKeyHash, expectedValue)
}
func TestProfilerProxy_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedKeyHash := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	mockDb.EXPECT().SetTransientState(expectedAddr, expectedKeyHash, expectedValue)
	p.SetTransientState(expectedAddr, expectedKeyHash, expectedValue)
}
func TestProfilerProxy_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedKeyHash := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetTransientState(expectedAddr, expectedKeyHash).Return(expectedValue)
	value := p.GetTransientState(expectedAddr, expectedKeyHash)
	assert.Equal(t, expectedValue, value)
}
func TestProfilerProxy_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().SelfDestruct(expectedAddr)
	p.SelfDestruct(expectedAddr)
}
func TestProfilerProxy_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().HasSelfDestructed(expectedAddr).Return(true)
	hasDestructed := p.HasSelfDestructed(expectedAddr)
	assert.True(t, hasDestructed)
}
func TestProfilerProxy_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().Exist(expectedAddr).Return(true)
	exists := p.Exist(expectedAddr)
	assert.True(t, exists)
}
func TestProfilerProxy_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().Empty(expectedAddr).Return(true)
	isEmpty := p.Empty(expectedAddr)
	assert.True(t, isEmpty)
}
func TestProfilerProxy_Prepare(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	rules := params.Rules{} // Use zero value or mock as needed
	sender := common.HexToAddress("0x1234")
	coinbase := common.HexToAddress("0x5678")
	dest := &common.Address{}
	precompiles := []common.Address{common.HexToAddress("0x1111")}
	txAccesses := types.AccessList{}

	mockDb.EXPECT().Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
	p.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}
func TestProfilerProxy_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().AddAddressToAccessList(expectedAddr)
	p.AddAddressToAccessList(expectedAddr)
}
func TestProfilerProxy_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().AddressInAccessList(expectedAddr).Return(true)
	inList := p.AddressInAccessList(expectedAddr)
	assert.True(t, inList)
}
func TestProfilerProxy_SlotInAccessList(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedAddr := common.HexToAddress("0x1234")
	expectedSlot := common.HexToHash("0x5678")
	mockDb.EXPECT().SlotInAccessList(expectedAddr, expectedSlot).Return(true, false)
	gotAddressOk, gotSlotOk := p.SlotInAccessList(expectedAddr, expectedSlot)
	assert.True(t, gotAddressOk)
	assert.False(t, gotSlotOk)
}
func TestProfilerProxy_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedAddr := common.HexToAddress("0x1234")
	expectedSlot := common.HexToHash("0x5678")
	mockDb.EXPECT().AddSlotToAccessList(expectedAddr, expectedSlot)
	p.AddSlotToAccessList(expectedAddr, expectedSlot)
}
func TestProfilerProxy_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedSnapshotID := 1
	mockDb.EXPECT().Snapshot().Return(expectedSnapshotID)
	snapshotID := p.Snapshot()
	assert.Equal(t, expectedSnapshotID, snapshotID)
}
func TestProfilerProxy_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedSnapshotID := 1
	mockDb.EXPECT().RevertToSnapshot(expectedSnapshotID)
	p.RevertToSnapshot(expectedSnapshotID)
}
func TestProfilerProxy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	mockDb.EXPECT().Error().Return(nil)
	err := p.Error()
	assert.Nil(t, err)
}
func TestProfilerProxy_do(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	value := 0
	mockOp := func() {
		value = 1
	}
	p.do(operation.CreateAccountID, mockOp)
	assert.Equal(t, value, 1)
}
func TestProfilerProxy_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedTxNumber := uint32(42)
	mockDb.EXPECT().BeginTransaction(expectedTxNumber).Return(nil)
	err := p.BeginTransaction(expectedTxNumber)
	assert.Nil(t, err)
}
func TestProfilerProxy_EndTransaction(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().EndTransaction().Return(nil)
	err := p.EndTransaction()
	assert.Nil(t, err)

}
func TestProfilerProxy_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedBlockNumber := uint64(123)
	mockDb.EXPECT().BeginBlock(expectedBlockNumber).Return(nil)
	err := p.BeginBlock(expectedBlockNumber)
	assert.Nil(t, err)
}

func TestProfilerProxy_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().EndBlock().Return(nil)
	err := p.EndBlock()
	assert.Nil(t, err)
}

func TestProfilerProxy_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedPeriodNumber := uint64(456)
	mockDb.EXPECT().BeginSyncPeriod(expectedPeriodNumber)
	p.BeginSyncPeriod(expectedPeriodNumber)
}

func TestProfilerProxy_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().EndSyncPeriod()
	p.EndSyncPeriod()
}

func TestProfilerProxy_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedHash := common.HexToHash("0x1234")
	mockDb.EXPECT().GetHash().Return(expectedHash, nil)
	hash, err := p.GetHash()
	assert.Equal(t, expectedHash, hash)
	assert.Nil(t, err)
}

func TestProfilerProxy_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedLog := &types.Log{
		Address: common.HexToAddress("0x1234"),
		Topics:  []common.Hash{common.HexToHash("0x5678")},
		Data:    []byte{0x01, 0x02, 0x03},
	}
	mockDb.EXPECT().AddLog(expectedLog)
	p.AddLog(expectedLog)
}

func TestProfilerProxy_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedHash := common.HexToHash("0x1234")
	expectedBlock := uint64(123)
	expectedBlockHash := common.HexToHash("0x5678")
	expectedTimestamp := uint64(9999)
	expectedLogs := []*types.Log{
		{Address: common.HexToAddress("0x1111"), Data: []byte{0x01}},
	}

	mockDb.EXPECT().GetLogs(expectedHash, expectedBlock, expectedBlockHash, expectedTimestamp).Return(expectedLogs)
	logs := p.GetLogs(expectedHash, expectedBlock, expectedBlockHash, expectedTimestamp)
	assert.Equal(t, expectedLogs, logs)
}

func TestProfilerProxy_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().PointCache().Return(nil)
	cache := p.PointCache()
	assert.Nil(t, cache)
}

func TestProfilerProxy_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().Witness().Return(nil)
	witness := p.Witness()
	assert.Nil(t, witness)
}

func TestProfilerProxy_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedHash := common.HexToHash("0x1234")
	expectedImage := []byte{0x01, 0x02, 0x03, 0x04}
	mockDb.EXPECT().AddPreimage(expectedHash, expectedImage)
	p.AddPreimage(expectedHash, expectedImage)
}

func TestProfilerProxy_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().AccessEvents().Return(nil)
	events := p.AccessEvents()
	assert.Nil(t, events)
}

func TestProfilerProxy_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedHash := common.HexToHash("0x1234")
	expectedIndex := 42
	mockDb.EXPECT().SetTxContext(expectedHash, expectedIndex)
	p.SetTxContext(expectedHash, expectedIndex)
}

func TestProfilerProxy_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().Finalise(true)
	p.Finalise(true)
}

func TestProfilerProxy_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedRoot := common.HexToHash("0x1234")
	mockDb.EXPECT().IntermediateRoot(true).Return(expectedRoot)
	root := p.IntermediateRoot(true)
	assert.Equal(t, expectedRoot, root)
}

func TestProfilerProxy_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedBlock := uint64(123)
	expectedHash := common.HexToHash("0x1234")
	mockDb.EXPECT().Commit(expectedBlock, true).Return(expectedHash, nil)
	hash, err := p.Commit(expectedBlock, true)
	assert.Equal(t, expectedHash, hash)
	assert.Nil(t, err)
}

func TestProfilerProxy_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().GetSubstatePostAlloc().Return(nil)
	result := p.GetSubstatePostAlloc()
	assert.Nil(t, result)
}

func TestProfilerProxy_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedBlock := uint64(123)
	mockDb.EXPECT().PrepareSubstate(nil, expectedBlock)
	p.PrepareSubstate(nil, expectedBlock)
}

func TestProfilerProxy_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().Close().Return(nil)
	err := p.Close()
	assert.Nil(t, err)
}
func TestProfilerProxy_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewMockLogger(ctrl)
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	mockLogger.EXPECT().Fatal(gomock.Any())
	expectedBlock := uint64(123)
	bulkLoad, err := p.StartBulkLoad(expectedBlock)
	assert.Nil(t, bulkLoad)
	assert.Nil(t, err)
}
func TestProfilerProxy_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	expectedBlock := uint64(123)
	archiveState, err := p.GetArchiveState(expectedBlock)
	assert.Nil(t, archiveState)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "archive states are not (yet) supported")
}
func TestProfilerProxy_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	height, exists, err := p.GetArchiveBlockHeight()
	assert.Equal(t, uint64(0), height)
	assert.False(t, exists)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "archive states are not (yet) supported")
}
func TestProfilerProxy_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().GetMemoryUsage().Return(nil)
	usage := p.GetMemoryUsage()
	assert.Nil(t, usage)
}
func TestProfilerProxy_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}

	mockDb.EXPECT().GetShadowDB().Return(nil)
	shadowDB := p.GetShadowDB()
	assert.Nil(t, shadowDB)
}
func TestProfilerProxy_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().CreateContract(expectedAddr)
	p.CreateContract(expectedAddr)
}
func TestProfilerProxy_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	mockDb.EXPECT().SelfDestruct6780(expectedAddr)
	p.SelfDestruct6780(expectedAddr)
}
func TestProfilerProxy_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockAnalytics := analytics.NewIncrementalAnalytics(operation.NumOperations)
	mockLogger := logger.NewLogger("info", "test")
	p := &ProfilerProxy{
		db:   mockDb,
		anlt: mockAnalytics,
		log:  mockLogger,
	}
	expectedAddr := common.HexToAddress("0x1234")
	expectedRoot := common.HexToHash("0x5678")
	mockDb.EXPECT().GetStorageRoot(expectedAddr).Return(expectedRoot)
	root := p.GetStorageRoot(expectedAddr)
	assert.Equal(t, expectedRoot, root)
}
