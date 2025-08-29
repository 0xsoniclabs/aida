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

package ethtest

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	etests "github.com/ethereum/go-ethereum/tests"
	"github.com/stretchr/testify/assert"
)

func TestStBlockEnvironment_GetBaseFee(t *testing.T) {
	baseFee := newBigInt(10)

	tests := []struct {
		name     string
		baseFee  *BigInt
		want     *big.Int
		fork     string
		chainCfg *params.ChainConfig
	}{
		{
			name:    "Use_Predefined_If_nil",
			baseFee: nil,
			fork:    "London",
			want:    big.NewInt(0x0a),
		},
		{
			name:    "Pre_London_Returns_nil",
			baseFee: baseFee,
			fork:    "Berlin",
			want:    nil,
		},
		{
			name:    "Post_London_Uses_Given",
			baseFee: baseFee,
			fork:    "London",
			want:    baseFee.Convert(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chainCfg, _, err := etests.GetChainConfig(test.fork)
			if err != nil {
				t.Fatalf("cannot get chain config: %v", err)
			}
			env := stBlockEnvironment{BaseFee: baseFee, chainCfg: chainCfg}
			if got, want := env.GetBaseFee(), test.want; got.Cmp(want) != 0 {
				t.Errorf("unexpected base fee\ngot: %d\nwant: %d", got.Uint64(), want.Uint64())
			}
		})
	}
}

func TestStBlockEnvironment_GetBlockHash_Correctly_Converts(t *testing.T) {
	blockNum := int64(10)
	want := common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(blockNum).String())))
	env := &stBlockEnvironment{Number: newBigInt(blockNum)}

	got, err := env.GetBlockHash(uint64(blockNum))
	if err != nil {
		t.Fatalf("cannot get block hash: %v", err)
	}
	if want.Cmp(got) != 0 {
		t.Errorf("unexpected block hash, got: %s, want: %s", got, want)
	}
}

func TestStBlockEnvironment_GetGasLimit(t *testing.T) {
	tests := []struct {
		name     string
		gasLimit int64
		want     uint64
	}{
		{
			name:     "0_Uses_GenesisGasLimit",
			gasLimit: 0,
			want:     params.GenesisGasLimit,
		},
		{
			name:     "Non_0_Converts",
			gasLimit: 10,
			want:     10,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := &stBlockEnvironment{GasLimit: newBigInt(test.gasLimit)}
			if got, want := env.GetGasLimit(), test.want; got != want {
				t.Errorf("incorrect gas limit, got: %v, want: %v", got, want)
			}
		})
	}
}

func TestStBlockEnvironment_GetDifficulty(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int64
		fork       string
		random     *BigInt
		want       uint64
	}{
		{
			name:       "PreLondon_Uses_Given",
			difficulty: 1,
			fork:       "Berlin",
			want:       1,
		},
		{
			name:       "PostLondon_With_NotNil_Random_Resets",
			difficulty: 1,
			fork:       "London",
			random:     newBigInt(1),
			want:       0,
		},
		{
			name:       "PostLondon_With_Nil_Random_Uses_Given",
			difficulty: 1,
			fork:       "London",
			want:       1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chainCfg, _, err := etests.GetChainConfig(test.fork)
			if err != nil {
				t.Fatalf("cannot get chain config: %v", err)
			}
			env := &stBlockEnvironment{
				Difficulty: newBigInt(test.difficulty),
				Random:     test.random,
				chainCfg:   chainCfg,
			}
			if got, want := env.GetDifficulty(), test.want; got.Uint64() != want {
				t.Errorf("incorrect gas limit, got: %v, want: %v", got, want)
			}
		})
	}
}

func TestStBlockEnvironment_GetBlobBaseFee(t *testing.T) {
	tests := []struct {
		name        string
		blobBaseFee int64
		fork        string
		want        *big.Int
	}{
		{
			name:        "PreCancun_Returns_Nil",
			blobBaseFee: 1,
			fork:        "London",
			want:        nil,
		},
		{
			name:        "PostCancun_Calculates",
			blobBaseFee: 1,
			fork:        "Cancun",
			want:        big.NewInt(1),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chainCfg, _, err := etests.GetChainConfig(test.fork)
			if err != nil {
				t.Fatalf("cannot get chain config: %v", err)
			}
			env := &stBlockEnvironment{
				ExcessBlobGas: newBigInt(test.blobBaseFee),
				chainCfg:      chainCfg,
				Timestamp:     newBigInt(1),
			}

			got, want := env.GetBlobBaseFee(), test.want
			if got == nil && want == nil {
				return
			}
			if got.Cmp(want) == 0 {
				return
			}

			t.Errorf("incorrect gas limit, got: %d, want: %d", got, want)
		})
	}
}

func TestStBlockEnvironment_CorrectBlockNumberIsReturned(t *testing.T) {
	blkNumber := uint64(1)
	env := &stBlockEnvironment{
		Number: newBigInt(int64(blkNumber)),
	}

	if got, want := env.GetNumber(), blkNumber; got != want {
		t.Errorf("unexpected block number, got: %v, want: %v", got, want)
	}
}

func TestStBlockEnvironment_GetTimestamp(t *testing.T) {
	env := &stBlockEnvironment{
		Timestamp: newBigInt(1234),
	}

	ts := env.GetTimestamp()
	assert.Equal(t, uint64(1234), ts)
}

func TestStBlockEnvironment_GetCoinbase(t *testing.T) {
	env := &stBlockEnvironment{
		Coinbase: common.HexToAddress("0x1234"),
	}

	output := env.GetCoinbase()
	assert.Equal(t, common.HexToAddress("0x1234"), output)
}

func TestStBlockEnvironment_GetRandom(t *testing.T) {
	env := &stBlockEnvironment{
		Random: newBigInt(1234),
		chainCfg: &params.ChainConfig{
			LondonBlock: big.NewInt(0),
		},
	}
	expected := common.HexToHash("0x04d2")
	h := env.GetRandom()
	assert.Equal(t, &expected, h)
}

func TestStBlockEnvironment_GetFork(t *testing.T) {
	env := &stBlockEnvironment{
		fork: "Berlin",
	}

	h := env.GetFork()
	assert.Equal(t, "Berlin", h)
}
