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

package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func GetTestSubstate(encoding string) *substate.Substate {
	txType := int32(substate.SetCodeTxType)
	ss := &substate.Substate{
		InputSubstate:  substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		OutputSubstate: substate.NewWorldState().Add(types.Address{2}, 2, new(uint256.Int).SetUint64(2), nil),
		Env: &substate.Env{
			Coinbase:   types.Address{1},
			Difficulty: new(big.Int).SetUint64(1),
			GasLimit:   1,
			Number:     1,
			Timestamp:  1,
			BaseFee:    new(big.Int).SetUint64(1),
			Random:     &types.Hash{1},
		},
		Message: substate.NewMessage(
			1,
			true,
			new(big.Int).SetUint64(1),
			1,
			types.Address{1},
			new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, &txType,
			types.AccessList{{Address: types.Address{1}, StorageKeys: []types.Hash{{1}, {2}}}}, new(big.Int).SetUint64(1),
			new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0),
			[]types.SetCodeAuthorization{
				{ChainID: *uint256.NewInt(1), Address: types.Address{1}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)},
			}),
		Result: substate.NewResult(1, types.Bloom{1}, []*types.Log{
			{
				Address: types.Address{1},
				Topics:  []types.Hash{{1}, {2}},
				Data:    []byte{1, 2, 3},
				// intentionally skipped: BlockNumber, TxIndex, Index - because protobuf Substate encoding doesn't use these values
				TxHash:    types.Hash{1},
				BlockHash: types.Hash{1},
				Removed:   false,
			},
		},
			// intentionally skipped: ContractAddress - because protobuf Substate encoding doesn't use this value,
			// instead the ContractAddress is derived from Message.From and Message.Nonce
			types.Address{},
			1),
		Block:       37_534_834,
		Transaction: 1,
	}

	// remove fields that are not supported in rlp encoding
	// TODO once protobuf becomes default add ' && encoding != "default" ' to the condition
	if encoding != "protobuf" {
		ss.Env.Random = nil
		ss.Message.AccessList = types.AccessList{}
		ss.Message.SetCodeAuthorizations = nil
	}
	return ss
}

// Must is a helper function that takes a value of any type and an error.
// If the error is nil, it returns the value; if the error is non-nil, it panics.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// CreateTestSubstateDb creates a test substate database with a predefined substate.
func CreateTestSubstateDb(t *testing.T, encoding substateDb.SubstateEncodingSchema) (*substate.Substate, string) {
	path := t.TempDir()
	db, err := substateDb.NewSubstateDB(path, nil, nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.SetSubstateEncoding(encoding))

	ss := GetTestSubstate(string(encoding))
	err = db.PutSubstate(ss)
	require.NoError(t, err)

	md := NewAidaDbMetadata(db, "CRITICAL")
	dbHash, err := hex.DecodeString("a0d4f7616f3007bf8c02f816a60b2526")
	require.NoError(t, err)

	require.NoError(t, md.genMetadata(ss.Block-1, ss.Block+1, 0, 0, SonicMainnetChainID, dbHash))

	require.NoError(t, db.Close())

	return ss, path
}

// ArgsBuilder helps create []string for CLI testing in a type-safe way
type ArgsBuilder struct {
	args []string
}

func NewArgs(cmd string) *ArgsBuilder {
	return &ArgsBuilder{args: []string{cmd}}
}

func (b *ArgsBuilder) Flag(name string, value interface{}) *ArgsBuilder {
	switch v := value.(type) {
	case string:
		b.args = append(b.args, "--"+name, v)
	case int:
		b.args = append(b.args, "--"+name, strconv.Itoa(v))
	case bool:
		if v {
			b.args = append(b.args, "--"+name)
		}
	// You can add more types here (float, time.Duration, etc.)
	default:
		panic(fmt.Sprintf("unsupported flag type %T", v))
	}
	return b
}

func (b *ArgsBuilder) Arg(value interface{}) *ArgsBuilder {
	switch v := value.(type) {
	case string:
		b.args = append(b.args, v)
	case int:
		b.args = append(b.args, strconv.Itoa(v))
	case bool:
		if v {
			b.args = append(b.args, "true")
		} else {
			b.args = append(b.args, "false")
		}
	default:
		panic(fmt.Sprintf("unsupported arg type %T", v))
	}
	return b
}

func (b *ArgsBuilder) Build() []string {
	return b.args
}
