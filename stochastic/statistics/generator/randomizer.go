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

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous_empirical"
	"github.com/0xsoniclabs/aida/stochastic/statistics/discrete_empirical"
)

// ArgSetRandomizer interface for argument sets
type ArgSetRandomizer interface {
	SampleArg(n ArgumentType) ArgumentType // sample argument distribution
	SampleQueue() int                      // sample queue distribution
}

// ProxyRandomizer struct
type EmpiricalArgSetRandomizer struct {
	rg   *rand.Rand
	pdf  []float64    // probability distribution function of queue from [1, QueueLen-1]
	ecdf [][2]float64 // empirical cumulative distribution function
}

// NewEmpiricalQueueRandomizer creates a new EmpiricalQueueRandomizer
func NewEmpiricalArgSetRandomizer(rg *rand.Rand, qpdf []float64, ecdf [][2]float64) *EmpiricalArgSetRandomizer {
	if len(qpdf) != stochastic.QueueLen {
		return nil
	}
	factor := 1.0 - qpdf[0]
	if factor <= 0 {
		return nil
	}
	cp := make([]float64, stochastic.QueueLen-1)
	for i := range stochastic.QueueLen - 1 {
		cp[i] = qpdf[i+1] / factor
	}
	return &EmpiricalArgSetRandomizer{
		rg:   rg,
		pdf:  cp,
		ecdf: ecdf,
	}
}

// SampleArg samples an argument from a distribution with n possible arguments
func (r *EmpiricalArgSetRandomizer) SampleArg(n ArgumentType) ArgumentType {
	return ArgumentType(continuous_empirical.Sample(r.rg, r.ecdf, int64(n)))
}

// SampleQueue samples an index for a queue
func (r *EmpiricalArgSetRandomizer) SampleQueue() int {
	return discrete_empirical.Sample(r.pdf, r.rg.Float64()) + 1
}

// SnapshotSet interface for snapshot arguments
type SnapshotSet interface {
	SampleSnapshot(n int) int // sample snapshot set for an active snapshot stack size
}

// ProxyRandomizer struct
type EmpiricalSnapshotRandomizer struct {
	rg   *rand.Rand
	ecdf [][2]float64 // empirical cumulative distribution function
}

// NewEmpiricalSnapshotRandomizer creates a new EmpiricalSnapshotRandomizer
func NewEmpiricalSnapshotRandomizer(rg *rand.Rand, ecdf [][2]float64) *EmpiricalSnapshotRandomizer {
	return &EmpiricalSnapshotRandomizer{
		rg:   rg,
		ecdf: ecdf,
	}
}

// SampleSnapshot samples an argument from a distribution with n possible arguments
func (r *EmpiricalSnapshotRandomizer) SampleSnapshot(n int) int {
	return int(continuous_empirical.Sample(r.rg, r.ecdf, int64(n)))
}
