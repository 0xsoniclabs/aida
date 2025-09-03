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

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/urfave/cli/v2"
)

// RunArchiveApp defines metadata and configuration options the vm-adb executable.
var RunArchiveApp = cli.App{
	Action:    RunVmAdb,
	Name:      "Aida Archive Evaluation Tool",
	HelpName:  "vm-adb",
	Usage:     "run VM on the archive",
	Copyright: "(c) 2025 Sonic Labs",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	// TODO: derive supported flags from utilized executor extensions (issue #664).
	Flags: []cli.Flag{
		// substate
		&config.WorkersFlag,

		// utils
		&config.CpuProfileFlag,
		&config.ChainIDFlag,
		&logger.LogLevelFlag,
		&config.StateDbLoggingFlag,
		&config.TrackProgressFlag,
		&config.NoHeartbeatLoggingFlag,
		&config.ErrorLoggingFlag,

		// StateDb
		&config.AidaDbFlag,
		&config.StateDbSrcFlag,
		&config.ValidateTxStateFlag,
		&config.ValidateFlag,

		// ShadowDb
		&config.ShadowDb,

		// VM
		&config.VmImplementation,
		&config.EvmImplementation,
	},
	Description: "Runs transactions on historic states derived from an archive DB",
}

// main implements vm-sdb cli.
func main() {
	if err := RunArchiveApp.Run(os.Args); err != nil {
		code := 1
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}
