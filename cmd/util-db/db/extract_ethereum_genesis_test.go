package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCmd_ExtractEthereumGenesisCommand(t *testing.T) {
	// given - create a simple test genesis.json file
	tmpDir := t.TempDir()
	genesisFile := filepath.Join(tmpDir, "genesis.json")
	updateDbPath := filepath.Join(tmpDir, "update-db")

	// Create a minimal genesis.json file for testing
	genesisContent := `{
	"config": {
		"chainId": 1,
		"homesteadBlock": 0,
		"byzantiumBlock": 0,
		"constantinopleBlock": 0,
		"petersburgBlock": 0,
		"istanbulBlock": 0
	},
	"alloc": {
		"0x1000000000000000000000000000000000000000": {
			"balance": "0xde0b6b3a7640000"
		}
	},
	"coinbase": "0x0000000000000000000000000000000000000000",
	"difficulty": "0x400",
	"extraData": "0x",
	"gasLimit": "0x2fefd8",
	"nonce": "0x0000000000000042",
	"mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"timestamp": "0x0"
}`

	err := os.WriteFile(genesisFile, []byte(genesisContent), 0644)
	assert.NoError(t, err, "Should be able to create genesis file")

	// Setup CLI app and command
	app := cli.NewApp()
	app.Commands = []*cli.Command{&ExtractEthereumGenesisCommand}

	args := utils.NewArgs("test").
		Arg(ExtractEthereumGenesisCommand.Name).
		Flag(utils.UpdateDbFlag.Name, updateDbPath).
		Flag(utils.ChainIDFlag.Name, 1).
		Arg(genesisFile). // path to genesis.json file
		Build()

	// when
	err = app.Run(args)

	// then
	assert.NoError(t, err)

	// verify update database was created
	_, err = os.Stat(updateDbPath)
	assert.NoError(t, err)
}
