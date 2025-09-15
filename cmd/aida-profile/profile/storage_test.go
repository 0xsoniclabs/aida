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

package profile

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunGetStorageUpdateSizeCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir(path.Join(testDataDir, "sample-pb-db"), aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&GetStorageUpdateSizeCommand}
	args := utils.NewArgs("test").
		Arg(GetStorageUpdateSizeCommand.Name).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.WorkersFlag.Name, 1).
		Arg("1").
		Arg("1000").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
