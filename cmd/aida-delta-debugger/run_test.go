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
	"flag"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func newRunContext(t *testing.T, traceFiles []string, output string) *cli.Context {
	t.Helper()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	for _, fl := range []cli.Flag{
		&utils.DeltaTraceFileFlag,
		&utils.DeltaOutputFlag,
		&utils.DeltaTimeoutFlag,
		&utils.AddressSampleRunsFlag,
		&utils.RandomSeedFlag,
		&utils.MaxFactorFlag,
		&utils.StateDbImplementationFlag,
		&utils.StateDbVariantFlag,
		&utils.DbTmpFlag,
		&utils.CarmenSchemaFlag,
		&utils.ChainIDFlag,
		&logger.LogLevelFlag,
	} {
		require.NoError(t, fl.Apply(fs))
	}

	for _, tf := range traceFiles {
		require.NoError(t, fs.Set(utils.DeltaTraceFileFlag.Name, tf))
	}
	if output != "" {
		require.NoError(t, fs.Set(utils.DeltaOutputFlag.Name, output))
	}

	return cli.NewContext(cli.NewApp(), fs, nil)
}

func TestRun_NoTraceFile(t *testing.T) {
	ctx := newRunContext(t, nil, "out.trace")

	err := run(ctx)
	require.Error(t, err)
	exitErr, ok := err.(cli.ExitCoder)
	require.True(t, ok)
	require.Contains(t, exitErr.Error(), "provide --trace-file")
}

func TestRun_MultipleTraceFiles(t *testing.T) {
	ctx := newRunContext(t, []string{"a", "b"}, "out.trace")

	err := run(ctx)
	require.Error(t, err)
	exitErr, ok := err.(cli.ExitCoder)
	require.True(t, ok)
	require.Contains(t, exitErr.Error(), "provide exactly one --trace-file")
}

func TestRun_MissingOutput(t *testing.T) {
	ctx := newRunContext(t, []string{"trace.txt"}, "")

	err := run(ctx)
	require.Error(t, err)
	exitErr, ok := err.(cli.ExitCoder)
	require.True(t, ok)
	require.Contains(t, exitErr.Error(), "specify --output")
}

func TestRun_LoadOperationsError(t *testing.T) {
	ctx := newRunContext(t, []string{filepath.Join(t.TempDir(), "missing.log")}, filepath.Join(t.TempDir(), "out.trace"))

	err := run(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "open trace")
}
