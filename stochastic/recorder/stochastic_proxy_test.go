// Copyright 2025 Fantom Foundation
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

package recorder

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

// TestStochasticProxy_CreateAccount tests the CreateAccount method of StochasticProxy.
func TestStochasticProxy_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().CreateAccount(addr)
	proxy.CreateAccount(addr)
}

// TestStochasticProxy_CreateContract tests the CreateContract method of StochasticProxy.
func TestStochasticProxy_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0xbeef")
	base.EXPECT().CreateContract(addr)
	proxy.CreateContract(addr)
}

// TestStochasticProxy_SubBalance tests the SubBalance method of StochasticProxy.
func TestStochasticProxy_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	amount := uint256.NewInt(100)
	reason := tracing.BalanceChangeUnspecified
	base.EXPECT().SubBalance(addr, amount, reason)
	proxy.SubBalance(addr, amount, reason)
}

// TestStochasticProxy_AddBalance tests the AddBalance method of StochasticProxy.
func TestStochasticProxy_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	amount := uint256.NewInt(100)
	reason := tracing.BalanceChangeUnspecified
	base.EXPECT().AddBalance(addr, amount, reason)
	proxy.AddBalance(addr, amount, reason)
}

// TestStochasticProxy_GetBalance tests the GetBalance method of StochasticProxy.
func TestStochasticProxy_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedBalance := uint256.NewInt(100)
	base.EXPECT().GetBalance(addr).Return(expectedBalance)
	balance := proxy.GetBalance(addr)
	assert.Equal(t, expectedBalance, balance)
}

// TestStochasticProxy_GetNonce tests the GetNonce method of StochasticProxy.
func TestStochasticProxy_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedNonce := uint64(42)
	base.EXPECT().GetNonce(addr).Return(expectedNonce)
	nonce := proxy.GetNonce(addr)
	assert.Equal(t, expectedNonce, nonce)
}

// TestStochasticProxy_SetNonce tests the SetNonce method of StochasticProxy.
func TestStochasticProxy_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	nonce := uint64(42)
	reason := tracing.NonceChangeUnspecified
	base.EXPECT().SetNonce(addr, nonce, reason)
	proxy.SetNonce(addr, nonce, reason)
}

// TestStochasticProxy_GetCodeHash tests the GetCodeHash method of StochasticProxy.
func TestStochasticProxy_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedHash := common.HexToHash("0x5678")
	base.EXPECT().GetCodeHash(addr).Return(expectedHash)
	hash := proxy.GetCodeHash(addr)
	assert.Equal(t, expectedHash, hash)
}

// TestStochasticProxy_GetCode tests the GetCode method of StochasticProxy.
func TestStochasticProxy_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedCode := []byte{0x01, 0x02, 0x03}
	base.EXPECT().GetCode(addr).Return(expectedCode)
	code := proxy.GetCode(addr)
	assert.Equal(t, expectedCode, code)
}

// TestStochasticProxy_SetCode tests the SetCode method of StochasticProxy.
func TestStochasticProxy_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	code := []byte{0x01, 0x02, 0x03}
	base.EXPECT().SetCode(addr, code)
	proxy.SetCode(addr, code)
}

// TestStochasticProxy_GetCodeSize tests the GetCodeSize method of StochasticProxy.
func TestStochasticProxy_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expectedSize := 3
	base.EXPECT().GetCodeSize(addr).Return(expectedSize)
	size := proxy.GetCodeSize(addr)
	assert.Equal(t, expectedSize, size)
}

// TestStochasticProxy_AddRefund tests the AddRefund method of StochasticProxy.
func TestStochasticProxy_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	amount := uint64(100)
	base.EXPECT().AddRefund(amount)
	proxy.AddRefund(amount)
}

// TestStochasticProxy_SubRefund tests the SubRefund method of StochasticProxy.
func TestStochasticProxy_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	amount := uint64(50)
	base.EXPECT().SubRefund(amount)
	proxy.SubRefund(amount)
}

