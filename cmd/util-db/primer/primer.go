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

package primer

import (
	"github.com/0xsoniclabs/aida/config"
	"github.com/urfave/cli/v2"

	"github.com/0xsoniclabs/aida/logger"
)

var RunPrimerCmd = cli.Command{
	Action:    RunPrimer,
	Name:      "priming",
	Usage:     "Performs priming of the specified database",
	ArgsUsage: "<blockNum>",
	Flags: []cli.Flag{
		// AidaDb
		&config.AidaDbFlag,

		// StateDb
		&config.CarmenSchemaFlag,
		&config.StateDbImplementationFlag,
		&config.StateDbVariantFlag,
		&config.StateDbSrcFlag,
		&config.DbTmpFlag,
		&config.StateDbLoggingFlag,

		// ArchiveDb
		&config.ArchiveModeFlag,
		&config.ArchiveQueryRateFlag,
		&config.ArchiveMaxQueryAgeFlag,
		&config.ArchiveVariantFlag,

		// Profiling
		&config.CpuProfileFlag,
		&config.CpuProfilePerIntervalFlag,
		&config.DiagnosticServerFlag,
		&config.MemoryBreakdownFlag,
		&config.MemoryProfileFlag,
		&config.RandomSeedFlag,
		&config.PrimeThresholdFlag,

		// Priming
		&config.RandomizePrimingFlag,
		&config.UpdateBufferSizeFlag,

		// Utils
		&config.CustomDbNameFlag,
		&logger.LogLevelFlag,
		&config.TrackProgressFlag,
		&config.ErrorLoggingFlag,
	},
	Description: `
The util-primer priming command requires one argument: <blockNum>

<blockNum> is the block to which the priming will start.`,
}
