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
	"math/rand"

	"github.com/0xsoniclabs/aida/stochastic/exponential"
	"github.com/0xsoniclabs/aida/stochastic/statistics"
)

// ArgSetRandomizer interface for argument sets
type ArgSetRandomizer interface {
	SampleArg(n ArgumentType) ArgumentType // sample argument distribution
	SampleQueue() int                      // sample queue distribution
}

// SampleArgRandomizer interface for argument sets
type SampleArgRandomizer interface {
	SampleArg(n ArgumentType) ArgumentType // sample argument distribution
}

// SampleQueueRandomizer interface for argument sets
type SampleQueueRandomizer interface {
	SampleQueue() int // sample queue distribution
}

// ProxyRandomizer struct
type ProxyRandomizer struct {
	sampleArg SampleArgRandomizer
	sampleQ   SampleQueueRandomizer
}

// NewProxyRandomizer creates a new ProxyRandomizer
func NewProxyRandomizer(argR SampleArgRandomizer, qR SampleQueueRandomizer) *ProxyRandomizer {
	return &ProxyRandomizer{
		sampleArg: argR,
		sampleQ:   qR,
	}
}

// SampleArg samples an argument from a distribution with n possible arguments
func (r *ProxyRandomizer) SampleArg(n ArgumentType) ArgumentType {
	return r.sampleArg.SampleArg(n)
}

// SampleQueue samples an index for a queue
func (r *ProxyRandomizer) SampleQueue() int {
	return r.sampleQ.SampleQueue()
}

// ExponentialArgRandomizer struct
type ExponentialArgRandomizer struct {
	rg     *rand.Rand
	lambda float64
}

// NewExponentialArgRandomizer creates a new ExponentialArgRandomizer
func NewExponentialArgRandomizer(rg *rand.Rand, lambda float64) *ExponentialArgRandomizer {
	return &ExponentialArgRandomizer{
		rg:     rg,
		lambda: lambda,
	}
}

// SampleArg samples an argument from a distribution with n possible arguments
func (r *ExponentialArgRandomizer) SampleArg(n ArgumentType) ArgumentType {
	return ArgumentType(exponential.DiscreteSample(r.rg, r.lambda, int64(n)))
}

// EmpiricalQueueRandomizer struct
type EmpiricalQueueRandomizer struct {
	rg   *rand.Rand // random generator
	qpdf []float64  // queue probability distribution function
}

// NewEmpiricalQueueRandomizer creates a new EmpiricalQueueRandomizer
func NewEmpiricalQueueRandomizer(rg *rand.Rand, qpdf []float64) *EmpiricalQueueRandomizer {
	if len(qpdf) != statistics.QueueLen {
		return nil
	}
	cp := make([]float64, statistics.QueueLen)
	copy(cp, qpdf)
	return &EmpiricalQueueRandomizer{
		rg:   rg,
		qpdf: cp,
	}
}

// SampleQueue samples an index for a queue
func (r *EmpiricalQueueRandomizer) SampleQueue() int {
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
