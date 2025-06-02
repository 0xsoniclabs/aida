// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/world_state.go
package txcontext

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestNewWorldState(t *testing.T) {
	// Create a new empty world state
	ws := NewWorldState(make(map[common.Address]Account))
	assert.Equal(t, 0, ws.Len())

	// Create a world state with accounts
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	acc1 := NewAccount([]byte{1, 2, 3}, nil, big.NewInt(100), 1)
	acc2 := NewAccount([]byte{4, 5, 6}, nil, big.NewInt(200), 2)

	accounts := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2,
	}

	ws = NewWorldState(accounts)
	assert.Equal(t, 2, ws.Len())
}

func TestWorldStateGetAndHas(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	nonExistentAddr := common.HexToAddress("0x3")

	acc1 := NewAccount([]byte{1, 2, 3}, nil, big.NewInt(100), 1)
	acc2 := NewAccount([]byte{4, 5, 6}, nil, big.NewInt(200), 2)

	accounts := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2,
	}

	ws := NewWorldState(accounts)

	// Test Has method
	assert.True(t, ws.Has(addr1))
	assert.True(t, ws.Has(addr2))
	assert.False(t, ws.Has(nonExistentAddr))

	// Test Get method
	assert.Equal(t, acc1, ws.Get(addr1))
	assert.Equal(t, acc2, ws.Get(addr2))
	assert.Nil(t, ws.Get(nonExistentAddr))
}

func TestWorldStateDelete(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")

	acc1 := NewAccount([]byte{1, 2, 3}, nil, big.NewInt(100), 1)
	acc2 := NewAccount([]byte{4, 5, 6}, nil, big.NewInt(200), 2)

	accounts := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2,
	}

	ws := NewWorldState(accounts)
	assert.Equal(t, 2, ws.Len())

	// Delete an account
	ws.Delete(addr1)
	assert.Equal(t, 1, ws.Len())
	assert.False(t, ws.Has(addr1))
	assert.True(t, ws.Has(addr2))

	// Delete non-existent account should not affect the state
	ws.Delete(common.HexToAddress("0x3"))
	assert.Equal(t, 1, ws.Len())
}

func TestWorldStateForEachAccount(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")

	acc1 := NewAccount([]byte{1, 2, 3}, nil, big.NewInt(100), 1)
	acc2 := NewAccount([]byte{4, 5, 6}, nil, big.NewInt(200), 2)

	accounts := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2,
	}

	ws := NewWorldState(accounts)

	// Use ForEachAccount to collect accounts
	visitedAccounts := make(map[common.Address]Account)
	ws.ForEachAccount(func(addr common.Address, acc Account) {
		visitedAccounts[addr] = acc
	})

	assert.Equal(t, 2, len(visitedAccounts))
	assert.Equal(t, acc1, visitedAccounts[addr1])
	assert.Equal(t, acc2, visitedAccounts[addr2])
}

func TestWorldStateEqual(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")

	acc1 := NewAccount([]byte{1, 2, 3}, nil, big.NewInt(100), 1)
	acc2 := NewAccount([]byte{4, 5, 6}, nil, big.NewInt(200), 2)

	// Create two identical world states
	accounts1 := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2,
	}
	ws1 := NewWorldState(accounts1)

	accounts2 := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2,
	}
	ws2 := NewWorldState(accounts2)

	// Test equality with identical world states
	assert.True(t, ws1.Equal(ws2))
	assert.True(t, ws2.Equal(ws1))

	// Test equality with same world state
	assert.True(t, ws1.Equal(ws1))

	// Create different world states for testing inequality
	// Different number of accounts
	accounts3 := map[common.Address]Account{
		addr1: acc1,
	}
	ws3 := NewWorldState(accounts3)
	assert.False(t, ws1.Equal(ws3))

	// Same number of accounts but different addresses
	addr3 := common.HexToAddress("0x3")
	accounts4 := map[common.Address]Account{
		addr1: acc1,
		addr3: acc2,
	}
	ws4 := NewWorldState(accounts4)
	assert.False(t, ws1.Equal(ws4))

	// Same addresses but different account data
	acc2Modified := NewAccount([]byte{4, 5, 6}, nil, big.NewInt(300), 2)
	accounts5 := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2Modified,
	}
	ws5 := NewWorldState(accounts5)
	assert.False(t, ws1.Equal(ws5))
}

func TestWorldStateString(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")

	storage := map[common.Hash]common.Hash{
		common.HexToHash("0x1"): common.HexToHash("0xabc"),
	}

	acc1 := NewAccount([]byte{1, 2, 3}, storage, big.NewInt(100), 1)
	acc2 := NewAccount([]byte{4, 5, 6}, nil, big.NewInt(200), 2)

	accounts := map[common.Address]Account{
		addr1: acc1,
		addr2: acc2,
	}

	ws := NewWorldState(accounts)

	str := ws.String()

	// Verify the string contains the important parts
	assert.Contains(t, str, "World State")
	assert.Contains(t, str, "size: 2")
	assert.Contains(t, str, "Accounts")
	assert.Contains(t, str, addr1.String()[2:]) // Address without 0x prefix
	assert.Contains(t, str, addr2.String()[2:]) // Address without 0x prefix
}
