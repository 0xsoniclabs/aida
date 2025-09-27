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
	"math"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/stat/distuv"
)

// TestDiscrete_CheckPMF checks if the given probability mass function (pmf) is valid.
func TestDiscrete_CheckPMF(t *testing.T) {
	pmf := []float64{0.2, 0.5, 0.3}
	if err := Check(pmf); err != nil {
		t.Fatalf("valid pmf: want nil, got %v", err)
	}
	pmf = []float64{0.0, 1.0, 0.0}
	if err := Check(pmf); err != nil {
		t.Fatalf("valid pmf with zeros: want nil, got %v", err)
	}
	pmf = []float64{0.0, 0.0, 0.0}
	if err := Check(pmf); err == nil {
		t.Fatalf("all zeros pmf: want error, got nil")
	}
	pmf = []float64{-1.0, 0.0, 0.0}
	if err := Check(pmf); err == nil {
		t.Fatalf("negative number in pmf: want error, got nil")
	}
	pmf = []float64{1.1, 0.0, 0.0}
	if err := Check(pmf); err == nil {
		t.Fatalf("probability greater than one: want error, got nil")
	}
	pmf = []float64{math.NaN(), 0.0, 0.0}
	if err := Check(pmf); err == nil {
		t.Fatalf("a probability as NaN: want error, got nil")
	}
}

// TestDiscrete_QuantileBasic tests the Quantile function.
func TestDiscrete_QuantileBasic(t *testing.T) {
	pmf := []float64{0.2, 0.3, 0.5}
	if got := Quantile(pmf, 0.0); got != 0 {
		t.Fatalf("u=0.0: want 0, got %d", got)
	}
	if got := Quantile(pmf, 0.2); got != 0 {
		t.Fatalf("u=0.2 (boundary): want 0, got %d", got)
	}
	if got := Quantile(pmf, 0.4); got != 1 {
		t.Fatalf("u=0.4: want 1, got %d", got)
	}
	if got := Quantile(pmf, 0.8); got != 2 {
		t.Fatalf("u=0.8: want 2, got %d", got)
	}
}

// TestDiscrete_QuantileReturnsLastPositiveWhenUSurpassesTotal tests the Sample function
func TestDiscrete_QuantileReturnsLastPositiveWhenUSurpassesTotal(t *testing.T) {
	pmf := []float64{0.1, 0.0, 0.2}
	if got := Quantile(pmf, 0.999); got != 2 {
		t.Fatalf("u>sum: want last positive index 2, got %d", got)
	}
	pmf = []float64{0.0, 0.7, 0.0}
	if got := Quantile(pmf, 0.9); got != 1 {
		t.Fatalf("u>sum: want last positive index 1, got %d", got)
	}
}

// TestDiscrete_QuantileAllZerosAndEmpty tests the Sample function with all-zero and empty pmfs.
func TestDiscrete_QuantileAllZerosAndEmpty(t *testing.T) {
	pmfZeros := []float64{0.0, 0.0, 0.0}
	if got := Quantile(pmfZeros, 0.5); got != 0 {
		t.Fatalf("all zeros: want 0, got %d", got)
	}
	var pmfEmpty []float64
	if got := Quantile(pmfEmpty, 0.3); got != 0 {
		t.Fatalf("empty pmf: want 0, got %d", got)
	}
}

// TestDiscrete_QuantileNumericalStabilityKahanPathIsExercised tests the Sample function for numerical stability.
func TestDiscrete_QuantileNumericalStabilityKahanPathIsExercised(t *testing.T) {
	pmf := []float64{
		1e-16, 1e-16, 1e-16, 1e-16,
		0.25, 0.25, 0.25, 0.25,
	}
	if got := Quantile(pmf, 5e-16); got != 4 {
		t.Fatalf("u~tiny: want 4, got %d", got)
	}
	if got := Quantile(pmf, 0.4); got != 5 {
		t.Fatalf("u=0.4: want 5, got %d", got)
	}
	if got := Quantile(pmf, 1.0-math.SmallestNonzeroFloat64); got != 7 {
		t.Fatalf("u≈1: want 7, got %d", got)
	}
}

// TestDiscrete_ShrinkPMFBasic tests the ShrinkPMF function with a basic pmf.
func TestDiscrete_ShrinkPMFBasic(t *testing.T) {
	pmf := []float64{0.2, 0.7}
	shrunk, err := Shrink(pmf)
	if err == nil {
		t.Fatalf("invalid pmf (sum<1): want error, got nil")
	}
	pmf = []float64{1.0, 0.0, 0.0}
	shrunk, err = Shrink(pmf)
	if err == nil {
		t.Fatalf("invalid scaling factor: want error, got nil")
	}
	pmf = []float64{1.0}
	shrunk, err = Shrink(pmf)
	if err == nil {
		t.Fatalf("too short pmf: want error, got nil")
	}
	pmf = []float64{0.1, 0.2, 0.7}
	shrunk, err = Shrink(pmf)
	if err != nil {
		t.Fatalf("valid pmf: want nil error, got %v", err)
	}
	expected := []float64{0.22222222222222224, 0.7777777777777778}
	if len(shrunk) != len(expected) {
		t.Fatalf("shrunk pmf length: want %d, got %d", len(expected), len(shrunk))
	}
	for i := range len(expected) {
		if math.Abs(shrunk[i]-expected[i]) > 1e-9 {
			t.Fatalf("shrunk pmf[%d]: want %g, got %g", i, expected[i], shrunk[i])
		}
	}
}

// testSample performs statistical tests on the Sample function to ensure it behaves as expected.
func testSample(f []float64, t *testing.T) {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	// check that the PMF is valid
	if err := Check(f); err != nil {
		t.Fatalf("The ECDF is not valid. Error: %v", err)
	}

	// parameters
	numSteps := 100000
	n := len(f)

	// populate buckets
	counts := make([]int64, n)
	for range numSteps {
		counts[Sample(rg, f)]++
	}

	// compute chi-squared value for observations
	chi2 := float64(0.0)
	for i, v := range counts {
		// compute expected value of bucket
		expected := float64(numSteps) * f[i]
		err := expected - float64(v)
		chi2 += (err * err) / expected
		//fmt.Printf("Err: %v %v %v\n", i, v, expected)
	}

	// Perform statistical test whether the sampling is unbiased
	// with an alpha of 0.05 and a degree of freedom of the number of buckets minus one.
	alpha := 0.05
	df := float64(n - 1)
	chi2Critical := distuv.ChiSquared{K: df, Src: nil}.Quantile(1.0 - alpha)
	//fmt.Printf("Chi^2 value: %v Chi^2 critical value: %v df: %v\n", chi2, chi2Critical, df)

	if chi2 > chi2Critical {
		t.Fatalf("The random index selection biased.")
	}
}

// TestSample_Statistical tests the Sample function with a statistical test.
func TestSample_Statistical(t *testing.T) {
	t.Run("PMF1", func(t *testing.T) {
		pmf := []float64{0.1, 0.2, 0.3, 0.4}
		testSample(pmf, t)
	})
	t.Run("PMF2", func(t *testing.T) {
		pmf := []float64{0.0, 0.0, 1.0, 0.0, 0.0}
		testSample(pmf, t)
	})
	t.Run("PMF3", func(t *testing.T) {
		pmf := []float64{0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.92}
		testSample(pmf, t)
	})
	t.Run("PMF4", func(t *testing.T) {
		pmf := []float64{0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.992}
		testSample(pmf, t)
	})
}
