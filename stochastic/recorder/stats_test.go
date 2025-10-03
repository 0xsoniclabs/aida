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

package recorder

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TestStatsUpdateFreq checks some operation labels with their argument classes.
func TestStatsUpdateFreq(t *testing.T) {
	r := NewStats()

	// check that frequencies of argument-encoded operations and
	// transit frequencies are zero.
	for i := 0; i < operations.NumArgOps; i++ {
		if r.argOpFreq[i] > 0 {
			t.Fatalf("Operation frequency must be zero")
		}
		for j := 0; j < operations.NumArgOps; j++ {
			if r.transitFreq[i][j] > 0 {
				t.Fatalf("Transit frequency must be zero")
			}
		}
	}

	// inject first operation
	op := operations.CreateAccountID
	addr := stochastic.RandArgID
	key := stochastic.NoArgID
	value := stochastic.NoArgID
	err := r.updateFreq(op, addr, key, value)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	argop1, _ := operations.EncodeArgOp(op, addr, key, value)

	// check updated operation/transit frequencies
	for i := 0; i < operations.NumArgOps; i++ {
		for j := 0; j < operations.NumArgOps; j++ {
			if r.transitFreq[i][j] > 0 {
				t.Fatalf("Transit frequency must be zero")
			}
		}
		if i != argop1 && r.argOpFreq[i] > 0 {
			t.Fatalf("Operation frequency must be zero")
		}
	}
	if r.argOpFreq[argop1] != 1 {
		t.Fatalf("Operation frequency must be one")
	}

	// inject second operation
	op = operations.SetStateID
	addr = stochastic.RandArgID
	key = stochastic.PrevArgID
	value = stochastic.ZeroArgID
	err = r.updateFreq(op, addr, key, value)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	argop2, _ := operations.EncodeArgOp(op, addr, key, value)
	for i := 0; i < operations.NumArgOps; i++ {
		for j := 0; j < operations.NumArgOps; j++ {
			if r.transitFreq[i][j] > 0 && i != argop1 && j != argop2 {
				t.Fatalf("Transit frequency must be zero")
			}
		}
	}
	for i := 0; i < operations.NumArgOps; i++ {
		if (i == argop1 || i == argop2) && r.argOpFreq[i] != 1 {
			t.Fatalf("Operation frequency must be one")
		}
		if (i != argop1 && i != argop2) && r.argOpFreq[i] > 0 {
			t.Fatalf("Operation frequency must be zero")
		}
	}
	if r.transitFreq[argop1][argop2] != 1 {
		t.Fatalf("Transit frequency must be one %v", r.transitFreq[argop2][argop1])
	}
}

// checkFrequencies checks whether the operation and transit frequencies match the expected ones.
func checkFrequencies(r *Stats, opFreq [operations.NumArgOps]uint64, transitFreq [operations.NumArgOps][operations.NumArgOps]uint64) bool {
	for i := 0; i < operations.NumArgOps; i++ {
		if r.argOpFreq[i] != opFreq[i] {
			return false
		}
		for j := 0; j < operations.NumArgOps; j++ {
			if r.transitFreq[i][j] != transitFreq[i][j] {
				return false
			}
		}
	}
	return true
}

// TestStatsOperation checks operation registrations and their argument classes.
func TestStatsOperation(t *testing.T) {
	// operation/transit frequencies
	var (
		opFreq      [operations.NumArgOps]uint64
		transitFreq [operations.NumArgOps][operations.NumArgOps]uint64
	)

	r := NewStats()

	// check that frequencies are zero.
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject first operation and check frequencies.
	addr := common.HexToAddress("0x000000010")
	r.CountAddressOp(operations.CreateAccountID, &addr)
	argop1, _ := operations.EncodeArgOp(operations.CreateAccountID, stochastic.NewArgID, stochastic.NoArgID, stochastic.NoArgID)
	opFreq[argop1]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject second operation and check frequencies.
	key := common.HexToHash("0x000000200")
	r.CountKeyOp(operations.GetStateID, &addr, &key)
	argop2, _ := operations.EncodeArgOp(operations.GetStateID, stochastic.PrevArgID, stochastic.NewArgID, stochastic.NoArgID)
	opFreq[argop2]++
	transitFreq[argop1][argop2]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject third operation and check frequencies.
	value := common.Hash{}
	r.CountValueOp(operations.SetStateID, &addr, &key, &value)
	argop3, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.PrevArgID, stochastic.PrevArgID, stochastic.ZeroArgID)
	opFreq[argop3]++
	transitFreq[argop2][argop3]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject forth operation and check frequencies.
	r.CountOp(operations.SnapshotID)
	argop4, _ := operations.EncodeArgOp(operations.SnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	opFreq[argop4]++
	transitFreq[argop3][argop4]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}
}

