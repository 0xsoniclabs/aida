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
