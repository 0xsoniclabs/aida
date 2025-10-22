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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/stochastic/recorder/arguments"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleStats() *recorder.StatsJSON {
	return &recorder.StatsJSON{
		SnapshotECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		Balance: recorder.ScalarStatsJSON{
			Max:  10,
			ECDF: [][2]float64{{0.0, 0.0}, {0.5, 0.6}, {1.0, 1.0}},
		},
		Nonce: recorder.ScalarStatsJSON{
			Max:  5,
			ECDF: [][2]float64{{0.0, 0.0}, {0.4, 0.3}, {1.0, 1.0}},
		},
		CodeSize: recorder.ScalarStatsJSON{
			Max:  20,
			ECDF: [][2]float64{{0.0, 0.0}, {0.25, 0.5}, {1.0, 1.0}},
		},
		Contracts: arguments.ClassifierJSON{
			Counting: arguments.ArgStatsJSON{
				ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
			},
			Queuing: arguments.QueueStatsJSON{Distribution: []float64{0.5, 0.5}},
		},
		Keys: arguments.ClassifierJSON{
			Counting: arguments.ArgStatsJSON{
				ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
			},
			Queuing: arguments.QueueStatsJSON{Distribution: []float64{0.5, 0.5}},
		},
		Values: arguments.ClassifierJSON{
			Counting: arguments.ArgStatsJSON{
				ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
			},
			Queuing: arguments.QueueStatsJSON{Distribution: []float64{0.5, 0.5}},
		},
		StochasticMatrix: [][]float64{
			{0.2, 0.4, 0.4},
			{0.3, 0.4, 0.3},
			{0.1, 0.5, 0.4},
		},
		Operations: []string{"BT", "BB", "BS"},
	}
}

func colorStats() *recorder.StatsJSON {
	ops := []string{
		operations.OpMnemo(operations.BeginSyncPeriodID),
		operations.OpMnemo(operations.BeginBlockID),
		operations.OpMnemo(operations.BeginTransactionID),
		operations.OpMnemo(operations.EndTransactionID),
	}
	matrix := [][]float64{
		{0.0, 0.2, 0.8, 0.0},
		{0.0, 0.0, 1.0, 0.0},
		{0.25, 0.75, 0.0, 0.0},
		{1.0, 0.0, 0.0, 0.0},
	}
	dist := make([]float64, stochastic.QueueLen)
	if len(dist) > 0 {
		dist[0] = 1.0
	}
	cls := arguments.ClassifierJSON{
		Counting: arguments.ArgStatsJSON{
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		},
		Queuing: arguments.QueueStatsJSON{Distribution: dist},
	}
	return &recorder.StatsJSON{
		SnapshotECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		Balance: recorder.ScalarStatsJSON{
			Max:  1,
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		},
		Nonce: recorder.ScalarStatsJSON{
			Max:  1,
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		},
		CodeSize: recorder.ScalarStatsJSON{
			Max:  1,
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		},
		Contracts:        cls,
		Keys:             cls,
		Values:           cls,
		Operations:       ops,
		StochasticMatrix: matrix,
	}
}

func mustSetView(t *testing.T, stats *recorder.StatsJSON) {
	t.Helper()
	require.NoError(t, setViewState(stats))
}

func clearView(t *testing.T) {
	t.Helper()
	currentMu.Lock()
	currentState = nil
	currentMu.Unlock()
}
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

func TestVisualizer_newScalarChart(t *testing.T) {
	balance := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}
	nonce := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}
	code := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}

	chart := newScalarChart(balance, nonce, code)

	assert.NotNil(t, chart)
}