func TestStatsScalarDistributions(t *testing.T) {
	stats := NewStats()
	stats.RecordBalance(-5)
	stats.RecordBalance(10)
	stats.RecordBalance(10)
	stats.RecordNonce(1)
	stats.RecordNonce(uint64(math.MaxInt64) + 1)
	stats.RecordCodeSize(-1)
	stats.RecordCodeSize(42)

	assert.Equal(t, uint64(1), stats.balance.freq[0])
	assert.Equal(t, uint64(2), stats.balance.freq[10])
	assert.Equal(t, uint64(1), stats.nonce.freq[1])
	assert.Equal(t, uint64(1), stats.nonce.freq[math.MaxInt64])
	if _, exists := stats.code.freq[-1]; exists {
		t.Fatalf("negative code sizes must be ignored")
	}
	assert.Equal(t, uint64(1), stats.code.freq[42])

	json := stats.JSON()
	assert.Equal(t, int64(10), json.Balance.Max)
	assert.Equal(t, int64(math.MaxInt64), json.Nonce.Max)
	assert.Equal(t, int64(42), json.CodeSize.Max)

	if len(json.Balance.ECDF) < 2 {
		t.Fatalf("balance ecdf must have at least two points")
	}
	if len(json.Nonce.ECDF) < 2 {
		t.Fatalf("nonce ecdf must have at least two points")
	}
	if len(json.CodeSize.ECDF) < 2 {
		t.Fatalf("code size ecdf must have at least two points")
	}
	if got, want := json.Balance.ECDF[0], [2]float64{0.0, 0.0}; got != want {
		t.Fatalf("balance ecdf start mismatch, got %v want %v", got, want)
	}
	if got, want := json.Balance.ECDF[len(json.Balance.ECDF)-1], [2]float64{1.0, 1.0}; got != want {
		t.Fatalf("balance ecdf end mismatch, got %v want %v", got, want)
	}
	if got, want := json.Nonce.ECDF[len(json.Nonce.ECDF)-1], [2]float64{1.0, 1.0}; got != want {
		t.Fatalf("nonce ecdf end mismatch, got %v want %v", got, want)
	}
	if got, want := json.CodeSize.ECDF[len(json.CodeSize.ECDF)-1], [2]float64{1.0, 1.0}; got != want {
		t.Fatalf("code size ecdf end mismatch, got %v want %v", got, want)
	}
}

// TestStatsZeroOperation checks zero value, new and previous argument classes.
func TestStatsZeroOperation(t *testing.T) {
	// operation/transit frequencies
	var (
		opFreq      [operations.NumArgOps]uint64
		transitFreq [operations.NumArgOps][operations.NumArgOps]uint64
	)

	r := NewStats()

	// check that frequencies are zero.
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject first operation and check frequencies.
	addr := common.Address{}
	key := common.Hash{}
	value := common.Hash{}
	r.CountValueOp(operations.SetStateID, &addr, &key, &value)
	argop1, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.ZeroArgID)
	opFreq[argop1]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject second operation and check frequencies.
	addr = common.HexToAddress("0x12312121212")
	key = common.HexToHash("0x232313123123213")
	value = common.HexToHash("0x2301238021830912830")
	r.CountValueOp(operations.SetStateID, &addr, &key, &value)
	argop2, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.NewArgID, stochastic.NewArgID, stochastic.NewArgID)
	opFreq[argop2]++
	transitFreq[argop1][argop2]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject third operation and check frequencies.
	r.CountValueOp(operations.SetStateID, &addr, &key, &value)
	argop3, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.PrevArgID, stochastic.PrevArgID, stochastic.PrevArgID)
	opFreq[argop3]++
	transitFreq[argop2][argop3]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}
}

func TestStatsUpdateFreqError(t *testing.T) {
	r := NewStats()
	if err := r.updateFreq(-1, 0, 0, 0); err == nil {
		t.Fatalf("expected error for invalid opcode")
	}
}

func TestStatsJSONMarshalSetsFileID(t *testing.T) {
	statsJSON := StatsJSON{}
	bytes, err := json.Marshal(statsJSON)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	var decoded StatsJSON
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if decoded.FileId != statsFileID {
		t.Fatalf("expected FileId %q, got %q", statsFileID, decoded.FileId)
	}
}

func TestStatsJSONUnmarshalRejectsInvalidFileID(t *testing.T) {
	data := []byte(`{"FileId":"invalid"}`)
	var statsJSON StatsJSON
	if err := json.Unmarshal(data, &statsJSON); err == nil {
		t.Fatalf("expected error for invalid FileId")
	}
}

