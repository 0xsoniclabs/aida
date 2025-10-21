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
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/0xsoniclabs/aida/delta"
	log "github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error {
	traceFiles := c.StringSlice("trace-file")
	outputPath := c.String("output")
	timeout := c.Duration("timeout")
	verbose := c.Bool("verbose")
	addressRuns := c.Int("address-sample-runs")
	seed := c.Int64("seed")
	maxFactor := c.Int("max-factor")

	dbImpl := c.String(utils.StateDbImplementationFlag.Name)
	dbVariant := c.String(utils.StateDbVariantFlag.Name)
	tmpDir := c.Path(utils.DbTmpFlag.Name)
	carmenSchema := c.Int(utils.CarmenSchemaFlag.Name)
	chainID := c.Int(utils.ChainIDFlag.Name)
	logLevel := c.String(log.LogLevelFlag.Name)

	if len(traceFiles) == 0 {
		return cli.Exit("provide --trace-file pointing to the logger output", 1)
	}
	if len(traceFiles) > 1 {
		return cli.Exit("provide exactly one --trace-file when using logger traces", 1)
	}
	if strings.TrimSpace(outputPath) == "" {
		return cli.Exit("specify --output to store the minimized trace", 1)
	}

	files := traceFiles

	ops, err := delta.LoadOperations(files, 0, 0)
	if err != nil {
		return err
	}

	loggerFn := func(string, ...any) {}
	if verbose {
		loggerFn = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}

	minimizer := delta.NewMinimizer(delta.MinimizerConfig{
		AddressSampleRuns: addressRuns,
		RandSeed:          seed,
		MaxFactor:         maxFactor,
		Logger:            loggerFn,
	})

	tester, err := delta.NewStateTester(delta.StateTesterConfig{
		DbImpl:       dbImpl,
		Variant:      dbVariant,
		TmpDir:       tmpDir,
		CarmenSchema: carmenSchema,
		LogLevel:     logLevel,
		ChainID:      chainID,
	})
	if err != nil {
		return err
	}

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	start := time.Now()
	minimized, err := minimizer.Minimize(ctx, ops, tester)
	if err != nil {
		if errors.Is(err, delta.ErrInputDoesNotFail) {
			return cli.Exit("delta-debugger: command succeeds on the original trace", 1)
		}
		if errors.Is(err, context.Canceled) {
			return cli.Exit("delta-debugger: operation cancelled", 1)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return cli.Exit("delta-debugger: timeout reached", 1)
		}
		return err
	}

	if err := delta.WriteTrace(outputPath, minimized); err != nil {
		return err
	}

	duration := time.Since(start)
	originalContracts := len(delta.UniqueContracts(ops))
	reducedContracts := len(delta.UniqueContracts(minimized))

	fmt.Fprintf(os.Stderr,
		"aida-delta-debugger: reduced operations %d -> %d, contracts %d -> %d in %.2fs\n",
		len(ops), len(minimized), originalContracts, reducedContracts, duration.Seconds())
	fmt.Fprintf(os.Stderr, "aida-delta-debugger: minimized trace written to %s\n", outputPath)

	return nil
}
