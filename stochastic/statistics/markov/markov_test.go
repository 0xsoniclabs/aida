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

package markov

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"gonum.org/v1/gonum/stat/distuv"
)

// checkStationaryDistribution tests stationary distribution of a uniform Markovian process
// whose transition probability is 1/n for n states.
func checkStationaryDistribution(t *testing.T, n int) {
	// Create a stochastic matrix where a_ij = 1/n
	A := make([][]float64, n)
	L := make([]string, n)
	for i := range n {
		L[i] = "s" + strconv.Itoa(i)
		A[i] = make([]float64, n)
		for j := range n {
			A[i][j] = 1.0 / float64(n)
		}
	}
	mc, err := New(A, L)
	if err != nil {
		t.Fatalf("Expected an markov chain. Error: %v", err)
	}
	eps := 1e-3
	dist, err := mc.Stationary()
	if err != nil {
		t.Fatalf("Failed to compute stationary distribution. Error: %v", err)
	}
	for i := range n {
		if dist[i] < 0.0 || dist[i] > 1.0 {
			t.Fatalf("Not a probability in distribution.")
		}
		if math.Abs(dist[i]-1.0/float64(n)) > eps {
			t.Fatalf("Failed to compute sufficiently precise stationary distribution.")
		}
	}
}

// TestStationaryDistribution of a Markov Chain
// TestEstimation checks the correcntness of approximating
// a lambda for a discrete CDF.
func TestStationaryDistribution(t *testing.T) {
	for n := 2; n < 10; n++ {
		checkStationaryDistribution(t, n)
	}
}

// TextDeterministicNextState checks transition of a deterministic Markovian process.
func TestDeterministicNextState(t *testing.T) {
	// create markov chain with two states s1 and s2 and their
	// deterministic transition, i.e., s1 -> s2 and s2 -> s1
	mc, err := New([][]float64{{0.0, 1.0}, {1.0, 0.0}}, []string{"s1", "s2"})
	if err != nil {
		t.Fatalf("Expected a markov chain. Error: %v", err)
	}

	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	// sample the transition with a uniform random variable
	// If in state s1, the resulting state should be s2.
	x := rg.Float64()
	if y, err := mc.Sample(0, x); y != 1 || err != nil {
		t.Fatalf("Illegal state transition (row 0)")
	}

	// sample the transition with a uniform random variable
	// If in state s2, the resulting state should be s1.
	x = rg.Float64()
	if y, err := mc.Sample(1, x); y != 0 || err != nil {
		t.Fatalf("Illegal state transition (row 1)")
	}
}

// TextDeterministicNextState2 checks transition of a deterministic Markovian process.
func TestDeterministicNextState2(t *testing.T) {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	mc, err := New(
		[][]float64{ // stochastic matrix
			{0.0, 1.0, 0.0},
			{0.0, 0.0, 1.0},
			{1.0, 0.0, 0.0},
		},
		[]string{ // state labels
			"s1",
			"s2",
			"s3",
		})
	if err != nil {
		t.Fatalf("Expected a markov chain. Error: %v", err)
	}

	x := rg.Float64()
	if y, err := mc.Sample(0, x); y != 1 || err != nil {
		t.Fatalf("Illegal state transition (row 0)")
	}

	x = rg.Float64()
	if y, err := mc.Sample(1, x); y != 2 || err != nil {
		t.Fatalf("Illegal state transition (row 1)")
	}

	x = rg.Float64()
	if y, err := mc.Sample(2, x); y != 0 || err != nil {
		t.Fatalf("Illegal state transition (row 1)")
	}
}

// TextNextStateFail checks whether nextState fails if
// stochastic matrix is broken.
func TestNextStateFail(t *testing.T) {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	var x float64
	mc, err := New([][]float64{
		{0.0, 0.0},
		{math.NaN(), 0.0},
	},
		[]string{"s1", "s2"})
	if err != nil {
		t.Fatalf("Expected a markov chain. Error: %v", err)
	}

	x = rg.Float64()
	if y, err := mc.Sample(0, x); y != -1 || err != nil {
		t.Fatalf("Could not capture faulty stochastic matrix")
	}

	x = rg.Float64()
	if y, err := mc.Sample(1, x); y != -1 || err != nil {
		t.Fatalf("Could not capture faulty stochastic matrix")
	}
}

