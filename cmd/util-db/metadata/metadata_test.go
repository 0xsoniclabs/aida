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

package metadata

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCmd_MetadataCommand(t *testing.T) {
	// given - test main metadata command structure
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	// Test with help flag to verify command structure without executing subcommands
	args := utils.NewArgs("test").
		Arg(Command.Name).
		Flag("help", true).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_PrintMetadataCommand(t *testing.T) {
	// given
	_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Arg(printCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_GenerateMetadataCommand(t *testing.T) {
	// given
	_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Arg(generateCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.OperaMainnetChainID)).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_GenerateMetadataCommand_EmptyAidaDb(t *testing.T) {
	aidaDbPath := t.TempDir() + "/empty.db"
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Arg(generateCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.OperaMainnetChainID)).
		Build()

	// when
	err := app.Run(args)

	// then
	require.ErrorContains(t, err, "cannot find block range in substate")
}

func TestCmd_InsertMetadataCommand(t *testing.T) {
	// given
	_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}
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
			Arg(Command.Name).
			Arg(insertCommand.Name).
			Flag(utils.AidaDbFlag.Name, aidaDbPath).
			Arg(param[2:]).
			Arg(params[param]).
			Build()

		err := app.Run(args)

		// then
		assert.NoError(t, err)
	}
}

func TestCmd_InsertMetadataCommand_Errors(t *testing.T) {
	_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)

	tests := []struct {
		name        string
		argsBuilder *utils.ArgsBuilder
		wantErr     string
	}{
		{
			name: "NotEnoughArguments",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name).
				Arg(insertCommand.Name).
				Flag(utils.AidaDbFlag.Name, aidaDbPath),
			wantErr: "this command requires two arguments",
		},
		{
			name: "UnknownKey",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name).
				Arg(insertCommand.Name).
				Flag(utils.AidaDbFlag.Name, aidaDbPath).
				Arg("unknownKey").
				Arg("123"),
			wantErr: "incorrect keyArg: unknownKey",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			app := cli.NewApp()
			app.Commands = []*cli.Command{&Command}
			err := app.Run(test.argsBuilder.Build())

			// then
			assert.ErrorContains(t, err, test.wantErr)
		})
	}

}

func TestCmd_InsertMetadataCommand_IncorrectArguments(t *testing.T) {
	// given
	_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}
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
			Arg(Command.Name).
			Arg(insertCommand.Name).
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
	_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Arg(removeCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
