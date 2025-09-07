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

	"github.com/0xsoniclabs/aida/stochastic/statistics/classifier"
	discrete_empiricial "github.com/0xsoniclabs/aida/stochastic/statistics/discrete_empirical"
	"github.com/0xsoniclabs/aida/stochastic/statistics/exponential"
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

// NewProxyRandomizer creates a proxy for SampleArgRandomizer and SampleQueueRandomizer
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
	return ArgumentType(exponential.Sample(r.rg, r.lambda, int64(n)))
}

// ExponentialQueueRandomizer struct
type ExponentialQueueRandomizer struct {
	rg     *rand.Rand
	lambda float64
}

// NewExponentialQueueRandomizer creates a new ExponentialQueueRandomizer
func NewExponentialQueueRandomizer(rg *rand.Rand, lambda float64) *ExponentialQueueRandomizer {
	return &ExponentialQueueRandomizer{
		rg:     rg,
		lambda: lambda,
	}
}

// SampleQueue samples an index for a queue
func (r *ExponentialQueueRandomizer) SampleQueue() int {
	return int(exponential.Sample(r.rg, r.lambda, int64(classifier.QueueLen-1))) + 1
}

// EmpiricalQueueRandomizer struct
type EmpiricalQueueRandomizer struct {
	rg  *rand.Rand // random generator
	pdf []float64  // probability distribution function of queue indices 1..QueueLen-1 (excluding zero)
}

// NewEmpiricalQueueRandomizer creates a new EmpiricalQueueRandomizer
func NewEmpiricalQueueRandomizer(rg *rand.Rand, qpdf []float64) *EmpiricalQueueRandomizer {
	if len(qpdf) != classifier.QueueLen {
		return nil
	}
	factor := 1.0 - qpdf[0]
	if factor <= 0 {
		return nil
	}
	cp := make([]float64, classifier.QueueLen-1)
	for i := range classifier.QueueLen - 1 {
		cp[i] = qpdf[i+1] / factor
	}
	return &EmpiricalQueueRandomizer{
		rg:  rg,
		pdf: cp,
	}
}

// SampleQueue samples an index for a queue
func (r *EmpiricalQueueRandomizer) SampleQueue() int {
	return discrete_empiricial.Sample(r.pdf, r.rg.Float64()) + 1
}

// SnapshotSet interface for snapshot arguments
type SnapshotSet interface {
	SampleSnapshot(n int) int // sample queue distribution
}

// ExponentialSnapshotRandomizer struct
type ExponentialSnapshotRandomizer struct {
	rg     *rand.Rand
	lambda float64
}

func NewExponentialSnapshotRandomizer(rg *rand.Rand, lambda float64) *ExponentialSnapshotRandomizer {
	return &ExponentialSnapshotRandomizer{
		rg:     rg,
		lambda: lambda,
	}
}

func (r *ExponentialSnapshotRandomizer) SampleSnapshot(n int) int {
	return int(exponential.Sample(r.rg, r.lambda, int64(n)))
}
