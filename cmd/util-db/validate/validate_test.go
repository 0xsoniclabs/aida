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

package validate

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"os"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

var testDataDir string

func TestMain(m *testing.M) {
	fmt.Println("Performing global setup...")

	// setup
	tempDir, err := os.MkdirTemp("", "profile_test_*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	testDataDir = tempDir
	err = utils.DownloadTestDataset(testDataDir)
	fmt.Printf("Downloaded test data: %s\n", testDataDir)
	if err != nil {
		fmt.Printf("Failed to download test dataset: %v\n", err)
		_ = os.RemoveAll(testDataDir)
		os.Exit(1)
	}

	// run
	exitCode := m.Run()

	// teardown
	err = os.RemoveAll(testDataDir)
	if err != nil {
		fmt.Printf("Failed to remove temp dir: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Performing global teardown...")
	os.Exit(exitCode)
}

func TestCmd_ValidateCommand(t *testing.T) {
	// given
	_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_ValidateCommandError(t *testing.T) {
	tests := []struct {
		name        string
		argsBuilder *utils.ArgsBuilder
		wantErr     string
		setup       func(aidaDbPath string)
	}{
		{
			name: "CannotParseCfg",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name).
				Flag(utils.ChainIDFlag.Name, 9990099).
				Flag(utils.AidaDbFlag.Name, ""),
			wantErr: "cannot parse config",
			setup:   func(aidaDbPath string) {},
		},
		{
			name: "WrongAidaDbType",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name),
			wantErr: fmt.Sprintf("your db type (%v) cannot be validated", utils.NoType),
			setup: func(aidaDbPath string) {
				aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
				require.NoError(t, err)
				md := utils.NewAidaDbMetadata(aidaDb, "CRITICAL")
				err = md.SetDbType(utils.NoType)
				require.NoError(t, err)
				err = aidaDb.Close()
				require.NoError(t, err)
			},
		},
		{
			name: "NoDbHashFound",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name),
			wantErr: "could not find expected db hash",
			setup: func(aidaDbPath string) {
				aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
				require.NoError(t, err)
				md := utils.NewAidaDbMetadata(aidaDb, "CRITICAL")
				err = md.SetDbHash([]byte{})
				require.NoError(t, err)
				err = aidaDb.Close()
				require.NoError(t, err)
			},
		},
		{
			name: "WrongDbHash",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name),
			wantErr: "hashes are different",
			setup: func(aidaDbPath string) {
				aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
				require.NoError(t, err)
				md := utils.NewAidaDbMetadata(aidaDb, "CRITICAL")
				err = md.SetDbHash([]byte("wrong-hash"))
				require.NoError(t, err)
				err = aidaDb.Close()
				require.NoError(t, err)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, aidaDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
			test.setup(aidaDbPath)
			app := cli.NewApp()
			app.Commands = []*cli.Command{&Command}
			// when
			test.argsBuilder.Flag(utils.AidaDbFlag.Name, aidaDbPath)
			err := app.Run(test.argsBuilder.Build())

			// then
			require.ErrorContains(t, err, test.wantErr)
		})
	}

}

func TestCmd_ValidateCommandRealDatabase(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir(path.Join(testDataDir, "sample-pb-db"), aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_ValidateCommandRealDatabaseError(t *testing.T) {

	t.Run("cannot parse config", func(t *testing.T) {
		app := cli.NewApp()
		app.Commands = []*cli.Command{&Command}

		args := utils.NewArgs("test").
			Arg(Command.Name).
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
		app.Commands = []*cli.Command{&Command}

		args := utils.NewArgs("test").
			Arg(Command.Name).
			Flag(utils.AidaDbFlag.Name, "").
			Build()

		// when
		err := app.Run(args)

		// then
		assert.Error(t, err)
	})

}
