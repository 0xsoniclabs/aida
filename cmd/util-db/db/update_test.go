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

package db

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_UpdateCommand(t *testing.T) {
	// given
	tmpDir := t.TempDir()
	aidaDbPath := filepath.Join(tmpDir, "aida-db")
	tmpDbPath := filepath.Join(tmpDir, "tmp-db")
	require.NoError(t, os.Mkdir(tmpDbPath, os.ModePerm))

	app := cli.NewApp()
	app.Commands = []*cli.Command{&UpdateCommand}

	args := utils.NewArgs("test").
		Arg(UpdateCommand.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Flag(utils.DbTmpFlag.Name, tmpDbPath).
		Flag(utils.UpdateTypeFlag.Name, "stable").
		Flag(utils.SubstateEncodingFlag.Name, "protobuf").
		Build()

	// Create a context with cancellation to control the app execution
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to communicate the result of app.Run
	resultChan := make(chan error, 1)

	// Run the app in a goroutine
	go func() {
		err := app.Run(args)
		resultChan <- err
	}()

	// Monitor tmp-db folder for gz files
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(1 * time.Minute) // 1 minute timeout
	defer timeout.Stop()

	for {
		select {
		case err := <-resultChan:
			// App completed naturally
			assert.NoError(t, err)
			return
		case <-ticker.C:
			// Check for gz files in tmp-db folder
			if hasGzFile(tmpDbPath) {
				// Found gz file, cancel context and terminate test successfully
				cancel()
				t.Log("Found gz file in tmp-db, terminating test early")
				return
			}
		case <-timeout.C:
			// Timeout reached
			cancel()
			t.Fatal("Test timed out waiting for gz file or completion")
		case <-ctx.Done():
			// Context cancelled
			return
		}
	}
}

// hasGzFile checks if there are any .gz files in the specified directory
func hasGzFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gz") {
			return true
		}
	}
	return false
}
