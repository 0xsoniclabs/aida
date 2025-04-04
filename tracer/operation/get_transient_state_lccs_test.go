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
	"math/rand"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/ethereum/go-ethereum/common"
)

func initGetTransientStateLccs(t *testing.T) (*context.Replay, *GetTransientStateLccs, common.Address, common.Hash, common.Hash) {
	rand.Seed(time.Now().UnixNano())
	pos := 0

	// create context context
	ctx := context.NewReplay()

	// create new operation
	op := NewGetTransientStateLccs(pos)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != GetTransientStateLccsID {
		t.Fatalf("wrong ID returned")
	}

	addr := getRandomAddress(t)
	ctx.EncodeContract(addr)

	storage := getRandomHash(t)
	ctx.EncodeKey(storage)

	storage2 := getRandomHash(t)

	return ctx, op, addr, storage, storage2
}

// TestGetTransientStateLccsReadWrite writes a new GetTransientStateLccs object into a buffer, reads from it,
// and checks equality.
func TestGetTransientStateLccsReadWrite(t *testing.T) {
	_, op1, _, _, _ := initGetTransientStateLccs(t)
	testOperationReadWrite(t, op1, ReadGetTransientStateLccs)
}

// TestGetTransientStateLccsDebug creates a new GetTransientStateLccs object and checks its Debug message.
func TestGetTransientStateLccsDebug(t *testing.T) {
	ctx, op, addr, storage, _ := initGetTransientStateLccs(t)
	testOperationDebug(t, ctx, op, fmt.Sprint(addr, storage))
}

// TestGetTransientStateLccsExecute
func TestGetTransientStateLccsExecute(t *testing.T) {
	ctx, op, addr, storage, storage2 := initGetTransientStateLccs(t)

	// check execution
	mock := NewMockStateDB()
	op.Execute(mock, ctx)

	ctx.EncodeKey(storage2)

	op.Execute(mock, ctx)

	// check whether methods were correctly called
	expected := []Record{{GetStateID, []any{addr, storage}}, {GetStateID, []any{addr, storage2}}}
	mock.compareRecordings(expected, t)
}
