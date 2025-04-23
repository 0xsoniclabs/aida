package utildb

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
)

func TestLoadEthereumGenesisWorldState(t *testing.T) {
	// Create a temporary file to store the genesis JSON
	tempFile, err := os.CreateTemp("", "genesis.json")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// minimum JSON data to test
	genesisData := `{
		"alloc": {
			"0000000000000000000000000000000000000000": {
				"balance": "0x1",
				"nonce":"0x1"
			},
			"efa7454f1116807975a4750b46695e967850de5d": {
				"balance": "0xd3c21bcecceda1000000",
				"storage":{"0x0000000000000000000000000000000000000000000000000000000000000022":"0xf5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b"},
				"nonce":"0x1"
			},
			"fbfd6fa9f73ac6a058e01259034c28001bef8247": {
				"balance": "0x52b7d2dcc80cd2e4000000",
				"code":"0x60806040526004361061003f5760003560e01c"
			}
		}
	}`

	// Write the JSON data to the temporary file
	_, err = tempFile.WriteString(genesisData)
	assert.NoError(t, err)

	// Close the file
	err = tempFile.Close()
	assert.NoError(t, err)

	// Call the function
	worldState, err := LoadEthereumGenesisWorldState(tempFile.Name())
	assert.NoError(t, err)

	// Validate the world state
	assert.NotNil(t, worldState)
	assert.Equal(t, 3, len(worldState))

	// Check specific accounts
	// Check specific accounts
	account1 := worldState[types.HexToAddress("0000000000000000000000000000000000000000")]
	assert.Equal(t, "0x1", account1.Balance.Hex())
	assert.Equal(t, uint64(0x1), account1.Nonce)

	account2 := worldState[types.HexToAddress("efa7454f1116807975a4750b46695e967850de5d")]
	assert.Equal(t, "0xd3c21bcecceda1000000", account2.Balance.Hex())
	assert.Equal(t, uint64(0x1), account2.Nonce)
	decodedKey, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000022")
	if err != nil {
		t.Fatalf("Failed to decode storage key hex string: %v", err)
	}
	decodedValue, err := hex.DecodeString("f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b")
	if err != nil {
		t.Fatalf("Failed to decode storage value hex string: %v", err)
	}
	assert.Equal(t, types.BytesToHash(decodedValue), account2.Storage[(types.BytesToHash(decodedKey))])

	account3 := worldState[types.HexToAddress("fbfd6fa9f73ac6a058e01259034c28001bef8247")]
	assert.Equal(t, "0x52b7d2dcc80cd2e4000000", account3.Balance.Hex())
	code, err := hex.DecodeString("60806040526004361061003f5760003560e01c")
	if err != nil {
		t.Fatalf("Failed to decode code hex string: %v", err)
	}
	assert.Equal(t, code, account3.Code)
}
