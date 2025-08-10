// Copyright 2024 Fantom Foundation
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
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
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

func TestCmd_ValidateCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
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

func TestCmd_PrintDbHashCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&PrintDbHashCommand}

	args := utils.NewArgs("test").
		Arg(PrintDbHashCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_PrintTableHashCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&PrintTableHashCommand}

	args := utils.NewArgs("test").
		Arg(PrintTableHashCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.DbComponentFlag.Name, "substate").
		Flag(utils.SubstateEncodingFlag.Name, "protobuf").
		Arg("0").
		Arg("1000").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_PrintPrefixHashCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&PrintPrefixHashCommand}

	args := utils.NewArgs("test").
		Arg(PrintPrefixHashCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Arg("1s"). // prefix argument for substates
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
