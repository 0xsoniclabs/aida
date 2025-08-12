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

func TestQueue_PlaceAndFind(t *testing.T) {
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

func TestQueue_IsCircularBuffer(t *testing.T) {
	q := NewQueue[int]()

	// Fill the queue with more items than its capacity
	totalItems := QueueLen + 50
	for i := 1; i <= totalItems; i++ {
		q.Classify(i)
	}

	// Oldest items should be classified as new (since they were overwritten)
	for i := 1; i <= 50; i++ {
		if idx := q.Find(i); idx != -1 {
			t.Errorf("Expected item %d to be overwritten, but found at index %d", i, idx)
		}
	}

	// Most recent items should be present and classified as RecentValueID or PreviousValueID
	for i := totalItems; i > totalItems-QueueLen; i-- {
		idx := q.Find(i)
		if idx == -1 {
			t.Errorf("Expected item %d to be in queue, but not found", i)
			continue
		}
		if idx == 0 {
			// Most recent item
			id, _ := q.Classify(i)
			if id != PreviousValueID {
				t.Errorf("Expected item %d to be classified as PreviousValueID, got id=%d", i, id)
			}
		} else {
			id, _ := q.Classify(i)
			if id != RecentValueID {
				t.Errorf("Expected item %d to be classified as RecentValueID, got id=%d", i, id)
			}
		}
	}
}
