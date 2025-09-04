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
		Contracts: statistics.AccessJSON{
			Counting: statistics.CountingJSON{
				ECdf: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
			},
		},
		Keys: statistics.AccessJSON{
			Counting: statistics.CountingJSON{
				ECdf: [][2]float64{{0.5, 0.6}, {0.7, 0.8}},
			},
		},
		Values: statistics.AccessJSON{
			Counting: statistics.CountingJSON{
				ECdf: [][2]float64{{0.9, 1.0}, {1.1, 1.2}},
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
	// SyncPeriod, Block and Tx operations are excluded
	assert.Equal(t, stochastic.NumOps-6, len(e.TxOperation))
	assert.Equal(t, stochastic.NumOps, len(e.SimplifiedMatrix))
}

func TestAccessData_PopulateAccess(t *testing.T) {
	d := &statistics.AccessJSON{
		Counting: statistics.CountingJSON{
			ECdf: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
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
