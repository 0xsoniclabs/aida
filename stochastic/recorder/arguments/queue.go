// Copyright 2025 Fantom Foundation
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

import "github.com/0xsoniclabs/aida/stochastic"

// queue data structure for a generic FIFO queue.
type queue[T comparable] struct {
	// queue structure
	top  int                    // index of first entry in queue
	rear int                    // index of last entry in queue
	data [stochastic.QueueLen]T // queue data

	// counting statistics for queue position
	// (counter for each position counting successful finds)
	freq [stochastic.QueueLen]uint64
}

// newQueue creates a new queue.
func newQueue[T comparable]() queue[T] {
	return queue[T]{
		top:  -1,
		rear: -1,
		data: [stochastic.QueueLen]T{},
		freq: [stochastic.QueueLen]uint64{},
	}
}

// place a new item into the queue.
func (q *queue[T]) place(item T) {
	// is the queue empty => initialize top/rear
	if q.top == -1 {
		q.top, q.rear = 0, 0
		q.data[q.top] = item
		return
	}

	// put new item into the queue
	q.top = (q.top + 1) % stochastic.QueueLen
	q.data[q.top] = item

	// update rear of queue
	if q.top == q.rear {
		q.rear = (q.rear + 1) % stochastic.QueueLen
	}
}

// findPos the position of an argument in the counting queue.
func (q *queue[T]) findPos(item T) int {
	if q.top == -1 {
		return -1 // if queue is empty, return -1 (not found)
	}
	i := q.top // for non-empty queues, find item by iterating from top
	for {
		if q.data[i] == item {
			// if found, return position in the FIFO queue
			idx := (q.top - i + stochastic.QueueLen) % stochastic.QueueLen
			q.freq[idx]++
			return idx
		}
		if i == q.rear {
			return -1 // if rear of queue reached, return not found
		}
		// go back one position in the queue
		i = (i - 1 + stochastic.QueueLen) % stochastic.QueueLen
	}
}

// QueueStatsJSON is the JSON output for queuing statistics.
type QueueStatsJSON struct {
	// probability of a position in the queue
	Distribution []float64 `json:"distribution"`
}

// json produces JSON output for for a queuing statistics.
func (q *queue[T]) json() QueueStatsJSON {
	// Compute total frequency over all positions
	total := uint64(0)
	for i := range stochastic.QueueLen {
		total += q.freq[i]
	}
	// compute position probabilities
	dist := make([]float64, stochastic.QueueLen)
	if total > 0 {
		for i := range stochastic.QueueLen {
			dist[i] = float64(q.freq[i]) / float64(total)
		}
	}
	// populate position probabilities
	return QueueStatsJSON{
		Distribution: dist,
	}
}
