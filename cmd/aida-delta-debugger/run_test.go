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
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func createTestContext(traceFiles []string, output string) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	traceSlice := cli.NewStringSlice(traceFiles...)
	set.Var(traceSlice, utils.DeltaTraceFileFlag.Name, "")
	set.String(utils.DeltaOutputFlag.Name, output, "")
	set.String(utils.StateDbImplementationFlag.Name, "geth", "")
	set.String(utils.StateDbVariantFlag.Name, "", "")
	set.String(utils.DbTmpFlag.Name, "", "")
	set.Int(utils.CarmenSchemaFlag.Name, 0, "")
	set.Int(utils.ChainIDFlag.Name, int(utils.SonicMainnetChainID), "")
	set.String("log-level", "", "")
	set.Duration(utils.DeltaTimeoutFlag.Name, 0, "")
	set.Int(utils.AddressSampleRunsFlag.Name, 0, "")
	set.Int64(utils.RandomSeedFlag.Name, 0, "")
	set.Int(utils.MaxFactorFlag.Name, 0, "")

	return cli.NewContext(nil, set, nil)
}

func TestRun_NoTraceFile(t *testing.T) {
	ctx := createTestContext([]string{}, "")
	err := run(ctx)
	require.Error(t, err)

	var exitErr cli.ExitCoder
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestRun_MultipleTraceFiles(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile1 := filepath.Join(tmpDir, "trace1.txt")
	traceFile2 := filepath.Join(tmpDir, "trace2.txt")

	require.NoError(t, os.WriteFile(traceFile1, []byte("BeginBlock, 1\n"), 0644))
	require.NoError(t, os.WriteFile(traceFile2, []byte("BeginBlock, 2\n"), 0644))

	ctx := createTestContext([]string{traceFile1, traceFile2}, "out.txt")
	err := run(ctx)
	require.Error(t, err)

	var exitErr cli.ExitCoder
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestRun_NoOutput(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "trace.txt")
	require.NoError(t, os.WriteFile(traceFile, []byte("BeginBlock, 1\n"), 0644))

	ctx := createTestContext([]string{traceFile}, "")
	err := run(ctx)
	require.Error(t, err)

	var exitErr cli.ExitCoder
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestRun_EmptyOutput(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "trace.txt")
	require.NoError(t, os.WriteFile(traceFile, []byte("BeginBlock, 1\n"), 0644))

	ctx := createTestContext([]string{traceFile}, "  ")
	err := run(ctx)
	require.Error(t, err)

	var exitErr cli.ExitCoder
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestRun_InvalidTraceFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	ctx := createTestContext([]string{"/nonexistent/file.txt"}, outputFile)
	err := run(ctx)
	require.Error(t, err)
}

func TestRun_EmptyTraceFile(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "empty.txt")
	outputFile := filepath.Join(tmpDir, "output.txt")

	require.NoError(t, os.WriteFile(traceFile, []byte(""), 0644))

	ctx := createTestContext([]string{traceFile}, outputFile)
	err := run(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not contain operations")
}
