package main

import (
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"math/big"
	"testing"
	"time"
)

func TestAidaSdbRecordAndReplay_AllCalls(t *testing.T) {
	file := t.TempDir() + "/test_trace"
	writer, err := tracer.NewFileWriter(file)
	require.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockState := state.NewMockStateDB(ctrl)
	argCtx, err := tracer.NewContext(writer, 0, 100)
	require.NoError(t, err)
	proxy := proxy.NewTracerProxy(mockState, argCtx)

	addr := common.Address{0x1}
	key := common.Hash{0x2}
	val := common.Hash{0x3}
	balance := uint256.NewInt(123)
	reason := tracing.BalanceChangeTransfer
	nonce := uint64(42)
	code := []byte{0xAA, 0xBB}
	rules := params.Rules{ChainID: big.NewInt(146), IsPrague: true}
	accessList := types.AccessList{{Address: addr, StorageKeys: []common.Hash{key}}}
	logEntry := &types.Log{Address: addr, BlockNumber: 1}
	image := []byte{0xCC, 0xDD}
	precompiles := []common.Address{addr}
	block := uint64(100)
	tx := uint32(33)
	snapshot := 1
	txIndex := 2
	boolVal := true
	ws := txcontext.NewWorldState(map[common.Address]txcontext.Account{addr: nil})

	mockState.EXPECT().CreateAccount(addr).Times(2)
	mockState.EXPECT().SubBalance(addr, balance, reason).Times(2)
	mockState.EXPECT().AddBalance(addr, balance, reason).Times(2)
	mockState.EXPECT().GetBalance(addr).Times(2)
	mockState.EXPECT().GetNonce(addr).Times(2)
	mockState.EXPECT().SetNonce(addr, nonce, tracing.NonceChangeNewContract).Times(2)
	mockState.EXPECT().GetCodeHash(addr).Times(2)
	mockState.EXPECT().GetCode(addr).Times(2)
	mockState.EXPECT().SetCode(addr, code).Times(2)
	mockState.EXPECT().GetCodeSize(addr).Times(2)
	mockState.EXPECT().AddRefund(uint64(100)).Times(2)
	mockState.EXPECT().SubRefund(uint64(50)).Times(2)
	mockState.EXPECT().GetRefund().Times(2)
	mockState.EXPECT().GetCommittedState(addr, key).Times(2)
	mockState.EXPECT().GetState(addr, key).Times(2)
	mockState.EXPECT().SetState(addr, key, val).Times(2)
	mockState.EXPECT().SetTransientState(addr, key, val).Times(2)
	mockState.EXPECT().GetTransientState(addr, key).Times(2)
	mockState.EXPECT().SelfDestruct(addr).Times(2)
	mockState.EXPECT().HasSelfDestructed(addr).Times(2)
	mockState.EXPECT().Exist(addr).Times(2)
	mockState.EXPECT().Empty(addr).Times(2)
	mockState.EXPECT().Prepare(rules, addr, addr, nil, precompiles, accessList).Times(2)
	mockState.EXPECT().AddAddressToAccessList(addr).Times(2)
	mockState.EXPECT().AddressInAccessList(addr).Times(2)
	mockState.EXPECT().SlotInAccessList(addr, key).Times(2)
	mockState.EXPECT().AddSlotToAccessList(addr, key).Times(2)
	mockState.EXPECT().RevertToSnapshot(snapshot).Times(2)
	mockState.EXPECT().Snapshot().Times(2)
	mockState.EXPECT().AddLog(logEntry).Times(2)
	mockState.EXPECT().GetLogs(key, block, key, block).Times(2)
	mockState.EXPECT().PointCache().Times(2)
	mockState.EXPECT().Witness().Times(2)
	mockState.EXPECT().AddPreimage(key, image).Times(2)
	mockState.EXPECT().AccessEvents().Times(2)
	mockState.EXPECT().SetTxContext(key, txIndex).Times(2)
	mockState.EXPECT().Finalise(boolVal).Times(2)
	mockState.EXPECT().IntermediateRoot(boolVal).Times(2)
	mockState.EXPECT().Commit(block, boolVal).Times(2)
	mockState.EXPECT().GetHash().Times(2)
	mockState.EXPECT().GetSubstatePostAlloc().Times(2)
	mockState.EXPECT().PrepareSubstate(ws, block).Times(2)
	mockState.EXPECT().BeginTransaction(tx).Times(2)
	mockState.EXPECT().EndTransaction().Times(2)
	mockState.EXPECT().BeginBlock(block).Times(2)
	mockState.EXPECT().EndBlock().Times(2)
	mockState.EXPECT().BeginSyncPeriod(block).Times(2)
	mockState.EXPECT().EndSyncPeriod().Times(2)
	mockState.EXPECT().CreateContract(addr).Times(2)
	mockState.EXPECT().SelfDestruct6780(addr).Times(2)
	mockState.EXPECT().GetStorageRoot(addr).Times(2)
	mockState.EXPECT().GetArchiveState(block).Times(2)
	mockState.EXPECT().GetArchiveBlockHeight().Times(2)
	mockState.EXPECT().Error()
	mockState.EXPECT().Close().Times(2)

	// Call every proxy method
	err = proxy.BeginBlock(block)
	require.NoError(t, err)
	err = proxy.BeginTransaction(tx)
	require.NoError(t, err)
	proxy.CreateAccount(addr)
	proxy.SubBalance(addr, balance, reason)
	proxy.AddBalance(addr, balance, reason)
	proxy.GetBalance(addr)
	proxy.GetNonce(addr)
	proxy.SetNonce(addr, nonce, tracing.NonceChangeNewContract)
	proxy.GetCodeHash(addr)
	proxy.GetCode(addr)
	proxy.SetCode(addr, code)
	proxy.GetCodeSize(addr)
	proxy.AddRefund(100)
	proxy.SubRefund(50)
	proxy.GetRefund()
	proxy.GetCommittedState(addr, key)
	proxy.GetState(addr, key)
	proxy.SetState(addr, key, val)
	proxy.SetTransientState(addr, key, val)
	proxy.GetTransientState(addr, key)
	proxy.SelfDestruct(addr)
	proxy.HasSelfDestructed(addr)
	proxy.Exist(addr)
	proxy.Empty(addr)
	proxy.Prepare(rules, addr, addr, nil, precompiles, accessList)
	proxy.AddAddressToAccessList(addr)
	proxy.AddressInAccessList(addr)
	proxy.SlotInAccessList(addr, key)
	proxy.AddSlotToAccessList(addr, key)
	proxy.RevertToSnapshot(snapshot)
	proxy.Snapshot()
	proxy.AddLog(logEntry)
	proxy.GetLogs(key, block, key, block)
	proxy.PointCache()
	proxy.Witness()
	proxy.AddPreimage(key, image)
	proxy.AccessEvents()
	proxy.SetTxContext(key, txIndex)
	proxy.Finalise(boolVal)
	proxy.IntermediateRoot(boolVal)
	_, err = proxy.Commit(block, boolVal)
	require.NoError(t, err)
	proxy.GetHash()
	proxy.GetSubstatePostAlloc()
	proxy.PrepareSubstate(ws, block)

	proxy.BeginSyncPeriod(block)
	proxy.EndSyncPeriod()

	proxy.CreateContract(addr)
	proxy.SelfDestruct6780(addr)
	proxy.GetStorageRoot(addr)
	_, err = proxy.GetArchiveState(block)
	require.NoError(t, err)
	_, _, err = proxy.GetArchiveBlockHeight()
	require.NoError(t, err)
	err = proxy.Error()
	require.NoError(t, err)
	err = proxy.EndTransaction()
	require.NoError(t, err)
	err = proxy.EndBlock()
	require.NoError(t, err)

	// Close must be last
	err = proxy.Close()
	require.NoError(t, err)

	reader, first, last, err := tracer.NewFileReader(file)
	require.NoError(t, err)
	require.Equal(t, first, uint64(0))
	require.Equal(t, last, uint64(100))

	tp := &traceProcessor{}
	ctx := &executor.Context{
		State: mockState,
	}

	provider := executor.NewTraceProvider(reader)

	// provider.Run might deadlock - we must ensure the test fails if that happens
	done := make(chan struct{}, 1)
	go func() {
		select {
		case <-done:
			return
		case <-time.After(10 * time.Second):
			t.Fail()
		}
	}()

	err = provider.Run(0, 101, func(info executor.TransactionInfo[tracer.Operation]) error {
		err = tp.Process(executor.State[tracer.Operation]{
			Block:       info.Block,
			Transaction: info.Transaction,
			Data:        info.Data,
		}, ctx)
		assert.NoErrorf(t, err, "%s failed", tracer.OpText[info.Data.Op])
		return nil
	})
	close(done)
	require.NoError(t, err)
}
