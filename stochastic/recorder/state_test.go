// Copyright 2024 Fantom Foundation
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
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TestEventRegistryUpdateFreq checks some operation labels with their argument classes.
func TestEventRegistryUpdateFreq(t *testing.T) {
	r := NewState()

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
	r.updateFreq(op, addr, key, value)
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
	r.updateFreq(op, addr, key, value)
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

// check frequencies
func checkFrequencies(r *State, opFreq [operations.NumArgOps]uint64, transitFreq [operations.NumArgOps][operations.NumArgOps]uint64) bool {
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

// TestEventRegistryOperation checks registration for operations
func TestEventRegistryOperation(t *testing.T) {
	// operation/transit frequencies
	var (
		opFreq      [operations.NumArgOps]uint64
		transitFreq [operations.NumArgOps][operations.NumArgOps]uint64
	)

	// create new event registry
	r := NewState()

	// check that frequencies are zero.
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject first operation and check frequencies.
	addr := common.HexToAddress("0x000000010")
	r.RegisterAddressOp(operations.CreateAccountID, &addr)
	argop1, _ := operations.EncodeArgOp(operations.CreateAccountID, stochastic.NewArgID, stochastic.NoArgID, stochastic.NoArgID)
	opFreq[argop1]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject second operation and check frequencies.
	key := common.HexToHash("0x000000200")
	r.RegisterKeyOp(operations.GetStateID, &addr, &key)
	argop2, _ := operations.EncodeArgOp(operations.GetStateID, stochastic.PrevArgID, stochastic.NewArgID, stochastic.NoArgID)
	opFreq[argop2]++
	transitFreq[argop1][argop2]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject third operation and check frequencies.
	value := common.Hash{}
	r.RegisterValueOp(operations.SetStateID, &addr, &key, &value)
	argop3, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.PrevArgID, stochastic.PrevArgID, stochastic.ZeroArgID)
	opFreq[argop3]++
	transitFreq[argop2][argop3]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject forth operation and check frequencies.
	r.RegisterOp(operations.SnapshotID)
	argop4, _ := operations.EncodeArgOp(operations.SnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	opFreq[argop4]++
	transitFreq[argop3][argop4]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}
}

// TestEventRegistryZeroOperation checks zero value, new and previous argument classes.
func TestEventRegistryZeroOperation(t *testing.T) {
	// operation/transit frequencies
	var (
		opFreq      [operations.NumArgOps]uint64
		transitFreq [operations.NumArgOps][operations.NumArgOps]uint64
	)

	// create new event registry
	r := NewState()

	// check that frequencies are zero.
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject first operation and check frequencies.
	addr := common.Address{}
	key := common.Hash{}
	value := common.Hash{}
	r.RegisterValueOp(operations.SetStateID, &addr, &key, &value)
	argop1, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.ZeroArgID, stochastic.ZeroArgID, stochastic.ZeroArgID)
	opFreq[argop1]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject second operation and check frequencies.
	addr = common.HexToAddress("0x12312121212")
	key = common.HexToHash("0x232313123123213")
	value = common.HexToHash("0x2301238021830912830")
	r.RegisterValueOp(operations.SetStateID, &addr, &key, &value)
	argop2, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.NewArgID, stochastic.NewArgID, stochastic.NewArgID)
	opFreq[argop2]++
	transitFreq[argop1][argop2]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject third operation and check frequencies.
	r.RegisterValueOp(operations.SetStateID, &addr, &key, &value)
	argop3, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.PrevArgID, stochastic.PrevArgID, stochastic.PrevArgID)
	opFreq[argop3]++
	transitFreq[argop2][argop3]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}
}

func TestStochastic_ReadState(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		input := &StateJSON{
			FileId: "events",
		}
		marshal, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("cannot marshal EventRegistryJSON; %v", err)
		}
		err = os.WriteFile(tempDir+"/events.json", marshal, 0644)
		if err != nil {
			t.Fatalf("cannot write EventRegistryJSON to file; %v", err)
		}

		events, err := Read(tempDir + "/events.json")
		assert.NoError(t, err)
		assert.NotNil(t, events)
	})

	t.Run("not events file", func(t *testing.T) {
		input := &StateJSON{}
		marshal, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("cannot marshal EventRegistryJSON; %v", err)
		}
		err = os.WriteFile(tempDir+"/events.json", marshal, 0644)
		if err != nil {
			t.Fatalf("cannot write EventRegistryJSON to file; %v", err)
		}

		events, err := Read(tempDir + "/events.json")
		assert.Error(t, err)
		assert.Nil(t, events)
	})

	t.Run("not json", func(t *testing.T) {
		err := os.WriteFile(tempDir+"/events.json", []byte{}, 0644)
		if err != nil {
			t.Fatalf("cannot write EventRegistryJSON to file; %v", err)
		}
		events, err := Read(tempDir + "/events.json")
		assert.Error(t, err)
		assert.Nil(t, events)
	})

	t.Run("not exist", func(t *testing.T) {
		events, err := Read(tempDir + "/1234.json")
		assert.Error(t, err)
		assert.Nil(t, events)
	})
}

