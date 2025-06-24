// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/substate/world_state.go
package substate

import (
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestWorldState_Has(t *testing.T) {
	// Create test addresses
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	nonExistentAddr := common.HexToAddress("0x3333333333333333333333333333333333333333")

	// Create substate accounts
	acc1 := &substate.Account{
		Balance: new(uint256.Int).SetUint64(100),
	}
	acc2 := &substate.Account{
		Balance: new(uint256.Int).SetUint64(200),
	}

	// Create world state
	alloc := substate.WorldState{
		substatetypes.Address(addr1): acc1,
		substatetypes.Address(addr2): acc2,
	}
	ws := NewWorldState(alloc)

	// Test Has method
	assert.True(t, ws.Has(addr1))
	assert.True(t, ws.Has(addr2))
	assert.False(t, ws.Has(nonExistentAddr))
}

func TestWorldState_Get(t *testing.T) {
	// Create test address
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	nonExistentAddr := common.HexToAddress("0x3333333333333333333333333333333333333333")

	// Create storage
	storage := map[substatetypes.Hash]substatetypes.Hash{
		substatetypes.Hash(common.HexToHash("0x1")): substatetypes.Hash(common.HexToHash("0xabc")),
	}

	// Create substate account
	acc1 := &substate.Account{
		Code:    []byte{1, 2, 3},
		Storage: storage,
		Balance: new(uint256.Int).SetUint64(100),
		Nonce:   5,
	}

	// Create world state
	alloc := substate.WorldState{
		substatetypes.Address(addr1): acc1,
	}
	ws := NewWorldState(alloc)

	// Test Get for existing account
	account := ws.Get(addr1)
	assert.NotNil(t, account)
	assert.Equal(t, uint64(5), account.GetNonce())
	assert.Equal(t, uint256.NewInt(100), account.GetBalance())
	assert.Equal(t, []byte{1, 2, 3}, account.GetCode())
	assert.Equal(t, 1, account.GetStorageSize())
	assert.True(t, account.HasStorageAt(common.HexToHash("0x1")))
	assert.Equal(t, common.HexToHash("0xabc"), account.GetStorageAt(common.HexToHash("0x1")))

	// Test Get for non-existent account
	nonExistentAcc := ws.Get(nonExistentAddr)
	assert.Nil(t, nonExistentAcc)
}

func TestWorldState_ForEachAccount(t *testing.T) {
	// Create test addresses
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Create substate accounts
	acc1 := &substate.Account{
		Balance: new(uint256.Int).SetUint64(100),
		Nonce:   1,
	}
	acc2 := &substate.Account{
		Balance: new(uint256.Int).SetUint64(200),
		Nonce:   2,
	}

	// Create world state
	alloc := substate.WorldState{
		substatetypes.Address(addr1): acc1,
		substatetypes.Address(addr2): acc2,
	}
	ws := NewWorldState(alloc)

	// Test ForEachAccount
	visitedAccounts := make(map[common.Address]txcontext.Account)
	ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		visitedAccounts[addr] = acc
	})

	// Verify all accounts were visited
	assert.Equal(t, 2, len(visitedAccounts))
	assert.NotNil(t, visitedAccounts[addr1])
	assert.NotNil(t, visitedAccounts[addr2])
	assert.Equal(t, uint64(1), visitedAccounts[addr1].GetNonce())
	assert.Equal(t, uint64(2), visitedAccounts[addr2].GetNonce())

	// Test ForEachAccount with empty world state
	emptyWS := NewWorldState(substate.WorldState{})
	emptyVisited := make(map[common.Address]txcontext.Account)
	emptyWS.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		emptyVisited[addr] = acc
	})
	assert.Equal(t, 0, len(emptyVisited))
}

func TestWorldState_Len(t *testing.T) {
	// Test empty world state
	emptyWS := NewWorldState(substate.WorldState{})
	assert.Equal(t, 0, emptyWS.Len())

	// Test non-empty world state
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	alloc := substate.WorldState{
		substatetypes.Address(addr1): &substate.Account{},
		substatetypes.Address(addr2): &substate.Account{},
	}
	ws := NewWorldState(alloc)
	assert.Equal(t, 2, ws.Len())
}

