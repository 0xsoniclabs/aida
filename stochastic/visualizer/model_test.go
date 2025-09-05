package visualizer

import (
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/statistics"
	"github.com/stretchr/testify/assert"
)

func TestVisualizer_GetEventsData(t *testing.T) {
	out := GetEventsData()
	assert.NotNil(t, out)
}

func TestEventData_PopulateEventData(t *testing.T) {
	d := &stochastic.EventRegistryJSON{
		SnapshotEcdf: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
		Contracts: statistics.ArgClassifierJSON{
			Counting: statistics.ArgStatsJSON{
				ECDF: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
			},
		},
		Keys: statistics.ArgClassifierJSON{
			Counting: statistics.ArgStatsJSON{
				ECDF: [][2]float64{{0.5, 0.6}, {0.7, 0.8}},
			},
		},
		Values: statistics.ArgClassifierJSON{
			Counting: statistics.ArgStatsJSON{
				ECDF: [][2]float64{{0.9, 1.0}, {1.1, 1.2}},
			},
		},
		StochasticMatrix: [][]float64{
			{1 / 3.0, 1 / 3.0, 1 / 3.0},
			{1 / 3.0, 1 / 3.0, 1 / 3.0},
			{1 / 3.0, 1 / 3.0, 1 / 3.0},
		},
		Operations: []string{
			"BT",
			"BB",
			"BS",
		},
	}
	e := &EventData{}
	e.PopulateEventData(d)
	expectedStationary := []OpData{
		{label: "BT", value: 0.3333333333333333},
		{label: "BB", value: 0.3333333333333333},
		{label: "BS", value: 0.3333333333333333},
	}
	assert.Equal(t, expectedStationary, e.Stationary)
	assert.Equal(t, float64(1), e.TxPerBlock)
	assert.Equal(t, float64(1), e.BlocksPerSyncPeriod)
	assert.Equal(t, d.Operations, e.OperationLabel)
	assert.Equal(t, d.StochasticMatrix, e.StochasticMatrix)
	assert.Equal(t, 24, len(e.TxOperation))
	assert.Equal(t, 30, len(e.SimplifiedMatrix))
}

func TestAccessData_PopulateAccess(t *testing.T) {
	d := &statistics.ArgClassifierJSON{
		Counting: statistics.ArgStatsJSON{
			ECDF: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
		},
	}
	a := &AccessData{}
	a.PopulateAccess(d)
	assert.Equal(t, 24.999999991320017, a.Lambda)
	assert.Equal(t, 101, len(a.Cdf))
	assert.Equal(t, [][2]float64{{0.1, 0.2}, {0.3, 0.4}}, a.ECdf)
	assert.Equal(t, []float64{}, a.QPdf)
}

func TestSnapshotData_PopulateSnapshot(t *testing.T) {
	d := &stochastic.EventRegistryJSON{
		SnapshotEcdf: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
	}
	s := &SnapshotData{}
	s.PopulateSnapshotStats(d)
	assert.Equal(t, [][2]float64{{0.1, 0.2}, {0.3, 0.4}}, s.ECdf)
	assert.Equal(t, 24.999999991320017, s.Lambda)
	assert.Equal(t, 101, len(s.Cdf))
}