func TestEventRegistry_RegisterSnapshotDelta(t *testing.T) {
	r := NewState()
	r.RegisterSnapshot(3)
	r.RegisterSnapshot(5)
	assert.Equal(t, uint64(1), r.snapshotFreq[3])
	assert.Equal(t, uint64(1), r.snapshotFreq[5])
}

func TestEventRegistry_WriteJSON_SuccessAndError(t *testing.T) {
	r := NewState()
	r.RegisterSnapshot(1)
	r.RegisterSnapshot(1)

	tmp := t.TempDir()
	file := tmp + "/events.json"
	err := r.Write(file)
	assert.NoError(t, err)
	_, err = os.Stat(file)
	assert.NoError(t, err)

	// error path: write to a directory
	err = r.Write(tmp)
	assert.Error(t, err)
}

func TestEventRegistry_NewEventRegistryJSON(t *testing.T) {
	r := NewState()

	argop1, _ := operations.EncodeArgOp(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	argop2, _ := operations.EncodeArgOp(operations.SetStateID, stochastic.NewArgID, stochastic.NewArgID, stochastic.NewArgID)

	r.argOpFreq[argop1] = 1
	r.argOpFreq[argop2] = 2

	r.transitFreq[argop1][argop2] = 1
	r.transitFreq[argop2][argop1] = 2

	r.RegisterSnapshot(0)
	r.RegisterSnapshot(1)

	events := r.JSON()
	assert.Equal(t, "events", events.FileId)
	assert.Len(t, events.Operations, 2)
	assert.Len(t, events.StochasticMatrix, 2)

	labelIndex := map[string]int{}
	for i, lab := range events.Operations {
		labelIndex[lab] = i
	}
	exp1, _ := operations.EncodeOpcode(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	exp2, _ := operations.EncodeOpcode(operations.SetStateID, stochastic.NewArgID, stochastic.NewArgID, stochastic.NewArgID)
	i1, ok1 := labelIndex[exp1]
	i2, ok2 := labelIndex[exp2]
	if !(ok1 && ok2) {
		t.Fatalf("expected labels %v and %v in %v", exp1, exp2, events.Operations)
	}

	assert.InDelta(t, 0.0, events.StochasticMatrix[i1][i1], 1e-9)
	assert.InDelta(t, 1.0, events.StochasticMatrix[i1][i2], 1e-9)
	assert.InDelta(t, 1.0, events.StochasticMatrix[i2][i1], 1e-9)
	assert.InDelta(t, 0.0, events.StochasticMatrix[i2][i2], 1e-9)

	if len(events.SnapshotECDF) > 0 {
		last := events.SnapshotECDF[len(events.SnapshotECDF)-1]
		assert.InDelta(t, 1.0, last[0], 1e-9)
		assert.InDelta(t, 1.0, last[1], 1e-9)
	}
}

func TestReadState_ReadErrorOnDirectory(t *testing.T) {
	dir := t.TempDir()
	events, err := Read(dir)
	assert.Error(t, err)
	assert.Nil(t, events)
}

func TestEventRegistry_WriteJSON_MarshalError(t *testing.T) {
	r := NewState()
	r.RegisterSnapshot(0)

	tmp := t.TempDir()
	err := r.Write(tmp + "/events.json")
	assert.Error(t, err)
}

func TestEventRegistry_WriteJSON_WriteError(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("/dev/full is Linux-specific")
	}
	r := NewState()
	// Avoid NaN in ecdf by using delta 1
	r.RegisterSnapshot(1)
	err := r.Write("/dev/full")
	assert.Error(t, err)
}

func TestEventRegistry_RegisterOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_REGISTER_OP") == "1" {
		r := NewState()
		r.RegisterOp(-1)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestEventRegistry_RegisterOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_REGISTER_OP=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in RegisterOp")
	}
}

func TestEventRegistry_RegisterAddressOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_REGISTER_ADDR") == "1" {
		r := NewState()
		addr := common.Address{}
		r.RegisterAddressOp(-1, &addr)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestEventRegistry_RegisterAddressOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_REGISTER_ADDR=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in RegisterAddressOp")
	}
}

func TestEventRegistry_RegisterKeyOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_REGISTER_KEY") == "1" {
		r := NewState()
		addr := common.Address{}
		key := common.Hash{}
		r.RegisterKeyOp(-1, &addr, &key)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestEventRegistry_RegisterKeyOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_REGISTER_KEY=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in RegisterKeyOp")
	}
}

func TestEventRegistry_RegisterValueOp_FatalIfInvalid(t *testing.T) {
	if os.Getenv("WANT_FATAL_REGISTER_VALUE") == "1" {
		r := NewState()
		addr := common.Address{}
		key := common.Hash{}
		val := common.Hash{}
		r.RegisterValueOp(-1, &addr, &key, &val)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestEventRegistry_RegisterValueOp_FatalIfInvalid")
	cmd.Env = append(os.Environ(), "WANT_FATAL_REGISTER_VALUE=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit due to log.Fatalf in RegisterValueOp")
	}
}
