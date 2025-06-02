// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/txgenerator/norma.go
package txgenerator

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestNewNormaTxContext(t *testing.T) {
	// Create a private key for signing
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Create addresses
	sender := crypto.PubkeyToAddress(privateKey.PublicKey)
	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create transaction parameters
	value := big.NewInt(1000)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1)
	data := []byte{1, 2, 3, 4}
	nonce := uint64(0)
	chainId := big.NewInt(1)

	// Create and sign the transaction
	tx := types.NewTransaction(nonce, recipient, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	assert.NoError(t, err)

	// Test with provided sender
	blkNumber := uint64(12345)
	fork := "shanghai"
	ctx, err := NewNormaTxContext(signedTx, blkNumber, &sender, fork)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// Test message fields
	msg := ctx.GetMessage()
	assert.Equal(t, sender, msg.From)
	assert.Equal(t, &recipient, msg.To)
	assert.Equal(t, nonce, msg.Nonce)
	assert.Equal(t, value, msg.Value)
	assert.Equal(t, gasLimit, msg.GasLimit)
	assert.Equal(t, gasPrice, msg.GasPrice)
	assert.Equal(t, data, msg.Data)

	// Test environment fields
	env := ctx.GetBlockEnvironment()
	assert.Equal(t, blkNumber, env.GetNumber())
	assert.Equal(t, fork, env.GetFork())

	// Test with derived sender
	ctx2, err := NewNormaTxContext(signedTx, blkNumber, nil, fork)
	assert.NoError(t, err)
	assert.NotNil(t, ctx2)

	// Verify sender was derived correctly
	msg2 := ctx2.GetMessage()
	assert.Equal(t, sender, msg2.From)

	// Test error case with invalid transaction signature
	invalidTx := types.NewTransaction(nonce, recipient, value, gasLimit, gasPrice, data)
	_, err = NewNormaTxContext(invalidTx, blkNumber, nil, fork)
	assert.Error(t, err)
}

func TestNormaTxBlockEnv(t *testing.T) {
	// Create block environment
	blkNumber := uint64(12345)
	fork := "shanghai"
	env := normaTxBlockEnv{
		blkNumber: blkNumber,
		fork:      fork,
	}

	// Test GetRandom
	assert.Nil(t, env.GetRandom())

	// Test GetCoinbase
	assert.Equal(t, common.HexToAddress("0x1"), env.GetCoinbase())

	// Test GetBlobBaseFee
	assert.Equal(t, big.NewInt(0), env.GetBlobBaseFee())

	// Test GetDifficulty
	assert.Equal(t, big.NewInt(1), env.GetDifficulty())

	// Test GetGasLimit
	assert.Equal(t, uint64(1_000_000_000_000), env.GetGasLimit())

	// Test GetNumber
	assert.Equal(t, blkNumber, env.GetNumber())

	// Test GetTimestamp
	now := uint64(time.Now().Unix())
	timestamp := env.GetTimestamp()
	// Timestamp should be close to current time (within 2 seconds)
	assert.True(t, timestamp >= now-2 && timestamp <= now+2)

	// Test GetBlockHash
	testBlockNumber := uint64(100)
	expectedHash := common.BigToHash(big.NewInt(int64(testBlockNumber)))
	actualHash, err := env.GetBlockHash(testBlockNumber)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, actualHash)

	// Test GetBaseFee
	assert.Equal(t, big.NewInt(0), env.GetBaseFee())

	// Test GetFork
	assert.Equal(t, fork, env.GetFork())
}
