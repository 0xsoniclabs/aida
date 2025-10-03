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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
)

func initSetNonce(t *testing.T) (*context.Replay, *SetNonce, common.Address, uint64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	addr := getRandomAddress(t)
	nonce := r.Uint64()

	// create context context
	ctx := context.NewReplay()
	contract := ctx.EncodeContract(addr)

	// create new operation
	op := NewSetNonce(contract, nonce, tracing.NonceChangeUnspecified)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != SetNonceID {
		t.Fatalf("wrong ID returned")
	}

	return ctx, op, addr, nonce
}

// TestSetNonceReadWrite writes a new SetNonce object into a buffer, reads from it,
// and checks equality.
func TestSetNonceReadWrite(t *testing.T) {
	_, op1, _, _ := initSetNonce(t)
	testOperationReadWrite(t, op1, ReadSetNonce)
}

// TestSetNonceDebug creates a new SetNonce object and checks its Debug message.
func TestSetNonceDebug(t *testing.T) {
	ctx, op, addr, value := initSetNonce(t)
	testOperationDebug(t, ctx, op, fmt.Sprint(addr, value, tracing.NonceChangeUnspecified))
}

// TestSetNonceExecute
func TestSetNonceExecute(t *testing.T) {
	ctx, op, addr, nonce := initSetNonce(t)

	// check execution
	mock := NewMockStateDB()
	op.Execute(mock, ctx)

	// check whether methods were correctly called
	expected := []Record{{SetNonceID, []any{addr, nonce, tracing.NonceChangeUnspecified}}}
	mock.compareRecordings(expected, t)
}
