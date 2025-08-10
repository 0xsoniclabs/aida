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
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_CloneCommand(t *testing.T) {
	// given - test main clone command structure
	app := cli.NewApp()
	app.Commands = []*cli.Command{&CloneCommand}

	// Test with help flag to verify command structure without executing subcommands
	args := utils.NewArgs("test").
		Arg(CloneCommand.Name).
		Flag("help", true).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_ClonePatchCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	targetDb := filepath.Join(tempDir, "target-patch-db")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&CloneCommand}

	args := utils.NewArgs("test").
		Arg(CloneCommand.Name).
		Arg(ClonePatch.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.TargetDbFlag.Name, targetDb).
		Arg("1").   // blockNumFirst
		Arg("100"). // blockNumLast
		Arg("0").   // epochNumFirst
		Arg("10").  // epochNumLast
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)

	// verify target database was created
	_, err = os.Stat(targetDb)
	assert.NoError(t, err)
}

func TestCmd_CloneDbCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	targetDb := filepath.Join(tempDir, "target-clone-db")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&CloneCommand}

	args := utils.NewArgs("test").
		Arg(CloneCommand.Name).
		Arg(CloneDb.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.TargetDbFlag.Name, targetDb).
		Arg("1").   // blockNumFirst
		Arg("100"). // blockNumLast
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)

	// verify target database was created
	_, err = os.Stat(targetDb)
	assert.NoError(t, err)
}

func TestCmd_CloneCustomCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	targetDb := filepath.Join(tempDir, "target-custom-db")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&CloneCommand}

	args := utils.NewArgs("test").
		Arg(CloneCommand.Name).
		Arg(CloneCustom.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.TargetDbFlag.Name, targetDb).
		Flag(utils.DbComponentFlag.Name, "substate"). // specify component to clone
		Arg("1").                                     // blockNumFirst
		Arg("100").                                   // blockNumLast
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)

	// verify target database was created
	_, err = os.Stat(targetDb)
	assert.NoError(t, err)
}