// TestStochasticProxy_GetRefund tests the GetRefund method of StochasticProxy.
func TestStochasticProxy_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedRefund := uint64(100)
	base.EXPECT().GetRefund().Return(expectedRefund)
	refund := proxy.GetRefund()
	assert.Equal(t, expectedRefund, refund)
}

// TestStochasticProxy_GetCommittedState tests the GetCommittedState method of StochasticProxy.
func TestStochasticProxy_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	base.EXPECT().GetCommittedState(addr, key).Return(key)
	out := proxy.GetCommittedState(addr, key)
	assert.Equal(t, key, out)
}

// TestStochasticProxy_GetStateAndCommittedState tests the GetStateAndCommittedState method of StochasticProxy.
func TestStochasticProxy_GetStateAndCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	base.EXPECT().GetStateAndCommittedState(addr, key).Return(key, key)
	out1, out2 := proxy.GetStateAndCommittedState(addr, key)
	assert.Equal(t, key, out1)
	assert.Equal(t, key, out2)
}

// TestStochasticProxy_GetState tests the GetState method of StochasticProxy.
func TestStochasticProxy_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	base.EXPECT().GetState(addr, key).Return(expectedValue)
	value := proxy.GetState(addr, key)
	assert.Equal(t, expectedValue, value)
}

// TestStochasticProxy_GetStorageRoot tests the GetStorageRoot method of StochasticProxy.
func TestStochasticProxy_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	expected := common.HexToHash("0xabcd")
	base.EXPECT().GetStorageRoot(addr).Return(expected)
	out := proxy.GetStorageRoot(addr)
	assert.Equal(t, expected, out)
}

// TestStochasticProxy_SetState	tests the SetState method of StochasticProxy.
func TestStochasticProxy_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	base.EXPECT().SetState(addr, key, value)
	proxy.SetState(addr, key, value)
}

// TestStochasticProxy_SetTransientState tests the SetTransientState method of StochasticProxy.
func TestStochasticProxy_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	base.EXPECT().SetTransientState(addr, key, value)
	proxy.SetTransientState(addr, key, value)
}

// TestStochasticProxy_GetTransientState tests the GetTransientState method of StochasticProxy.
func TestStochasticProxy_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expectedValue := common.HexToHash("0x9abc")
	base.EXPECT().GetTransientState(addr, key).Return(expectedValue)
	value := proxy.GetTransientState(addr, key)
	assert.Equal(t, expectedValue, value)
}

// TestStochasticProxy_SelfDestruct tests the SelfDestruct method of StochasticProxy.
func TestStochasticProxy_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().SelfDestruct(addr)
	out := proxy.SelfDestruct(addr)
	assert.Equal(t, uint256.Int{0x0, 0x0, 0x0, 0x0}, out)
}

// TestStochasticProxy_SelfDestruct6780 tests the SelfDestruct6780 method of StochasticProxy.
func TestStochasticProxy_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().SelfDestruct6780(addr).Return(uint256.Int{}, true)
	val, ok := proxy.SelfDestruct6780(addr)
	assert.Equal(t, uint256.Int{}, val)
	assert.True(t, ok)
}

// TestStochasticProxy_CancelSelfDestruct tests the CancelSelfDestruct method of StochasticProxy.
func TestStochasticProxy_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().HasSelfDestructed(addr).Return(true)
	result := proxy.HasSelfDestructed(addr)
	assert.True(t, result)
}

// TestStochasticProxy_CancelSelfDestruct tests the CancelSelfDestruct method of StochasticProxy.
func TestStochasticProxy_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().Exist(addr).Return(true)
	result := proxy.Exist(addr)
	assert.True(t, result)
}

// TestStochasticProxy_Empty tests the Empty method of StochasticProxy.
func TestStochasticProxy_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().Empty(addr).Return(false)
	result := proxy.Empty(addr)
	assert.False(t, result)
}

