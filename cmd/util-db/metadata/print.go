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
	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/urfave/cli/v2"
)

var printCommand = cli.Command{
	Action: printAction,
	Name:   "print",
	Usage:  "Prints metadata",
	Flags: []cli.Flag{
		&config.AidaDbFlag,
	},
}

func printAction(ctx *cli.Context) error {
	cfg, argErr := config.NewConfig(ctx, config.NoArgs)
	if argErr != nil {
		return argErr
	}

	return utildb.PrintMetadata(cfg.AidaDb)
}
