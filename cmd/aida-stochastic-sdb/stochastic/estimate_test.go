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
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunGetAddressStatsCommand(t *testing.T) {
	// given
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test_events.json")
	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticEstimateCommand}
	args := utils.NewArgs("test").
		Arg(StochasticEstimateCommand.Name).
		Flag(utils.OutputFlag.Name, outputFile).
		Arg(path.Join(testDataDir, "events.json")).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	stat, err := os.Stat(outputFile)
	require.NoError(t, err)
	assert.NotZero(t, stat.Size())
}
