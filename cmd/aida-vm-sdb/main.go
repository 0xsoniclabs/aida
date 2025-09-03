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

// RunVMApp data structure
var RunVMApp = cli.App{
	Name:      "Aida Storage Run VM Manager",
	Copyright: "(c) 2025 Sonic Labs",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Commands: []*cli.Command{
		&RunSubstateCmd,
		&RunEthTestsCmd,
		&RunTxGeneratorCmd,
	},
	Description: `
The aida-vm-sdb command requires two arguments: <blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and last block of
the inclusive range of blocks.`,
}

var RunSubstateCmd = cli.Command{
	Action:    RunSubstate,
	Name:      "substate",
	Usage:     "Iterates over substates that are executed into a StateDb",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		// AidaDb
		&config.AidaDbFlag,

		// StateDb
		&config.CarmenCheckpointInterval,
		&config.CarmenCheckpointPeriod,
		&config.CarmenSchemaFlag,
		&config.StateDbImplementationFlag,
		&config.StateDbVariantFlag,
		&config.StateDbSrcFlag,
		&config.StateDbSrcOverwriteFlag,
		&config.DbTmpFlag,
		&config.StateDbLoggingFlag,
		&config.ValidateStateHashesFlag,

		// ArchiveDb
		&config.ArchiveModeFlag,
		&config.ArchiveQueryRateFlag,
		&config.ArchiveMaxQueryAgeFlag,
		&config.ArchiveVariantFlag,

		// ShadowDb
		&config.ShadowDb,
		&config.ShadowDbImplementationFlag,
		&config.ShadowDbVariantFlag,

		// VM
		&config.EvmImplementation,
		&config.VmImplementation,

		// Profiling
		&config.CpuProfileFlag,
		&config.CpuProfilePerIntervalFlag,
		&config.DiagnosticServerFlag,
		&config.MemoryBreakdownFlag,
		&config.MemoryProfileFlag,
		&config.RandomSeedFlag,
		&config.PrimeThresholdFlag,
		&config.ProfileFlag,
		&config.ProfileDepthFlag,
		&config.ProfileFileFlag,
		&config.ProfileSqlite3Flag,
		&config.ProfileIntervalFlag,
		&config.ProfileDBFlag,
		&config.ProfileBlocksFlag,

		// RegisterRun
		&config.RegisterRunFlag,
		&config.OverwriteRunIdFlag,

		// Priming
		&config.RandomizePrimingFlag,
		&config.SkipPrimingFlag,
		&config.UpdateBufferSizeFlag,

		// Utils
		&config.WorkersFlag,
		&config.ChainIDFlag,
		&config.ContinueOnFailureFlag,
		&config.SyncPeriodLengthFlag,
		&config.KeepDbFlag,
		&config.CustomDbNameFlag,
		//&utils.MaxNumTransactionsFlag,
		&config.ValidateTxStateFlag,
		&config.ValidateFlag,
		&config.OverwritePreWorldStateFlag,
		&logger.LogLevelFlag,
		&config.NoHeartbeatLoggingFlag,
		&config.TrackProgressFlag,
		&config.ErrorLoggingFlag,
		&config.TrackerGranularityFlag,
		&config.SubstateEncodingFlag,
	},
	Description: `
The aida-vm-sdb substate command requires two arguments: <blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and last block of
the inclusive range of blocks.`,
}

var RunTxGeneratorCmd = cli.Command{
	Action: RunTxGenerator,
	Name:   "tx-generator",
	Usage:  "Generates transactions for specified block range and executes them over StateDb",
	Flags: []cli.Flag{
		// TxGenerator specific flags
		&config.TxGeneratorTypeFlag,

		// StateDb
		&config.CarmenSchemaFlag,
		&config.StateDbImplementationFlag,
		&config.StateDbVariantFlag,
		&config.StateDbSrcFlag,
		&config.StateDbSrcOverwriteFlag,
		&config.DbTmpFlag,
		&config.StateDbLoggingFlag,
		&config.ValidateStateHashesFlag,

		// ShadowDb
		&config.ShadowDb,
		&config.ShadowDbImplementationFlag,
		&config.ShadowDbVariantFlag,

		// RegisterRun
		&config.RegisterRunFlag,
		&config.OverwriteRunIdFlag,

		// VM
		&config.EvmImplementation,
		&config.VmImplementation,

		// Profiling
		&config.CpuProfileFlag,
		&config.CpuProfilePerIntervalFlag,
		&config.DiagnosticServerFlag,
		&config.MemoryBreakdownFlag,
		&config.MemoryProfileFlag,

		// Utils
		&config.WorkersFlag,
		&config.ChainIDFlag,
		&config.ContinueOnFailureFlag,
		&config.KeepDbFlag,
		&config.ValidateFlag,
		&logger.LogLevelFlag,
		&config.NoHeartbeatLoggingFlag,
		&config.BlockLengthFlag,
		&config.TrackerGranularityFlag,
		&config.ForkFlag,
	},
	Description: `
The aida-vm-sdb tx-generator command requires two arguments: <blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and last block of
the inclusive range of blocks.`,
}

// main implements vm-sdb cli.
func main() {
	if err := RunVMApp.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
