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

package discrete_empirical

import (
	"fmt"
	"math"

	"github.com/0xsoniclabs/aida/stochastic"
)

// For a given probability density function (pdf) and a uniform random number u in [0,1),
// Sample returns an index i such that the cumulative distribution function (CDF)
// at i is greater than or equal to u. The pdf is represented as a slice of float64,
// where each element corresponds to the probability of the respective index.
// The function uses Kahn's summation algorithm to ensure numerical stability when
// summing the probabilities. If all probabilities are zero, it defaults to returning 1.
func Sample(pdf []float64, u float64) int {
	sum := 0.0 // Kahn's summation algorithm for probability sum
	c := 0.0   // Compensation term of Kahn's algorithm
	lastPositive := -1
	for i := range len(pdf) {
		p := pdf[i]
		// skip if not a probability in the range [0,1]
		if p != math.NaN() && p >= 0.0 && p <= 1.0 {
			y := p - c
			t := sum + y
			c = (t - sum) - y
			sum = t
			if u <= sum {
				return i
			}
			if pdf[i] > 0.0 {
				lastPositive = i
			}
		}
	}
	if lastPositive != -1 {
		return lastPositive
	}
	return 0 // default position if all probabilities are zero
}

// ShrinkPdf removes the first element from the given probability density function (pdf)
func ShrinkPdf(pdf []float64) ([]float64, error) {
	if len(pdf) != stochastic.QueueLen {
		return nil, fmt.Errorf("invalid pdf length: %d, expected: %d", len(pdf), stochastic.QueueLen)
	}
	factor := 1.0 - pdf[0]
	//if factor > 1.0 {
	//	return nil, fmt.Errorf("Invalid scaling propability (%v)", factor)
	//	}
	scaled_pdf := make([]float64, stochastic.QueueLen-1)
	for i := range stochastic.QueueLen - 1 {
		if pdf[i+1] <= 0 || pdf[i+1] >= 1.0 || math.IsNaN(pdf[i+1]) {
			return nil, fmt.Errorf("Invalid scaling propability (%v)", factor)
		}
		scaled_pdf[i] = pdf[i+1] / factor
	}
	return scaled_pdf, nil
}
