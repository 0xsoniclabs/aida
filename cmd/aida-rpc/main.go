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
	"log"
	"os"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/urfave/cli/v2"
)

var rpcApp = &cli.App{
	Action: RunRpc,
	Name:   "Replay-RPC",
	Usage: "Sends real API requests recorded on rpcapi.fantom.network to StateDB then compares recorded" +
		"result with result returned by DB.",
	Copyright: "(c) 2025 Sonic Labs",
	Flags: []cli.Flag{
		&config.RpcRecordingFileFlag,
		&config.WorkersFlag,

		// VM
		&config.VmImplementation,

		// Config
		&logger.LogLevelFlag,
		&config.ChainIDFlag,
		&config.ContinueOnFailureFlag,
		&config.ValidateFlag,
		&config.NoHeartbeatLoggingFlag,
		&config.ErrorLoggingFlag,
		&config.TrackProgressFlag,

		// Register
		&config.RegisterRunFlag,
		&config.OverwriteRunIdFlag,

		// ShadowDB
		&config.ShadowDb,

		// StateDB
		&config.StateDbSrcFlag,
		&config.StateDbLoggingFlag,

		// Trace
		&config.TraceFlag,
		&config.TraceFileFlag,
		&config.TraceDebugFlag,

		// Performance
		&config.CpuProfileFlag,
		&config.MemoryProfileFlag,
		&config.ProfileFlag,
		&config.ProfileFileFlag,
	},
}

func main() {
	if err := rpcApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
