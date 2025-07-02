package ethtest

import (
	"math/big"
	"testing"

	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func makeTestAlloc() types.GenesisAlloc {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	return types.GenesisAlloc{
		addr1: {Code: []byte{0x01}, Storage: map[common.Hash]common.Hash{}, Balance: big.NewInt(100), Nonce: 1},
		addr2: {Code: []byte{0x02}, Storage: map[common.Hash]common.Hash{}, Balance: big.NewInt(200), Nonce: 2},
	}
}

func TestWorldStateAlloc_NewWorldState(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	assert.NotNil(t, ws)
}

func TestWorldStateAlloc_Get_Exists(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	addr := common.HexToAddress("0x1")
	acc := ws.Get(addr)
	assert.NotNil(t, acc)
	assert.Equal(t, uint256.NewInt(100), acc.GetBalance())
	assert.Equal(t, uint64(1), acc.GetNonce())
}

func TestWorldStateAlloc_Get_NotExists(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	addr := common.HexToAddress("0x3")
	acc := ws.Get(addr)
	assert.Equal(t, txcontext.NewNilAccount(), acc)
}

func TestWorldStateAlloc_Has(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	addr := common.HexToAddress("0x1")
	assert.True(t, ws.Has(addr))
	addr2 := common.HexToAddress("0x3")
	assert.False(t, ws.Has(addr2))
}

func TestWorldStateAlloc_ForEachAccount(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	found := map[string]bool{"0x1": false, "0x2": false}
	ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		if addr.Hex() == "0x0000000000000000000000000000000000000001" {
			found["0x1"] = true
		}
		if addr.Hex() == "0x0000000000000000000000000000000000000002" {
			found["0x2"] = true
		}
	})
	assert.True(t, found["0x1"])
	assert.True(t, found["0x2"])
}

func TestWorldStateAlloc_Len(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	assert.Equal(t, 2, ws.Len())
}

func TestWorldStateAlloc_Equal(t *testing.T) {
	alloc := makeTestAlloc()
	ws1 := NewWorldState(alloc)
	ws2 := NewWorldState(alloc)
	assert.True(t, ws1.Equal(ws2))
}

func TestWorldStateAlloc_Delete(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	addr := common.HexToAddress("0x1")
	ws.Delete(addr)
	assert.False(t, ws.Has(addr))
}

func TestWorldStateAlloc_String(t *testing.T) {
	alloc := makeTestAlloc()
	ws := NewWorldState(alloc)
	str := ws.String()
	assert.Contains(t, str, "World State")
	assert.Contains(t, str, "Accounts")
}
