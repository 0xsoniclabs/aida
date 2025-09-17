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

package exponential

import (
	"errors"
	"fmt"
	"math"
	"math/rand"

	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous"
)

// Package for the one-sided truncated exponential distribution with a bound of one.

const (
	newtonError      = 1e-9  // epsilon for Newton's convergences criteria
	newtonMaxStep    = 10000 // maximum number of iteration in the Newtonian
	newtonInitLambda = 1.0   // initial parameter in Newtonion's search
)

// CDF is the cumulative distribution function for the truncated exponential distribution with a bound of 1.
func CDF(lambda float64, x float64) float64 {
	return (math.Exp(-lambda*x) - 1.0) / (math.Exp(-lambda) - 1.0)
}

// ToECDF is a piecewise linear representation of the cumulative distribution function.
func ToECDF(lambda float64, n int) [][2]float64 {
	// The points are equi-distantly spread, i.e., 1/n.
	fn := [][2]float64{}
	for i := 0; i <= n; i++ {
		x := float64(i) / float64(n)
		y := CDF(lambda, x)
		fn = append(fn, [2]float64{x, y})
	}
	return fn
}

// Quantile is the inverse cumulative distribution function.
func Quantile(lambda float64, p float64) float64 {
	return math.Log(p*math.Exp(-lambda)-p+1) / -lambda
}

// Sample samples the distribution and discretizes the result for numbers in the range between 0 and n-1.
func Sample(rg *rand.Rand, lambda float64, n int64) int64 {
	y := int64(float64(n) * Quantile(lambda, rg.Float64()))
	if y < 0 {
		return 0
	} else if y >= n {
		return n - 1
	} else {
		return y
	}
}

// mle is the Maximum Likelihood Estimation function for finding a suitable lambda.
func mle(lambda float64, mean float64) (float64, error) {
	if math.IsNaN(lambda) || math.IsNaN(mean) {
		return 0, errors.New("lambda or mean values are not a number")
	}
	t := 1 / (math.Exp(lambda) - 1)
	// ensure that exponent calculation is stable
	if math.IsNaN(t) {
		// If numerical limits are reached, replace with symbolic limits.
		if lambda >= 1.0 {
			t = 0
		} else {
			// assuming that for very small values of lambda, a NaN is produced.
			t = 1.0
		}
	}
	return 1/lambda - t - mean, nil
}

// dMLE computes the derivative of the Maximum Likelihood Estimation function.
func dMLE(lambda float64) (float64, error) {
	if math.IsNaN(lambda) {
		return 0, errors.New("lambda is not a number")
	}

	t := math.Exp(lambda) / math.Pow(math.Exp(lambda)-1, 2)
	// ensure that exponent calculation is stable
	if math.IsNaN(t) {
		// If numerical limits are reached, replace by symbolic limits.
		t = 1.0
	}
	return t - 1/(lambda*lambda), nil
}

// ApproximateLambda performs a classical Newtonian to determine
// the lambda value since the MLE function is a transcendental
// functions and no closed form can be found. The function returns either
// lambda if it is in the epsilon environment (newtonError) or
// an error if the maximal number of steps for the convergence criteria
// is exceeded.
func ApproximateLambda(points [][2]float64) (float64, error) {
	m := continuous.Mean(points)
	l := newtonInitLambda
	for range newtonMaxStep {
		mleValue, err := mle(l, m)
		if err != nil {
			return 0, err
		}

		dMleValue, err := dMLE(l)
		if err != nil {
			return 0, err
		}

		l = l - mleValue/dMleValue
		if math.Abs(mleValue) < newtonError {
			return l, nil
		}
	}
	return 0.0, fmt.Errorf("ApproximateLambda: failed to converge after %v steps", newtonMaxStep)
}
