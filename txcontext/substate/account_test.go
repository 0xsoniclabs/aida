// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/substate/account.go
package substate

import (
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestAccount_NewAccount(t *testing.T) {
	// Create a substate account
	code := []byte{1, 2, 3, 4}
	storage := map[substatetypes.Hash]substatetypes.Hash{
		substatetypes.Hash(common.HexToHash("0x1")): substatetypes.Hash(common.HexToHash("0xabc")),
		substatetypes.Hash(common.HexToHash("0x2")): substatetypes.Hash(common.HexToHash("0xdef")),
	}
	balance := uint256.NewInt(100)
	nonce := uint64(5)

	substateAccount := &substate.Account{
		Code:    code,
		Storage: storage,
		Balance: balance,
		Nonce:   nonce,
	}

	// Create txcontext.Account from substate.Account
	acc := NewAccount(substateAccount)

	// Test GetNonce
	assert.Equal(t, nonce, acc.GetNonce())

	// Test GetBalance
	assert.Equal(t, balance, acc.GetBalance())

	// Test GetCode
	assert.Equal(t, code, acc.GetCode())

	// Test GetStorageSize
	assert.Equal(t, len(storage), acc.GetStorageSize())

	// Test HasStorageAt
	assert.True(t, acc.HasStorageAt(common.HexToHash("0x1")))
	assert.True(t, acc.HasStorageAt(common.HexToHash("0x2")))
	assert.False(t, acc.HasStorageAt(common.HexToHash("0x3")))

	// Test GetStorageAt
	assert.Equal(t, common.HexToHash("0xabc"), acc.GetStorageAt(common.HexToHash("0x1")))
	assert.Equal(t, common.HexToHash("0xdef"), acc.GetStorageAt(common.HexToHash("0x2")))
	assert.Equal(t, common.Hash{}, acc.GetStorageAt(common.HexToHash("0x3")))
}

func TestAccount_ForEachStorage(t *testing.T) {
	// Create a substate account with storage
	storage := map[substatetypes.Hash]substatetypes.Hash{
		substatetypes.Hash(common.HexToHash("0x1")): substatetypes.Hash(common.HexToHash("0xabc")),
		substatetypes.Hash(common.HexToHash("0x2")): substatetypes.Hash(common.HexToHash("0xdef")),
	}

	substateAccount := &substate.Account{
		Storage: storage,
		Balance: uint256.NewInt(0),
	}

	acc := NewAccount(substateAccount)

	// Test ForEachStorage
	visitedKeys := make(map[common.Hash]common.Hash)
	acc.ForEachStorage(func(key, value common.Hash) {
		visitedKeys[key] = value
	})

	// Verify all storage items were visited
	assert.Equal(t, 2, len(visitedKeys))
	assert.Equal(t, common.HexToHash("0xabc"), visitedKeys[common.HexToHash("0x1")])
	assert.Equal(t, common.HexToHash("0xdef"), visitedKeys[common.HexToHash("0x2")])
}

func TestAccount_String(t *testing.T) {
	// Create a substate account with storage
	code := []byte{1, 2, 3}
	storage := map[substatetypes.Hash]substatetypes.Hash{
		substatetypes.Hash(common.HexToHash("0x1")): substatetypes.Hash(common.HexToHash("0xabc")),
	}
	balance := uint256.NewInt(100)
	nonce := uint64(5)

	substateAccount := &substate.Account{
		Code:    code,
		Storage: storage,
		Balance: balance,
		Nonce:   nonce,
	}

	acc := NewAccount(substateAccount)

	// Test String method
	str := acc.String()

	// Verify the string contains the important parts
	assert.Contains(t, str, "nonce: 5")
	assert.Contains(t, str, "balance")
	assert.Contains(t, str, "100") // Balance value
	assert.Contains(t, str, "Storage")
	assert.Contains(t, str, "0x0000000000000000000000000000000000000000000000000000000000000001")
	assert.Contains(t, str, "0x0000000000000000000000000000000000000000000000000000000000000abc")
}

func TestAccount_WithEmptyStorage(t *testing.T) {
	// Create a substate account with empty storage
	code := []byte{1, 2, 3, 4}
	storage := map[substatetypes.Hash]substatetypes.Hash{}
	balance := uint256.NewInt(100)
	nonce := uint64(5)

	substateAccount := &substate.Account{
		Code:    code,
		Storage: storage,
		Balance: balance,
		Nonce:   nonce,
	}

	acc := NewAccount(substateAccount)

	// Test GetStorageSize with empty storage
	assert.Equal(t, 0, acc.GetStorageSize())

	// Test ForEachStorage with empty storage
	visitCount := 0
	acc.ForEachStorage(func(key, value common.Hash) {
		visitCount++
	})
	assert.Equal(t, 0, visitCount)
}