// TestStochastic_ReadStats tests reading stats from a JSON file.
func TestStochastic_ReadStats(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		input := &StatsJSON{
			FileId: "stats",
		}
		marshal, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("cannot marshal StatsJSON; %v", err)
		}
		err = os.WriteFile(tempDir+"/stats.json", marshal, 0644)
		if err != nil {
			t.Fatalf("cannot write StatsJSON to file; %v", err)
		}

		stats, err := Read(tempDir + "/stats.json")
		assert.NoError(t, err)
		assert.NotNil(t, stats)
	})

	t.Run("no stats file", func(t *testing.T) {
		err := os.WriteFile(tempDir+"/stats.json", []byte(`{"operations":[]}`), 0644)
		if err != nil {
			t.Fatalf("cannot write StatsJSON to file; %v", err)
		}

		stats, err := Read(tempDir + "/stats.json")
		assert.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("no json", func(t *testing.T) {
		err := os.WriteFile(tempDir+"/stats.json", []byte{}, 0644)
		if err != nil {
			t.Fatalf("cannot write StatsJSON to file; %v", err)
		}
		stats, err := Read(tempDir + "/stats.json")
		assert.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("no exist", func(t *testing.T) {
		stats, err := Read(tempDir + "/1234.json")
		assert.Error(t, err)
		assert.Nil(t, stats)
	})
}

// TestStats_CountSnapshotDelta checks snapshot registrations.
func TestStats_CountSnapshotDelta(t *testing.T) {
	r := NewStats()
	r.CountSnapshot(3)
	r.CountSnapshot(5)
	assert.Equal(t, uint64(1), r.snapshotFreq[3])
	assert.Equal(t, uint64(1), r.snapshotFreq[5])
}

// TestStats_WriteJSON_SuccessAndError tests writing stats to a JSON file.
func TestStats_WriteJSON_SuccessAndError(t *testing.T) {
	r := NewStats()
	r.CountSnapshot(1)
	r.CountSnapshot(1)

	tmp := t.TempDir()
	file := tmp + "/stats.json"
	err := r.Write(file)
	assert.NoError(t, err)
	_, err = os.Stat(file)
	assert.NoError(t, err)

	// error path: write to a directory
	err = r.Write(tmp)
	assert.Error(t, err)
}

// TestStats_JSON checks JSON output of stats.
func TestStats_JSON(t *testing.T) {
	r := NewStats()

	argop1, _ := operations.EncodeArgOp(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	argop2, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.NewArgID, stochastic.NewArgID, stochastic.NewArgID)

	r.argOpFreq[argop1] = 1
	r.argOpFreq[argop2] = 2

	r.transitFreq[argop1][argop2] = 1
	r.transitFreq[argop2][argop1] = 2

	r.CountSnapshot(0) // implicit RevertToSnapshot op
	r.CountSnapshot(1)

	stats := r.JSON()
	assert.Equal(t, "stats", stats.FileId)
	assert.Len(t, stats.Operations, 3)
	assert.Len(t, stats.StochasticMatrix, 3)

	labelIndex := map[string]int{}
	for i, lab := range stats.Operations {
		labelIndex[lab] = i
	}
	exp1, _ := operations.EncodeOpcode(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	exp2, _ := operations.EncodeOpcode(operations.SetStateID, stochastic.NewArgID, stochastic.NewArgID, stochastic.NewArgID)
	exp3, _ := operations.EncodeOpcode(operations.RevertToSnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	i1, ok1 := labelIndex[exp1]
	i2, ok2 := labelIndex[exp2]
	i3, ok3 := labelIndex[exp3]
	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("expected labels %v, %v and %v in %v", exp1, exp2, exp3, stats.Operations)
	}

	assert.InDelta(t, 0.0, stats.StochasticMatrix[i1][i1], 1e-9)
	assert.InDelta(t, 1.0, stats.StochasticMatrix[i1][i2], 1e-9)
	assert.InDelta(t, 0.0, stats.StochasticMatrix[i1][i3], 1e-9)
	assert.InDelta(t, 1.0, stats.StochasticMatrix[i2][i1], 1e-9)
	assert.InDelta(t, 0.0, stats.StochasticMatrix[i2][i2], 1e-9)
	assert.InDelta(t, 0.0, stats.StochasticMatrix[i2][i3], 1e-9)
	assert.InDelta(t, 0.0, stats.StochasticMatrix[i3][i1], 1e-9)
	assert.InDelta(t, 0.0, stats.StochasticMatrix[i3][i2], 1e-9)

	if len(stats.SnapshotECDF) > 0 {
		last := stats.SnapshotECDF[len(stats.SnapshotECDF)-1]
		assert.InDelta(t, 1.0, last[0], 1e-9)
		assert.InDelta(t, 1.0, last[1], 1e-9)
	}
}

// TestReadStats_ReadErrorOnDirectory tests reading stats from a directory instead of a file.
func TestReadStats_ReadErrorOnDirectory(t *testing.T) {
	dir := t.TempDir()
	stats, err := Read(dir)
	assert.Error(t, err)
	assert.Nil(t, stats)
}

// TestStats_WriteJSON_MarshalError tests error handling during JSON marshalling.
func TestStats_WriteJSON_MarshalError(t *testing.T) {
	r := NewStats()
	r.CountSnapshot(0)

	tmp := t.TempDir()
	err := r.Write(tmp + "/stats.json")
	assert.Nil(t, err)
}

// TestStats_WriteJSON_WriteError tests error handling during file writing.
func TestStats_WriteJSON_WriteError(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("/dev/full is Linux-specific")
	}
	r := NewStats()
	// Avoid NaN in ecdf by using delta 1
	r.CountSnapshot(1)
	err := r.Write("/dev/full")
	assert.Error(t, err)
}

// The following tests check that invalid operation registrations cause a fatal error.
func TestStats_CountOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_COUNT_OP") == "1" {
		r := NewStats()
		r.CountOp(-1)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestStats_CountOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_COUNT_OP=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in CountOp")
	}
}

// TestStats_CountAddressOp_FatalIfInvalid checks that CountAddressOp with an invalid operation ID causes a fatal error.
func TestStats_CountAddressOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_COUNT_ADDR") == "1" {
		r := NewStats()
		addr := common.Address{}
		r.CountAddressOp(-1, &addr)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestStats_CountAddressOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_COUNT_ADDR=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in CountAddressOp")
	}
}

