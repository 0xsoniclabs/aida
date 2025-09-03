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

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension/logger"
	"github.com/0xsoniclabs/aida/executor/extension/primer"
	"github.com/0xsoniclabs/aida/executor/extension/profiler"
	"github.com/0xsoniclabs/aida/executor/extension/statedb"
	"github.com/0xsoniclabs/aida/executor/extension/validator"
	log "github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/urfave/cli/v2"
)

var RunEthTestsCmd = cli.Command{
	Action:    RunEthereumTest,
	Name:      "ethereum-test",
	Usage:     "Execute ethereum tests",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Aliases:   []string{"ethtest"},
	Flags: []cli.Flag{
		// StateDb
		&config.CarmenSchemaFlag,
		&config.StateDbImplementationFlag,
		&config.StateDbVariantFlag,
		&config.DbTmpFlag,
		&config.StateDbLoggingFlag,

		//// ShadowDb
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

		// Utils
		&config.WorkersFlag,
		&config.ChainIDFlag,
		&config.ContinueOnFailureFlag,
		&config.ValidateFlag,
		&config.ValidateStateHashesFlag,
		&log.LogLevelFlag,
		&config.ErrorLoggingFlag,
		&config.MaxNumErrorsFlag,

		// Ethereum execution tests
		&config.EthTestTypeFlag,
		&config.ForkFlag,
	},
	Description: `
The aida-vm-sdb geth-state-tests command requires one argument: <pathToJsonTest or pathToDirWithJsonTests>`,
}

// RunEthereumTest performs sequential block processing on a StateDb
func RunEthereumTest(ctx *cli.Context) error {
	cfg, err := config.NewConfig(ctx, config.PathArg)
	if err != nil {
		return err
	}

	cfg.StateValidationMode = config.SubsetCheck
	cfg.ValidateTxState = true

	processor, err := executor.MakeEthTestProcessor(cfg)
	if err != nil {
		return err
	}
	if !ctx.IsSet(config.ChainIDFlag.Name) {
		return fmt.Errorf("please specify chain ID using --%s flag (1337 for most cases for this tool)", config.ChainIDFlag.Name)
	}

	return runEth(cfg, executor.NewEthStateTestProvider(cfg), nil, processor, nil)
}

func runEth(
	cfg *config.Config,
	provider executor.Provider[txcontext.TxContext],
	stateDb state.StateDB,
	processor executor.Processor[txcontext.TxContext],
	extra []executor.Extension[txcontext.TxContext],
) error {
	// order of extensionList has to be maintained
	var extensionList = []executor.Extension[txcontext.TxContext]{
		profiler.MakeCpuProfiler[txcontext.TxContext](cfg),
		profiler.MakeDiagnosticServer[txcontext.TxContext](cfg),
		logger.MakeErrorLogger[txcontext.TxContext](cfg),
	}

	if stateDb == nil {
		extensionList = append(
			extensionList,
			statedb.MakeEthStateTestDbPrepper(cfg),
			statedb.MakeLiveDbBlockChecker[txcontext.TxContext](cfg),
			logger.MakeDbLogger[txcontext.TxContext](cfg),
			primer.MakeEthStateTestDbPrimer(cfg), // < to be placed after the DbLogger to log priming operations
		)
	}

	extensionList = append(
		extensionList,
		logger.MakeEthStateTestLogger(cfg, 0),
		validator.MakeShadowDbValidator(cfg),
		validator.MakeEthStateTestStateHashValidator(cfg),
		statedb.MakeEthStateScopeTestEventEmitter(),
		validator.MakeEthStateTestErrorValidator(cfg),
		validator.MakeEthStateTestLogHashValidator(cfg),
	)

	extensionList = append(extensionList, extra...)

	return executor.NewExecutor(provider, cfg.LogLevel).Run(
		executor.Params{
			From:                   int(cfg.First),
			To:                     int(cfg.Last) + 1,
			NumWorkers:             1,
			State:                  stateDb,
			ParallelismGranularity: executor.BlockLevel,
		},
		processor,
		extensionList,
		nil,
	)
}
