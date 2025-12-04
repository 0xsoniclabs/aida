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
	"bufio"
	"bytes"
	"errors"
	"strings"
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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDeltaLogger_ProducesDeltaTraceFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLog := logger.NewMockLogger(ctrl)
	mockDB := state.NewMockStateDB(ctrl)

	buf := new(bytes.Buffer)
	sink := NewDeltaLogSink(mockLog, bufio.NewWriter(buf), nil)

	addr := common.HexToAddress("0x1")
	key := common.HexToHash("0x2")
	hash := common.HexToHash("0x3")
	amount := new(uint256.Int).SetUint64(42)
	code := []byte{0xde, 0xad}
	preimage := []byte{0xbe, 0xef}

	proxyDB := NewDeltaLoggerProxy(mockDB, sink)

	expectedAddBalance := "AddBalance, 0x0000000000000000000000000000000000000001, 42, 0, Unspecified, 42"
	expectedSetNonce := "SetNonce, 0x0000000000000000000000000000000000000001, 7, Authorization"
	expectedSetCode := "SetCode, 0x0000000000000000000000000000000000000001, 0xdead"
	expectedCommit := "Commit, true"
	expectedPreimage := "AddPreimage, 0x0000000000000000000000000000000000000000000000000000000000000003, 0xbeef"

	gomock.InOrder(
		mockLog.EXPECT().Debug("BeginBlock, 1"),
		mockDB.EXPECT().BeginBlock(uint64(1)),
		mockLog.EXPECT().Debug(expectedAddBalance),
		mockDB.EXPECT().AddBalance(addr, amount, tracing.BalanceChangeUnspecified),
		mockLog.EXPECT().Debug(expectedSetNonce),
		mockDB.EXPECT().SetNonce(addr, uint64(7), tracing.NonceChangeAuthorization),
		mockLog.EXPECT().Debug("SetState, 0x0000000000000000000000000000000000000001, 0x0000000000000000000000000000000000000000000000000000000000000002, 0x0000000000000000000000000000000000000000000000000000000000000002"),
		mockDB.EXPECT().SetState(addr, key, key),
		mockLog.EXPECT().Debug(expectedSetCode),
		mockDB.EXPECT().SetCode(addr, code, tracing.CodeChangeUnspecified),
		mockLog.EXPECT().Debug(expectedCommit),
		mockDB.EXPECT().Commit(uint64(9), true),
		mockLog.EXPECT().Debug(expectedPreimage),
		mockDB.EXPECT().AddPreimage(hash, preimage),
		mockLog.EXPECT().Debug("EndBlock"),
		mockDB.EXPECT().EndBlock(),
	)

	require.NoError(t, proxyDB.BeginBlock(1))
	proxyDB.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)
	proxyDB.SetNonce(addr, 7, tracing.NonceChangeAuthorization)
	proxyDB.SetState(addr, key, key)
	proxyDB.SetCode(addr, code, tracing.CodeChangeUnspecified)
	_, _ = proxyDB.Commit(9, true)
	proxyDB.AddPreimage(hash, preimage)
	require.NoError(t, proxyDB.EndBlock())
	require.NoError(t, sink.Flush())

	got := strings.TrimSpace(buf.String())
	expected := strings.Join([]string{
		"BeginBlock, 1",
		expectedAddBalance,
		expectedSetNonce,
		"SetState, 0x0000000000000000000000000000000000000001, 0x0000000000000000000000000000000000000000000000000000000000000002, 0x0000000000000000000000000000000000000000000000000000000000000002",
		expectedSetCode,
		expectedCommit,
		expectedPreimage,
		"EndBlock",
	}, "\n")

	require.Equal(t, expected, got)
}

type fakeSyncCloser struct {
	bytes.Buffer
	syncCalled  bool
	closeCalled bool
	syncErr     error
	closeErr    error
	failWrite   bool
}

func (f *fakeSyncCloser) Write(p []byte) (int, error) {
	if f.failWrite {
		return 0, errors.New("write")
	}
	return f.Buffer.Write(p)
}

