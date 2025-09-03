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

var runVmApp = &cli.App{
	Action:    RunVm,
	Name:      "EVM evaluation tool",
	HelpName:  "aida-vm",
	Copyright: "(c) 2025 Sonic Labs",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	// TODO: derive supported flags from utilized executor extensions.
	Flags: []cli.Flag{
		&config.WorkersFlag,
		//&substate.SkipTransferTxsFlag,
		//&substate.SkipCallTxsFlag,
		//&substate.SkipCreateTxsFlag,
		&config.ChainIDFlag,
		//&utils.ProfileEVMCallFlag,
		//&utils.MicroProfilingFlag,
		//&utils.BasicBlockProfilingFlag,
		//&utils.ProfilingDbNameFlag,
		&config.ChannelBufferSizeFlag,
		&config.EvmImplementation,
		&config.VmImplementation,
		&config.ValidateTxStateFlag,
		&config.ValidateFlag,
		//&utils.OnlySuccessfulFlag,
		&config.CpuProfileFlag,
		&config.DiagnosticServerFlag,
		&config.AidaDbFlag,
		&logger.LogLevelFlag,
		&config.ErrorLoggingFlag,
		&config.StateDbImplementationFlag,
		&config.StateDbLoggingFlag,
		&config.CacheFlag,
		&config.SubstateEncodingFlag,
	},
}

func main() {
	if err := runVmApp.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
