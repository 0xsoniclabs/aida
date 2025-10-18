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
	"github.com/stretchr/testify/assert"
)

func initSetCode(t *testing.T) (*context.Replay, *SetCode, common.Address, []byte) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	addr := getRandomAddress(t)
	code := make([]byte, 100)
	r.Read(code)

	// create context context
	ctx := context.NewReplay()
	contract := ctx.EncodeContract(addr)

	// create new operation
	op := NewSetCode(contract, code)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != SetCodeID {
		t.Fatalf("wrong ID returned")
	}

	return ctx, op, addr, code
}

// TestSetCodeReadWrite writes a new SetCode object into a buffer, reads from it,
// and checks equality.
func TestSetCodeReadWrite(t *testing.T) {
	_, op1, _, _ := initSetCode(t)
	testOperationReadWrite(t, op1, ReadSetCode)
}

// TestSetCodeDebug creates a new SetCode object and checks its Debug message.
func TestSetCodeDebug(t *testing.T) {
	ctx, op, addr, value := initSetCode(t)
	testOperationDebug(t, ctx, op, fmt.Sprintf("%v%v", addr, value))
}

// TestSetCodeExecute
func TestSetCodeExecute(t *testing.T) {
	ctx, op, addr, code := initSetCode(t)

	// check execution
	mock := NewMockStateDB()
	execute, err := op.Execute(mock, ctx)
	assert.NoError(t, err)
	assert.True(t, execute > 0)

	// check whether methods were correctly called
	expected := []Record{{SetCodeID, []any{addr, code}}}
	mock.compareRecordings(expected, t)
}
