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

package stochastic

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TestEventRegistryUpdateFreq checks some operation labels with their argument classes.
func TestEventRegistryUpdateFreq(t *testing.T) {
	r := NewEventRegistry()

	// check that frequencies of argument-encoded operations and
	// transit frequencies are zero.
	for i := 0; i < numArgOps; i++ {
		if r.argOpFreq[i] > 0 {
			t.Fatalf("Operation frequency must be zero")
		}
		for j := 0; j < numArgOps; j++ {
			if r.transitFreq[i][j] > 0 {
				t.Fatalf("Transit frequency must be zero")
			}
		}
	}

	// inject first operation
	op := CreateAccountID
	addr := statistics.RandArgID
	key := statistics.NoArgID
	value := statistics.NoArgID
	r.updateFreq(op, addr, key, value)
	argop1 := EncodeArgOp(op, addr, key, value)

	// check updated operation/transit frequencies
	for i := 0; i < numArgOps; i++ {
		for j := 0; j < numArgOps; j++ {
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
	op = SetStateID
	addr = statistics.RandArgID
	key = statistics.PrevArgID
	value = statistics.ZeroArgID
	r.updateFreq(op, addr, key, value)
	argop2 := EncodeArgOp(op, addr, key, value)
	for i := 0; i < numArgOps; i++ {
		for j := 0; j < numArgOps; j++ {
			if r.transitFreq[i][j] > 0 && i != argop1 && j != argop2 {
				t.Fatalf("Transit frequency must be zero")
			}
		}
	}
	for i := 0; i < numArgOps; i++ {
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
func checkFrequencies(r *EventRegistry, opFreq [numArgOps]uint64, transitFreq [numArgOps][numArgOps]uint64) bool {
	for i := 0; i < numArgOps; i++ {
		if r.argOpFreq[i] != opFreq[i] {
			return false
		}
		for j := 0; j < numArgOps; j++ {
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
		opFreq      [numArgOps]uint64
		transitFreq [numArgOps][numArgOps]uint64
	)

	// create new event registry
	r := NewEventRegistry()

	// check that frequencies are zero.
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject first operation and check frequencies.
	addr := common.HexToAddress("0x000000010")
	r.RegisterAddressOp(CreateAccountID, &addr)
	argop1 := EncodeArgOp(CreateAccountID, statistics.NewArgID, statistics.NoArgID, statistics.NoArgID)
	opFreq[argop1]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject second operation and check frequencies.
	key := common.HexToHash("0x000000200")
	r.RegisterKeyOp(GetStateID, &addr, &key)
	argop2 := EncodeArgOp(GetStateID, statistics.PrevArgID, statistics.NewArgID, statistics.NoArgID)
	opFreq[argop2]++
	transitFreq[argop1][argop2]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject third operation and check frequencies.
	value := common.Hash{}
	r.RegisterValueOp(SetStateID, &addr, &key, &value)
	argop3 := EncodeArgOp(SetStateID, statistics.PrevArgID, statistics.PrevArgID, statistics.ZeroArgID)
	opFreq[argop3]++
	transitFreq[argop2][argop3]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject forth operation and check frequencies.
	r.RegisterOp(SnapshotID)
	argop4 := EncodeArgOp(SnapshotID, statistics.NoArgID, statistics.NoArgID, statistics.NoArgID)
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
		opFreq      [numArgOps]uint64
		transitFreq [numArgOps][numArgOps]uint64
	)

	// create new event registry
	r := NewEventRegistry()

	// check that frequencies are zero.
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject first operation and check frequencies.
	addr := common.Address{}
	key := common.Hash{}
	value := common.Hash{}
	r.RegisterValueOp(SetStateID, &addr, &key, &value)
	argop1 := EncodeArgOp(SetStateID, statistics.ZeroArgID, statistics.ZeroArgID, statistics.ZeroArgID)
	opFreq[argop1]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject second operation and check frequencies.
	addr = common.HexToAddress("0x12312121212")
	key = common.HexToHash("0x232313123123213")
	value = common.HexToHash("0x2301238021830912830")
	r.RegisterValueOp(SetStateID, &addr, &key, &value)
	argop2 := EncodeArgOp(SetStateID, statistics.NewArgID, statistics.NewArgID, statistics.NewArgID)
	opFreq[argop2]++
	transitFreq[argop1][argop2]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}

	// inject third operation and check frequencies.
	r.RegisterValueOp(SetStateID, &addr, &key, &value)
	argop3 := EncodeArgOp(SetStateID, statistics.PrevArgID, statistics.PrevArgID, statistics.PrevArgID)
	opFreq[argop3]++
	transitFreq[argop2][argop3]++
	if !checkFrequencies(&r, opFreq, transitFreq) {
		t.Fatalf("operation/transit frequency diverges")
	}
}

func TestStochastic_ReadEvents(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		input := &EventRegistryJSON{
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

		events, err := ReadEvents(tempDir + "/events.json")
		assert.NoError(t, err)
		assert.NotNil(t, events)
	})

	t.Run("not events file", func(t *testing.T) {
		input := &EventRegistryJSON{}
		marshal, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("cannot marshal EventRegistryJSON; %v", err)
		}
		err = os.WriteFile(tempDir+"/events.json", marshal, 0644)
		if err != nil {
			t.Fatalf("cannot write EventRegistryJSON to file; %v", err)
		}

		events, err := ReadEvents(tempDir + "/events.json")
		assert.Error(t, err)
		assert.Nil(t, events)
	})

	t.Run("not json", func(t *testing.T) {
		err := os.WriteFile(tempDir+"/events.json", []byte{}, 0644)
		if err != nil {
			t.Fatalf("cannot write EventRegistryJSON to file; %v", err)
		}
		events, err := ReadEvents(tempDir + "/events.json")
		assert.Error(t, err)
		assert.Nil(t, events)
	})

	t.Run("not exist", func(t *testing.T) {
		events, err := ReadEvents(tempDir + "/1234.json")
		assert.Error(t, err)
		assert.Nil(t, events)
	})
}
