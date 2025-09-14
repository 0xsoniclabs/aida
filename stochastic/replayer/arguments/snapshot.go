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
	"math/rand"

	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous_empirical"
)

// SnapshotSet interface for snapshot arguments
type SnapshotSet interface {
	SampleSnapshot(n int) int // sample snapshot set for an active snapshot stack size
}

// EmpiricalSnapshotRandomizer struct for snapshot arguments
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
