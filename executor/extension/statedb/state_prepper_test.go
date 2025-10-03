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

package statedb

import (
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"go.uber.org/mock/gomock"
)

func TestStatePrepper_PreparesStateBeforeEachTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	allocA := substatecontext.NewTxContext(&substate.Substate{InputSubstate: substate.WorldState{substatetypes.Address{1}: &substate.Account{}}})
	allocB := substatecontext.NewTxContext(&substate.Substate{InputSubstate: substate.WorldState{substatetypes.Address{2}: &substate.Account{}}})
	ctx := &executor.Context{State: db}

	gomock.InOrder(
		db.EXPECT().PrepareSubstate(allocA.GetInputState(), uint64(5)),
		db.EXPECT().PrepareSubstate(allocB.GetInputState(), uint64(7)),
	)

	prepper := MakeStateDbPrepper()

	// Check and handle error return value for PreTransaction
	err := prepper.PreTransaction(executor.State[txcontext.TxContext]{
		Block: 5,
		Data:  allocA,
	}, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = prepper.PreTransaction(executor.State[txcontext.TxContext]{
		Block: 7,
		Data:  allocB,
	}, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStatePrepper_DoesNotCrashOnMissingStateOrSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)
	ctx := &executor.Context{State: db}

	prepper := MakeStateDbPrepper()
	// Check error return values (if any) for PreTransaction
	err := prepper.PreTransaction(executor.State[txcontext.TxContext]{Block: 5}, nil) // misses both
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = prepper.PreTransaction(executor.State[txcontext.TxContext]{Block: 5}, ctx) // misses the data
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = prepper.PreTransaction(executor.State[txcontext.TxContext]{Block: 5, Data: substatecontext.NewTxContext(&substate.Substate{})}, nil) // misses the state
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
