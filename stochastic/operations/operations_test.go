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

package operations

import (
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOperationDecoding checks whether number encoding/decoding of operations with their arguments works.
func TestOperationDecoding(t *testing.T) {
	// enumerate whole operation space with arguments
	// and check encoding/decoding whether it is symmetric.
	for op := range NumOps {
		for addr := range stochastic.NumArgKinds {
			for key := range stochastic.NumArgKinds {
				for value := range stochastic.NumArgKinds {
					// check legality of argument/op combination
					if (OpNumArgs[op] == 0 && addr == stochastic.NoArgID && key == stochastic.NoArgID && value == stochastic.NoArgID) ||
						(OpNumArgs[op] == 1 && addr != stochastic.NoArgID && key == stochastic.NoArgID && value == stochastic.NoArgID) ||
						(OpNumArgs[op] == 2 && addr != stochastic.NoArgID && key != stochastic.NoArgID && value == stochastic.NoArgID) ||
						(OpNumArgs[op] == 3 && addr != stochastic.NoArgID && key != stochastic.NoArgID && value != stochastic.NoArgID) {

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
		for addr := range stochastic.NumArgKinds {
			for key := range stochastic.NumArgKinds {
				for value := range stochastic.NumArgKinds {
					// check legality of argument/op combination
					if (OpNumArgs[op] == 0 && addr == stochastic.NoArgID && key == stochastic.NoArgID && value == stochastic.NoArgID) ||
						(OpNumArgs[op] == 1 && addr != stochastic.NoArgID && key == stochastic.NoArgID && value == stochastic.NoArgID) ||
						(OpNumArgs[op] == 2 && addr != stochastic.NoArgID && key != stochastic.NoArgID && value == stochastic.NoArgID) ||
						(OpNumArgs[op] == 3 && addr != stochastic.NoArgID && key != stochastic.NoArgID && value != stochastic.NoArgID) {

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

func TestOperations_OpMnemo(t *testing.T) {
	// case 1
	out := OpMnemo(SnapshotID)
	assert.Equal(t, "SN", out)

	// case 2
	assert.Panics(t, func() {
		OpMnemo(-1)
	})
}

func TestOperations_EncodeArgOp(t *testing.T) {
	argop, err := EncodeArgOp(SetStateID, stochastic.PrevArgID, stochastic.NewArgID, stochastic.NewArgID)
	assert.Nil(t, err)
	op, addr, key, value, err := DecodeArgOp(argop)
	assert.Nil(t, err)
	assert.Equal(t, SetStateID, op)
	assert.Equal(t, stochastic.PrevArgID, addr)
	assert.Equal(t, stochastic.NewArgID, key)
	assert.Equal(t, stochastic.NewArgID, value)

	_, err = EncodeArgOp(SetStateID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NewArgID)
	assert.NotNil(t, err)
}

func TestOperations_DecodeArgOp(t *testing.T) {
	_, _, _, _, err := DecodeArgOp(NumArgOps)
	assert.NotNil(t, err)

	argop := (((int(SetCodeID)*stochastic.NumArgKinds)+stochastic.NoArgID)*stochastic.NumArgKinds+stochastic.NoArgID)*stochastic.NumArgKinds + stochastic.NewArgID
	_, _, _, _, err = DecodeArgOp(argop)
	assert.NotNil(t, err)
}

func TestOperations_EncodeOpcode(t *testing.T) {
	_, err := EncodeOpcode(SetStateID, stochastic.PrevArgID, stochastic.NewArgID, stochastic.NewArgID)
	assert.Nil(t, err)

	_, err = EncodeOpcode(SetStateID, stochastic.NoArgID, stochastic.NewArgID, stochastic.NewArgID)
	assert.NotNil(t, err)
}

func TestOperations_checkArgOp(t *testing.T) {
	err := checkArgOp(SnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)

	err = checkArgOp(SnapshotID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.ZeroArgID)
	assert.NotNil(t, err)

	err = checkArgOp(CreateAccountID, stochastic.ZeroArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)

	err = checkArgOp(CreateAccountID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.NoArgID)
	assert.NotNil(t, err)

	err = checkArgOp(GetStateID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.NoArgID)
	assert.Nil(t, err)

	err = checkArgOp(GetStateID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.NoArgID)
	assert.Nil(t, err)

	err = checkArgOp(GetStateID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.ZeroArgID)
	assert.NotNil(t, err)

	err = checkArgOp(SetStateID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.ZeroArgID)
	assert.Nil(t, err)

	err = checkArgOp(SetStateID, stochastic.NoArgID, stochastic.ZeroArgID, stochastic.ZeroArgID)
	assert.NotNil(t, err)

	err = checkArgOp(-1, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.NotNil(t, err)

	err = checkArgOp(SnapshotID, -1, stochastic.NoArgID, stochastic.NoArgID)
	assert.NotNil(t, err)

	err = checkArgOp(SnapshotID, stochastic.NoArgID, -1, stochastic.NoArgID)
	assert.NotNil(t, err)

	err = checkArgOp(SnapshotID, stochastic.NoArgID, stochastic.NoArgID, -1)
	assert.NotNil(t, err)
}

func TestOperations_DecodeOpcode(t *testing.T) {
	_, _, _, _, err := DecodeOpcode("XX")
	assert.NotNil(t, err)

	_, _, _, _, err = DecodeOpcode("CAz")
	assert.Nil(t, err)

	_, _, _, _, err = DecodeOpcode("GSzn")
	assert.Nil(t, err)

	_, _, _, _, err = DecodeOpcode("SSnpz")
	assert.Nil(t, err)

	_, _, _, _, err = DecodeOpcode("SS")
	assert.NotNil(t, err)

	_, _, _, _, err = DecodeOpcode("SSl")
	assert.NotNil(t, err)

	_, _, _, _, err = DecodeOpcode("SSll")
	assert.NotNil(t, err)

	_, _, _, _, err = DecodeOpcode("SSlll")
	assert.NotNil(t, err)

	_, _, _, _, err = DecodeOpcode("CAl")
	assert.NotNil(t, err)

	_, _, _, _, err = DecodeOpcode("GSll")
	assert.NotNil(t, err)

	_, _, _, _, err = DecodeOpcode("SS")
	assert.NotNil(t, err)
}

func TestOperations_IsValidArgOp(t *testing.T) {
	// encode to an argument-encoded operation
	argop, err := EncodeArgOp(SetStateID, stochastic.PrevArgID, stochastic.NewArgID, stochastic.NewArgID)
	if err != nil {
		t.Fatalf("Encoding failed")
	}
	valid := IsValidArgOp(argop)
	assert.True(t, valid)

	invalid := IsValidArgOp(-1)
	assert.False(t, invalid)

	invalid = IsValidArgOp(NumArgOps)
	assert.False(t, invalid)
}

func TestOperations_ToAddress(t *testing.T) {
	// case 1
	key, err := ToAddress(0)
	assert.Nil(t, err)
	assert.Equal(t, "0x0000000000000000000000000000000000000000", key.Hex())

	// case 2
	key, err = ToAddress(1)
	assert.Nil(t, err)
	assert.Equal(t, "0xe73e0539db9dEb8D2e32FF19dD634AA67Ef69Fd6", key.Hex())

	// case 3
	key, err = ToAddress(16)
	assert.Nil(t, err)
	assert.Equal(t, "0xE42dc42fA2F6e826e4CF42cF3Ef168729B691eD1", key.Hex())

	// case 4
	_, err = ToAddress(-1)
	assert.NotNil(t, err)
}

func TestOperations_ToHash(t *testing.T) {
	// case 1
	h, err := ToHash(0)
	assert.Nil(t, err)
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", h.String())

	// case 2
	h, err = ToHash(1)
	assert.Nil(t, err)
	assert.Equal(t, "0x054a70e1e64dddae740f584be73e0539db9deb8d2e32ff19dd634aa67ef69fd6", h.String())

	// case 3
	h, err = ToHash(16)
	assert.Nil(t, err)
	assert.Equal(t, "0xfba0f49e88150664b8512808e42dc42fa2f6e826e4cf42cf3ef168729b691ed1", h.String())

	// case 4
	h, err = ToHash(-1)
	assert.NotNil(t, err)
	assert.Equal(t, common.Hash{}, h)
}

func TestOperations_checkArgOpInvalidArity(t *testing.T) {
	op := SnapshotID
	original := OpNumArgs[op]
	OpNumArgs[op] = 4
	t.Cleanup(func() {
		OpNumArgs[op] = original
	})

	err := checkArgOp(op, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid number of arguments")
}

func TestOperations_EncodeOpcodeLengthMismatch(t *testing.T) {
	original := argMnemo[stochastic.NoArgID]
	argMnemo[stochastic.NoArgID] = "x"
	t.Cleanup(func() {
		argMnemo[stochastic.NoArgID] = original
	})

	_, err := EncodeOpcode(SnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong opcode length")
}
