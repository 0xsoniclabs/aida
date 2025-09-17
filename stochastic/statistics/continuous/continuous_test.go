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

package continuous

import (
	"math"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/stat/distuv"
)

func almostEqual(a, b float64) bool {
	const eps = 1e-12
	return math.Abs(a-b) <= eps
}

func TestCDF_PiecewiseInterpolationAndBoundaries(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.25, 0.1},
		{0.6, 0.7},
		{1.0, 1.0},
	}
	if v := CDF(f, 0.0); !almostEqual(v, 0.0) {
		t.Fatalf("CDF at x=0.0: want 0.0, got %g", v)
	}
	if v := CDF(f, 0.125); !almostEqual(v, 0.05) {
		t.Fatalf("CDF at x=0.125: want 0.05, got %g", v)
	}
	if v := CDF(f, 0.25); !almostEqual(v, 0.1) {
		t.Fatalf("CDF at x=0.25 (boundary): want 0.1, got %g", v)
	}
	if v := CDF(f, 0.40); !almostEqual(v, 0.35714285714285715) {
		t.Fatalf("CDF at x=0.40: want ~0.3571428571, got %g", v)
	}
	if v := CDF(f, 1.2); !almostEqual(v, 1.0) {
		t.Fatalf("CDF at x=1.2 (>1): want 1.0, got %g", v)
	}
}

func TestCDF_ExtrapolatesForNegativeX(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
	}
	if v := CDF(f, -0.1); !almostEqual(v, -0.1) {
		t.Fatalf("CDF at x=-0.1: want -0.1 (extrapolated), got %g", v)
	}
}

func TestCheckPiecewiseLinearCDF_Valid(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.2, 0.1},
		{0.8, 0.9},
		{1.0, 1.0},
	}
	if err := CheckECDF(f); err != nil {
		t.Fatalf("expected valid CDF definition. Error: %v", err)
	}
}

func TestCheckPiecewiseLinearCDF_BadStart(t *testing.T) {
	f := [][2]float64{
		{0.001, 0.0},
		{0.5, 0.4},
		{1.0, 1.0},
	}
	if err := CheckECDF(f); err == nil {
		t.Fatalf("expected invalid CDF due to bad start point. Error: %v", err)
	}
}

func TestCheckPiecewiseLinearCDF_BadEnd(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.4},
		{0.999, 0.999},
	}
	if err := CheckECDF(f); err == nil {
		t.Fatalf("expected invalid CDF due to bad end point")
	}
}

func TestCheckPiecewiseLinearCDF_NonIncreasingX(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.4},
		{0.5, 0.6},
		{1.0, 1.0},
	}
	if err := CheckECDF(f); err != nil {
		t.Fatalf("expected invalid CDF due to non-increasing x. Error: %v", err)
	}
}

// TestEmpiricalCDF checks the evaluation of the CDF for a uniform distribution.
func TestEmpiricalCDF(t *testing.T) {
	ecdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}} // uniform distribution
	n := 10000
	for i := range n {
		x := float64(i) / float64(n)
		if v := CDF(ecdf, x); !almostEqual(v, x) {
			t.Fatalf("CDF at x=%v: want %v, got %v", x, x, v)
		}
	}
}

// TestEmpiricalQuantile checks the evaluation of the quantile function for a uniform distribution.
func TestEmpiricalQuantile(t *testing.T) {
	ecdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}} // uniform distribution
	n := 10000
	for i := range n {
		y := float64(i) / float64(n)
		if v := Quantile(ecdf, y); !almostEqual(v, y) {
			t.Fatalf("Quantile at y=%v: want %v, got %v", y, y, v)
		}
	}
}

