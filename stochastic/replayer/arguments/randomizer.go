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

import (
	"fmt"
	"math/rand"

	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous"
	"github.com/0xsoniclabs/aida/stochastic/statistics/discrete"
)

// Randomizer interface for argument sets
type Randomizer interface {
	SampleArg(n int64) int64 // sample argument distribution
	SampleQueue() int        // sample queue distribution
}

// EmpiricalRandomizer struct for randomizing arguments
type EmpiricalRandomizer struct {
	rg    *rand.Rand
	q_pmf []float64    // probability distribution function of queue from [1, QueueLen-1] excluding first element in queue
	a_cdf [][2]float64 // empirical cumulative distribution function for arguments
}

// NewEmpiricalRandomizer creates a new randomizer
func NewEmpiricalRandomizer(rg *rand.Rand, q_pmf []float64, a_cdf [][2]float64) (*EmpiricalRandomizer, error) {
	nq_pmf, err := discrete.Shrink(q_pmf)
	if err != nil {
		return nil, fmt.Errorf("NewEmpiricalRandomizer: cannot shrink pdf by one. Error: %v", err)
	}
	return &EmpiricalRandomizer{
		rg:    rg,
		q_pmf: nq_pmf,
		a_cdf: a_cdf,
	}, nil
}

// SampleArg samples an argument from a distribution with n possible arguments
func (r *EmpiricalRandomizer) SampleArg(n int64) int64 {
	return continuous.Sample(r.rg, r.a_cdf, int64(n))
}

// SampleQueue samples an index for a queue
func (r *EmpiricalRandomizer) SampleQueue() int {
	return discrete.Sample(r.rg, r.q_pmf) + 1
}
