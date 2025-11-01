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
	"fmt"
	"math"

	"github.com/0xsoniclabs/aida/stochastic/coverage"
	"github.com/0xsoniclabs/aida/stochastic/statistics/markov"
)

type coverageBias struct {
	chain          *markov.Chain
	weights        []float64
	maxWeight      float64
	minWeight      float64
	baseBoost      float64
	unitFactor     float64
	lineFactor     float64
	coverageFactor float64
	totalBoosts    uint64
	totalNewUnits  uint64
	totalNewLines  uint64
}

// newCoverageBias creates a new coverage-guided bias for Markov chain sampling.
// All states start with equal weight of 1.0. Weights are boosted when transitions
// discover new code coverage.
func newCoverageBias(chain *markov.Chain) (*coverageBias, error) {
	if chain == nil {
		return nil, fmt.Errorf("newCoverageBias: chain cannot be nil")
	}

	numStates := chain.NumStates()
	if numStates <= 0 {
		return nil, fmt.Errorf("newCoverageBias: chain must have at least one state, got %d", numStates)
	}

	weights := make([]float64, numStates)
	for i := range weights {
		weights[i] = 1.0
	}

	return &coverageBias{
		chain:          chain,
		weights:        weights,
		maxWeight:      10.0,
		minWeight:      0.1,
		baseBoost:      0.05,
		unitFactor:     0.05,
		lineFactor:     0.01,
		coverageFactor: 1.0,
	}, nil
}

// sample performs weighted sampling of the next state in the Markov chain.
// Returns an error if the state is invalid or all weighted transitions are zero.
func (cb *coverageBias) sample(state int, u float64) (int, error) {
	if cb == nil || cb.chain == nil {
		return 0, fmt.Errorf("sample: coverageBias not properly initialized")
	}

	nextState, err := cb.chain.WeightedSample(state, u, cb.weights)
	if err != nil {
		// If weighted sampling fails due to zero weights, fall back to unweighted sampling
		if nextState, fallbackErr := cb.chain.Sample(state, u); fallbackErr == nil {
			return nextState, nil
		}
		return 0, fmt.Errorf("sample: %w", err)
	}

	return nextState, nil
}

// boost increases the weight of a state based on coverage improvements.
// States that discover new coverage get higher weights for future sampling.
func (cb *coverageBias) boost(state int, delta coverage.Delta) error {
	if cb == nil {
		return fmt.Errorf("boost: coverageBias not properly initialized")
	}

	if state < 0 || state >= len(cb.weights) {
		return fmt.Errorf("boost: state %d out of range [0, %d)", state, len(cb.weights))
	}

	// No coverage improvement means no boost
	if delta.NewUnits == 0 && delta.NewLines == 0 {
		return nil
	}

	// Calculate boost based on coverage metrics
	boost := cb.baseBoost +
		float64(delta.NewUnits)*cb.unitFactor +
		float64(delta.NewLines)*cb.lineFactor +
		delta.CoverageIncrease*cb.coverageFactor

	if boost <= 0 {
		return nil
	}

	// Update weight with min/max bounds
	newWeight := cb.weights[state] + boost
	cb.weights[state] = math.Max(cb.minWeight, math.Min(newWeight, cb.maxWeight))

	// Track statistics
	cb.totalBoosts++
	cb.totalNewUnits += uint64(delta.NewUnits)
	cb.totalNewLines += uint64(delta.NewLines)

	return nil
}

// weight returns the current weight for a state.
// Returns 1.0 (neutral weight) if the state is out of range.
func (cb *coverageBias) weight(state int) float64 {
	if cb == nil || state < 0 || state >= len(cb.weights) {
		return 1.0
	}
	return cb.weights[state]
}

// resetWeights resets all state weights to 1.0.
func (cb *coverageBias) resetWeights() {
	if cb == nil {
		return
	}
	for i := range cb.weights {
		cb.weights[i] = 1.0
	}
}

// stats returns statistics about the coverage-guided exploration.
func (cb *coverageBias) stats() (totalBoosts, totalNewUnits, totalNewLines uint64) {
	if cb == nil {
		return 0, 0, 0
	}
	return cb.totalBoosts, cb.totalNewUnits, cb.totalNewLines
}
