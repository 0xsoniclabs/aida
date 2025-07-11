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

package tracer

import (
	"testing"
)

// TestQueueSimple tests for existence/non-existence of elements.
func TestQueueSimple(t *testing.T) {
	// create index queue
	queue := NewQueue[int]()

	// place first element
	queue.Place(0)

	// find first element
	pos := queue.Find(0)
	if pos != 0 {
		t.Fatalf("element cannot be found")
	}

	// unknown element must not be found
	pos = queue.Find(1)
	if pos != -1 {
		t.Fatalf("element must not be found")
	}
}

// TestQueueSimple1 tests for existence/non-existence of elements.
func TestQueueSimple1(t *testing.T) {
	// create index queue
	queue := NewQueue[int]()

	// find first element
	pos := queue.Find(0)
	if pos != -1 {
		t.Fatalf("Queue must be empty")
	}

	// place first element
	queue.Place(0)

	// place second element
	queue.Place(1)

	// find first element
	pos = queue.Find(1)
	if pos != 0 {
		t.Fatalf("first element cannot be found")
	}
	pos = queue.Find(0)
	if pos != 1 {
		t.Fatalf("second element cannot be found")
	}
}

// TestQueueSimple2 tests for existence/non-existence of elements.
func TestQueueSimple2(t *testing.T) {
	// create index queue
	queue := NewQueue[int]()

	// place first element
	for i := 0; i < QueueLen+1; i++ {
		queue.Place(i)
	}

	// find first element
	pos := queue.Find(0)
	if pos != -1 {
		t.Fatalf("first element must not be found")
	}
	pos = queue.Find(1)
	if pos != QueueLen-1 {
		t.Fatalf("second element must be found: %v", pos)
	}
	pos = queue.Find(QueueLen)
	if pos != 0 {
		t.Fatalf("last element must be found")
	}

	queue.Place(QueueLen + 1)

	pos = queue.Find(1)
	if pos != -1 {
		t.Fatalf("second element must not be found")
	}
	pos = queue.Find(2)
	if pos != QueueLen-1 {
		t.Fatalf("third element must be found: %v", pos)
	}
	pos = queue.Find(QueueLen + 1)
	if pos != 0 {
		t.Fatalf("last element must be found")
	}
}

func TestQueue_Classify(t *testing.T) {
	queue := NewQueue[int]()

	// Zero value
	id, idx := queue.Classify(0)
	if id != ZeroValueID || idx != -1 {
		t.Fatalf("expected ZeroValueID for zero value, got id=%d idx=%d", id, idx)
	}

	// New value (non-zero)
	id, idx = queue.Classify(1)
	if id != NewValueID || idx != -1 {
		t.Fatalf("expected NewValueID for new value, got id=%d idx=%d", id, idx)
	}

	// Previous value (just placed)
	id, idx = queue.Classify(1)
	if id != PreviousValueID || idx != -1 {
		t.Fatalf("expected PreviousValueID for previous value, got id=%d idx=%d", id, idx)
	}

	// Add another value
	id, idx = queue.Classify(2)
	if id != NewValueID || idx != -1 {
		t.Fatalf("expected NewValueID for new value, got id=%d idx=%d", id, idx)
	}

	// Recent value (not the most recent)
	id, idx = queue.Classify(1)
	if id != RecentValueID || idx != 0 {
		t.Fatalf("expected RecentValueID for recent value, got id=%d idx=%d", id, idx)
	}
}