// TestStochasticProxy_Prepare tests the Prepare method of StochasticProxy.
func TestStochasticProxy_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	rule := params.Rules{}
	sender := common.HexToAddress("0x1234")
	coinbase := common.HexToAddress("0x5678")
	dest := common.HexToAddress("0xabcd")
	base.EXPECT().Prepare(rule, sender, coinbase, &dest, nil, nil)
	proxy.Prepare(rule, sender, coinbase, &dest, nil, nil)
}

// TestStochasticProxy_AddAddressToAccessList tests the AddAddressToAccessList method of StochasticProxy.
func TestStochasticProxy_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().AddAddressToAccessList(addr)
	proxy.AddAddressToAccessList(addr)
}

// TestStochasticProxy_AddAddressToAccessList tests the AddAddressToAccessList method of StochasticProxy.
func TestStochasticProxy_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	base.EXPECT().AddressInAccessList(addr).Return(true)
	result := proxy.AddressInAccessList(addr)
	assert.True(t, result)
}

// TestStochasticProxy_SlotInAccessList tests the SlotInAccessList method of StochasticProxy.
func TestStochasticProxy_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	base.EXPECT().SlotInAccessList(addr, slot).Return(true, false)
	res1, res2 := proxy.SlotInAccessList(addr, slot)
	assert.True(t, res1)
	assert.False(t, res2)
}

// TestStochasticProxy_AddSlotToAccessList tests the AddSlotToAccessList method of StochasticProxy.
func TestStochasticProxy_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	base.EXPECT().AddSlotToAccessList(addr, slot)
	proxy.AddSlotToAccessList(addr, slot)
}

// TestStochasticProxy_RevertToSnapshot tests the RevertToSnapshot method of StochasticProxy.
func TestStochasticProxy_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	snapshotID := 1
	base.EXPECT().RevertToSnapshot(snapshotID)
	proxy.RevertToSnapshot(snapshotID)
}

// TestStochasticProxy_RevertToSnapshot_WithStack exercises snapshot delta recording.
func TestStochasticProxy_RevertToSnapshot_WithStack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)

	base.EXPECT().Snapshot().Return(10)
	_ = proxy.Snapshot()
	base.EXPECT().Snapshot().Return(11)
	_ = proxy.Snapshot()

	base.EXPECT().RevertToSnapshot(10)
	proxy.RevertToSnapshot(10)
	assert.Equal(t, uint64(1), reg.snapshotFreq[1])
}

// TestStochasticProxy_Snapshot tests the Snapshot method of StochasticProxy.
func TestStochasticProxy_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().Snapshot().Return(1)
	snapshotID := proxy.Snapshot()
	assert.Equal(t, 1, snapshotID)
}

// TestStochasticProxy_Snapshot tests the Snapshot method of StochasticProxy.
func TestStochasticProxy_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	log := &types.Log{
		Address: common.HexToAddress("0x1234"),
		Topics:  []common.Hash{common.HexToHash("0x5678")},
		Data:    []byte{0x01, 0x02, 0x03},
	}
	base.EXPECT().AddLog(log)
	proxy.AddLog(log)
}

// TestStochasticProxy_GetLogs tests the GetLogs method of StochasticProxy.
func TestStochasticProxy_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	hash := common.Hash{0x12}
	blk := uint64(2)
	blkHash := common.Hash{2}
	blkTimestamp := uint64(13)
	base.EXPECT().GetLogs(hash, blk, blkHash, blkTimestamp)
	proxy.GetLogs(hash, blk, blkHash, blkTimestamp)
}

// TestStochasticProxy_PointCache tests the PointCache method of StochasticProxy.
func TestStochasticProxy_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expected := utils.PointCache{}
	base.EXPECT().PointCache().Return(&expected)
	out := proxy.PointCache()
	assert.Equal(t, &expected, out)
}

