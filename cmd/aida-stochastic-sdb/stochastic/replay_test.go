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

package stochastic

import (
	"bufio"
	"flag"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunStochasticReplayCommand(t *testing.T) {
	// given
	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticReplayCommand}
	args := utils.NewArgs("test").
		Arg(StochasticReplayCommand.Name).
		Flag(utils.BalanceRangeFlag.Name, 100).
		Flag(utils.NonceRangeFlag.Name, 100).
		Flag(utils.MemoryBreakdownFlag.Name, true).
		Flag(utils.ShadowDbImplementationFlag.Name, "geth").
		Flag(utils.StateDbImplementationFlag.Name, "carmen").
		Flag(utils.StateDbVariantFlag.Name, "go-file").
		Arg(10).
		Arg(path.Join(testDataDir, "stats.json")).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestStochasticReplayCommand_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing args", args: []string{}, wantErr: "missing simulation file"},
		{name: "non integer length", args: []string{"not-a-number", "stats.json"}, wantErr: "simulation length is not an integer"},
		{name: "non positive length", args: []string{"0", "stats.json"}, wantErr: "simulation length must be greater than zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			require.NoError(t, fs.Parse(tt.args))

			ctx := cli.NewContext(cli.NewApp(), fs, nil)
			err := stochasticReplayAction(ctx)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCmd_RunStochasticReplayCommandWithDbLogging(t *testing.T) {
	tempDir := t.TempDir()
	dbLogFile := filepath.Join(tempDir, "db-trace.txt")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticReplayCommand}
	args := utils.NewArgs("test").
		Arg(StochasticReplayCommand.Name).
		Flag(utils.BalanceRangeFlag.Name, 100).
		Flag(utils.NonceRangeFlag.Name, 100).
		Flag(utils.StateDbLoggingFlag.Name, dbLogFile).
		Flag(utils.MemoryBreakdownFlag.Name, true).
		Flag(utils.StateDbImplementationFlag.Name, "geth").
		Arg(10).
		Arg(path.Join(testDataDir, "stats.json")).
		Build()

	err := app.Run(args)

	assert.NoError(t, err)

	traceStat, err := os.Stat(dbLogFile)
	require.NoError(t, err)
	assert.NotZero(t, traceStat.Size())

	f, err := os.Open(dbLogFile)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	foundBeginBlock := false
	foundEndBlock := false
	foundBeginSyncPeriod := false
	foundEndSyncPeriod := false
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if strings.HasPrefix(line, "BeginBlock") {
			foundBeginBlock = true
		}
		if strings.HasPrefix(line, "EndBlock") {
			foundEndBlock = true
		}
		if strings.HasPrefix(line, "BeginSyncPeriod") {
			foundBeginSyncPeriod = true
		}
		if strings.HasPrefix(line, "EndSyncPeriod") {
			foundEndSyncPeriod = true
		}
	}
	require.NoError(t, scanner.Err())

	assert.True(t, foundBeginBlock, "Trace should contain BeginBlock operation")
	assert.True(t, foundEndBlock, "Trace should contain EndBlock operation")
	assert.True(t, foundBeginSyncPeriod, "Trace should contain BeginSyncPeriod operation")
	assert.True(t, foundEndSyncPeriod, "Trace should contain EndSyncPeriod operation")
	assert.Greater(t, lineCount, 0, "Trace file should not be empty")
}
