// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/substate/block_environment.go
package substate

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockEnvironment(t *testing.T) {
	// Create a substate.Env with test values
	random := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	randomBytes := random.Bytes()

	blockHashes := map[uint64]substatetypes.Hash{
		100: substatetypes.Hash(common.HexToHash("0x100")),
		101: substatetypes.Hash(common.HexToHash("0x101")),
	}

	coinbase := common.HexToAddress("0x1111111111111111111111111111111111111111")
	difficulty := big.NewInt(1000000)
	gasLimit := uint64(30000000)
	number := uint64(12345)
	timestamp := uint64(1621234567)
	baseFee := big.NewInt(1000)
	blobBaseFee := big.NewInt(2000)

	randomHash := substatetypes.BytesToHash(randomBytes)
	env := &substate.Env{
		Random:      &randomHash,
		BlockHashes: blockHashes,
		Coinbase:    substatetypes.Address(coinbase),
		Difficulty:  difficulty,
		GasLimit:    gasLimit,
		Number:      number,
		Timestamp:   timestamp,
		BaseFee:     baseFee,
		BlobBaseFee: blobBaseFee,
	}

	// Create BlockEnvironment from substate.Env
	blockEnv := NewBlockEnvironment(env)

	// Test GetRandom
	assert.Equal(t, &random, blockEnv.GetRandom())

	// Test GetBlockHash for existing block hash
	hash100, err := blockEnv.GetBlockHash(100)
	assert.NoError(t, err)
	assert.Equal(t, common.HexToHash("0x100"), hash100)

	// Test GetBlockHash for non-existing block hash
	_, err = blockEnv.GetBlockHash(999)
	assert.Error(t, err)

	// Test GetCoinbase
	assert.Equal(t, coinbase, blockEnv.GetCoinbase())

	// Test GetDifficulty
	assert.Equal(t, difficulty, blockEnv.GetDifficulty())

	// Test GetGasLimit
	assert.Equal(t, gasLimit, blockEnv.GetGasLimit())

	// Test GetNumber
	assert.Equal(t, number, blockEnv.GetNumber())

	// Test GetTimestamp
	assert.Equal(t, timestamp, blockEnv.GetTimestamp())

	// Test GetBaseFee
	assert.Equal(t, baseFee, blockEnv.GetBaseFee())

	// Test GetBlobBaseFee
	assert.Equal(t, blobBaseFee, blockEnv.GetBlobBaseFee())

	// Test GetFork - currently returns empty string
	assert.Equal(t, "", blockEnv.GetFork())
}

func TestBlockEnvironmentWithNilRandom(t *testing.T) {
	// Create a substate.Env with nil Random
	env := &substate.Env{
		Random:     nil,
		Difficulty: big.NewInt(1000000),
		GasLimit:   30000000,
		Number:     12345,
		Timestamp:  1621234567,
	}

	// Create BlockEnvironment from substate.Env
	blockEnv := NewBlockEnvironment(env)

	// Test GetRandom returns nil
	assert.Nil(t, blockEnv.GetRandom())
}

func TestBlockEnvironmentWithNilBlockHashes(t *testing.T) {
	// Create a substate.Env with nil BlockHashes
	env := &substate.Env{
		BlockHashes: nil,
		Difficulty:  big.NewInt(1000000),
		GasLimit:    30000000,
		Number:      12345,
		Timestamp:   1621234567,
	}

	// Create BlockEnvironment from substate.Env
	blockEnv := NewBlockEnvironment(env)

	// Test GetBlockHash returns error
	_, err := blockEnv.GetBlockHash(100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no blockhashes provided")
}
