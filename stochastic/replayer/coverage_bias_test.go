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

package replayer

import (
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/coverage"
	"github.com/0xsoniclabs/aida/stochastic/statistics/markov"
	"github.com/stretchr/testify/require"
)

func TestCoverageBiasBoost(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)
	require.Equal(t, 1.0, bias.weight(0))

	err = bias.boost(0, coverage.Delta{
		NewUnits:         3,
		NewLines:         1,
		CoverageIncrease: 0.05,
	})
	require.NoError(t, err)
	require.Greater(t, bias.weight(0), 1.0)

	// Verify stats tracking
	totalBoosts, totalNewUnits, totalNewLines := bias.stats()
	require.Equal(t, uint64(1), totalBoosts)
	require.Equal(t, uint64(3), totalNewUnits)
	require.Equal(t, uint64(1), totalNewLines)
}

func TestCoverageBiasClamp(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		err = bias.boost(1, coverage.Delta{
			NewUnits:         10,
			NewLines:         5,
			CoverageIncrease: 0.5,
		})
		require.NoError(t, err)
	}
	require.InDelta(t, bias.maxWeight, bias.weight(1), 1e-9)
}

func TestCoverageBiasNilChain(t *testing.T) {
	_, err := newCoverageBias(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be nil")
}

func TestCoverageBiasSample(t *testing.T) {
	matrix := [][]float64{
		{0.3, 0.7},
		{0.6, 0.4},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Test normal sampling
	state, err := bias.sample(0, 0.5)
	require.NoError(t, err)
	require.True(t, state >= 0 && state < 2)

	// Boost state 1 and verify weighted sampling still works
	err = bias.boost(1, coverage.Delta{
		NewUnits:         5,
		NewLines:         3,
		CoverageIncrease: 0.1,
	})
	require.NoError(t, err)

	state, err = bias.sample(0, 0.5)
	require.NoError(t, err)
	require.True(t, state >= 0 && state < 2)
}

func TestCoverageBiasResetWeights(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Boost a state
	err = bias.boost(0, coverage.Delta{
		NewUnits:         5,
		NewLines:         2,
		CoverageIncrease: 0.1,
	})
	require.NoError(t, err)
	require.Greater(t, bias.weight(0), 1.0)

	// Reset weights
	bias.resetWeights()
	require.Equal(t, 1.0, bias.weight(0))
	require.Equal(t, 1.0, bias.weight(1))
}

func TestCoverageBiasBoostOutOfRange(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Try to boost invalid state indices
	err = bias.boost(-1, coverage.Delta{NewUnits: 5})
	require.Error(t, err)
	require.Contains(t, err.Error(), "out of range")

	err = bias.boost(999, coverage.Delta{NewUnits: 5})
	require.Error(t, err)
	require.Contains(t, err.Error(), "out of range")
}

func TestCoverageBiasBoostNoCoverage(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	initialWeight := bias.weight(0)

	// Boost with no coverage increase
	err = bias.boost(0, coverage.Delta{
		NewUnits:         0,
		NewLines:         0,
		CoverageIncrease: 0,
	})
	require.NoError(t, err)

	// Weight should not change
	require.Equal(t, initialWeight, bias.weight(0))
}

func TestCoverageBiasBoostNegative(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	initialWeight := bias.weight(0)

	// Boost with negative coverage (should be ignored)
	err = bias.boost(0, coverage.Delta{
		NewUnits:         -5,
		NewLines:         -10,
		CoverageIncrease: -0.05,
	})
	require.NoError(t, err)

	// Weight should not increase with negative boost
	require.Equal(t, initialWeight, bias.weight(0))
}

func TestCoverageBiasMinWeight(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Manually set weight below minimum
	bias.weights[0] = 0.05

	// Boost should bring it to at least minWeight
	err = bias.boost(0, coverage.Delta{
		NewUnits: 1,
	})
	require.NoError(t, err)

	require.GreaterOrEqual(t, bias.weight(0), bias.minWeight)
}

func TestCoverageBiasWeightOutOfRange(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Out of range states should return 1.0
	require.Equal(t, 1.0, bias.weight(-1))
	require.Equal(t, 1.0, bias.weight(999))
}

func TestCoverageBiasSampleErrors(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Invalid state
	_, err = bias.sample(-1, 0.5)
	require.Error(t, err)

	// Invalid probability
	_, err = bias.sample(0, -0.5)
	require.Error(t, err)

	_, err = bias.sample(0, 1.5)
	require.Error(t, err)
}

func TestCoverageBiasSampleFallback(t *testing.T) {
	// Test fallback to unweighted sampling when all weights are zero
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Set all weights to zero to trigger fallback
	bias.weights[0] = 0.0
	bias.weights[1] = 0.0

	// Should still work via fallback to unweighted sampling
	state, err := bias.sample(0, 0.5)
	require.NoError(t, err)
	require.True(t, state >= 0 && state < 2)
}

func TestCoverageBiasNilOperations(t *testing.T) {
	// Test nil bias behavior
	var bias *coverageBias

	// weight should return 1.0
	require.Equal(t, 1.0, bias.weight(0))

	// boost should return error
	err := bias.boost(0, coverage.Delta{NewUnits: 5})
	require.Error(t, err)

	// sample should return error
	_, err = bias.sample(0, 0.5)
	require.Error(t, err)

	// resetWeights should not panic
	require.NotPanics(t, func() {
		bias.resetWeights()
	})

	// stats should return zeros
	b, u, l := bias.stats()
	require.Equal(t, uint64(0), b)
	require.Equal(t, uint64(0), u)
	require.Equal(t, uint64(0), l)
}

func TestCoverageBiasStatsAccumulation(t *testing.T) {
	matrix := [][]float64{
		{0.5, 0.5},
		{0.5, 0.5},
	}
	labels := []string{"A", "B"}
	chain, err := markov.New(matrix, labels)
	require.NoError(t, err)

	bias, err := newCoverageBias(chain)
	require.NoError(t, err)

	// Multiple boosts should accumulate stats
	for i := 0; i < 5; i++ {
		err = bias.boost(0, coverage.Delta{
			NewUnits: 2,
			NewLines: 3,
		})
		require.NoError(t, err)
	}

	boosts, units, lines := bias.stats()
	require.Equal(t, uint64(5), boosts)
	require.Equal(t, uint64(10), units) // 2 * 5
	require.Equal(t, uint64(15), lines) // 3 * 5
}

func TestCoverageBiasEmptyChain(t *testing.T) {
	// Chain with no states
	matrix := [][]float64{}
	labels := []string{}
	chain, err := markov.New(matrix, labels)

	// If markov.New accepts empty chains, newCoverageBias should reject it
	if err == nil {
		_, err = newCoverageBias(chain)
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one state")
	}
	// If markov.New rejects empty chains, that's also fine
}
