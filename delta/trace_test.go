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

package delta

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOperations_SingleFile(t *testing.T) {
	dir := t.TempDir()
	tracePath := filepath.Join(dir, "test.txt")

	content := `BeginBlock, 1000
	BeginTransaction, 0
	CreateAccount, 0x1234567890123456789012345678901234567890
	EndTransaction
	EndBlock
	`
	require.NoError(t, os.WriteFile(tracePath, []byte(content), 0644))

	ops, err := LoadOperations([]string{tracePath}, 0, 0)
	require.NoError(t, err)
	require.Len(t, ops, 5)
	require.Equal(t, "BeginBlock", ops[0].Kind)
	require.Equal(t, uint64(1000), ops[0].Block)
	require.True(t, ops[0].HasBlock)
}

func TestLoadOperations_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	tracePath := filepath.Join(dir, "empty.txt")
	require.NoError(t, os.WriteFile(tracePath, []byte(""), 0644))

	_, err := LoadOperations([]string{tracePath}, 0, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not contain operations")
}

func TestLoadOperations_NoFiles(t *testing.T) {
	_, err := LoadOperations([]string{}, 0, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no trace files provided")
}

func TestWriteTrace(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "output.txt")

	ops := []TraceOp{
		{Raw: "BeginBlock, 1000\n", Kind: "BeginBlock"},
		{Raw: "EndBlock\n", Kind: "EndBlock"},
	}

	require.NoError(t, WriteTrace(outputPath, ops))

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "BeginBlock, 1000")
	require.Contains(t, string(content), "EndBlock")
}

func TestWriteTrace_EmptyOps(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "empty.txt")

	err := WriteTrace(outputPath, []TraceOp{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot write empty trace")
}

func TestFirstBlockNumber(t *testing.T) {
	ops := []TraceOp{
		{Kind: "CreateAccount", HasBlock: false},
		{Kind: "BeginBlock", HasBlock: true, Block: 42},
		{Kind: "EndBlock", HasBlock: true, Block: 42},
	}

	block, found := FirstBlockNumber(ops)
	require.True(t, found)
	require.Equal(t, uint64(42), block)
}

func TestFirstBlockNumber_NoBlock(t *testing.T) {
	ops := []TraceOp{
		{Kind: "CreateAccount", HasBlock: false},
		{Kind: "EndTransaction", HasBlock: false},
	}

	_, found := FirstBlockNumber(ops)
	require.False(t, found)
}

func TestParseTraceLine_BeginBlock(t *testing.T) {
	line := "BeginBlock, 1234"
	op, err := parseTraceLine(line)
	require.NoError(t, err)
	require.Equal(t, "BeginBlock", op.Kind)
	require.True(t, op.HasBlock)
	require.Equal(t, uint64(1234), op.Block)
}

func TestParseTraceLine_WithAddress(t *testing.T) {
	line := "CreateAccount, 0x1234567890123456789012345678901234567890"
	op, err := parseTraceLine(line)
	require.NoError(t, err)
	require.Equal(t, "CreateAccount", op.Kind)
	require.True(t, op.HasContract)
	require.Equal(t, "0x1234567890123456789012345678901234567890", op.Contract.Hex())
}

func TestParseTraceLine_InvalidBlockNumber(t *testing.T) {
	line := "BeginBlock, invalid"
	_, err := parseTraceLine(line)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid block number")
}

func TestParseTraceLine_MissingBlockNumber(t *testing.T) {
	line := "BeginBlock"
	_, err := parseTraceLine(line)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing block number")
}

func TestParseTraceLine_EmptyLine(t *testing.T) {
	line := ""
	_, err := parseTraceLine(line)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing operation kind")
}

func TestWriteTrace_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "subdir", "output.txt")

	ops := []TraceOp{
		{Raw: "BeginBlock, 1000", Kind: "BeginBlock"},
	}

	require.NoError(t, WriteTrace(outputPath, ops))
	require.FileExists(t, outputPath)
}

func TestWriteTrace_AddsNewlines(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "output.txt")

	ops := []TraceOp{
		{Raw: "BeginBlock, 1000", Kind: "BeginBlock"},
		{Raw: "EndBlock\n", Kind: "EndBlock"},
	}

	require.NoError(t, WriteTrace(outputPath, ops))

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	lines := string(content)
	require.Contains(t, lines, "BeginBlock, 1000\n")
	require.Contains(t, lines, "EndBlock\n")
}

func TestLoadOperations_BlockFilters(t *testing.T) {
	dir := t.TempDir()
	tracePath := filepath.Join(dir, "test.txt")

	content := "BeginBlock, 1000\n"
	require.NoError(t, os.WriteFile(tracePath, []byte(content), 0644))

	_, err := LoadOperations([]string{tracePath}, 100, 200)
	require.Error(t, err)
	require.Contains(t, err.Error(), "block filters are not supported")
}

func TestLoadOperations_FileNotFound(t *testing.T) {
	_, err := LoadOperations([]string{"/nonexistent/file.txt"}, 0, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "open trace")
}
