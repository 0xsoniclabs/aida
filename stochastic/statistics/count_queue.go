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

package statistics

// QueueLen sets the length of counting queue (must be greater than one).
const QueueLen = 32

// countQueue data structure for a generic FIFO countQueue.
type countQueue[T comparable] struct {
	// queue structure
	top  int         // index of first entry in queue
	rear int         // index of last entry in queue
	data [QueueLen]T // queue data

	// counting statistics for queue position
	// (counter for each position counting successful finds)
	freq [QueueLen]uint64
}

// NewCountQueue creates a new queue.
func NewCountQueue[T comparable]() countQueue[T] {
	return countQueue[T]{
		top:  -1,
		rear: -1,
		data: [QueueLen]T{},
		freq: [QueueLen]uint64{},
	}
}

// place a new item into the queue.
func (q *countQueue[T]) place(item T) {
	// is the queue empty => initialize top/rear
	if q.top == -1 {
		q.top, q.rear = 0, 0
		q.data[q.top] = item
		return
	}

	// put new item into the queue
	q.top = (q.top + 1) % QueueLen
	q.data[q.top] = item

	// update rear of queue
	if q.top == q.rear {
		q.rear = (q.rear + 1) % QueueLen
	}
}

// findPos the position of an argument in the counting queue.
func (q *countQueue[T]) findPos(item T) int {
	if q.top == -1 {
		return -1 // if queue is empty, return -1 (not found)
	}
	i := q.top // for non-empty queues, find item by iterating from top
	for {
		if q.data[i] == item {
			// if found, return position in the FIFO queue
			idx := (q.top - i + QueueLen) % QueueLen
			q.freq[idx]++
			return idx
		}
		if i == q.rear {
			return -1 // if rear of queue reached, return not found
		}
		// go back one position in the queue
		i = (i - 1 + QueueLen) % QueueLen
	}
}

// QueueStatsJSON is the JSON output for queuing statistics.
type QueueStatsJSON struct {
	// probability of a position in the queue
	Distribution []float64 `json:"distribution"`
}

// newQueueStatsJSON produces JSON output for for a queuing statistics.
func (q *countQueue[T]) newQueueStatsJSON() QueueStatsJSON {
	// Compute total frequency over all positions
	total := uint64(0)
	for i := range QueueLen {
		total += q.freq[i]
	}
	// compute position probabilities
	dist := make([]float64, QueueLen)
	if total > 0 {
		for i := range QueueLen {
			dist[i] = float64(q.freq[i]) / float64(total)
		}
	}
	// populate position probabilities
	return QueueStatsJSON{
		Distribution: dist,
	}
}
