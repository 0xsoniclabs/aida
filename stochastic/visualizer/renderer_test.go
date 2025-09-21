// Copyright 2025 Fantom Foundation
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/stochastic/recorder/arguments"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/stretchr/testify/assert"
)

func TestVisualizer_renderMain(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderMain)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, MainHtml, rr.Body.String())
}

func TestVisualizer_convertCountingData(t *testing.T) {
	testData := [][2]float64{{1.0, 2.0}, {3.0, 4.0}, {5.0, 6.0}}

	result := convertCountingData(testData)

	assert.Len(t, result, 3)
	assert.Equal(t, opts.LineData{Value: [2]float64{1.0, 2.0}}, result[0])
	assert.Equal(t, opts.LineData{Value: [2]float64{3.0, 4.0}}, result[1])
	assert.Equal(t, opts.LineData{Value: [2]float64{5.0, 6.0}}, result[2])
}

func TestVisualizer_newCountingChart(t *testing.T) {
	title := "Test Title"
	contracts := [][2]float64{{1.0, 0.5}, {2.0, 0.8}}
	values := [][2]float64{{1.0, 0.5}, {2.0, 0.8}}
	keys := [][2]float64{{1.0, 0.5}, {2.0, 0.8}}

	chart := newCountingChart(title, contracts, values, keys)

	assert.NotNil(t, chart)
}

func TestVisualizer_renderCounting(t *testing.T) {
	data := GetData()
	data.Contracts.A_CDF = [][2]float64{{1.0, 0.5}}
	data.Keys.A_CDF = [][2]float64{{1.0, 0.6}}
	data.Values.A_CDF = [][2]float64{{1.0, 0.7}}

	req, err := http.NewRequest("GET", "/counting-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderCounting)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestVisualizer_renderSnapshotStats(t *testing.T) {
	data := GetData()
	data.Snapshot.ECdf = [][2]float64{{1.0, 0.5}}

	req, err := http.NewRequest("GET", "/snapshot-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderSnapshotStats)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestVisualizer_convertQueuingData(t *testing.T) {
	testData := []float64{0.1, 0.2, 0.3, 0.4}

	result := convertQueuingData(testData)

	assert.Len(t, result, 4)
	assert.Equal(t, opts.ScatterData{Value: [2]float64{0.0, 0.1}, SymbolSize: 5}, result[0])
	assert.Equal(t, opts.ScatterData{Value: [2]float64{1.0, 0.2}, SymbolSize: 5}, result[1])
	assert.Equal(t, opts.ScatterData{Value: [2]float64{2.0, 0.3}, SymbolSize: 5}, result[2])
	assert.Equal(t, opts.ScatterData{Value: [2]float64{3.0, 0.4}, SymbolSize: 5}, result[3])
}

func TestVisualizer_renderQueuing(t *testing.T) {
	e := &StateData{}
	e.Contracts.Q_PMF = []float64{0.1, 0.2}
	e.Keys.Q_PMF = []float64{0.3, 0.4}
	e.Values.Q_PMF = []float64{0.5, 0.6}

	req, err := http.NewRequest("GET", "/queuing-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderQueuing)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Len(t, rr.Body.String(), 1877)
}

func TestVisualizer_convertOperationData(t *testing.T) {
	testData := []OpData{
		{label: "op1", value: 0.1},
		{label: "op2", value: 0.2},
		{label: "op3", value: 0.3},
	}

	result := convertOperationData(testData)

	assert.Len(t, result, 3)
	assert.Equal(t, opts.BarData{Value: 0.1}, result[0])
	assert.Equal(t, opts.BarData{Value: 0.2}, result[1])
	assert.Equal(t, opts.BarData{Value: 0.3}, result[2])
}

func TestVisualizer_convertOperationLabel(t *testing.T) {
	testData := []OpData{
		{label: "operation1", value: 0.1},
		{label: "operation2", value: 0.2},
		{label: "operation3", value: 0.3},
	}

	result := convertOperationLabel(testData)

	assert.Len(t, result, 3)
	assert.Equal(t, "operation1", result[0])
	assert.Equal(t, "operation2", result[1])
	assert.Equal(t, "operation3", result[2])
}

func TestVisualizer_renderOperationStats(t *testing.T) {
	e := &StateData{}
	e.Stationary = []OpData{
		{label: "op1", value: 0.3},
		{label: "op2", value: 0.7},
	}

	req, err := http.NewRequest("GET", "/operation-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderOperationStats)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Len(t, rr.Body.String(), 1441)
}

func TestVisualizer_renderTransactionalOperationStats(t *testing.T) {
	e := &StateData{}
	e.TxOperation = []OpData{
		{label: "tx_op1", value: 1.5},
		{label: "tx_op2", value: 2.5},
	}
	e.TxPerBlock = 100.5
	e.BlocksPerSyncPeriod = 50.3

	req, err := http.NewRequest("GET", "/tx-operation-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderTransactionalOperationStats)
	handler.ServeHTTP(rr, req)
	response := rr.Body.String()

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Len(t, response, 1444)
}

func TestVisualizer_renderSimplifiedMarkovChain(t *testing.T) {
	e := &StateData{}
	// Initialize a simple matrix with some non-zero values
	for i := 0; i < operations.NumOps; i++ {
		for j := 0; j < operations.NumOps; j++ {
			if i == j {
				e.SimplifiedMatrix[i][j] = 0.5
			} else if j == (i+1)%operations.NumOps {
				e.SimplifiedMatrix[i][j] = 0.3
			}
		}
	}

	req, err := http.NewRequest("GET", "/simplified-markov-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderSimplifiedMarkovChain)
	handler.ServeHTTP(rr, req)
	response := rr.Body.String()

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Len(t, response, 2507)
	assert.Contains(t, response, "StateDB Simplified Markov-Chain")
}

func TestVisualizer_renderMarkovChain(t *testing.T) {
	e := &StateData{}
	e.OperationLabel = []string{"op1", "op2", "op3"}
	e.StochasticMatrix = [][]float64{
		{0.5, 0.3, 0.2},
		{0.4, 0.4, 0.2},
		{0.3, 0.3, 0.4},
	}

	req, err := http.NewRequest("GET", "/markov-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderMarkovChain)
	handler.ServeHTTP(rr, req)
	response := rr.Body.String()

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Len(t, response, 637)
	assert.Contains(t, response, "StateDB Markov-Chain")
}

func TestVisualizer_FireUpWeb(t *testing.T) {
	stateJSON := &recorder.StatsJSON{
		SnapshotECDF: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
		Contracts: arguments.ClassifierJSON{
			Counting: arguments.ArgStatsJSON{
				ECDF: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
			},
		},
		Keys: arguments.ClassifierJSON{
			Counting: arguments.ArgStatsJSON{
				ECDF: [][2]float64{{0.5, 0.6}, {0.7, 0.8}},
			},
		},
		Values: arguments.ClassifierJSON{
			Counting: arguments.ArgStatsJSON{
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

	assert.NotPanics(t, func() {
		go func() {
			FireUpWeb(stateJSON, "0")
		}()
	})
}