// TestStats_CountKeyOp_FatalIfInvalid checks that CountKeyOp with an invalid operation ID causes a fatal error.
func TestStats_CountKeyOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_COUNT_KEY") == "1" {
		r := NewStats()
		addr := common.Address{}
		key := common.Hash{}
		r.CountKeyOp(-1, &addr, &key)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestStats_CountKeyOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_COUNT_KEY=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in CountKeyOp")
	}
}

func TestStats_CountValueOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_COUNT_VALUE") == "1" {
		r := NewStats()
		addr := common.Address{}
		key := common.Hash{}
		val := common.Hash{}
		r.CountValueOp(-1, &addr, &key, &val)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestStats_CountValueOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_COUNT_VALUE=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in CountValueOp")
	}
}

func TestStats_CountOpPanicsForInvalidOp(t *testing.T) {
	r := NewStats()
	assert.Panics(t, func() {
		r.CountOp(operations.NumOps)
	})
}

func TestStats_CountAddressOpPanicsForInvalidOp(t *testing.T) {
	r := NewStats()
	addr := common.Address{}
	assert.Panics(t, func() {
		r.CountAddressOp(operations.NumOps, &addr)
	})
}

func TestStats_CountKeyOpPanicsForInvalidOp(t *testing.T) {
	r := NewStats()
	addr := common.Address{}
	key := common.Hash{}
	assert.Panics(t, func() {
		r.CountKeyOp(operations.NumOps, &addr, &key)
	})
}

func TestStats_CountValueOpPanicsForInvalidOp(t *testing.T) {
	r := NewStats()
	addr := common.Address{}
	key := common.Hash{}
	val := common.Hash{}
	assert.Panics(t, func() {
		r.CountValueOp(operations.NumOps, &addr, &key, &val)
	})
}

func TestStats_MarshalJSONSetsDefaultFileID(t *testing.T) {
	stats := StatsJSON{}
	data, err := stats.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	var decoded StatsJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded.FileId != statsFileID {
		t.Fatalf("expected FileId %q, got %q", statsFileID, decoded.FileId)
	}
}

func TestStats_UnmarshalRejectsMissingFileID(t *testing.T) {
	payload := []byte(`{"operations":[]}`)
	var statsJSON StatsJSON
	if err := json.Unmarshal(payload, &statsJSON); err == nil {
		t.Fatalf("expected error for missing FileId")
	}
}

func TestStat_UnmarshalInvalidJSON(t *testing.T) {
	var statsJSON StatsJSON
	if err := statsJSON.UnmarshalJSON([]byte("{invalid")); err == nil {
		t.Fatalf("expected unmarshal error for malformed input")
	}
}
