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
	"sort"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/simplify"
)

// Argument counting statistics counts the occurrence of arguments in StateDB operations
// (e.g. account addresses, storage keys, storage values).  The counting statistics provides
// an empirical cumulative distribution function (ECDF) of the argument frequencies. The
// ECDF is computed using the Visvalingam-Whyatt algorithm to reduce the number of points
// to a manageable size. The number of arguments in a StateDB can be very large and
// hence we compress it the distribution function to a fixed number of points.
// See: https://en.wikipedia.org/wiki/Visvalingam-Whyatt_algorithm

// NumECDFPoints sets the number of points in the empirical cumulative distribution function.
const NumECDFPoints = 300

// argCount data struct for a counting statistics of StateDB operations' arguments.
type argCount[T comparable] struct {
	freq map[T]uint64 // frequency counts per argument
}

// newArgCount creates a new counting statistics for numbers.
func newArgCount[T comparable]() argCount[T] {
	return argCount[T]{map[T]uint64{}}
}

// Places an item into the counting statistics.
func (s *argCount[T]) Place(data T) {
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

// newArgStatsJSON computes the ECDF of the counting stats.
func (s *argCount[T]) newArgStatsJSON() ArgStatsJSON {
	return s.produceJSON(NumECDFPoints)
}

// produceJSON computes the ECDF and set the number field in the JSON struct.
func (s *argCount[T]) produceJSON(numPoints int) ArgStatsJSON {
	// sort frequency entries for arguments by frequency (highest frequency first)
	n := len(s.freq)
	args := make([]T, 0, n)
	total := uint64(0)
	for arg, freq := range s.freq {
		args = append(args, arg)
		total += freq
	}
	sort.SliceStable(args, func(i, j int) bool {
		return s.freq[args[i]] > s.freq[args[j]]
	})
	var compressFreqs orb.LineString
	if n > 0 {
		ls := orb.LineString{}
		// print points of the empirical cumulative freq
		sumP := float64(0.0)
		// Correction term for Kahan's sum
		cP := float64(0.0)
		// add first point to line string
		ls = append(ls, orb.Point{0.0, 0.0})
		// iterate through all items
		for i := range n {
			// Implement Kahan's summation to avoid errors
			// for accumulated probabilities (they might be very small)
			// https://en.wikipedia.org/wiki/Kahan_summation_algorithm
			f := float64(s.freq[args[i]]) / float64(total)
			x := (float64(i) + 0.5) / float64(n)
			yP := f - cP
			tP := sumP + yP
			cP = (tP - sumP) - yP
			sumP = tP
			// add new point to Ecdf
			ls = append(ls, orb.Point{x, sumP})
		}
		// add last point
		ls = append(ls, orb.Point{1.0, 1.0})
		// reduce full ecdf using Visvalingam-Whyatt algorithm to
		// "numPoints" points. See:
		// https://en.wikipedia.org/wiki/Visvalingam-Whyatt_algorithm
		simplifier := simplify.VisvalingamKeep(numPoints)
		compressFreqs = simplifier.Simplify(ls).(orb.LineString)
	}
	ecdf := make([][2]float64, len(compressFreqs))
	for i := range compressFreqs {
		ecdf[i] = [2]float64(compressFreqs[i])
	}
	return ArgStatsJSON{
		N:    int64(n),
		ECDF: ecdf,
	}
}
