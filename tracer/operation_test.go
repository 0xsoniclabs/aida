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

package tracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOperationDecoding checks whether number encoding/decoding of operations with their arguments works.
func TestOperationDecoding(t *testing.T) {
	// enumerate whole operation space with arguments
	// and check encoding/decoding whether it is symmetric.
	for op := uint8(0); op < NumOps; op++ {
		for addr := uint8(0); addr < NumClasses; addr++ {
			for key := uint8(0); key < NumClasses; key++ {
				for value := uint8(0); value < NumClasses; value++ {
					// check legality of argument/op combination
					if (opNumArgs[op] == 0 && addr == NoArgID && key == NoArgID && value == NoArgID) ||
						(opNumArgs[op] == 1 && addr != NoArgID && key == NoArgID && value == NoArgID) ||
						(opNumArgs[op] == 2 && addr != NoArgID && key != NoArgID && value == NoArgID) ||
						(opNumArgs[op] == 3 && addr != NoArgID && key != NoArgID && value != NoArgID) {

						// encode to an argument-encoded operation
						argop, err := EncodeArgOp(op, addr, key, value)
						require.NoError(t, err)

						// decode argument-encoded operation
						dop, daddr, dkey, dvalue, err := DecodeArgOp(argop)
						require.NoError(t, err)

						if op != dop || addr != daddr || key != dkey || value != dvalue {
							t.Fatalf("Encoding/decoding failed")
						}
					}
				}
			}
		}
	}
}

// TestOperationOpcode checks the mnemonic encoding/decoding of operations with their argument classes as opcode.
func TestOperationOpcode(t *testing.T) {
	// enumerate whole operation space with arguments
	// and check encoding/decoding whether it is symmetric.
	for op := uint8(0); op < NumOps; op++ {
		for addr := uint8(0); addr < NumClasses; addr++ {
			for key := uint8(0); key < NumClasses; key++ {
				for value := uint8(0); value < NumClasses; value++ {
					// check legality of argument/op combination
					if (opNumArgs[op] == 0 && addr == NoArgID && key == NoArgID && value == NoArgID) ||
						(opNumArgs[op] == 1 && addr != NoArgID && key == NoArgID && value == NoArgID) ||
						(opNumArgs[op] == 2 && addr != NoArgID && key != NoArgID && value == NoArgID) ||
						(opNumArgs[op] == 3 && addr != NoArgID && key != NoArgID && value != NoArgID) {

						// encode to an argument-encoded operation
						opcode, err := EncodeOpcode(op, addr, key, value)
						require.NoErrorf(t, err, "op: %d, addr: %d, key: %d, value: %d", op, addr, key, value)

						// decode argument-encoded operation
						dop, daddr, dkey, dvalue, err := DecodeOpcode(opcode)
						require.NoError(t, err)

						if op != dop || addr != daddr || key != dkey || value != dvalue {
							t.Fatalf("Encoding/decoding failed for %v", opcode)
						}
					}
				}
			}
		}
	}
}

func TestOpMnemo(t *testing.T) {
	for op := range uint8(NumOps) {
		require.Equal(t, OpMnemo(op), opMnemo[op])
	}
}

func TestOpMnemo_OverflowPanicks(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()
	OpMnemo(NumOps)
}
func Test_EncodeArgOp_OverflowError(t *testing.T) {
	_, err := EncodeArgOp(NumOps, 0, 0, 0)
	assert.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")
	_, err = EncodeArgOp(0, NumClasses, 0, 0)
	assert.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")
	_, err = EncodeArgOp(0, 0, NumClasses, 0)
	assert.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")
	_, err = EncodeArgOp(0, 0, 0, NumClasses)
	assert.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")
}
