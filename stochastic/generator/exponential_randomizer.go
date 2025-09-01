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
//

package generator

import (
	"math/rand"

	"github.com/0xsoniclabs/aida/stochastic/exponential"
	"github.com/0xsoniclabs/aida/stochastic/statistics"
)

type ExpRandomizer struct {
	rg     *rand.Rand
	lambda float64
	qpdf   []float64
	Randomizer
}

func NewExpRandomizer(rg *rand.Rand, lambda float64, qpdf []float64) *ExpRandomizer {
	cp := make([]float64, statistics.QueueLen)
	copy(cp, qpdf)
	return &ExpRandomizer{
		rg:     rg,
		lambda: lambda,
		qpdf:   cp,
	}
}

func (r *ExpRandomizer) SampleDistribution(n int64) int64 {
	return exponential.DiscreteSample(r.rg, r.lambda, n+1)
}

func (r *ExpRandomizer) SampleQueue() int {
	u := r.rg.Float64()

	factor := 1.0 - r.qpdf[0]
	if factor <= 0 {
		for i := 1; i < statistics.QueueLen; i++ {
			if r.qpdf[i] > 0 {
				return i
			}
		}
		return 1
	}

	sum := 0.0
	c := 0.0
	lastPositive := -1

	for i := 1; i < statistics.QueueLen; i++ {
		pi := r.qpdf[i] / factor
		y := pi - c
		t := sum + y
		c = (t - sum) - y
		sum = t

		if u <= sum {
			return i
		}
		if r.qpdf[i] > 0 {
			lastPositive = i
		}
	}

	if lastPositive != -1 {
		return lastPositive
	}

	if statistics.QueueLen > 1 {
		return 1
	}

	return 0
}
