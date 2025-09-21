// Copyright 2025 Sonic Labs
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

package discrete

import (
	"fmt"
	"math"
	"math/rand"
)

// Check if the given probability mass function (pmf) of a discrete
// finite random variable is valid.  A valid pmf has all probabilities
// in the range [0,1], and the sum of all probabilities must be 1.
func Check(f []float64) error {
	total := 0.0
	for i := range len(f) {
		x := f[i]
		if x < 0.0 || x > 1.0 || math.IsNaN(x) {
			return fmt.Errorf("Invalid probability (%v) in the pmf", x)
		}
		total += x
	}
	if math.Abs(total-1.0) > 1e-9 {
		return fmt.Errorf("Total is not one (%v)", total)
	}
	return nil
}

// Quantile computes the quantile (inverse CDF) for a discrete finite random variable.
// The discrete finite random variable is given by probability mass functions (pmf).
// It returns the index i such that the cumulative probability up to and including i
// is at least u. Kahn's summation is used to reduce numerical errors in the summation.
func Quantile(f []float64, u float64) int {
	sum := 0.0 // Kahn's summation algorithm for probability sum
	c := 0.0   // Compensation term of Kahn's algorithm
	lastPositive := -1
	for i := range len(f) {
		p := f[i]
		y := p - c
		t := sum + y
		c = (t - sum) - y
		sum = t
		if u <= sum {
			return i
		}
		if f[i] > 0.0 {
			lastPositive = i
		}
	}
	if lastPositive != -1 {
		return lastPositive
	}
	return 0 // default position if all probabilities are zero
}

// Sample the discrete finite random variable defined by the given pmf.
func Sample(rg *rand.Rand, f []float64) int {
	return Quantile(f, rg.Float64())
}

// Shrink removes the first element from the given probability mass function (pmf)
// and rescales the remaining elements so that they are a pmf again.
func Shrink(f []float64) ([]float64, error) {
	n := len(f)
	if n < 2 {
		return nil, fmt.Errorf("PMF is too short (%d)", n)
	}
	if err := Check(f); err != nil {
		return nil, err
	}
	factor := 1.0 - f[0]
	if math.Abs(factor) < 1e-9 || math.IsNaN(factor) {
		return nil, fmt.Errorf("Invalid scaling factor (%v)", factor)
	}
	scaled_pmf := make([]float64, n-1)
	for i := range n - 1 {
		x := f[i+1] / factor
		scaled_pmf[i] = x
	}
	if err := Check(scaled_pmf); err != nil {
		return nil, err
	}
	return scaled_pmf, nil
}
