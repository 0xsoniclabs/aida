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

package continuous_empirical

import (
	"math/rand"
	"sort"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/simplify"
)

// CDF computes the Cumulative Distribution Function of parameter x
// for a given piecewise linear function. The piecewise linear function
// approximates a continuous CDF with linear segments in the range [0,1].
// The piecewise linear function is given as a list of points (x_i, y_i).
// The first point is (0,0) and the last point is (1,1). Other points must
// be strict increasingly monotonic. For all i: x_i < x_{i+1} and y_i < y_{i+1}.
// TODO: for large number of points, use binary search.
func CDF(f [][2]float64, x float64) float64 {
	for i := range len(f) - 1 {
		if f[i+1][0] >= x {
			scale := (x - f[i][0]) / (f[i+1][0] - f[i][0])
			return f[i][1] + scale*(f[i+1][1]-f[i][1])
		}
	}
	return 1.0 // x is 1.0 or greater
}

// Inverse CDF
func Quantile(f [][2]float64, y float64) float64 {
	for i := range len(f) - 1 {
		if f[i+1][1] >= y {
			scale := (y - f[i][1]) / (f[i+1][1] - f[i][1])
			return f[i][0] + scale*(f[i+1][0]-f[i][0])
		}
	}
	return 1.0 // x is 1.0 or greater
}

// Sample
func Sample(rg *rand.Rand, ecdf [][2]float64, n int64) int64 {
	return int64(float64(n) * Quantile(ecdf, rg.Float64()))
}

// Check whether the piecewise linear function is valid as a CDF.
// The function must start at (0,0) and end at (1,1).
// The points of the function must be monotonically increasing.
func CheckPiecewiseLinearCDF(f [][2]float64) bool {
	// check start point; must be the coordinate (0,0)
	if f[0][0] != 0.0 || f[0][1] != 0 {
		return false
	}
	// check end point; must be the coordinate (1,1)
	last := len(f) - 1
	if f[last][0] != 1 || f[last][1] != 1 {
		return false
	}
	// check monotonicity condition of points
	for i := range len(f) - 1 {
		if f[i][0] >= f[i+1][0] {
			return false
		}
		if f[i][0] >= f[i+1][0] {
			return false
		}
	}
	return true
}

// ToECDF computes the empirical cumulative distribution function (eCDF)
// from a counting statistics. The eCDF is represented as a piecewise linear
// function with a fixed number of points (NumECDFPoints). The eCDF is
// computed using the Visvalingam-Whyatt algorithm to reduce the number of
// points in the eCDF. See:
// https://en.wikipedia.org/wiki/Visvalingam-Whyatt_algorithm
func ToCountECDF(count *map[int]uint64) [][2]float64 {

	// determine the maximum argument and total frequency
	totalFreq := uint64(0)
	maxArg := 0
	for arg, freq := range *count {
		totalFreq += freq
		if maxArg < arg {
			maxArg = arg
		}
	}

	var simplified orb.LineString

	// if no data-points, nothing to plot
	if len(*count) > 0 {

		// construct full eCdf as LineString
		ls := orb.LineString{}

		// print points of the empirical cumulative freq
		sumP := float64(0.0)

		// Correction term for Kahan's sum
		cP := float64(0.0)

		// add first point to line string
		ls = append(ls, orb.Point{0.0, 0.0})

		// iterate through all deltas
		for arg := 0; arg <= maxArg; arg++ {
			// Implement Kahan's summation to avoid errors
			// for accumulated probabilities (they might be very small)
			// https://en.wikipedia.org/wiki/Kahan_summation_algorithm
			f := float64((*count)[arg]) / float64(totalFreq)
			x := float64(arg) / float64(maxArg)

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
		simplifier := simplify.VisvalingamKeep(stochastic.NumECDFPoints)
		simplified = simplifier.Simplify(ls).(orb.LineString)
	}

	// convert orb.LineString to [][2]float64
	ecdf := make([][2]float64, len(simplified))
	for i := range simplified {
		ecdf[i] = [2]float64(simplified[i])
	}
	return ecdf
}

func ToECDF[T comparable](count *map[T]uint64) [][2]float64 {
	// sort frequency entries for arguments by frequency (highest frequency first)
	n := len(*count)
	args := make([]T, 0, n)
	total := uint64(0)
	for arg, freq := range *count {
		args = append(args, arg)
		total += freq
	}
	sort.SliceStable(args, func(i, j int) bool {
		return (*count)[args[i]] > (*count)[args[j]]
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
			f := float64((*count)[args[i]]) / float64(total)
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
		simplifier := simplify.VisvalingamKeep(stochastic.NumECDFPoints)
		compressFreqs = simplifier.Simplify(ls).(orb.LineString)
	}
	ecdf := make([][2]float64, len(compressFreqs))
	for i := range compressFreqs {
		ecdf[i] = [2]float64(compressFreqs[i])
	}
	return ecdf
}
