package db

import (
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_GenerateDbHashCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&GenerateDbHashCommand}

	args := utils.NewArgs("test").
		Arg(GenerateDbHashCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_GenerateDbHashCommandError(t *testing.T) {
	t.Run("cannot open db", func(t *testing.T) {
		// given
		app := cli.NewApp()
		app.Commands = []*cli.Command{&GenerateDbHashCommand}

		args := utils.NewArgs("test").
			Arg(GenerateDbHashCommand.Name).
			Flag(utils.AidaDbFlag.Name, "").
			Build()

		// when
		err := app.Run(args)

		// then
		assert.Error(t, err)
	})

}

func TestCmd_ValidateCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&ValidateCommand}

	args := utils.NewArgs("test").
		Arg(ValidateCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_ValidateCommandError(t *testing.T) {

	t.Run("cannot parse config", func(t *testing.T) {
		app := cli.NewApp()
		app.Commands = []*cli.Command{&ValidateCommand}
		ValidateCommand.Flags = append(ValidateCommand.Flags, &utils.ChainIDFlag)

		args := utils.NewArgs("test").
			Arg(ValidateCommand.Name).
			Flag(utils.ChainIDFlag.Name, 9990099).
			Flag(utils.AidaDbFlag.Name, "").
			Build()

		// when
		err := app.Run(args)

		// then
		assert.Error(t, err)
	})

	t.Run("cannot open db", func(t *testing.T) {
		app := cli.NewApp()
		app.Commands = []*cli.Command{&ValidateCommand}

		args := utils.NewArgs("test").
			Arg(ValidateCommand.Name).
			Flag(utils.AidaDbFlag.Name, "").
			Build()

		// when
		err := app.Run(args)

		// then
		assert.Error(t, err)
	})

}
