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
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
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
