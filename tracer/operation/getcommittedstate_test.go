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

package operation

import (
	"fmt"
	"testing"

	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/ethereum/go-ethereum/common"
)

func initGetCommittedState(t *testing.T) (*context.Replay, *GetCommittedState, common.Address, common.Hash) {
	addr := getRandomAddress(t)
	storage := getRandomHash(t)

	// create context context
	ctx := context.NewReplay()
	contract := ctx.EncodeContract(addr)
	sIdx, _ := ctx.EncodeKey(storage)

	// create new operation
	op := NewGetCommittedState(contract, sIdx)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != GetCommittedStateID {
		t.Fatalf("wrong ID returned")
	}
	return ctx, op, addr, storage
}

// TestGetCommittedStateReadWrite writes a new GetCommittedState object into a buffer, reads from it,
// and checks equality.
func TestGetCommittedStateReadWrite(t *testing.T) {
	_, op1, _, _ := initGetCommittedState(t)
	testOperationReadWrite(t, op1, ReadGetCommittedState)
}

// TestGetCommittedStateDebug creates a new GetCommittedState object and checks its Debug message.
func TestGetCommittedStateDebug(t *testing.T) {
	ctx, op, addr, storage := initGetCommittedState(t)
	testOperationDebug(t, ctx, op, fmt.Sprint(addr, storage))
}

// TestGetCommittedStateExecute
func TestGetCommittedStateExecute(t *testing.T) {
	ctx, op, addr, storage := initGetCommittedState(t)

	// check execution
	mock := NewMockStateDB()
	op.Execute(mock, ctx)

	// check whether methods were correctly called
	expected := []Record{{GetCommittedStateID, []any{addr, storage}}}
	mock.compareRecordings(expected, t)
}
