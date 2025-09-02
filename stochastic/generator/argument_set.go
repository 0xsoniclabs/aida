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

// ArgumentType defines the integer type of arguments
type ArgumentType = int64

// MaxArgumentType is the maximum value of the argument type
const MaxArgumentType = math.MaxInt64

// minCardinality is the minimum cardinality of the argument set and
// must be substantially larger than statistics.QueueLen.
// (Otherwise sampling for arguments with class RandomValueID may
// take a very long time and would slow down the simulation.)
const minCardinality = 10 * statistics.QueueLen

// ArgumentSet data structure for producing random arguments
// for StateDB operations. An argument set meshes a sample distribution
// with a queue of recently used arguments to produce arguments.
// The argument set always contains the zero argument.
type ArgumentSet struct {
	n     ArgumentType   // cardinality of argument set including the zero argument
	queue []ArgumentType // queue for recently chosen arguments
	rand  Randomizer     // random generator
}

// NewArgumentSet creates a random set of arguments for StateDB operations
func NewArgumentSet(n ArgumentType, rand Randomizer) *ArgumentSet {
	if n < minCardinality {
		return nil
	}
	queue := []ArgumentType{}
	for range statistics.QueueLen {
		v := rand.SampleDistribution(n-1) + 1
		queue = append(queue, v)
	}
	return &ArgumentSet{
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
func (a *ArgumentSet) Choose(kind int) (ArgumentType, error) {
	switch kind {

	// choose no argument
	case statistics.NoArgID:
		return 0, fmt.Errorf("Choose: illegal invocation for no argument")

	// choose zero argument (only way to return a zero value argument; other argument kinds
	// will result in a non-zero result).
	case statistics.ZeroArgID:
		return 0, nil

	// choose a new argument that hasn't been used before
	case statistics.NewArgID:
		if a.n == MaxArgumentType {
			return 0, fmt.Errorf("Choose: new value exceeds cardinality range")
		}
		v := a.n
		a.placeQ(v)
		a.n++
		return v, nil

	// choose a randomised argument that is not contained in the queue
	case statistics.RandArgID:
		for {
			// ensure that zero argument is never returned
			v := a.rand.SampleDistribution(a.n-1) + 1
			if !a.findQElem(v) {
				a.placeQ(v)
				return v, nil
			}
		}

	// choose the previous argument
	case statistics.PrevArgID:
		v := a.lastQ()
		a.placeQ(v)
		return v, nil

	// choose a recent argument that is not the previous one
	case statistics.RecentArgID:
		if v, err := a.recentQ(); err == nil {
			a.placeQ(v)
			return v, nil
		} else {
			return 0, err
		}

	default:
		return 0, fmt.Errorf("Unknown argument kind")
	}
}

// Remove an argument from set and shrink argument set by one
func (a *ArgumentSet) Remove(v ArgumentType) error {
	if v <= 0 || v >= a.n {
		return fmt.Errorf("Remove: argument (%v) out of range", v)
	}
	a.n--
	if a.n < minCardinality {
		return fmt.Errorf("Remove: cardinality (%v) of argument set too low", a.n)
	}
	// replace deleted last element by new element in queue
	// (to ensure that queue elements are always in range [0,n-1])
	j := a.rand.SampleDistribution(a.n-1) + 1
	for i := range statistics.QueueLen {
		if a.queue[i] >= a.n {
			a.queue[i] = j
		}
	}
	return nil
}

// Size returns the current size of the argument set.
func (a *ArgumentSet) Size() ArgumentType {
	return a.n
}

// findQElem finds an element in the queue.
func (a *ArgumentSet) findQElem(elem ArgumentType) bool {
	for i := range statistics.QueueLen {
		if a.queue[i] == elem {
			return true
		}
	}
	return false
}

// placeQ places element in the queue.
func (a *ArgumentSet) placeQ(elem ArgumentType) {
	a.queue = append([]ArgumentType{elem}, a.queue[0:statistics.QueueLen-1]...)
}

// lastQ returns previously queued element.
func (a *ArgumentSet) lastQ() ArgumentType {
	return a.queue[0]
}

// recentQ returns randomly an argument in the queue but not the previous one.
func (a *ArgumentSet) recentQ() (ArgumentType, error) {
	i := a.rand.SampleQueue()
	if i <= 0 || i >= statistics.QueueLen {
		return 0, fmt.Errorf("recentQ: queue index (%v) out of range for recent access", i)
	}
	return a.queue[i], nil
}
