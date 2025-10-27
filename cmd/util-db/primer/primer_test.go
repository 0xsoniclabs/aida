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

package primer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
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

func TestCmd_RunPrimerCmd(t *testing.T) {
	// given - basic priming test with default settings
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir(path.Join(testDataDir, "sample-pb-db"), aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&RunPrimerCmd}

	args := utils.NewArgs("test").
		Arg(RunPrimerCmd.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.StateDbImplementationFlag.Name, "carmen").
		Flag(utils.StateDbVariantFlag.Name, "go-file").
		Arg("100"). // block number to prime to
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
