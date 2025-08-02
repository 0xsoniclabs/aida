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

package statedb

import (
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEthStateScopeEventEmitter_PreBlockCallsBeginBlockAndBeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	ext := ethStateScopeEventEmitter{}

	db.EXPECT().BeginBlock(uint64(1))
	db.EXPECT().BeginTransaction(uint32(1))

	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: ethtest.CreateTestTransaction(t)}
	ctx := &executor.Context{State: db}
	err := ext.PreBlock(st, ctx)
	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}
}

func TestEthStateScopeEventEmitter_PostBlockCallsEndBlockAndEndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	ext := ethStateScopeEventEmitter{}

	db.EXPECT().EndTransaction()
	db.EXPECT().EndBlock()

	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: ethtest.CreateTestTransaction(t)}
	ctx := &executor.Context{State: db}
	err := ext.PostBlock(st, ctx)
	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}
}

func TestEthStateScopeEventEmitter_MakeEthStateScopeTestEventEmitter(t *testing.T) {
	ext := MakeEthStateScopeTestEventEmitter()
	_, ok := ext.(executor.Extension[txcontext.TxContext])
	assert.True(t, ok)
}

func TestEthStateScopeEventEmitter_PreBlockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	ext := ethStateScopeEventEmitter{}

	db.EXPECT().BeginBlock(uint64(1)).Return(assert.AnError)

	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: ethtest.CreateTestTransaction(t)}
	ctx := &executor.Context{State: db}
	err := ext.PreBlock(st, ctx)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}
func TestEthStateScopeEventEmitter_PostBlockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	ext := ethStateScopeEventEmitter{}

	db.EXPECT().EndTransaction().Return(assert.AnError)

	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: ethtest.CreateTestTransaction(t)}
	ctx := &executor.Context{State: db}
	err := ext.PostBlock(st, ctx)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}
