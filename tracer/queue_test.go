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
	"github.com/stretchr/testify/require"
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

func TestQueue_Get_Errors(t *testing.T) {
	queue := NewQueue[int]()
	_, err := queue.Get(0)
	require.ErrorContains(t, err, "queue is empty")
	queue.Place(1)
	_, err = queue.Get(1)
	require.ErrorContains(t, err, "index out of bounds")
}

func TestQueue_ClassifyGet(t *testing.T) {
	const m = 10

	queue := NewQueue[int]()
	item1 := 1
	item2 := 3
	item3 := 5
	item4 := 7
	item5 := 9
	item6 := 11
	// first place the items
	cl, _ := queue.Classify(item1)
	require.Equal(t, NewValueID, cl)
	cl, _ = queue.Classify(item2)
	require.Equal(t, NewValueID, cl)
	cl, _ = queue.Classify(item3)
	require.Equal(t, NewValueID, cl)
	cl, _ = queue.Classify(item4)
	require.Equal(t, NewValueID, cl)
	cl, _ = queue.Classify(item5)
	require.Equal(t, NewValueID, cl)
	cl, _ = queue.Classify(item6)
	require.Equal(t, NewValueID, cl)

	//idxFind := queue.Find(item1)
	//require.Equal(t, 0, idxFind)

	// then find the indexes
	cl, idx1 := queue.Classify(item1)
	require.Equal(t, RecentValueID, cl)
	cl, idx2 := queue.Classify(item2)
	require.Equal(t, RecentValueID, cl)
	cl, idx3 := queue.Classify(item3)
	require.Equal(t, RecentValueID, cl)
	cl, idx4 := queue.Classify(item4)
	require.Equal(t, RecentValueID, cl)
	cl, idx5 := queue.Classify(item5)
	require.Equal(t, RecentValueID, cl)

	item, err := queue.Get(idx1)
	require.NoError(t, err)
	require.Equal(t, item1, item)

	item, err = queue.Get(idx2)
	require.NoError(t, err)
	require.Equal(t, item2, item)

	item, err = queue.Get(idx3)
	require.NoError(t, err)
	require.Equal(t, item3, item)

	item, err = queue.Get(idx4)
	require.NoError(t, err)
	require.Equal(t, item4, item)

	item, err = queue.Get(idx5)
	require.NoError(t, err)
	require.Equal(t, item5, item)

}
