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
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// TraceOp represents a single line emitted by the logger proxy.
type TraceOp struct {
	Raw         string
	Kind        string
	SubKind     string
	Args        []string
	HasBlock    bool
	Block       uint64
	HasContract bool
	Contract    common.Address
}

// LoadOperations reads operations from textual trace files emitted by the logger proxy.
// Memory-optimized: pre-allocates based on estimated line count to minimize allocations.
func LoadOperations(files []string, firstBlock, lastBlock uint64) ([]TraceOp, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("delta: no trace files provided")
	}
	if firstBlock != 0 || lastBlock != 0 {
		return nil, fmt.Errorf("delta: block filters are not supported for logger traces")
	}

	// Estimate total capacity to minimize slice reallocations
	estimatedOps := estimateTotalOperations(files)
	ops := make([]TraceOp, 0, estimatedOps)

	for _, path := range files {
		if err := readTraceFileAppend(path, &ops); err != nil {
			return nil, err
		}
	}

	if len(ops) == 0 {
		return nil, fmt.Errorf("delta: trace does not contain operations")
	}

	return ops, nil
}

// estimateTotalOperations estimates the number of operations across all files
// to enable pre-allocation and reduce memory reallocations.
func estimateTotalOperations(files []string) int {
	const avgBytesPerOp = 80 // Average bytes per operation line
	totalBytes := int64(0)

	for _, path := range files {
		if stat, err := os.Stat(path); err == nil {
			totalBytes += stat.Size()
		}
	}

	if totalBytes == 0 {
		return 1024 // Default capacity
	}

	return int(totalBytes / avgBytesPerOp)
}

// readTraceFileAppend reads a trace file and appends operations directly to the provided slice.
// This avoids intermediate allocations and copies, making it memory-efficient for large files.
func readTraceFileAppend(path string, ops *[]TraceOp) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("delta: open trace %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	// Use a larger buffer for better I/O performance on large files
	scanner := bufio.NewScanner(f)
	const (
		initialBufSize = 1024 * 1024      // 1MB initial buffer
		maxBufSize     = 16 * 1024 * 1024 // 16MB max buffer
	)
	buf := make([]byte, initialBufSize)
	scanner.Buffer(buf, maxBufSize)

	var (
		currentBlock uint64
		hasBlockCtx  bool
	)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		op, err := parseTraceLine(line)
		if err != nil {
			return fmt.Errorf("delta: parse %s:%d: %w", path, lineNo, err)
		}

		if op.Kind == "BeginBlock" && op.HasBlock {
			currentBlock = op.Block
			hasBlockCtx = true
		}

		if hasBlockCtx {
			op.HasBlock = true
			op.Block = currentBlock
		}

		*ops = append(*ops, op)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("delta: scan %s: %w", path, err)
	}

	return nil
}

func parseTraceLine(line string) (TraceOp, error) {
	parts := splitAndTrim(line, ",")
	if len(parts) == 0 {
		return TraceOp{}, fmt.Errorf("empty trace line")
	}

	kind := parts[0]
	if kind == "" {
		return TraceOp{}, fmt.Errorf("missing operation kind")
	}
	args := parts[1:]

	op := TraceOp{
		Raw:  line,
		Kind: kind,
		Args: args,
	}

	switch kind {
	case "Bulk":
		if len(args) > 0 {
			op.SubKind = args[0]
		}
	case "BeginBlock":
		if len(args) == 0 {
			return TraceOp{}, fmt.Errorf("missing block number")
		}
		block, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return TraceOp{}, fmt.Errorf("invalid block number %q: %w", args[0], err)
		}
		op.HasBlock = true
		op.Block = block
	}

	if addr, ok := extractAddress(kind, args); ok {
		op.HasContract = true
		op.Contract = addr
	}

	return op, nil
}

func extractAddress(kind string, args []string) (common.Address, bool) {
	candidate := ""

	switch kind {
	case "Bulk":
		if len(args) > 1 {
			candidate = args[1]
		}
	default:
		if len(args) > 0 {
			candidate = args[0]
		}
	}

	if !looksLikeAddress(candidate) {
		return common.Address{}, false
	}

	return common.HexToAddress(candidate), true
}

func looksLikeAddress(input string) bool {
	if len(input) != 42 || !strings.HasPrefix(input, "0x") {
		return false
	}
	_, err := hex.DecodeString(input[2:])
	return err == nil
}

func splitAndTrim(input, sep string) []string {
	raw := strings.Split(input, sep)
	for i := range raw {
		raw[i] = strings.TrimSpace(raw[i])
	}
	return raw
}

// WriteTrace writes operations to the specified destination file.
func WriteTrace(path string, ops []TraceOp) error {
	if len(ops) == 0 {
		return fmt.Errorf("delta: cannot write empty trace")
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("delta: ensure output directory: %w", err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("delta: create trace: %w", err)
	}
	defer func() { _ = file.Close() }()

	writer := bufio.NewWriter(file)
	for _, op := range ops {
		if _, err := writer.WriteString(op.Raw); err != nil {
			return fmt.Errorf("delta: write trace: %w", err)
		}
		if !strings.HasSuffix(op.Raw, "\n") {
			if err := writer.WriteByte('\n'); err != nil {
				return fmt.Errorf("delta: write trace newline: %w", err)
			}
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("delta: flush trace: %w", err)
	}

	return nil
}

// firstBlockNumber returns the first block number present in operations.
func firstBlockNumber(ops []TraceOp) (uint64, bool) {
	for _, op := range ops {
		if op.Kind == "BeginBlock" && op.HasBlock {
			return op.Block, true
		}
	}
	return 0, false
}
