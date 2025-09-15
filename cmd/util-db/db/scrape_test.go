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
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCmd_ScrapeCommand(t *testing.T) {
	// given
	tmpDir := t.TempDir()
	targetDbPath := filepath.Join(tmpDir, "target-db")
	clientDbPath := filepath.Join(tmpDir, "client-db")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&ScrapeCommand}

	args := utils.NewArgs("test").
		Arg(ScrapeCommand.Name).
		Flag(utils.TargetDbFlag.Name, targetDbPath).
		Flag(utils.ClientDbFlag.Name, clientDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Arg("1"). // blockNumFirst
		Arg("5"). // blockNumLast - small range for testing
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
