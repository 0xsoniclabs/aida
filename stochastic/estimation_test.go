package stochastic

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
	"github.com/stretchr/testify/assert"
)

var mockEventRegistryJSON = &EventRegistryJSON{
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

func TestStochastic_NewEstimationModelJSON(t *testing.T) {
	model := NewEstimationModelJSON(mockEventRegistryJSON)
	assert.Equal(t, "simulation", model.FileId)
	assert.Equal(t, mockEventRegistryJSON.Operations, model.Operations)
	assert.Equal(t, mockEventRegistryJSON.StochasticMatrix, model.StochasticMatrix)
	assert.NotEmpty(t, model.Contracts)
	assert.NotEmpty(t, model.Keys)
	assert.NotEmpty(t, model.Values)
	assert.Equal(t, 24.999999991320017, model.SnapshotLambda)
}

func TestStochastic_ReadSimulation(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := tempDir + "/simulation.json"
	t.Run("success", func(t *testing.T) {
		model := NewEstimationModelJSON(mockEventRegistryJSON)
		marshal, err := json.Marshal(model)
		if err != nil {
			t.Fatalf("Failed to marshal mock events: %v", err)
		}
		err = os.WriteFile(tempFile, marshal, 0644)
		if err != nil {
			t.Fatalf("Failed to write mock events to file: %v", err)
		}
		simulation, err := ReadSimulation(tempFile)
		assert.Equal(t, &model, simulation)
		assert.NoError(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := ReadSimulation(tempDir + "/nonexistent.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidFile := tempDir + "/invalid.json"
		err := os.WriteFile(invalidFile, []byte("invalid json"), 0644)
		assert.NoError(t, err)
		_, err = ReadSimulation(invalidFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})
}

func TestEstimationModelJSON_WriteJSON(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := tempDir + "/simulation.json"
	model := NewEstimationModelJSON(mockEventRegistryJSON)
	err := model.WriteJSON(tempFile)
	assert.NoError(t, err)
	_, err = os.Stat(tempFile)
	assert.NoError(t, err)
}

func TestStochastic_NewEstimationStats(t *testing.T) {
	stats := NewEstimationStats(&mockEventRegistryJSON.Contracts)
	assert.Equal(t, int64(0), stats.NumKeys)
	assert.Equal(t, 24.999999991320017, stats.Lambda)
	assert.Equal(t, []float64{}, stats.QueueDistribution)
}
