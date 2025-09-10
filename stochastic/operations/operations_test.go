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

package operations

import (
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/statistics/classifier"
	"github.com/stretchr/testify/assert"
)

// TestOperationDecoding checks whether number encoding/decoding of operations with their arguments works.
func TestOperationDecoding(t *testing.T) {
	// enumerate whole operation space with arguments
	// and check encoding/decoding whether it is symmetric.
	for op := range NumOps {
		for addr := range classifier.NumArgKinds {
			for key := range classifier.NumArgKinds {
				for value := range classifier.NumArgKinds {
					// check legality of argument/op combination
					if (OpNumArgs[op] == 0 && addr == classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
						(OpNumArgs[op] == 1 && addr != classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
						(OpNumArgs[op] == 2 && addr != classifier.NoArgID && key != classifier.NoArgID && value == classifier.NoArgID) ||
						(OpNumArgs[op] == 3 && addr != classifier.NoArgID && key != classifier.NoArgID && value != classifier.NoArgID) {

						// encode to an argument-encoded operation
						argop, err := EncodeArgOp(op, addr, key, value)
						if err != nil {
							t.Fatalf("Encoding failed for %v", argop)
						}

						// decode argument-encoded operation
						dop, daddr, dkey, dvalue, err := DecodeArgOp(argop)
						if err != nil {
							t.Fatalf("Decoding failed for %v %v %v %v", dop, daddr, dkey, dvalue)
						}

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
	for op := range NumOps {
		for addr := range classifier.NumArgKinds {
			for key := range classifier.NumArgKinds {
				for value := range classifier.NumArgKinds {
					// check legality of argument/op combination
					if (OpNumArgs[op] == 0 && addr == classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
						(OpNumArgs[op] == 1 && addr != classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
						(OpNumArgs[op] == 2 && addr != classifier.NoArgID && key != classifier.NoArgID && value == classifier.NoArgID) ||
						(OpNumArgs[op] == 3 && addr != classifier.NoArgID && key != classifier.NoArgID && value != classifier.NoArgID) {

						// encode to an argument-encoded operation
						argop, err := EncodeOpcode(op, addr, key, value)
						if err != nil {
							t.Fatalf("Encoding failed for %v %v %v %v", op, addr, key, value)
						}

						// decode argument-encoded operation
						dop, daddr, dkey, dvalue, err := DecodeOpcode(argop)
						if err != nil {
							t.Fatalf("Decoding failed for %v", argop)
						}
						if op != dop || addr != daddr || key != dkey || value != dvalue {
							t.Fatalf("Encoding/decoding failed for %v", argop)
						}
					}
				}
			}
		}
	}
}

func TestStochastic_OpMnemo(t *testing.T) {
	// case 1
	out := OpMnemo(SnapshotID)
	assert.Equal(t, "SN", out)

	// case 2
	assert.Panics(t, func() {
		OpMnemo(-1)
	})
}

func TestStochastic_checkArgOp(t *testing.T) {
	// case 1
	err := checkArgOp(SnapshotID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
	assert.Nil(t, err)

	// case 2
	err = checkArgOp(-1, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
	assert.NotNil(t, err)

	// case 3
	err = checkArgOp(SnapshotID, -1, classifier.NoArgID, classifier.NoArgID)
	assert.NotNil(t, err)

	// case 4
	err = checkArgOp(SnapshotID, classifier.NoArgID, -1, classifier.NoArgID)
	assert.NotNil(t, err)

	// case 5
	err = checkArgOp(SnapshotID, classifier.NoArgID, classifier.NoArgID, -1)
	assert.NotNil(t, err)
}

func TestStochastic_IsValidArgOp(t *testing.T) {
	// encode to an argument-encoded operation
	argop, err := EncodeArgOp(SetStateID, classifier.PrevArgID, classifier.NewArgID, classifier.NewArgID)
	if err != nil {
		t.Fatalf("Encoding failed")
	}
	valid := IsValidArgOp(argop)
	assert.False(t, valid)

	invalid := IsValidArgOp(-1)
	assert.False(t, invalid)

	invalid = IsValidArgOp(NumArgOps)
	assert.False(t, invalid)
}
