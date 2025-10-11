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
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var testAddr = common.Address{0}

func TestDbLoggerExtension_CorrectClose(t *testing.T) {
	cfg := &utils.Config{}
	ext := MakeDbLogger[any](cfg)

	// start the report thread
	err := ext.PreRun(executor.State[any]{}, nil)
	assert.NoError(t, err)

	// make sure PostRun is not blocking.
	done := make(chan bool)
	go func() {
		err = ext.PostRun(executor.State[any]{}, nil, nil)
		assert.NoError(t, err)
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(time.Second):
		t.Fatalf("PostRun blocked unexpectedly")
	}
}

func TestDbLoggerExtension_NoLoggerIsCreatedIfNotEnabled(t *testing.T) {
	cfg := &utils.Config{}
	ext := MakeDbLogger[any](cfg)
	_, ok := ext.(extension.NilExtension[any])
	assert.True(t, ok)
}

func TestDbLoggerExtension_LoggingHappens(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	fileName := t.TempDir() + "test-log"
	cfg := &utils.Config{}
	cfg.DbLogging = fileName

	ext := makeDbLogger[any](cfg, log)

	ctx := &executor.Context{State: db}

	err := ext.PreRun(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	err = ext.PreTransaction(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	balance := new(uint256.Int).SetUint64(10)

	beginBlock := fmt.Sprintf("BeginBlock, %v", 1)
	beginTransaction := fmt.Sprintf("BeginTransaction, %v", 0)
	getBalance := fmt.Sprintf("GetBalance, %v, %v", testAddr, balance)
	endTransaction := "EndTransaction"
	endBlock := "EndBlock"

	gomock.InOrder(
		log.EXPECT().Debug(beginBlock),
		db.EXPECT().BeginBlock(uint64(1)),
		log.EXPECT().Debug(beginTransaction),
		db.EXPECT().BeginTransaction(uint32(0)),
		db.EXPECT().GetBalance(testAddr).Return(balance),
		log.EXPECT().Debug(getBalance),
		log.EXPECT().Debug(endTransaction),
		db.EXPECT().EndTransaction(),
		log.EXPECT().Debug(endBlock),
		db.EXPECT().EndBlock(),
	)

	err = ctx.State.BeginBlock(1)
	assert.NoError(t, err)

	err = ctx.State.BeginTransaction(0)
	assert.NoError(t, err)

	ctx.State.GetBalance(testAddr)

	err = ctx.State.EndTransaction()
	assert.NoError(t, err)

	err = ctx.State.EndBlock()
	assert.NoError(t, err)

	err = ext.PostRun(executor.State[any]{}, ctx, nil)
	assert.NoError(t, err)

	// signal and await the close
	close(ext.input)
	ext.wg.Wait()

	file, err := os.Open(fileName)
	assert.NoError(t, err)
	defer func() {
		err = file.Close()
		assert.NoError(t, err)
	}()

	fileContent, err := io.ReadAll(file)
	assert.NoError(t, err)

	got := strings.TrimSpace(string(fileContent))
	want := strings.TrimSpace("BeginBlock, 1\nBeginTransaction, 0\nGetBalance, 0x0000000000000000000000000000000000000000, 10\nEndTransaction\nEndBlock")

	assert.Equal(t, want, got)
}

func TestDbLoggerExtension_PreTransactionCreatesNewLoggerProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.DbLogging = t.TempDir() + "test-log"
	cfg.LogLevel = "critical"

	ctx := new(executor.Context)
	ctx.State = db

	ext := MakeDbLogger[any](cfg)

	// ctx.State is not yet a LoggerProxy hence PreTransaction assigns it
	err := ext.PreTransaction(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	_, ok := ctx.State.(*proxy.LoggingStateDb)
	assert.True(t, ok)
}

func TestDbLoggerExtension_PreTransactionDoesNotCreateNewLoggerProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.DbLogging = t.TempDir() + "test-log"
	cfg.LogLevel = "critical"

	ctx := new(executor.Context)
	ctx.State = db

	ext := MakeDbLogger[any](cfg)

	// first call PreTransaction to assign the proxy
	err := ext.PreTransaction(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	// save original state to make sure next call to PreTransaction will not have changed the ctx.State
	originalDb := ctx.State

	// then make sure it is not re-assigned again
	err = ext.PreTransaction(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	assert.Equal(t, originalDb, ctx.State)
}

func TestDbLoggerExtension_PreRunCreatesNewLoggerProxyIfStateIsNotNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.DbLogging = t.TempDir() + "test-log"
	cfg.LogLevel = "critical"

	ctx := new(executor.Context)
	ctx.State = db

	ext := MakeDbLogger[any](cfg)

	err := ext.PreRun(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	_, ok := ctx.State.(*proxy.LoggingStateDb)
	assert.True(t, ok)
}

func TestDbLoggerExtension_PreRunDoesNotCreateNewLoggerProxyIfStateIsNil(t *testing.T) {
	cfg := &utils.Config{}
	cfg.DbLogging = t.TempDir() + "test-log"
	cfg.LogLevel = "critical"

	ctx := new(executor.Context)

	ext := MakeDbLogger[any](cfg)

	err := ext.PreRun(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	_, ok := ctx.State.(*proxy.LoggingStateDb)
	assert.False(t, ok)
}

func TestDbLoggerExtension_StateDbCloseIsWrittenInTheFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	fileName := t.TempDir() + "test-log"
	cfg := &utils.Config{}
	cfg.DbLogging = fileName

	ext := makeDbLogger[any](cfg, log)

	ctx := &executor.Context{State: db}

	err := ext.PreRun(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	err = ext.PreTransaction(executor.State[any]{}, ctx)
	assert.NoError(t, err)

	want := "Close"
	gomock.InOrder(
		db.EXPECT().Close().Return(nil),
		log.EXPECT().Debug(want),
	)

	err = ctx.State.Close()
	assert.NoError(t, err)

	file, err := os.Open(fileName)
	assert.NoError(t, err)
	defer func() {
		err = file.Close()
		assert.NoError(t, err)
	}()

	fileContent, err := io.ReadAll(file)
	assert.NoError(t, err)

	assert.Contains(t, string(fileContent), want)
}