func TestVisualizer_renderCounting(t *testing.T) {
	stats := sampleStats()
	stats.Contracts.Counting.ECDF = [][2]float64{{0.0, 0.0}, {1.0, 0.5}}
	stats.Keys.Counting.ECDF = [][2]float64{{0.0, 0.0}, {1.0, 0.6}}
	stats.Values.Counting.ECDF = [][2]float64{{0.0, 0.0}, {1.0, 0.7}}
	mustSetView(t, stats)

	req, err := http.NewRequest("GET", "/counting-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderCounting)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestVisualizer_renderSnapshotStats(t *testing.T) {
	stats := sampleStats()
	stats.SnapshotECDF = [][2]float64{{0.0, 0.0}, {1.0, 0.5}}
	mustSetView(t, stats)

	req, err := http.NewRequest("GET", "/snapshot-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderSnapshotStats)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestVisualizer_renderScalarStats(t *testing.T) {
	stats := sampleStats()
	mustSetView(t, stats)

	req, err := http.NewRequest("GET", "/scalar-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderScalarStats)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Body.String())
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
	stats := sampleStats()
	stats.Contracts.Queuing.Distribution = []float64{0.1, 0.2}
	stats.Keys.Queuing.Distribution = []float64{0.3, 0.4}
	stats.Values.Queuing.Distribution = []float64{0.5, 0.6}
	mustSetView(t, stats)

	req, err := http.NewRequest("GET", "/queuing-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderQueuing)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Body.String())
}

func TestVisualizer_convertOperationData(t *testing.T) {
	testData := []opDatum{
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
	testData := []opDatum{
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
	mustSetView(t, sampleStats())

	req, err := http.NewRequest("GET", "/operation-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderOperationStats)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Body.String())
}

func TestVisualizer_renderTransactionalOperationStats(t *testing.T) {
	mustSetView(t, sampleStats())

	req, err := http.NewRequest("GET", "/tx-operation-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderTransactionalOperationStats)
	handler.ServeHTTP(rr, req)
	response := rr.Body.String()

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, response)
}

func TestVisualizer_renderSimplifiedMarkovChain(t *testing.T) {
	mustSetView(t, sampleStats())

	req, err := http.NewRequest("GET", "/simplified-markov-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderSimplifiedMarkovChain)
	handler.ServeHTTP(rr, req)
	response := rr.Body.String()

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotZero(t, len(response))
	assert.Contains(t, response, "StateDB Simplified Markov-Chain")
}

func TestVisualizer_renderMarkovChain(t *testing.T) {
	mustSetView(t, sampleStats())

	req, err := http.NewRequest("GET", "/markov-stats", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(renderMarkovChain)
	handler.ServeHTTP(rr, req)
	response := rr.Body.String()

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, response)
	assert.Contains(t, response, "StateDB Markov-Chain")
}

func TestVisualizer_renderMarkovVariantsColorCoverage(t *testing.T) {
	mustSetView(t, colorStats())

	req, err := http.NewRequest("GET", "/simplified-markov-stats", nil)
	require.NoError(t, err)
	simplified := httptest.NewRecorder()
	renderSimplifiedMarkovChain(simplified, req)
	assert.Equal(t, http.StatusOK, simplified.Code)

	markov := httptest.NewRecorder()
	renderMarkovChain(markov, req)
	assert.Equal(t, http.StatusOK, markov.Code)
}

func TestVisualizer_handlersWithoutState(t *testing.T) {
	handlers := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"renderCounting", renderCounting},
		{"renderQueuing", renderQueuing},
		{"renderSnapshotStats", renderSnapshotStats},
		{"renderScalarStats", renderScalarStats},
		{"renderOperationStats", renderOperationStats},
		{"renderTransactionalOperationStats", renderTransactionalOperationStats},
		{"renderSimplifiedMarkovChain", renderSimplifiedMarkovChain},
		{"renderMarkovChain", renderMarkovChain},
	}
	for _, tc := range handlers {
		t.Run(tc.name, func(t *testing.T) {
			clearView(t)
			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			tc.handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
		})
	}
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

	done := make(chan error, 1)
	go func() {
		done <- FireUpWeb(stateJSON, "0")
	}()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		// If no error after 1 seconds, pass the test
	}
}

func TestVisualizer_FireUpWebPanicsOnNilStats(t *testing.T) {
	err := FireUpWeb(nil, "0")
	assert.Error(t, err)
}
