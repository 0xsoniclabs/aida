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
	"flag"
	"path"
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
