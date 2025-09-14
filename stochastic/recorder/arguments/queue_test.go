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

package arguments

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
)

// TestQueuingSimple tests for existence/non-existence of elements.
func TestQueuingSimple(t *testing.T) {
	// create index queue
	queue := newQueue[int]()

	// place first element
	queue.place(0)

	// find first element
	pos := queue.findPos(0)
	if pos != 0 {
		t.Fatalf("element cannot be found")
	}

	// unknown element must not be found
	pos = queue.findPos(1)
	if pos != -1 {
		t.Fatalf("element must not be found")
	}
}

// TestQueuingSimple1 tests for existence/non-existence of elements.
func TestQueuingSimple1(t *testing.T) {
	// create index queue
	queue := newQueue[int]()

	// find first element
	pos := queue.findPos(0)
	if pos != -1 {
		t.Fatalf("Queue must be empty")
	}

	// place first element
	queue.place(0)

	// place second element
	queue.place(1)

	// find first element
	pos = queue.findPos(1)
	if pos != 0 {
		t.Fatalf("first element cannot be found")
	}
	pos = queue.findPos(0)
	if pos != 1 {
		t.Fatalf("second element cannot be found")
	}
}

// TestQueuingSimple2 tests for existence/non-existence of elements.
func TestQueuingSimple2(t *testing.T) {
	// create index queue
	queue := newQueue[int]()

	// place first element
	for i := range stochastic.QueueLen + 1 {
		queue.place(i)
	}

	// find first element
	pos := queue.findPos(0)
	if pos != -1 {
		t.Fatalf("first element must not be found")
	}
	pos = queue.findPos(1)
	if pos != stochastic.QueueLen-1 {
		t.Fatalf("second element must be found: %v", pos)
	}
	pos = queue.findPos(stochastic.QueueLen)
	if pos != 0 {
		t.Fatalf("last element must be found")
	}

	queue.place(stochastic.QueueLen + 1)

	pos = queue.findPos(1)
	if pos != -1 {
		t.Fatalf("second element must not be found")
	}
	pos = queue.findPos(2)
	if pos != stochastic.QueueLen-1 {
		t.Fatalf("third element must be found: %v", pos)
	}
	pos = queue.findPos(stochastic.QueueLen + 1)
	if pos != 0 {
		t.Fatalf("last element must be found")
	}
}

// TestQueueJSON tests JSON output for a queue statistics.
// It marshals the JSON output and unmarshals it again and checks whether
// the original and unmarshaled JSON output are identical.
func testQueueJSON(stats queue[int], t *testing.T) {
	jsonX := stats.json()
	jOut, err := json.Marshal(jsonX)
	if err != nil {
		t.Fatalf("Marshalling failed to produce distribution")
	}
	var jsonY QueueStatsJSON
	if err := json.Unmarshal(jOut, &jsonY); err != nil {
		t.Fatalf("Unmarshalling failed to produce JSON")
	}
	if !reflect.DeepEqual(jsonX, jsonY) {
		t.Errorf("Unmarshaling mismatch. Expected:\n%+v\nActual:\n%+v", jsonX, jsonY)
	}
}

// TestQueuingJSON tests JSON output of distribution.
func TestQueuingJSON(t *testing.T) {
	// create index queue
	queue := newQueue[int]()

	// check empty queue JSON output
	testQueueJSON(queue, t)

	// place first element
	for i := range 300 {
		queue.place(i)
		// find first element
		pos := queue.findPos(i)
		if pos != 0 {
			t.Fatalf("first element must be found")
		}
		pos = queue.findPos(i - 1)
		pos = queue.findPos(i - 2)
		pos = queue.findPos(i - 3)
	}

	// check populated queue JSON output
	testQueueJSON(queue, t)
}
