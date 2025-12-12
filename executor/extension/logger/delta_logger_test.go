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

package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMakeDeltaLogger_NoOpWhenNotEnabled(t *testing.T) {
	cfg := &utils.Config{}
	ext := MakeDeltaLogger[any](cfg)
	_, ok := ext.(extension.NilExtension[any])
	require.True(t, ok)
}

func TestDeltaLogger_WrapsAndWritesTrace(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	tracePath := filepath.Join(t.TempDir(), "delta.log")
	cfg := &utils.Config{
		DeltaLogging: tracePath,
		LogLevel:     "critical",
	}

	ext := makeDeltaLogger[any](cfg, log)
	ctx := &executor.Context{State: db}

	// PreRun should create sink and wrap existing state
	require.NoError(t, ext.PreRun(executor.State[any]{}, ctx))
	require.NoError(t, ext.PreTransaction(executor.State[any]{}, ctx))

	addr := common.HexToAddress("0x1")
	amount := uint256.NewInt(7)

	expectedBegin := "BeginBlock, 1"
	expectedBalance := "AddBalance, 0x0000000000000000000000000000000000000001, 7, 0, Unspecified, 7"
	expectedEnd := "EndBlock"

	gomock.InOrder(
		log.EXPECT().Debug(expectedBegin),
		db.EXPECT().BeginBlock(uint64(1)),
		log.EXPECT().Debug(expectedBalance),
		db.EXPECT().AddBalance(addr, amount, tracing.BalanceChangeUnspecified),
		log.EXPECT().Debug(expectedEnd),
		db.EXPECT().EndBlock(),
	)

	require.NoError(t, ctx.State.BeginBlock(1))
	_ = ctx.State.AddBalance(addr, amount, tracing.BalanceChangeUnspecified)
	require.NoError(t, ctx.State.EndBlock())

	require.NoError(t, ext.PostRun(executor.State[any]{}, ctx, nil))

	content, err := os.ReadFile(tracePath)
	require.NoError(t, err)

	got := strings.TrimSpace(string(content))
	require.Equal(t, strings.Join([]string{
		expectedBegin,
		expectedBalance,
		expectedEnd,
	}, "\n"), got)
}

func TestDeltaLogger_PreTransactionNoDoubleWrap(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{
		DeltaLogging: filepath.Join(t.TempDir(), "delta.log"),
		LogLevel:     "critical",
	}

	ext := MakeDeltaLogger[any](cfg)
	ctx := &executor.Context{State: db}

	require.NoError(t, ext.PreRun(executor.State[any]{}, ctx))
	require.NoError(t, ext.PreTransaction(executor.State[any]{}, ctx))
	original := ctx.State

	require.NoError(t, ext.PreTransaction(executor.State[any]{}, ctx))
	require.Equal(t, original, ctx.State)
}
