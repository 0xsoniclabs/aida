// Copyright 2024 Fantom Foundation
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

package classifier

import (
	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous_empirical"
)

// Argument counting statistics counts the occurrence of arguments in StateDB operations
// (e.g. account addresses, storage keys, storage values).  The counting statistics provides
// an empirical cumulative distribution function (ECDF) of the argument frequencies. The
// ECDF is computed using the Visvalingam-Whyatt algorithm to reduce the number of points
// to a manageable size. The number of arguments in a StateDB can be very large and
// hence we compress it the distribution function to a fixed number of points.
// See: https://en.wikipedia.org/wiki/Visvalingam-Whyatt_algorithm

// argCount data struct for a counting statistics of StateDB operations' arguments.
type argCount[T comparable] struct {
	freq map[T]uint64 // frequency counts per argument
}

// newArgCount creates a new counting statistics for numbers.
func newArgCount[T comparable]() argCount[T] {
	return argCount[T]{map[T]uint64{}}
}

// Places an item into the counting statistics.
func (s *argCount[T]) place(data T) {
	s.freq[data]++
}

// exists check whether data item exists in the counting statistics.
func (s *argCount[T]) exists(data T) bool {
	_, ok := s.freq[data]
	return ok
}

// JSON output for a argument counting statistics.
type ArgStatsJSON struct {
	N    int64        `json:"n"`    // Number of data entries
	ECDF [][2]float64 `json:"ecdf"` // Empirical cumulative distribution function
}

// json computes the ECDF of the counting stats.
func (s *argCount[T]) json() ArgStatsJSON {
	ecdf := continuous_empirical.ToECDF(&s.freq)
	return ArgStatsJSON{
		N:    int64(len(s.freq)),
		ECDF: ecdf,
	}
}
