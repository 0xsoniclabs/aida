package db

import (
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_GenDeletedAccountsCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	deletionDbPath := filepath.Join(tempDir, "deletion-db")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&GenDeletedAccountsCommand}

	args := utils.NewArgs("test").
		Arg(GenDeletedAccountsCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.OutputFlag.Name, deletionDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Flag(utils.WorkersFlag.Name, 1).
		Arg("1").
		Arg("5").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	assert.DirExists(t, deletionDbPath)
}

func TestCmd_GenDeletedAccountsCommandError(t *testing.T) {
	t.Run("no block range", func(t *testing.T) {
		app := cli.NewApp()
		app.Commands = []*cli.Command{&GenDeletedAccountsCommand}

		args := utils.NewArgs("test").
			Arg(GenDeletedAccountsCommand.Name).
			Flag(utils.AidaDbFlag.Name, "").
			Arg("ab").
			Arg("cd").
			Build()

		// when
		err := app.Run(args)

		// then
		assert.Error(t, err)
	})

	t.Run("no output", func(t *testing.T) {
		app := cli.NewApp()
		app.Commands = []*cli.Command{&GenDeletedAccountsCommand}

		args := utils.NewArgs("test").
			Arg(GenDeletedAccountsCommand.Name).
			Flag(utils.AidaDbFlag.Name, "").
			Arg("1").
			Arg("100").
			Build()

		// when
		err := app.Run(args)

		// then
		assert.Error(t, err)
	})

	t.Run("no aida", func(t *testing.T) {

		app := cli.NewApp()
		app.Commands = []*cli.Command{&GenDeletedAccountsCommand}

		args := utils.NewArgs("test").
			Arg(GenDeletedAccountsCommand.Name).
			Flag(utils.OutputFlag.Name, "abcd").
			Flag(utils.AidaDbFlag.Name, "").
			Arg("1").
			Arg("100").
			Build()

		// when
		err := app.Run(args)

		// then
		assert.Error(t, err)
	})
}
