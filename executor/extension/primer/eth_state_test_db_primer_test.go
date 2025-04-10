// Copyright 2024 Fantom Foundation
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

package primer

import (
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/mock/gomock"
)

func Test_EthStateTestDbPrimer_PreBlockPriming(t *testing.T) {
	cfg := &utils.Config{}
	ext := ethStateTestDbPrimer{cfg: cfg, log: logger.NewLogger(cfg.LogLevel, "EthStatePrimer")}

	testData := ethtest.CreateTestTransaction(t)
	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: testData}
	ctx := &executor.Context{}

	mockCtrl := gomock.NewController(t)
	mockState := state.NewMockStateDB(mockCtrl)
	mockLoad := state.NewMockBulkLoad(mockCtrl)

	mockState.EXPECT().BeginBlock(uint64(0))
	mockState.EXPECT().BeginTransaction(uint32(0))
	mockState.EXPECT().EndTransaction()
	mockState.EXPECT().EndBlock()
	mockState.EXPECT().StartBulkLoad(uint64(1)).Return(mockLoad, nil)
	testData.GetInputState().ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		mockState.EXPECT().Exist(addr).Return(false)
		mockLoad.EXPECT().CreateAccount(addr)
		mockLoad.EXPECT().SetBalance(addr, acc.GetBalance())
		mockLoad.EXPECT().SetNonce(addr, acc.GetNonce())
		mockLoad.EXPECT().SetCode(addr, acc.GetCode())
	})
	mockLoad.EXPECT().Close()

	ctx.State = mockState

	err := ext.PreBlock(st, ctx)
	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}
}

func Test_EthStateTestDbPrimer_PreBlockPrimingWorksWithPreExistedStateDb(t *testing.T) {
	cfg := &utils.Config{}
	ext := ethStateTestDbPrimer{cfg: cfg, log: logger.NewLogger(cfg.LogLevel, "EthStatePrimer")}

	testData := ethtest.CreateTestTransaction(t)
	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: testData}
	ctx := &executor.Context{}

	mockCtrl := gomock.NewController(t)
	mockState := state.NewMockStateDB(mockCtrl)
	mockLoad := state.NewMockBulkLoad(mockCtrl)

	mockState.EXPECT().BeginBlock(uint64(0))
	mockState.EXPECT().BeginTransaction(uint32(0))
	mockState.EXPECT().EndTransaction()
	mockState.EXPECT().EndBlock()
	mockState.EXPECT().StartBulkLoad(uint64(1)).Return(mockLoad, nil)

	testData.GetInputState().ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		mockState.EXPECT().Exist(addr).Return(true)
		mockLoad.EXPECT().SetBalance(addr, acc.GetBalance())
		mockLoad.EXPECT().SetNonce(addr, acc.GetNonce())
		mockLoad.EXPECT().SetCode(addr, acc.GetCode())
	})

	mockLoad.EXPECT().Close()

	ctx.State = mockState

	err := ext.PreBlock(st, ctx)
	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}
}
