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

package visualizer

import (
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetViewStateRejectsNil(t *testing.T) {
	err := setViewState(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stats are nil")
}

func TestSetViewStatePropagatesBuildError(t *testing.T) {
	stats := &recorder.StatsJSON{
		Operations: []string{"AA", "AA"},
		StochasticMatrix: [][]float64{
			{1.0, 0.0},
			{0.0, 1.0},
		},
	}
	err := setViewState(stats)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create markov chain")
}

func TestBuildViewStateInvalidMatrix(t *testing.T) {
	stats := &recorder.StatsJSON{
		Operations:       []string{"AA", "BB"},
		StochasticMatrix: [][]float64{{1.0}},
	}
	_, err := buildViewState(stats)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create markov chain")
}

func TestComputeTxOperationDecodeError(t *testing.T) {
	stats := &recorder.StatsJSON{
		Operations:       []string{"??"},
		StochasticMatrix: [][]float64{{1.0}},
	}
	_, _, _, err := computeTxOperation(stats, []float64{1.0})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode opcode")
}

func TestComputeSimplifiedMatrixDecodeError(t *testing.T) {
	stats := &recorder.StatsJSON{
		Operations: []string{"??"},
		StochasticMatrix: [][]float64{
			{1.0},
		},
	}
	_, err := computeSimplifiedMatrix(stats)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode opcode")
}

func TestBuildViewStatePropagatesTxErrors(t *testing.T) {
	stats := &recorder.StatsJSON{
		Operations:       []string{"??"},
		StochasticMatrix: [][]float64{{1.0}},
	}
	_, err := buildViewState(stats)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode opcode")
}

func TestCurrentViewWithoutState(t *testing.T) {
	clearView(t)
	_, err := currentView()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "statistics not initialised")
}
