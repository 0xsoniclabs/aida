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

package operation

import (
	"fmt"
	"testing"

	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/ethereum/go-ethereum/common"
)

func initGetStateAndCommittedState(t *testing.T) (*context.Replay, *GetStateAndCommittedState, common.Address, common.Hash) {
	addr := getRandomAddress(t)
	storage := getRandomHash(t)

	// create context context
	ctx := context.NewReplay()
	contract := ctx.EncodeContract(addr)
	sIdx, _ := ctx.EncodeKey(storage)

	// create new operation
	op := NewGetStateAndCommittedState(contract, sIdx)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != GetStateAndCommittedStateID {
		t.Fatalf("wrong ID returned")
	}

	return ctx, op, addr, storage
}

// TestGetStateAndCommittedStateReadWrite writes a new GetState object into a buffer, reads from it,
// and checks equality.
func TestGetStateAndCommittedStateReadWrite(t *testing.T) {
	_, op1, _, _ := initGetStateAndCommittedState(t)
	testOperationReadWrite(t, op1, ReadGetStateAndCommittedState)
}

// TestGetStateAndCommittedStateDebug creates a new GetState object and checks its Debug message.
func TestGetStateAndCommittedStateDebug(t *testing.T) {
	ctx, op, addr, storage := initGetStateAndCommittedState(t)
	testOperationDebug(t, ctx, op, fmt.Sprint(addr, storage))
}

// TestGetStateAndCommittedStateExecute
func TestGetStateAndCommittedStateExecute(t *testing.T) {
	ctx, op, addr, storage := initGetStateAndCommittedState(t)

	// check execution
	mock := NewMockStateDB()
	execute, err := op.Execute(mock, ctx)
	if err != nil {
		t.Fatalf("failed to execute operation; %v", err)
	}
	if execute <= 0 {
		t.Fatalf("expected execution to be > 0; got %v", execute)
	}

	// check whether methods were correctly called
	expected := []Record{{GetCommittedStateID, []any{addr, storage}}}
	mock.compareRecordings(expected, t)
}
