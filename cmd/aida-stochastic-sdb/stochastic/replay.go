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

package stochastic

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/stochastic/replayer"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// StochasticReplayCommand data structure for the replay app.
var StochasticReplayCommand = cli.Command{
	Action:    stochasticReplayAction,
	Name:      "replay",
	Usage:     "Simulates StateDB operations using a Markovian Process",
	ArgsUsage: "<simulation-length> <stats-file>",
	Flags: []cli.Flag{
		&utils.BalanceRangeFlag,
		&utils.CarmenSchemaFlag,
		&utils.ContinueOnFailureFlag,
		&utils.CpuProfileFlag,
		&utils.DebugFromFlag,
		&utils.MemoryBreakdownFlag,
		&utils.NonceRangeFlag,
		&utils.RandomSeedFlag,
		&utils.StateDbImplementationFlag,
		&utils.StateDbVariantFlag,
		&utils.DbTmpFlag,
		&utils.StateDbLoggingFlag,
		&utils.DeltaLoggingFlag,
		&utils.TraceFileFlag,
		&utils.TraceDebugFlag,
		&utils.TraceFlag,
		&utils.ShadowDbImplementationFlag,
		&utils.ShadowDbVariantFlag,
		&utils.ValidateStateHashesFlag,
		&logger.LogLevelFlag,
	},
	Description: `
The stochastic replay command requires two argument:
<simulation-length> <stats.json>

<simulation-length> determines the number of blocks
<stats.json> contains the stats for the Markovian Process.`,
}

// stochasticReplayAction implements the replay command. The user provides simulation file and
// the number of blocks that should be replayed as arguments.
func stochasticReplayAction(ctx *cli.Context) error {
	// parse command-line arguments
	if ctx.Args().Len() != 2 {
		return fmt.Errorf("missing simulation file and simulation length as parameter")
	}

	simLength, perr := strconv.Atoi(ctx.Args().Get(0))
	if perr != nil {
		return fmt.Errorf("simulation length is not an integer; %v", perr)
	}
	if simLength <= 0 {
		return fmt.Errorf("simulation length must be greater than zero")
	}

	// process configuration
	cfg, err := utils.NewConfig(ctx, utils.LastBlockArg)
	if err != nil {
		return err
	}
	if cfg.DbImpl == "memory" {
		return fmt.Errorf("db-impl memory is not supported")
	}
	log := logger.NewLogger(cfg.LogLevel, "Stochastic Replay")

	// start CPU profiling if requested.
	if err := utils.StartCPUProfile(cfg); err != nil {
		return err
	}
	defer utils.StopCPUProfile(cfg)

	// read simulation file
	simulation, serr := recorder.Read(ctx.Args().Get(1))
	if serr != nil {
		return fmt.Errorf("failed reading simulation; %v", serr)
	}

	// create a directory for the store to place all its files, and
	// instantiate the state DB under testing.
	log.Notice("Create StateDB")
	db, stateDbDir, err := utils.PrepareStateDB(cfg)
	if err != nil {
		return err
	}
	defer func(path string) {
		err = errors.Join(err, os.RemoveAll(path))
	}(stateDbDir)

	var loggerOutput chan string
	var loggerWg sync.WaitGroup
	var loggerFile *os.File
	var deltaSink *proxy.DeltaLogSink

	deltaLoggingPath := ctx.Path(utils.DeltaLoggingFlag.Name)
	dbLoggingPath := ctx.Path(utils.StateDbLoggingFlag.Name)

	if deltaLoggingPath != "" {
		var err error
		loggerFile, err = os.Create(deltaLoggingPath)
		if err != nil {
			return fmt.Errorf("cannot create delta logging output file: %w", err)
		}
		writer := bufio.NewWriter(loggerFile)
		deltaSink = proxy.NewDeltaLogSink(log, writer, loggerFile)
		db = proxy.NewDeltaLoggerProxy(db, deltaSink)
		log.Noticef("Delta logging enabled: %s", deltaLoggingPath)
	} else if dbLoggingPath != "" {
		var err error
		loggerFile, err = os.Create(dbLoggingPath)
		if err != nil {
			return fmt.Errorf("cannot create db logging output file: %w", err)
		}
		loggerOutput = make(chan string, 1000)
		loggerWg.Add(1)
		go func() {
			defer loggerWg.Done()
			writer := bufio.NewWriter(loggerFile)
			defer func() {
				if err := writer.Flush(); err != nil {
					log.Errorf("cannot flush db-logging writer; %v", err)
				}
			}()
			for line := range loggerOutput {
				if _, err := fmt.Fprintln(writer, line); err != nil {
					log.Errorf("cannot write db log line; %v", err)
					return
				}
			}
		}()
		log.Noticef("StateDB logging enabled: %s", dbLoggingPath)
		db = proxy.NewLoggerProxy(db, log, loggerOutput, &loggerWg)
	}

	defer func() {
		if deltaSink != nil {
			_ = deltaSink.Close()
			loggerFile = nil
		}
		if loggerOutput != nil {
			close(loggerOutput)
			loggerWg.Wait()
		}
		if loggerFile != nil {
			_ = loggerFile.Close()
		}
	}()

	// Enable tracing if debug flag is set
	if cfg.Trace {
		rCtx, err := context.NewRecord(cfg.TraceFile, uint64(0))
		if err != nil {
			return err
		}
		defer rCtx.Close()
		db = proxy.NewRecorderProxy(db, rCtx)
	}

	// run simulation.
	log.Info("Run simulation")
	runErr := replayer.RunStochasticReplay(db, simulation, simLength, cfg, logger.NewLogger(cfg.LogLevel, "Stochastic"))

	// print memory usage after simulation
	if cfg.MemoryBreakdown {
		if usage := db.GetMemoryUsage(); usage != nil {
			log.Noticef("State DB memory usage: %d byte\n%s", usage.UsedBytes, usage.Breakdown)
		} else {
			log.Info("Utilized storage solution does not support memory breakdowns")
		}
	}

	// close the DB and print disk usage
	start := time.Now()
	if err := db.Close(); err != nil {
		log.Criticalf("Failed to close database; %v", err)
	}
	log.Infof("Closing DB took %v", time.Since(start))

	if deltaLoggingPath != "" {
		log.Noticef("Delta trace written to: %s", deltaLoggingPath)
	} else if dbLoggingPath != "" {
		log.Noticef("StateDB trace written to: %s", dbLoggingPath)
	}

	size, err := utils.GetDirectorySize(stateDbDir)
	if err != nil {
		return fmt.Errorf("cannot size of state-db (%v); %v", stateDbDir, err)
	}
	log.Noticef("Final disk usage: %v MiB", float32(size)/float32(1024*1024))

	return runErr
}