// testSample checks the randomness of sampling for an empirical cumulative distribution function
func testSample(ecdf [][2]float64, t *testing.T) {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	// check that the ECDF is valid
	if err := CheckECDF(ecdf); err != nil {
		t.Fatalf("The ECDF is not valid. Error: %v", err)
	}

	// parameters
	numSteps := 10000
	idxRange := int64(10)

	// populate buckets
	counts := make([]int64, idxRange)
	for range numSteps {
		counts[Sample(rg, ecdf, idxRange)]++
	}

	// compute chi-squared value for observations
	chi2 := float64(0.0)
	for i, v := range counts {
		// compute expected value of bucket
		p := CDF(ecdf, float64(i+1)/float64(idxRange)) - CDF(ecdf, float64(i)/float64(idxRange))
		expected := float64(numSteps) * p
		err := expected - float64(v)
		chi2 += (err * err) / expected
		// fmt.Printf("Err: %v %v %v\n", i, v, expected)
	}

	// Perform statistical test whether the empirical distribution is unbiased
	// with an alpha of 0.05 and a degree of freedom of the number of buckets minus one.
	alpha := 0.05
	df := float64(idxRange - 1)
	chi2Critical := distuv.ChiSquared{K: df, Src: nil}.Quantile(1.0 - alpha)
	// fmt.Printf("Chi^2 value: %v Chi^2 critical value: %v df: %v\n", chi2, chi2Critical, df)

	if chi2 > chi2Critical {
		t.Fatalf("The random index selection biased.")
	}
}

// testCDFQuantileInverse checks the inverse property of the CDF and the quantile function.
func testCDFQuantileInverse(ecdf [][2]float64, t *testing.T) {
	n := 100000
	for i := range n {
		y := float64(i) / float64(n)
		x := Quantile(ecdf, y)
		y2 := CDF(ecdf, x)
		if !almostEqual(y, y2) {
			t.Fatalf("Quantile/CDF inverse at y=%v: want %v, got %v", y, y, y2)
		}
	}
}

// testQuantileCDFInverse checks the inverse property of the CDF and the quantile function.
func testQuantileCDFInverse(ecdf [][2]float64, t *testing.T) {
	n := 100000
	for i := range n {
		x := float64(i) / float64(n)
		y := CDF(ecdf, x)
		x2 := Quantile(ecdf, y)
		if !almostEqual(x, x2) {
			t.Fatalf("CDF/Quantile inverse at x=%v: want %v, got %v", x, x, x2)
		}
	}
}

// TestEmpirical checks the sampling using the chi2 test for various
// empirical cumulative distribution functions and the inverse property of
// the CDF and the quantile function.
func TestEmpirical(t *testing.T) {
	ecdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}
	t.Run("Uniform", func(t *testing.T) {
		testSample(ecdf, t)
		testCDFQuantileInverse(ecdf, t)
		testQuantileCDFInverse(ecdf, t)
	})
	ecdf = [][2]float64{{0.0, 0.0}, {0.5, 0.1}, {1.0, 1.0}}
	t.Run("Skewed", func(t *testing.T) {
		testSample(ecdf, t)
		testCDFQuantileInverse(ecdf, t)
		testQuantileCDFInverse(ecdf, t)
	})
	ecdf = [][2]float64{{0.0, 0.0}, {0.1, 0.5}, {1.0, 1.0}}
	t.Run("SkewedOtherWay", func(t *testing.T) {
		testSample(ecdf, t)
		testCDFQuantileInverse(ecdf, t)
		testQuantileCDFInverse(ecdf, t)
	})
	ecdf = [][2]float64{{0.0, 0.0}, {0.1, 0.1}, {0.5, 0.5}, {0.9, 0.9}, {1.0, 1.0}}
	t.Run("PiecewiseLinear", func(t *testing.T) {
		testSample(ecdf, t)
		testCDFQuantileInverse(ecdf, t)
		testQuantileCDFInverse(ecdf, t)
	})
	ecdf = [][2]float64{}
	for i := range 1001 {
		x := float64(i) / float64(1000)
		ecdf = append(ecdf, [2]float64{x, math.Sqrt(x)})
	}
	t.Run("SquareRoot", func(t *testing.T) {
		testSample(ecdf, t)
		testCDFQuantileInverse(ecdf, t)
		testQuantileCDFInverse(ecdf, t)
	})
	ecdf = [][2]float64{}
	for i := range 1001 {
		x := float64(i) / float64(1000)
		ecdf = append(ecdf, [2]float64{x, x * x})
	}
	t.Run("Square", func(t *testing.T) {
		testSample(ecdf, t)
		testCDFQuantileInverse(ecdf, t)
		testQuantileCDFInverse(ecdf, t)
	})
}