func (f *fakeSyncCloser) WriteString(s string) (int, error) {
	if f.failWrite {
		return 0, errors.New("write")
	}
	return f.Buffer.WriteString(s)
}

func (f *fakeSyncCloser) Sync() error {
	f.syncCalled = true
	return f.syncErr
}

func (f *fakeSyncCloser) Close() error {
	f.closeCalled = true
	return f.closeErr
}

func TestDeltaLogSink_LogfEdgeCases(t *testing.T) {
	var nilSink *DeltaLogSink
	require.NotPanics(t, func() { nilSink.Logf("ignored") })
	require.NoError(t, nilSink.Flush())
	require.NoError(t, nilSink.Close())

	ctrl := gomock.NewController(t)
	mockLog := logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Debug("only-debug").Times(1)

	sink := &DeltaLogSink{log: mockLog}
	sink.Logf("only-debug\n")
}

func TestDeltaLogSink_FlushAndClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLog := logger.NewMockLogger(ctrl)
	errorLogged := 0
	mockLog.EXPECT().Errorf(gomock.Any(), gomock.Any()).Do(func(string, ...any) {
		errorLogged++
	}).AnyTimes()
	mockLog.EXPECT().Debug("content-fail").Times(1)
	mockLog.EXPECT().Debug("content-ok").Times(1)

	syncCloserFail := &fakeSyncCloser{failWrite: true}
	failSink := NewDeltaLogSink(mockLog, bufio.NewWriterSize(syncCloserFail, 1), syncCloserFail)
	failSink.Logf("content-fail")

	syncCloser := &fakeSyncCloser{}
	sink := NewDeltaLogSink(mockLog, bufio.NewWriterSize(syncCloser, 4), syncCloser)
	sink.Logf("content-ok")
	require.NoError(t, sink.Flush())
	require.Contains(t, syncCloser.String(), "content-ok")

	syncCloser.syncErr = errors.New("sync")
	syncCloser.closeErr = errors.New("close")

	err := sink.Close()
	require.Error(t, err)
	require.ErrorContains(t, err, "sync")
	require.ErrorContains(t, err, "close")
	require.True(t, syncCloser.syncCalled)
	require.True(t, syncCloser.closeCalled)
	require.Greater(t, errorLogged, 0)
	require.Nil(t, sink.writer)
	require.Nil(t, sink.closer)
	require.NoError(t, sink.Close())
}

func TestNewDeltaLoggerProxy_NilSink(t *testing.T) {
	db := &state.MockStateDB{}

	res := NewDeltaLoggerProxy(db, nil)
	require.Same(t, db, res)
}

