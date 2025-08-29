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

package compact

import (
	"testing"

	"github.com/0xsoniclabs/substate/db"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_Compact(t *testing.T) {
	_, path := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	app := cli.NewApp()
	app.Action = compactAction
	app.Flags = []cli.Flag{
		&utils.TargetDbFlag,
	}

	err := app.Run([]string{Command.Name, "--target-db", path})
	require.NoError(t, err)
}
