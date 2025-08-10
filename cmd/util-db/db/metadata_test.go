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
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
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
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
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
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MetadataCommand}

	args := utils.NewArgs("test").
		Arg(MetadataCommand.Name).
		Arg(InsertMetadataCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Arg("ty"). // key argument
		Arg("0").  // value argument
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_RemoveMetadataCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
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
