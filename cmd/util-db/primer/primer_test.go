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

package primer

import (
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunPrimerCmd(t *testing.T) {
	// given - basic priming test with default settings
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
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
