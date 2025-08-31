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

package generator

import (
	"fmt"
	"math"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
)

// minSize must be substantially larger than statistics.QueueLen
// (Otherwise sampling for arguments with class RandomValueID may
// take a very long time and would slow down the simulation.)
const minSize = 100 * statistics.QueueLen

// Randomizer interface
type Randomizer interface {
	RandRange(l int64, u int64) int64 // produce a random value in the interval [l,u]
	Sample(n int64) int64             // sample distribution
	SampleQueue() int                 // sample queue distribution
}

// RandomAccess data structure for producing random index accesses.
type RandomAccess struct {
	n     int64      // cardinality of set
	queue []int64    // queue for indexes whose length is less than qStatslen
	rand  Randomizer // random generator
}

// NewAccess creates a new random access instance for arguments
func NewRandomAccess(n int64, rand Randomizer) *RandomAccess {
	queue := []int64{}
	for i := 0; i < statistics.QueueLen; i++ {
		queue = append(queue, rand.RandRange(1, n))
	}
	return &RandomAccess{
		n:     n,
		queue: queue,
		rand:  rand,
	}
}

// NextIndex returns the next random value for the given argument class
func (a *RandomAccess) NextIndex(class int) (int64, error) {
    switch class {

	case statistics.NoArgID:
		return 0, fmt.Errorf("NextIndex: illegal invocation for no-argument class")

	// only way to return zero value/all other access classes
	// will result in a non-zero result.
	case statistics.ZeroValueID:
		return 0, nil

	// increment population size of access set
	// and return newly introduced element.
	case statistics.NewValueID:
		if a.n == math.MaxInt64 {
			return 0, fmt.Errorf("NextIndex: new value exceeds cardinality range")
		}
		v := a.n
		a.placeQ(v)
		a.n++
		return v + 1, nil

	// use randomised value that is not contained in the queue
	case statistics.RandomValueID:
		for {
			v := a.rand.Sample(a.n - 1)
			if !a.findQElem(v) {
				a.placeQ(v)
				return v + 1, nil
			}
		}

	// return the value of the first position in the queue
	case statistics.PreviousValueID:
		v := a.lastQ()
		a.placeQ(v)
		return v + 1, nil

	// return the first position in the queue
    case statistics.RecentValueID:
        if v, err := a.recentQ(); err == nil {
            a.placeQ(v)
            return v + 1, nil
        } else {
            return 0, err
        }

	default:
		return 0, fmt.Errorf("Unknown argument class")
	}
}

// DeleteIndex deletes a value reducing the cardinality by one
func (a *RandomAccess) DeleteIndex(v int64) error {
	// check index range
	if v < 0 || v >= a.n {
		return fmt.Errorf("DeleteIndex: index (%v) out of index range", v)
	}
	// reduce cardinality by one
	if a.n <= minSize {
		return fmt.Errorf("DeleteIndex: cardinality of set too low")
	}
	a.n--
	// replace deleted last element by new element
	// note that the actual deleted element may be
	// in range, but there might elements in the queue
	// that exceed the new range limit. They need to
	// be replaced.
	j := a.rand.Sample(a.n - 1)
	for i := 0; i < statistics.QueueLen; i++ {
		if a.queue[i] >= a.n {
			a.queue[i] = j
		}
	}
	return nil
}

// findQElem finds an element in the queue.
func (a *RandomAccess) findQElem(elem int64) bool {
	for i := 0; i < statistics.QueueLen; i++ {
		if a.queue[i] == elem {
			return true
		}
	}
	return false
}

// placeQ places element in the queue.
func (a *RandomAccess) placeQ(elem int64) {
	a.queue = append([]int64{elem}, a.queue[0:statistics.QueueLen-1]...)
}

// lastQ returns previously queued element.
func (a *RandomAccess) lastQ() int64 {
	return a.queue[0]
}

// recentQ returns some element in the queue but not the previous one.
func (a *RandomAccess) recentQ() (int64, error) {
	i := a.rand.SampleQueue()
	if i <= 0 || i >= statistics.QueueLen {
		return 0, fmt.Errorf("recentQ: queue index out of range for recent access")
	}
	return a.queue[i], nil
}