// TestStochasticProxy_Witness tests the Witness method of StochasticProxy.
func TestStochasticProxy_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedWitness := &stateless.Witness{}
	base.EXPECT().Witness().Return(expectedWitness)
	out := proxy.Witness()
	assert.Equal(t, expectedWitness, out)
}

// TestStochasticProxy_Witness tests the Witness method of StochasticProxy.
func TestStochasticProxy_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	preimage := common.HexToHash("0x1234")
	data := []byte{0x01, 0x02, 0x03}
	base.EXPECT().AddPreimage(preimage, data)
	proxy.AddPreimage(preimage, data)
}

// TestStochasticProxy_Witness tests the Witness method of StochasticProxy.
func TestStochasticProxy_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedEvents := &gethstate.AccessEvents{}
	base.EXPECT().AccessEvents().Return(expectedEvents)
	events := proxy.AccessEvents()
	assert.Equal(t, expectedEvents, events)
}

// TestStochasticProxy_AccessEvents tests the AccessEvents method of StochasticProxy.
func TestStochasticProxy_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	hash := common.HexToHash("0x1234")
	ti := 1
	base.EXPECT().SetTxContext(hash, ti).Return()
	proxy.SetTxContext(hash, ti)
}

// TestStochasticProxy_Finalise tests the Finalise method of StochasticProxy.
func TestStochasticProxy_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().Finalise(true).Return()
	proxy.Finalise(true)
}

// TestStochasticProxy_IntermediateRoot tests the IntermediateRoot method of StochasticProxy.
func TestStochasticProxy_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedRoot := common.HexToHash("0x1234")
	base.EXPECT().IntermediateRoot(true).Return(expectedRoot)
	root := proxy.IntermediateRoot(true)
	assert.Equal(t, expectedRoot, root)
}

// TestStochasticProxy_Commit tests the Commit method of StochasticProxy.
func TestStochasticProxy_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	block := uint64(1)
	expectedRoot := common.HexToHash("0x1234")
	base.EXPECT().Commit(block, true).Return(expectedRoot, nil)
	root, err := proxy.Commit(block, true)
	assert.Equal(t, expectedRoot, root)
	assert.NoError(t, err)
}

// TestStochasticProxy_GetHash tests the GetHash method of StochasticProxy.
func TestStochasticProxy_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedHash := common.HexToHash("0x1234")
	base.EXPECT().GetHash().Return(expectedHash, nil)
	hash, err := proxy.GetHash()
	assert.Equal(t, expectedHash, hash)
	assert.NoError(t, err)
}

// TestStochasticProxy_Error tests the Error method of StochasticProxy.
func TestStochasticProxy_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedError := errors.New("test error")
	base.EXPECT().Error().Return(expectedError)
	err := proxy.Error()
	assert.Equal(t, err, expectedError)
}

// TestStochasticProxy_GetSubstatePostAlloc tests the GetSubstatePostAlloc method of StochasticProxy.
func TestStochasticProxy_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedSubstate := txcontext.NewWorldState(nil)
	base.EXPECT().GetSubstatePostAlloc().Return(expectedSubstate)
	substate := proxy.GetSubstatePostAlloc()
	assert.Equal(t, expectedSubstate, substate)
}

// TestStochasticProxy_PrepareSubstate tests the PrepareSubstate method of StochasticProxy.
func TestStochasticProxy_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	substate := txcontext.NewWorldState(nil)
	base.EXPECT().PrepareSubstate(substate, uint64(1))
	proxy.PrepareSubstate(substate, uint64(1))
}

// TestStochasticProxy_CommitSubstate tests the CommitSubstate method of StochasticProxy.
func TestStochasticProxy_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().BeginTransaction(uint32(32)).Return(nil)
	err := proxy.BeginTransaction(uint32(32))
	assert.NoError(t, err)
}

