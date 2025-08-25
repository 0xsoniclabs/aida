package db

import (
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_MetadataCommand(t *testing.T) {
	// given - test main metadata command structure
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MetadataCommand}

	// Test with help flag to verify command structure without executing subcommands
	args := utils.NewArgs("test").
		Arg(MetadataCommand.Name).
		Flag("help", true).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_PrintMetadataCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MetadataCommand}

	args := utils.NewArgs("test").
		Arg(MetadataCommand.Name).
		Arg(cmdPrintMetadata.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_GenerateMetadataCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MetadataCommand}

	args := utils.NewArgs("test").
		Arg(MetadataCommand.Name).
		Arg(cmdGenerateMetadata.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_InsertMetadataCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MetadataCommand}
	params := map[string]string{
		utils.FirstBlockPrefix:  "0",
		utils.LastBlockPrefix:   "0",
		utils.FirstEpochPrefix:  "0",
		utils.LastEpochPrefix:   "0",
		utils.TypePrefix:        "0",
		utils.ChainIDPrefix:     "0",
		utils.TimestampPrefix:   "0",
		utils.DbHashPrefix:      "1234",
		db.UpdatesetIntervalKey: "0",
		db.UpdatesetSizeKey:     "0",
	}
	for param := range params {
		args := utils.NewArgs("test").
			Arg(MetadataCommand.Name).
			Arg(InsertMetadataCommand.Name).
			Flag(utils.AidaDbFlag.Name, aidaDbPath).
			Arg(param[2:]).
			Arg(params[param]).
			Build()

		err := app.Run(args)

		// then
		assert.NoError(t, err)
	}
}

func TestCmd_InsertMetadataCommandError(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MetadataCommand}
	params := map[string]string{
		utils.FirstBlockPrefix:  "a",
		utils.LastBlockPrefix:   "b",
		utils.FirstEpochPrefix:  "c",
		utils.LastEpochPrefix:   "d",
		utils.TypePrefix:        "e",
		utils.ChainIDPrefix:     "f",
		utils.DbHashPrefix:      "0",
		db.UpdatesetIntervalKey: "h",
		db.UpdatesetSizeKey:     "i",
	}
	for param := range params {
		args := utils.NewArgs("test").
			Arg(MetadataCommand.Name).
			Arg(InsertMetadataCommand.Name).
			Flag(utils.AidaDbFlag.Name, aidaDbPath).
			Arg(param[2:]).
			Arg(params[param]).
			Build()

		err := app.Run(args)

		// then
		assert.Error(t, err)
	}
}

func TestCmd_RemoveMetadataCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MetadataCommand}

	args := utils.NewArgs("test").
		Arg(MetadataCommand.Name).
		Arg(RemoveMetadataCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
