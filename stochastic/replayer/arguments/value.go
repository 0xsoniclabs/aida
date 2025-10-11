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
	"math/rand"

	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous"
)

// ValueSampler samples scalar values from an empirical distribution.
type ValueSampler struct {
	rg   *rand.Rand
	ecdf [][2]float64
}

// NewValueSampler constructs a sampler for scalar arguments.
func NewValueSampler(rg *rand.Rand, ecdf [][2]float64) *ValueSampler {
	return &ValueSampler{
		rg:   rg,
		ecdf: ecdf,
	}
}

// Sample returns a value scaled to the provided upper bound. The result is in [0, limit).
func (s *ValueSampler) Sample(limit int64) int64 {
	if limit <= 0 {
		return 0
	}
	if len(s.ecdf) < 2 {
		return s.rg.Int63n(limit)
	}
	value := continuous.Sample(s.rg, s.ecdf, limit)
	if value >= limit {
		return limit - 1
	}
	if value < 0 {
		return 0
	}
	return value
}

// Replace updates the sampler with a new empirical CDF.
func (s *ValueSampler) Replace(ecdf [][2]float64) {
	s.ecdf = ecdf
}
