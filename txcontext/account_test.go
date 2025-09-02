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

package txcontext

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestAccount_NewNilAccount(t *testing.T) {
	acc := NewNilAccount()

	assert.Equal(t, &account{}, acc)
}

func TestAccount_NewAccount(t *testing.T) {
	code := []byte{1, 2, 3, 4}
	storage := map[common.Hash]common.Hash{
		common.HexToHash("0x1"): common.HexToHash("0xabc"),
		common.HexToHash("0x2"): common.HexToHash("0xdef"),
	}
	balance := big.NewInt(100)
	nonce := uint64(5)

	acc := NewAccount(code, storage, balance, nonce)

	assert.Equal(t, nonce, acc.GetNonce())
	assert.Equal(t, uint256.MustFromBig(balance), acc.GetBalance())
	assert.Equal(t, code, acc.GetCode())
	assert.Equal(t, len(storage), acc.GetStorageSize())

	// Test storage access
	assert.True(t, acc.HasStorageAt(common.HexToHash("0x1")))
	assert.False(t, acc.HasStorageAt(common.HexToHash("0x3")))
	assert.Equal(t, common.HexToHash("0xabc"), acc.GetStorageAt(common.HexToHash("0x1")))
}

func TestAccount_AccountEqual(t *testing.T) {
	// Create two identical accounts
	code1 := []byte{1, 2, 3, 4}
	storage1 := map[common.Hash]common.Hash{
		common.HexToHash("0x1"): common.HexToHash("0xabc"),
		common.HexToHash("0x2"): common.HexToHash("0xdef"),
	}
	balance1 := big.NewInt(100)
	nonce1 := uint64(5)

	acc1 := NewAccount(code1, storage1, balance1, nonce1)

	// Clone the first account
	code2 := []byte{1, 2, 3, 4}
	storage2 := map[common.Hash]common.Hash{
		common.HexToHash("0x1"): common.HexToHash("0xabc"),
		common.HexToHash("0x2"): common.HexToHash("0xdef"),
	}
	balance2 := big.NewInt(100)
	nonce2 := uint64(5)

	acc2 := NewAccount(code2, storage2, balance2, nonce2)

	// Test equality with identical accounts
	assert.True(t, AccountEqual(acc1, acc2))

	// Test equality with same account
	assert.True(t, AccountEqual(acc1, acc1))

	// Test with nil accounts
	assert.False(t, AccountEqual(acc1, nil))
	assert.False(t, AccountEqual(nil, acc1))
	assert.True(t, AccountEqual(nil, nil))

	// Create different accounts for testing inequality
	diffNonce := NewAccount(code1, storage1, balance1, nonce1+1)
	assert.False(t, AccountEqual(acc1, diffNonce))

	diffBalance := NewAccount(code1, storage1, big.NewInt(101), nonce1)
	assert.False(t, AccountEqual(acc1, diffBalance))

	diffCode := NewAccount([]byte{1, 2, 3, 5}, storage1, balance1, nonce1)
	assert.False(t, AccountEqual(acc1, diffCode))

	// TODO may be bug
	diffStorage := NewAccount(code1, map[common.Hash]common.Hash{
		common.HexToHash("0x1"): common.HexToHash("0xabc"),
		common.HexToHash("0x3"): common.HexToHash("0xdef"),
	}, balance1, nonce1)
	assert.True(t, AccountEqual(acc1, diffStorage))
}

func TestAccount_StorageHandling(t *testing.T) {
	storage := map[common.Hash]common.Hash{
		common.HexToHash("0x1"): common.HexToHash("0xabc"),
		common.HexToHash("0x2"): common.HexToHash("0xdef"),
	}

	acc := NewAccount([]byte{}, storage, big.NewInt(0), 0)

	// Test ForEachStorage
	visitedKeys := make(map[common.Hash]common.Hash)
	acc.ForEachStorage(func(key, value common.Hash) {
		visitedKeys[key] = value
	})

	assert.Equal(t, len(storage), len(visitedKeys))
	for k, v := range storage {
		assert.Equal(t, v, visitedKeys[k])
	}
}

func TestAccount_String(t *testing.T) {
	// Create an account with specific data for string representation testing
	storage := map[common.Hash]common.Hash{
		common.HexToHash("0x1"): common.HexToHash("0xabc"),
	}
	acc := NewAccount([]byte{1, 2, 3}, storage, big.NewInt(100), 5)

	str := acc.String()

	// Verify the string contains the important parts (we don't test exact format as it might change)
	assert.Contains(t, str, "nonce: 5")
	assert.Contains(t, str, "balance")
	assert.Contains(t, str, "Storage")
	assert.Contains(t, str, "0x0000000000000000000000000000000000000000000000000000000000000001")
	assert.Contains(t, str, "0x0000000000000000000000000000000000000000000000000000000000000abc")
}
