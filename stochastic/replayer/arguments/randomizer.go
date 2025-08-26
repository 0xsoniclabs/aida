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

// Randomizer interface for argument and queue sampling
type Randomizer interface {
	SampleArg(n int64) int64 // sample argument distribution
	SampleQueue() int        // sample queue distribution
}

// RandomizerData struct
type RandomizerData struct {
	rg       *rand.Rand
	queuePMF []float64    // probability mass function of queue indexes from [1, QueueLen-1]
	argCDF   [][2]float64 // cumulative distribution function for arguments
}

// NewRandomizer creates a new randomizer instance
func NewRandomizer(rg *rand.Rand, queuePMF []float64, argCDF [][2]float64) (*RandomizerData, error) {
	nqPMF, err := discrete.Shrink(queuePMF)
	if err != nil {
		return nil, fmt.Errorf("NewRandomizer: cannot shrink pdf by one. Error: %v", err)
	}
	return &RandomizerData{
		rg:       rg,
		queuePMF: nqPMF,
		argCDF:   argCDF,
	}, nil
}

// SampleArg samples an argument from a distribution with n possible arguments
func (r *RandomizerData) SampleArg(n int64) int64 {
	return continuous.Sample(r.rg, r.argCDF, int64(n))
}

// SampleQueue samples an index for a queue
func (r *RandomizerData) SampleQueue() int {
	return discrete.Sample(r.rg, r.queuePMF) + 1
}
