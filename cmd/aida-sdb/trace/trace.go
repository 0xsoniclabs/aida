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

package trace

import (
	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/urfave/cli/v2"
)

// TraceReplayCommand data structure for the replay app
var TraceReplayCommand = cli.Command{
	Action:    ReplayTrace,
	Name:      "replay",
	Usage:     "executes storage trace",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&config.CarmenSchemaFlag,
		&config.ChainIDFlag,
		&config.CpuProfileFlag,
		&config.SyncPeriodLengthFlag,
		&config.KeepDbFlag,
		&config.MemoryBreakdownFlag,
		&config.MemoryProfileFlag,
		&config.RandomSeedFlag,
		&config.PrimeThresholdFlag,
		&config.ProfileFlag,
		&config.ProfileFileFlag,
		&config.ProfileIntervalFlag,
		&config.RandomizePrimingFlag,
		&config.SkipPrimingFlag,
		&config.StateDbImplementationFlag,
		&config.StateDbVariantFlag,
		&config.StateDbSrcFlag,
		&config.StateDbSrcOverwriteFlag,
		&config.VmImplementation,
		&config.DbTmpFlag,
		&config.UpdateBufferSizeFlag,
		&config.StateDbLoggingFlag,
		&config.ShadowDb,
		&config.ShadowDbImplementationFlag,
		&config.ShadowDbVariantFlag,
		&config.WorkersFlag,
		&config.TraceFileFlag,
		&config.TraceDirectoryFlag,
		&config.TraceDebugFlag,
		&config.DebugFromFlag,
		//&utils.ValidateFlag,
		//&utils.ValidateTxStateFlag,
		&config.AidaDbFlag,
		&logger.LogLevelFlag,
	},
	Description: `
The trace replay command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay storage traces.`,
}

// TraceReplaySubstateCommand data structure for the replay-substate app
var TraceReplaySubstateCommand = cli.Command{
	Action:    ReplaySubstate,
	Name:      "replay-substate",
	Usage:     "executes storage trace using substates",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&config.ChainIDFlag,
		&config.CpuProfileFlag,
		&config.RandomizePrimingFlag,
		&config.RandomSeedFlag,
		&config.PrimeThresholdFlag,
		&config.ProfileFlag,
		&config.StateDbImplementationFlag,
		&config.StateDbVariantFlag,
		&config.StateDbLoggingFlag,
		&config.ShadowDbImplementationFlag,
		&config.ShadowDbVariantFlag,
		&config.SyncPeriodLengthFlag,
		&config.WorkersFlag,
		&config.TraceFileFlag,
		&config.TraceDirectoryFlag,
		&config.TraceDebugFlag,
		&config.DebugFromFlag,
		//&utils.ValidateFlag,
		//&utils.ValidateTxStateFlag,
		&config.AidaDbFlag,
		&logger.LogLevelFlag,
	},
	Description: `
The trace replay-substate command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay storage traces.`,
}