func TestDeltaLoggingStateDB_DelegatesMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLog := logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Debug(gomock.Any()).AnyTimes()

	mockDB := state.NewMockStateDB(ctrl)
	addr := common.HexToAddress("0x1")
	key := common.HexToHash("0x2")
	hash := common.HexToHash("0x3")
	value := uint256.NewInt(9)

	mockDB.EXPECT().Error().Return(errors.New("err"))
	mockDB.EXPECT().BeginBlock(uint64(1)).Return(nil)
	mockDB.EXPECT().EndBlock().Return(nil)
	mockDB.EXPECT().BeginSyncPeriod(uint64(2))
	mockDB.EXPECT().EndSyncPeriod()
	mockDB.EXPECT().GetHash().Return(hash, nil)
	mockDB.EXPECT().Close().Return(nil)
	mockBulk := state.NewMockBulkLoad(ctrl)
	mockDB.EXPECT().StartBulkLoad(uint64(3)).Return(mockBulk, nil)

	archive := state.NewMockNonCommittableStateDB(ctrl)
	archive.EXPECT().GetHash().Return(hash, nil)
	archive.EXPECT().Release().Return(nil)
	mockDB.EXPECT().GetArchiveState(uint64(4)).Return(archive, nil)
	mockDB.EXPECT().GetArchiveBlockHeight().Return(uint64(5), true, nil)
	mockDB.EXPECT().GetMemoryUsage().Return(&state.MemoryUsage{UsedBytes: 10})
	mockDB.EXPECT().GetShadowDB().Return(nil)
	mockDB.EXPECT().Finalise(true)
	mockDB.EXPECT().IntermediateRoot(true).Return(hash)
	mockDB.EXPECT().Commit(uint64(6), true).Return(hash, nil)
	mockDB.EXPECT().PrepareSubstate(txcontext.WorldState(nil), uint64(7))

	mockDB.EXPECT().CreateAccount(addr)
	mockDB.EXPECT().CreateContract(addr)
	mockDB.EXPECT().Exist(addr).Return(true)
	mockDB.EXPECT().Empty(addr).Return(false)
	mockDB.EXPECT().SelfDestruct(addr).Return(uint256.Int{})
	mockDB.EXPECT().SelfDestruct6780(addr).Return(uint256.Int{}, true)
	mockDB.EXPECT().HasSelfDestructed(addr).Return(true)
	mockDB.EXPECT().GetBalance(addr).Return(value)
	mockDB.EXPECT().AddBalance(addr, (*uint256.Int)(nil), tracing.BalanceChangeUnspecified).Return(uint256.Int{})
	mockDB.EXPECT().SubBalance(addr, value, tracing.BalanceChangeUnspecified).Return(uint256.Int{})
	mockDB.EXPECT().GetNonce(addr).Return(uint64(11))
	mockDB.EXPECT().SetNonce(addr, uint64(12), tracing.NonceChangeAuthorization)
	mockDB.EXPECT().GetCommittedState(addr, key).Return(hash)
	mockDB.EXPECT().GetStateAndCommittedState(addr, key).Return(hash, hash)
	mockDB.EXPECT().GetState(addr, key).Return(hash)
	mockDB.EXPECT().SetState(addr, key, hash).Return(hash)
	mockDB.EXPECT().SetTransientState(addr, key, hash)
	mockDB.EXPECT().GetTransientState(addr, key).Return(hash)
	mockDB.EXPECT().GetCodeHash(addr).Return(hash)
	mockDB.EXPECT().GetCode(addr).Return([]byte{1, 2})
	mockDB.EXPECT().SetCode(addr, gomock.Any(), tracing.CodeChangeUnspecified).Return([]byte{1})
	mockDB.EXPECT().GetCodeSize(addr).Return(2)
	mockDB.EXPECT().Snapshot().Return(1)
	mockDB.EXPECT().RevertToSnapshot(1)
	mockDB.EXPECT().BeginTransaction(uint32(13)).Return(nil)
	mockDB.EXPECT().EndTransaction().Return(nil)
	mockDB.EXPECT().Finalise(false)
	mockDB.EXPECT().AddRefund(uint64(3))
	mockDB.EXPECT().SubRefund(uint64(1))
	mockDB.EXPECT().GetRefund().Return(uint64(2))
	mockDB.EXPECT().Prepare(params.Rules{}, addr, addr, &addr, nil, types.AccessList{})
	mockDB.EXPECT().AddressInAccessList(addr).Return(true)
	mockDB.EXPECT().SlotInAccessList(addr, key).Return(true, false)
	mockDB.EXPECT().AddAddressToAccessList(addr)
	mockDB.EXPECT().AddSlotToAccessList(addr, key)
	mockDB.EXPECT().AddLog(gomock.AssignableToTypeOf(&types.Log{}))
	mockDB.EXPECT().GetLogs(hash, uint64(8), hash, uint64(9)).Return([]*types.Log{{Index: 1}})
	mockDB.EXPECT().PointCache().Return((*utils.PointCache)(nil))
	mockDB.EXPECT().Witness().Return((*stateless.Witness)(nil))
	mockDB.EXPECT().SetTxContext(hash, 14)
	mockDB.EXPECT().GetSubstatePostAlloc().Return(txcontext.WorldState(nil))
	mockDB.EXPECT().AddPreimage(hash, []byte{0xaa})
	mockDB.EXPECT().AccessEvents().Return(&geth.AccessEvents{})
	mockDB.EXPECT().GetStorageRoot(addr).Return(hash)

	proxyDB := NewDeltaLoggerProxy(mockDB, &DeltaLogSink{log: mockLog}).(*DeltaLoggingStateDB)

	require.Error(t, proxyDB.Error())
	require.NoError(t, proxyDB.BeginBlock(1))
	proxyDB.BeginSyncPeriod(2)
	proxyDB.EndSyncPeriod()
	_, err := proxyDB.GetHash()
	require.NoError(t, err)
	require.NotNil(t, proxyDB.GetMemoryUsage())
	require.Nil(t, proxyDB.GetShadowDB())
	_, err = proxyDB.StartBulkLoad(3)
	require.NoError(t, err)

	archiveProxy, err := proxyDB.GetArchiveState(4)
	require.NoError(t, err)
	_, err = archiveProxy.GetHash()
	require.NoError(t, err)
	require.NoError(t, archiveProxy.Release())

	_, _, err = proxyDB.GetArchiveBlockHeight()
	require.NoError(t, err)
	proxyDB.Finalise(true)
	proxyDB.IntermediateRoot(true)
	proxyDB.PrepareSubstate(txcontext.WorldState(nil), 7)
	proxyDB.CreateAccount(addr)
	proxyDB.CreateContract(addr)
	require.True(t, proxyDB.Exist(addr))
	require.False(t, proxyDB.Empty(addr))
	proxyDB.SelfDestruct(addr)
	proxyDB.SelfDestruct6780(addr)
	require.True(t, proxyDB.HasSelfDestructed(addr))
	require.Equal(t, value, proxyDB.GetBalance(addr))
	proxyDB.AddBalance(addr, nil, tracing.BalanceChangeUnspecified)
	proxyDB.SubBalance(addr, value, tracing.BalanceChangeUnspecified)
	require.Equal(t, uint64(11), proxyDB.GetNonce(addr))
	proxyDB.SetNonce(addr, 12, tracing.NonceChangeAuthorization)
	require.Equal(t, hash, proxyDB.GetCommittedState(addr, key))
	proxyDB.GetStateAndCommittedState(addr, key)
	proxyDB.GetState(addr, key)
	proxyDB.SetState(addr, key, hash)
	proxyDB.SetTransientState(addr, key, hash)
	proxyDB.GetTransientState(addr, key)
	proxyDB.GetCodeHash(addr)
	proxyDB.GetCode(addr)
	proxyDB.SetCode(addr, []byte{0x1}, tracing.CodeChangeUnspecified)
	proxyDB.GetCodeSize(addr)
	snap := proxyDB.Snapshot()
	proxyDB.RevertToSnapshot(snap)
	require.NoError(t, proxyDB.BeginTransaction(13))
	require.NoError(t, proxyDB.EndTransaction())
	proxyDB.deltaLoggingVmStateDb.Finalise(false)
	proxyDB.AddRefund(3)
	proxyDB.SubRefund(1)
	require.Equal(t, uint64(2), proxyDB.GetRefund())
	proxyDB.Prepare(params.Rules{}, addr, addr, &addr, nil, types.AccessList{})
	require.True(t, proxyDB.AddressInAccessList(addr))
	proxyDB.SlotInAccessList(addr, key)
	proxyDB.AddAddressToAccessList(addr)
	proxyDB.AddSlotToAccessList(addr, key)
	proxyDB.AddLog(&types.Log{})
	proxyDB.GetLogs(hash, 8, hash, 9)
	proxyDB.PointCache()
	proxyDB.Witness()
	proxyDB.SetTxContext(hash, 14)
	proxyDB.GetSubstatePostAlloc()
	proxyDB.AddPreimage(hash, []byte{0xaa})
	proxyDB.AccessEvents()
	proxyDB.GetStorageRoot(addr)
	require.NoError(t, proxyDB.EndBlock())
	_, err = proxyDB.Commit(6, true)
	require.NoError(t, err)
	require.NoError(t, proxyDB.Close())
}
