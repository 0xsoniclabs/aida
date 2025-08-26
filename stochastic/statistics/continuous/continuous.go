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

package continuous

import (
	"fmt"
	"math/rand"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/simplify"
)

// CDF computes the Cumulative Distribution Function of parameter x
// for a given piecewise linear function. The piecewise linear function
// approximates a continuous CDF with linear segments in the range [0,1].
// The piecewise linear function is given as a list of points (x_i, y_i).
// The first point is (0,0) and the last point is (1,1). Other points must
// be strictly increasing monotonic, i.e., for all i: x_i < x_{i+1} and
// y_i < y_{i+1}.
func CDF(f [][2]float64, x float64) float64 {
	if x <= 0 {
		return 0.0
	}
	for i := range len(f) - 1 {
		if f[i+1][0] >= x {
			scale := (x - f[i][0]) / (f[i+1][0] - f[i][0])
			return f[i][1] + scale*(f[i+1][1]-f[i][1])
		}
	}
	return 1.0 // x is 1.0 or greater
}

// Quantile computes the inverse Cumulative Distribution Function of parameter y
// for a cdf given as a piecewise linear function.
func Quantile(f [][2]float64, y float64) float64 {
	if y <= 0 {
		return 0.0
	}
	for i := range len(f) - 1 {
		if f[i+1][1] >= y {
			scale := (y - f[i][1]) / (f[i+1][1] - f[i][1])
			return f[i][0] + scale*(f[i+1][0]-f[i][0])
		}
	}
	return 1.0 // y is 1.0 or greater
}

// Sample draws a random sample from a piecewise linear CDF using inverse transform sampling.
func Sample(rg *rand.Rand, ecdf [][2]float64, n int64) int64 {
	return int64(float64(n) * Quantile(ecdf, rg.Float64()))
}

// Check whether the piecewise linear function is valid as a CDF.
// The function must start at (0,0) and end at (1,1).
// The points of the function must be monotonically increasing.
func Check(f [][2]float64) error {
	if len(f) < 2 {
		return fmt.Errorf("CDF must have at least start and end point")
	}
	// check start point; must be the coordinate (0,0)
	if f[0] != [2]float64{0.0, 0.0} {
		return fmt.Errorf("CDF must start at (0,0), but starts at (%v,%v)", f[0][0], f[0][1])
	}
	// check end point; must be the coordinate (1,1)
	last := len(f) - 1
	if f[last] != [2]float64{1.0, 1.0} {
		return fmt.Errorf("CDF must end at (1,1), but ends at (%v,%v)", f[last][0], f[last][1])
	}
	// check monotonicity condition of points
	for i := range len(f) - 1 {
		if f[i][0] >= f[i+1][0] && f[i][1] >= f[i+1][1] {
			return fmt.Errorf("CDF points must be strictly monotonically increasing, but point %v (%v,%v) is not smaller than point %v (%v,%v)", i, f[i][0], f[i][1], i+1, f[i+1][0], f[i+1][1])
		}
	}
	return nil
}

// PDFtoCDF computes the empirical cumulative distribution function (ECDF) from a given
// probability density function (PDF). The PDF is given as a list of points (x_i, p_i),
// where x_i is the domain of the random variable (in the range [0,1]) and p_i is the
// probability. The function returns the CDF as a list of points (x_i, p_i),
// where p_i is the cumulative propability up to x_i. The CDF is compressed using
// the Visvalingam-Whyatt algorithm
// to reduce the number of points to a manageable size defined by stochastic.NumECDFPoints.
// The function panics if the resulting ECDF is not valid.
func PDFtoCDF(pdf [][2]float64) [][2]float64 {
	// sort frequency entries for arguments by frequency (highest frequency first)
	n := len(pdf)
	var compressedECDF orb.LineString
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
		x := pdf[i][0]
		f := pdf[i][1]
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
	compressedECDF = simplifier.Simplify(ls).(orb.LineString)
	ecdf := make([][2]float64, len(compressedECDF))
	for i := range compressedECDF {
		ecdf[i] = [2]float64(compressedECDF[i])
	}
	if err := Check(ecdf); err != nil {
		panic(fmt.Sprintf("PDFtoCDF: cannot create valid CDF from counting statistics; Error %v", err))
	}
	return ecdf
}
