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
	"github.com/0xsoniclabs/aida/stochastic"
)

// minCardinality is the minimum cardinality of the argument set and
// must be substantially larger than stochastic.QueueLen.
// (Otherwise sampling for arguments with class RandomValueID may
// take a very long time and would slow down the simulation.)
const minCardinality = 10 * stochastic.QueueLen

// Set data structure for producing random arguments
// for StateDB operations. An argument set meshes a sample distribution
// with a queue of recently used arguments to produce arguments.
// The argument set always contains the zero argument.
type Set interface {

	// Choose the a random argument depending on the argument kind. There are
	// following argument kinds: (1) no argument, (2) argument with zero value,
	// (3) a new argument increasing the cardinality of the argument set, (4)
	// a random argument not contained in the queue, (5) the previous argument
	// (6) a recent argument contained in the queue but not the previous one.
	Choose(kind int) (int64, error)

	// Remove an argument from set and shrink argument set by one argument.
	Remove(v int64) error

	// Size returns the current size of the argument set.
	Size() int64
}
