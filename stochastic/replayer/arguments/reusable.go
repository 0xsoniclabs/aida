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

package arguments

import (
	"fmt"
	"math"

	"github.com/0xsoniclabs/aida/stochastic"
)

// Reusable data structure for producing random arguments
// for StateDB operations. An argument set meshes a sample distribution
// with a queue of recently used arguments to produce arguments.
// The argument set always contains the zero argument.
type Reusable struct {
	n     int64      // cardinality of argument set including the zero argument
	queue []int64    // queue for recently chosen arguments
	rand  Randomizer // random generator
	Set
}

// NewReusable creates a random set of arguments for StateDB operations
func NewReusable(n int64, rand Randomizer) *Reusable {
	if n < minCardinality {
		n = minCardinality
	}
	queue := []int64{}
	for range stochastic.QueueLen {
		v := rand.SampleArg(n-1) + 1
		queue = append(queue, v)
	}
	return &Reusable{
		n:     n,
		queue: queue,
		rand:  rand,
	}
}

// Choose the a random argument depending on the argument kind. There are
// following argument kinds: (1) no argument, (2) argument with zero value,
// (3) a new argument increasing the cardinality of the argument set, (4)
// a random argument not contained in the queue, (5) the previous argument
// (6) a recent argument contained in the queue but not the previous one.
func (a *Reusable) Choose(kind int) (int64, error) {
	switch kind {

	// choose no argument
	case stochastic.NoArgID:
		return 0, fmt.Errorf("Choose: illegal invocation for no argument")

	// choose zero argument (only way to return a zero value argument; other argument kinds
	// will result in a non-zero result).
	case stochastic.ZeroArgID:
		return 0, nil

	// choose a new argument that hasn't been used before
	case stochastic.NewArgID:
		if a.n == math.MaxInt64 {
			return 0, fmt.Errorf("Choose: new value exceeds cardinality range")
		}
		v := a.n
		a.placeQ(v)
		a.n++
		return v, nil

	// choose a randomised argument that is not contained in the queue
	case stochastic.RandArgID:
		for {
			// ensure that zero argument is never returned
			v := a.rand.SampleArg(a.n-1) + 1
			if !a.findQElem(v) {
				a.placeQ(v)
				return v, nil
			}
		}

	// choose the previous argument
	case stochastic.PrevArgID:
		v := a.lastQ()
		a.placeQ(v)
		return v, nil

	// choose a recent argument that is not the previous one
	case stochastic.RecentArgID:
		if v, err := a.recentQ(); err == nil {
			a.placeQ(v)
			return v, nil
		} else {
			return 0, err
		}

	default:
		return 0, fmt.Errorf("unknown argument kind")
	}
}

// Remove an argument from set and shrink argument set by one
func (a *Reusable) Remove(v int64) error {
	if v <= 0 || v >= a.n {
		return fmt.Errorf("remove: argument (%v) out of range", v)
	}
	a.n--
	if a.n < minCardinality {
		return fmt.Errorf("remove: cardinality (%v) of argument set too low", a.n)
	}
	// replace deleted last element by new element in queue
	// (to ensure that queue elements are always in range [0,n-1])
	j := a.rand.SampleArg(a.n-1) + 1
	for i := range stochastic.QueueLen {
		if a.queue[i] >= a.n {
			a.queue[i] = j
		}
	}
	return nil
}

// Size returns the current size of the argument set.
func (a *Reusable) Size() int64 {
	return a.n
}

// findQElem finds an element in the queue.
func (a *Reusable) findQElem(elem int64) bool {
	for i := range stochastic.QueueLen {
		if a.queue[i] == elem {
			return true
		}
	}
	return false
}

// placeQ places element in the queue.
func (a *Reusable) placeQ(elem int64) {
	a.queue = append([]int64{elem}, a.queue[0:stochastic.QueueLen-1]...)
}

// lastQ returns previously queued element.
func (a *Reusable) lastQ() int64 {
	return a.queue[0]
}

// recentQ returns randomly an argument in the queue but not the previous one.
func (a *Reusable) recentQ() (int64, error) {
	i := a.rand.SampleQueue()
	if i <= 0 || i >= stochastic.QueueLen {
		return 0, fmt.Errorf("recentQ: queue index (%v) out of range for recent access", i)
	}
	return a.queue[i], nil
}
