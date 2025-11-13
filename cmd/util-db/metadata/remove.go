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

package metadata

import (
	"errors"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// removeCommand is a command used for creating testing environment without metadata
var removeCommand = cli.Command{
	Action: removeAction,
	Name:   "remove",
	Usage:  "remove metadata from aidaDb",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
	Description: `
Removes block and epoch range and ChainID from metadata for given AidaDb.
`,
}

// removeAction command is used for testing scenario where AidaDb does not have metadata and a patch
// is applied onto it
func removeAction(ctx *cli.Context) (finalErr error) {
	aidaDbPath := ctx.String(utils.AidaDbFlag.Name)

	// open db
	base, err := db.NewDefaultSubstateDB(aidaDbPath)
	if err != nil {
		return err
	}

	defer func() {
		finalErr = errors.Join(finalErr, base.Close())
	}()
	md := utils.NewAidaDbMetadata(base, "DEBUG")
	md.DeleteMetadata()

	return nil
}