// Test error path for BeginTransaction
func TestStochasticProxy_BeginTransaction_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().BeginTransaction(uint32(1)).Return(errors.New("boom"))
	err := proxy.BeginTransaction(1)
	assert.Error(t, err)
}

// TestStochasticProxy_EndTransaction tests the EndTransaction method of StochasticProxy.
func TestStochasticProxy_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().EndTransaction().Return(nil)
	err := proxy.EndTransaction()
	assert.NoError(t, err)
}

// Test error path for EndTransaction
func TestStochasticProxy_EndTransaction_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().EndTransaction().Return(errors.New("boom"))
	err := proxy.EndTransaction()
	assert.Error(t, err)
}

// TestStochasticProxy_BeginBlock tests the BeginBlock method of StochasticProxy.
func TestStochasticProxy_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	blockNumber := uint64(1)
	base.EXPECT().BeginBlock(blockNumber).Return(nil)
	err := proxy.BeginBlock(blockNumber)
	assert.NoError(t, err)
}

// TestStochasticProxy_EndBlock tests the EndBlock method of StochasticProxy.
func TestStochasticProxy_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().EndBlock().Return(nil)
	err := proxy.EndBlock()
	assert.NoError(t, err)
}

// TestStochasticProxy_BeginSyncPeriod tests the BeginSyncPeriod method of StochasticProxy.
func TestStochasticProxy_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	blockNumber := uint64(1)
	base.EXPECT().BeginSyncPeriod(blockNumber).Return()
	proxy.BeginSyncPeriod(blockNumber)
}

// TestStochasticProxy_EndSyncPeriod tests the EndSyncPeriod method of StochasticProxy.
func TestStochasticProxy_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().EndSyncPeriod().Return()
	proxy.EndSyncPeriod()
}

// TestStochasticProxy_Close tests the Close method of StochasticProxy.
func TestStochasticProxy_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().Close().Return(nil)
	err := proxy.Close()
	assert.NoError(t, err)
}

// TestStochasticProxy_StartBulkLoad tests the StartBulkLoad method of StochasticProxy.
func TestStochasticProxy_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	assert.Panics(t, func() {
		_, _ = proxy.StartBulkLoad(uint64(1))
	})
}

// TestStochasticProxy_StartBulkLoad tests the StartBulkLoad method of StochasticProxy.
func TestStochasticProxy_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedUsage := &state.MemoryUsage{}
	base.EXPECT().GetMemoryUsage().Return(expectedUsage)
	usage := proxy.GetMemoryUsage()
	assert.Equal(t, expectedUsage, usage)
}

// TestStochasticProxy_GetArchiveState tests the GetArchiveState method of StochasticProxy.
func TestStochasticProxy_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	blockNumber := uint64(10)
	expectedState := state.NewMockNonCommittableStateDB(ctrl)
	base.EXPECT().GetArchiveState(blockNumber).Return(expectedState, nil)
	st, err := proxy.GetArchiveState(blockNumber)
	assert.NoError(t, err)
	assert.Equal(t, expectedState, st)
}

// TestStochasticProxy_GetArchiveState tests the GetArchiveState method of StochasticProxy.
func TestStochasticProxy_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	expectedHeight := uint64(100)
	base.EXPECT().GetArchiveBlockHeight().Return(expectedHeight, false, nil)
	height, _, err := proxy.GetArchiveBlockHeight()
	assert.NoError(t, err)
	assert.Equal(t, expectedHeight, height)
}

// TestStochasticProxy_GetShadowDB tests the GetShadowDB method of StochasticProxy.
func TestStochasticProxy_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	base := state.NewMockStateDB(ctrl)
	reg := NewStats()
	proxy := NewStochasticProxy(base, &reg)
	base.EXPECT().GetShadowDB().Return(base)
	shadowDB := proxy.GetShadowDB()
	assert.Equal(t, base, shadowDB)
}
