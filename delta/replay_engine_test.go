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

package delta

import (
	"context"
	"testing"

	"github.com/0xsoniclabs/aida/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestStateReplayer_CreateAccount(t *testing.T) {
	addr := common.HexToAddress("0x1")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "CreateAccount", Args: []string{addr.Hex()}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
	require.True(t, db.Exist(addr), "expected account to exist after replay")
	require.Equal(t, []common.Address{addr}, db.createAccountCalls, "unexpected create account calls")
}

func TestStateReplayer_AddBalance(t *testing.T) {
	addr := common.HexToAddress("0x2")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"5"}},
		{Kind: "CreateAccount", Args: []string{addr.Hex()}},
		{
			Kind: "AddBalance",
			Args: []string{
				addr.Hex(),
				"42",
				"0",
				tracing.BalanceChangeTransfer.String(),
				"42",
			},
		},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
	require.Len(t, db.addBalanceCalls, 1, "expected one AddBalance call")

	call := db.addBalanceCalls[0]
	require.Equal(t, addr, call.addr)
	require.Equal(t, tracing.BalanceChangeTransfer, call.reason)
	require.Equal(t, uint64(42), call.value.Uint64())

	balance := db.GetBalance(addr)
	require.Equal(t, uint64(42), balance.Uint64(), "unexpected final balance")
}

func TestStateReplayer_CommitUsesCurrentBlock(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"7"}},
		{Kind: "Commit", Args: []string{"true"}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
	require.Len(t, db.commitCalls, 1, "expected one Commit call")
	call := db.commitCalls[0]
	require.Equal(t, uint64(7), call.block, "commit should use current block number")
	require.True(t, call.deleteEmpty, "commit should pass deleteEmpty flag from trace")
}

func TestStateReplayer_UnsupportedOperation(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)
	op := TraceOp{Kind: "Unknown"}

	err := replayer.Execute(context.Background(), []TraceOp{op})
	require.Error(t, err, "unsupported operation should return an error")
}

type trackingStateDB struct {
	state.StateDB

	createAccountCalls []common.Address
	addBalanceCalls    []balanceCall
	commitCalls        []commitCall
	beginBlocks        []uint64
}

type balanceCall struct {
	addr   common.Address
	value  uint256.Int
	reason tracing.BalanceChangeReason
}

type commitCall struct {
	block       uint64
	deleteEmpty bool
}

func newTrackingStateDB(t *testing.T) *trackingStateDB {
	t.Helper()
	dir := t.TempDir()
	inner, err := state.MakeGethStateDB(dir, "", types.EmptyRootHash, false, nil)
	require.NoError(t, err)

	db := &trackingStateDB{StateDB: inner}
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	return db
}

func (t *trackingStateDB) CreateAccount(addr common.Address) {
	t.createAccountCalls = append(t.createAccountCalls, addr)
	t.StateDB.CreateAccount(addr)
}

func (t *trackingStateDB) AddBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	t.addBalanceCalls = append(t.addBalanceCalls, balanceCall{
		addr:   addr,
		value:  *value.Clone(),
		reason: reason,
	})
	return t.StateDB.AddBalance(addr, value, reason)
}

func (t *trackingStateDB) BeginBlock(block uint64) error {
	t.beginBlocks = append(t.beginBlocks, block)
	return t.StateDB.BeginBlock(block)
}

func (t *trackingStateDB) Commit(block uint64, deleteEmpty bool) (common.Hash, error) {
	t.commitCalls = append(t.commitCalls, commitCall{
		block:       block,
		deleteEmpty: deleteEmpty,
	})
	return t.StateDB.Commit(block, deleteEmpty)
}

func TestStateReplayer_SetState(t *testing.T) {
	addr := common.HexToAddress("0x1")
	key := common.HexToHash("0x2")
	value := common.HexToHash("0x3")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "CreateAccount", Args: []string{addr.Hex()}},
		{Kind: "SetState", Args: []string{addr.Hex(), key.Hex(), value.Hex()}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
	require.Equal(t, value, db.GetState(addr, key))
}

func TestStateReplayer_GetState(t *testing.T) {
	addr := common.HexToAddress("0x1")
	key := common.HexToHash("0x2")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "GetState", Args: []string{addr.Hex(), key.Hex()}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_SetNonce(t *testing.T) {
	addr := common.HexToAddress("0x1")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "CreateAccount", Args: []string{addr.Hex()}},
		{Kind: "SetNonce", Args: []string{addr.Hex(), "42", "NonceChangeUnspecified"}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
	require.Equal(t, uint64(42), db.GetNonce(addr))
}

func TestStateReplayer_GetNonce(t *testing.T) {
	addr := common.HexToAddress("0x1")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "GetNonce", Args: []string{addr.Hex()}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_BeginEndTransaction(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "BeginTransaction", Args: []string{"0"}},
		{Kind: "EndTransaction"},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_Snapshot(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "Snapshot"},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_GetHash(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "GetHash"},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_GetArchiveBlockHeight(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "GetArchiveBlockHeight"},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_SubBalance(t *testing.T) {
	addr := common.HexToAddress("0x1")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "CreateAccount", Args: []string{addr.Hex()}},
		{Kind: "AddBalance", Args: []string{addr.Hex(), "100", "0", "BalanceChangeUnspecified", "100"}},
		{Kind: "SubBalance", Args: []string{addr.Hex(), "50", "0", "BalanceChangeUnspecified", "50"}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
	balance := db.GetBalance(addr)
	require.Equal(t, uint64(50), balance.Uint64())
}

func TestStateReplayer_GetBalance(t *testing.T) {
	addr := common.HexToAddress("0x1")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "GetBalance", Args: []string{addr.Hex()}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_Exist(t *testing.T) {
	addr := common.HexToAddress("0x1")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "Exist", Args: []string{addr.Hex()}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_Empty(t *testing.T) {
	addr := common.HexToAddress("0x1")
	db := newTrackingStateDB(t)

	replayer := NewStateReplayer(db)
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "Empty", Args: []string{addr.Hex()}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_EndBlock(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "EndBlock"},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_IntermediateRoot(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "IntermediateRoot", Args: []string{"true"}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_Finalise(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "Finalise", Args: []string{"false"}},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_NoOps(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "AddLog"},
		{Kind: "Prepare"},
		{Kind: "PrepareSubstate"},
		{Kind: "Close"},
		{Kind: "Error"},
		{Kind: "Release"},
	}

	require.NoError(t, replayer.Execute(context.Background(), ops))
}

func TestStateReplayer_UnsupportedOperations(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	unsupportedOps := []string{
		"GetCodeHashLc",
		"GetCodeHashLcS",
		"GetStateLccs",
		"GetStateLc",
		"GetStateLcls",
	}

	for _, opKind := range unsupportedOps {
		ops := []TraceOp{
			{Kind: "BeginBlock", Args: []string{"1"}},
			{Kind: opKind},
		}

		err := replayer.Execute(context.Background(), ops)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not supported in logger traces")
	}
}

func TestStateReplayer_BulkOperation(t *testing.T) {
	db := newTrackingStateDB(t)
	replayer := NewStateReplayer(db)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "Bulk"},
	}

	err := replayer.Execute(context.Background(), ops)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bulk operations are not supported")
}