func TestWorldState_Delete(t *testing.T) {
	// Create test addresses
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Create world state
	alloc := substate.WorldState{
		substatetypes.Address(addr1): &substate.Account{},
		substatetypes.Address(addr2): &substate.Account{},
	}
	ws := NewWorldState(alloc)
	assert.Equal(t, 2, ws.Len())

	// Test Delete
	ws.Delete(addr1)
	assert.Equal(t, 1, ws.Len())
	assert.False(t, ws.Has(addr1))
	assert.True(t, ws.Has(addr2))

	// Test Delete on non-existent address
	nonExistentAddr := common.HexToAddress("0x3333333333333333333333333333333333333333")
	ws.Delete(nonExistentAddr) // Should not panic
	assert.Equal(t, 1, ws.Len())
}

func TestWorldState_Equal(t *testing.T) {
	// Create test addresses
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Create identical accounts
	acc1 := &substate.Account{
		Balance: new(uint256.Int).SetUint64(100),
		Nonce:   1,
	}
	acc2 := &substate.Account{
		Balance: new(uint256.Int).SetUint64(200),
		Nonce:   2,
	}

	// Create identical world states
	alloc1 := substate.WorldState{
		substatetypes.Address(addr1): acc1,
		substatetypes.Address(addr2): acc2,
	}
	alloc2 := substate.WorldState{
		substatetypes.Address(addr1): acc1,
		substatetypes.Address(addr2): acc2,
	}

	ws1 := NewWorldState(alloc1)
	ws2 := NewWorldState(alloc2)

	// Test Equal with identical world states
	assert.True(t, ws1.Equal(ws2))
	assert.True(t, ws2.Equal(ws1))

	// Test Equal with same world state
	assert.True(t, ws1.Equal(ws1))

	// Test Equal with different world states (different account count)
	alloc3 := substate.WorldState{
		substatetypes.Address(addr1): acc1,
	}
	ws3 := NewWorldState(alloc3)
	assert.False(t, ws1.Equal(ws3))

	// Test Equal with different world states (same count, different addresses)
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	alloc4 := substate.WorldState{
		substatetypes.Address(addr1): acc1,
		substatetypes.Address(addr3): acc2,
	}
	ws4 := NewWorldState(alloc4)
	assert.False(t, ws1.Equal(ws4))

	// Test Equal with different world states (same addresses, different account data)
	acc2Modified := &substate.Account{
		Balance: new(uint256.Int).SetUint64(300), // Different balance
		Nonce:   2,
	}
	alloc5 := substate.WorldState{
		substatetypes.Address(addr1): acc1,
		substatetypes.Address(addr2): acc2Modified,
	}
	ws5 := NewWorldState(alloc5)
	assert.False(t, ws1.Equal(ws5))
}

func TestWorldState_String(t *testing.T) {
	// Create a world state with accounts
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")

	storage := map[substatetypes.Hash]substatetypes.Hash{
		substatetypes.Hash(common.HexToHash("0x1")): substatetypes.Hash(common.HexToHash("0xabc")),
	}

	acc1 := &substate.Account{
		Code:    []byte{1, 2, 3},
		Storage: storage,
		Balance: new(uint256.Int).SetUint64(100),
		Nonce:   5,
	}

	alloc := substate.WorldState{
		substatetypes.Address(addr1): acc1,
	}

	ws := NewWorldState(alloc)

	// Test String method
	str := ws.String()

	// Verify the string contains the important parts
	assert.Contains(t, str, "World State")
	assert.Contains(t, str, "size: 1")
	assert.Contains(t, str, addr1.String()[2:]) // Address without 0x prefix
	assert.Contains(t, str, "nonce: 5")
	assert.Contains(t, str, "100") // Balance value
}
