package generate

import (
	"encoding/hex"
	"os"
	"strconv"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGenerate_GenerateDeletedAccountsCommand(t *testing.T) {
	// Protobuf is not yet supported
	ss, sdbPath := utils.CreateTestSubstateDb(t, db.RLPEncodingSchema)
	ddbPath := t.TempDir()
	tests := []struct {
		name    string
		wantErr string
		args    []string
	}{
		{
			name: "Success",
			// Transaction fail
			wantErr: "intrinsic gas too low",
			args: []string{
				generateDeletedAccountsCommand.Name,
				"--aida-db",
				sdbPath,
				"--deletion-db",
				ddbPath,
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
			},
		},
		{
			name:    "NoDeletionDb",
			wantErr: "you need to specify where you want deletion-db to save (--deletion-db)",
			args: []string{
				generateDeletedAccountsCommand.Name,
				"--aida-db",
				sdbPath,
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
			},
		},
		{
			name:    "IncorrectConfig",
			wantErr: "cannot set chain id: unknown chain id 11111",
			args: []string{
				generateDeletedAccountsCommand.Name,
				"--aida-db",
				sdbPath,
				"--chainid",
				strconv.FormatInt(11111, 10),
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			app.Action = generateDeletedAccountsCommand.Action
			app.Flags = generateDeletedAccountsCommand.Flags

			err := app.Run(test.args)
			require.ErrorContains(t, err, test.wantErr)
		})
	}

}

func TestGenerateDbHash_Command(t *testing.T) {
	_, path := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Action = generateDbHashCommand.Action
	app.Flags = generateDbHashCommand.Flags

	err := app.Run([]string{generateDbHashCommand.Name, "--aida-db", path})
	require.NoError(t, err)

	aidaDb, err := db.NewDefaultBaseDB(path)
	require.NoError(t, err)
	md := utils.NewAidaDbMetadata(aidaDb, "CRITICAL")
	got := md.GetDbHash()
	require.Equal(t, "a0d4f7616f3007bf8c02f816a60b2526", hex.EncodeToString(got))
	err = aidaDb.Close()
	require.NoError(t, err)
}

func TestExtractEthereumGenesis_Command(t *testing.T) {
	// Create a temporary file to store the genesis JSON
	genesisPath := t.TempDir() + "genesis.json"
	genesisFile, err := os.Create(genesisPath)
	require.NoError(t, err)
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

	_, err = genesisFile.WriteString(genesisData)
	require.NoError(t, err)
	err = genesisFile.Close()
	require.NoError(t, err)

	udbPath := t.TempDir()
	app := cli.NewApp()
	app.Action = generateEthereumGenesisCommand.Action
	app.Flags = generateEthereumGenesisCommand.Flags

	err = app.Run([]string{
		generateEthereumGenesisCommand.Name,
		"--update-db",
		udbPath,
		"-l",
		"CRITICAL",
		"--chainid",
		strconv.FormatInt(int64(utils.EthereumChainID), 10),
		genesisPath,
	})
	require.NoError(t, err)

	udb, err := db.NewDefaultUpdateDB(udbPath)
	require.NoError(t, err)
	updateSet, err := udb.GetUpdateSet(0)
	require.NoError(t, err)
	_, exists := updateSet.WorldState[types.HexToAddress("0x0000000000000000000000000000000000000000")]
	assert.True(t, exists)
	_, exists = updateSet.WorldState[types.HexToAddress("0xefa7454f1116807975a4750b46695e967850de5d")]
	assert.True(t, exists)
	_, exists = updateSet.WorldState[types.HexToAddress("0xfbfd6fa9f73ac6a058e01259034c28001bef8247")]
	assert.True(t, exists)
}

func TestExtractEthereumGenesis_Command_Error(t *testing.T) {
	genesisPath := t.TempDir() + "wrong.json"
	genesisFile, err := os.Create(genesisPath)
	require.NoError(t, err)
	_, err = genesisFile.WriteString("some text")
	require.NoError(t, err)
	err = genesisFile.Close()
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "ArgNotPassed",
			args: []string{
				generateEthereumGenesisCommand.Name,
				"-l",
				"CRITICAL",
				"--chainid",
				strconv.FormatInt(int64(utils.EthereumChainID), 10),
			},
			wantErr: "ethereum-update command requires exactly 1 argument",
		},
		{
			name: "IncorrectConfig",
			args: []string{
				generateEthereumGenesisCommand.Name,
				"-l",
				"CRITICAL",
				"--chainid",
				strconv.FormatInt(11111, 10),
				genesisPath,
			},
			wantErr: "cannot set chain id: unknown chain id 11111",
		},
		{
			name: "WrongJsonFormat",
			args: []string{
				generateEthereumGenesisCommand.Name,
				"-l",
				"CRITICAL",
				"--chainid",
				strconv.FormatInt(int64(utils.EthereumChainID), 10),
				"--update-db",
				t.TempDir() + "/update.db",
				genesisPath,
			},
			wantErr: "failed to unmarshal genesis file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			app.Action = generateEthereumGenesisCommand.Action
			app.Flags = generateEthereumGenesisCommand.Flags

			err = app.Run(test.args)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}
