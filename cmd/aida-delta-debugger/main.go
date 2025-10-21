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

package main

import (
	"fmt"
	"os"

	log "github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "Aida Delta Debugger",
		HelpName:  "aida-delta-debugger",
		Usage:     "minimize failing state traces via delta debugging",
		Copyright: "(c) 2025 Sonic Labs",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "trace-file",
				Usage:   "path to a trace file (repeatable)",
				Aliases: []string{"f"},
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "write the minimized trace to the given path",
			},
			&cli.IntFlag{
				Name:  "address-sample-runs",
				Usage: "number of attempts per sampling factor when reducing contracts",
				Value: 5,
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Usage: "overall timeout for the minimization run",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "print progress information for each test run",
			},
			&cli.Int64Flag{
				Name:  "seed",
				Usage: "seed for random address sampling (default: time-based)",
			},
			&cli.IntFlag{
				Name:  "max-factor",
				Usage: "maximum sampling factor when reducing addresses",
			},
			&utils.StateDbImplementationFlag,
			&utils.StateDbVariantFlag,
			&utils.CarmenSchemaFlag,
			&utils.DbTmpFlag,
			&utils.ChainIDFlag,
			&log.LogLevelFlag,
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
