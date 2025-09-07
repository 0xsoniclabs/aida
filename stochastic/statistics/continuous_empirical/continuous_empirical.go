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

package continous_empiricial

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