// checkMarkovChain checks via chi-squared test whether
// transitions are independent using the number of
// observed states. For this test, we assume that all
// rows are identical to avoid the calculation of a stationary
// distribution for an arbitrary matrix. Also the convergence
// is too slow for an arbitrary matrix.
func checkMarkovChain(mc *Chain, numSteps int) error {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	// get number of states
	n := len(mc.a)

	// frequency counts for states
	counts := make([]int, n)

	// run Markovian process for numSteps times
	var err error
	state := 0
	for range numSteps {
		x := rg.Float64()
		state, err = mc.Sample(state, x)
		if state < 0 || state >= n || err != nil {
			return fmt.Errorf("Illegal state %v", state)
		}
		counts[state]++
	}

	// compute chi-squared value for observations
	// We assume that all rows are identical.
	// For arbitrary stochastic matrix, the stationary
	// distribution must be used instead of A[0].
	chi2 := float64(0.0)
	for i, v := range counts {
		expected := float64(numSteps) * mc.a[0][i]
		err := expected - float64(v)
		chi2 += (err * err) / expected
	}

	// Perform statistical test whether uniform Markovian process is unbiased
	// with an alpha of 0.05 and a degree of freedom of n-1 where n is the
	// number of states in the uniform Markovian process.
	alpha := 0.05
	df := float64(n - 1)
	chi2Critical := distuv.ChiSquared{K: df, Src: nil}.Quantile(1.0 - alpha)

	if chi2 > chi2Critical {
		return fmt.Errorf("Statistical test failed. Degree of freedom is %v and chi^2 value is %v; chi^2 critical value is %v", n, chi2, chi2Critical)
	}
	return nil
}

// TestRandomNextState checks whether a uniform Markovian process produces a uniform
// state distribution via a chi-squared test for various number of states.
func TestRandomNextState(t *testing.T) {
	// test small Markov chain by setting up a uniform Markovian process with
	// uniform distributions. The stationary distribution of the uniform
	// Markovian process is (1/n, , ... , 1/n).
	n := 10
	A := make([][]float64, n)
	L := make([]string, n)
	for i := 0; i < n; i++ {
		L[i] = "s" + strconv.Itoa(i)
		A[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			A[i][j] = 1.0 / float64(n)
		}
	}
	mc, err := New(A, L)
	if err != nil {
		t.Fatalf("Expected a markov chain. Error: %v", err)
	}
	if err = checkMarkovChain(mc, n*n); err != nil {
		t.Fatalf("Uniform Markovian process is not unbiased for a small test-case. Error: %v", err)
	}

	// test larger uniform markov chain
	n = 5400
	A = make([][]float64, n)
	L = make([]string, n)
	for i := 0; i < n; i++ {
		L[i] = "s" + strconv.Itoa(i)
		A[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			A[i][j] = 1.0 / float64(n)
		}
	}
	mc, err = New(A, L)
	if err != nil {
		t.Fatalf("Expected markov chain. Error: %v", err)
	}
	if err := checkMarkovChain(mc, 25*n); err != nil {
		t.Fatalf("Uniform Markovian process is not unbiased for a larger test-case. Error: %v", err)
	}

	// Setup a Markovian process with a truncated geometric distributions for
	// next states. The distribution has the following formula:
	//  Pr(X=x_j) = (1-beta)*beta^n * (1-beta^n) / -beta ^ j
	// for values {x_1, ..., x_n}  of random variable X and
	// with distribution parameter beta.
	n = 10
	beta := 0.6
	A = make([][]float64, n)
	L = make([]string, n)
	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
		L[i] = "s" + strconv.Itoa(i)
		for j := 0; j < n; j++ {
			A[i][j] = ((1.0 - beta) * math.Pow(beta, float64(n)) /
				(1.0 - math.Pow(beta, float64(n)))) *
				math.Pow(beta, -float64(j+1))
		}
	}
	mc, err = New(A, L)
	if err != nil {
		t.Fatalf("Expected a markov chain. Error: %v", err)
	}
	if err := checkMarkovChain(mc, n*n); err != nil {
		t.Fatalf("Geometric Markovian process is not unbiased for a small experiment. Error: %v", err)
	}
}

// TestInitialState checks function find
// for returning the correct intial state.
func TestInitialState(t *testing.T) {
	mc, err := New(
		[][]float64{ // stochastic matrix
			{0.0, 1.0, 0.0},
			{0.0, 0.0, 1.0},
			{1.0, 0.0, 0.0},
		},
		[]string{ // state labels
			"A",
			"B",
			"C",
		})
	if err != nil {
		t.Fatalf("Expected a markov chain. Error: %v", err)
	}

	if x, err := mc.FindState("A"); x != 0 || err != nil {
		t.Fatalf("Cannot find first state A")
	}
	if x, err := mc.FindState("B"); x != 1 || err != nil {
		t.Fatalf("Cannot find first state B")
	}
	if x, err := mc.FindState("C"); x != 2 || err != nil {
		t.Fatalf("Cannot find first state C")
	}
	if x, err := mc.FindState("D"); x != -1 || err != nil {
		t.Fatalf("Should not find first state D")
	}
	if x, err := mc.FindState(""); x != -1 || err != nil {
		t.Fatalf("Should not find first state")
	}
}
