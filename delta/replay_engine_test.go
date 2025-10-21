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
