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
	"math/rand"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/tracer/context"
)

func initBeginSyncPeriod(t *testing.T) (*context.Replay, *BeginSyncPeriod) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := rng.Uint64()

	// create context context
	ctx := context.NewReplay()

	// create new operation
	op := NewBeginSyncPeriod(num)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != BeginSyncPeriodID {
		t.Fatalf("wrong ID returned")
	}

	return ctx, op
}

// TestBeginSyncPeriodReadWrite writes a new BeginSyncPeriod object into a buffer, reads from it,
// and checks equality.
func TestBeginSyncPeriodReadWrite(t *testing.T) {
	_, op1 := initBeginSyncPeriod(t)
	testOperationReadWrite(t, op1, ReadBeginSyncPeriod)
}

// TestBeginSyncPeriodDebug creates a new BeginSyncPeriod object and checks its Debug message.
func TestBeginSyncPeriodDebug(t *testing.T) {
	ctx, op := initBeginSyncPeriod(t)
	testOperationDebug(t, ctx, op, fmt.Sprintf("%v", op.SyncPeriodNumber))
}

// TestBeginSyncPeriodExecute
func TestBeginSyncPeriodExecute(t *testing.T) {
	ctx, op := initBeginSyncPeriod(t)

	// check execution
	mock := NewMockStateDB()
	execute, err := op.Execute(mock, ctx)
	if err != nil {
		t.Fatalf("failed to execute operation; %v", err)
	}
	if execute <= 0 {
		t.Fatalf("execution time is not positive")
	}
	// check whether methods were correctly called
	expected := []Record{{BeginSyncPeriodID, []any{op.SyncPeriodNumber}}}
	mock.compareRecordings(expected, t)
}
