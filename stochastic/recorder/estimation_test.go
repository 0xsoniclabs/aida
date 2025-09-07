package recorder

import (
    "encoding/json"
    "math"
    "os"
    "os/exec"
    "runtime"
    "testing"

    "github.com/0xsoniclabs/aida/stochastic/statistics/classifier"
    "github.com/stretchr/testify/assert"
)

var mockEventRegistryJSON = &EventRegistryJSON{
	SnapshotEcdf: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
	Contracts: classifier.ArgClassifierJSON{
		Counting: classifier.ArgStatsJSON{
			ECDF: [][2]float64{{0.1, 0.2}, {0.3, 0.4}},
		},
	},
	Keys: classifier.ArgClassifierJSON{
		Counting: classifier.ArgStatsJSON{
			ECDF: [][2]float64{{0.5, 0.6}, {0.7, 0.8}},
		},
	},
	Values: classifier.ArgClassifierJSON{
		Counting: classifier.ArgStatsJSON{
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

    t.Run("not a simulation file id", func(t *testing.T) {
        input := &EstimationModelJSON{FileId: "not-simulation"}
        marshal, err := json.Marshal(input)
        if err != nil {
            t.Fatalf("Failed to marshal input: %v", err)
        }
        path := tempDir + "/wrong-id.json"
        err = os.WriteFile(path, marshal, 0644)
        if err != nil {
            t.Fatalf("Failed to write mock file: %v", err)
        }
        simulation, err := ReadSimulation(path)
        assert.Error(t, err)
        assert.Nil(t, simulation)
    })

    t.Run("read error on directory", func(t *testing.T) {
        dirPath := tempDir + "/dir"
        err := os.Mkdir(dirPath, 0o755)
        assert.NoError(t, err)
        simulation, err := ReadSimulation(dirPath)
        assert.Error(t, err)
        assert.Nil(t, simulation)
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
    err = model.WriteJSON(tempDir)
    assert.Error(t, err)
}

func TestEstimationModelJSON_WriteJSON_MarshalError(t *testing.T) {
    tempDir := t.TempDir()
    badModel := &EstimationModelJSON{
        FileId:           "simulation",
        SnapshotLambda:   math.NaN(),
        Operations:       []string{"BB"},
        StochasticMatrix: [][]float64{{1}},
        Contracts:        EstimationStatsJSON{},
        Keys:             EstimationStatsJSON{},
        Values:           EstimationStatsJSON{},
    }
    err := badModel.WriteJSON(tempDir + "/bad.json")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unsupported value: NaN")
}

func TestFatal_NewEstimationModelJSON(t *testing.T) {
    if os.Getenv("WANT_FATAL_NEW_ESTIMATION_MODEL") == "1" {
        bad := &EventRegistryJSON{SnapshotEcdf: [][2]float64{{0, math.NaN()}, {1, 1}}}
        _ = NewEstimationModelJSON(bad)
        return
    }
    cmd := exec.Command(os.Args[0], "-test.run=TestFatal_NewEstimationModelJSON")
    cmd.Env = append(os.Environ(), "WANT_FATAL_NEW_ESTIMATION_MODEL=1")
    err := cmd.Run()
    if err == nil {
        t.Fatalf("expected process to exit due to log.Fatalf")
    }
}

func TestFatal_NewEstimationStats(t *testing.T) {
    if os.Getenv("WANT_FATAL_NEW_ESTIMATION_STATS") == "1" {
        bad := &classifier.ArgClassifierJSON{Counting: classifier.ArgStatsJSON{ECDF: [][2]float64{{0, math.NaN()}, {1, 1}}}}
        _ = NewEstimationStats(bad)
        return
    }
    cmd := exec.Command(os.Args[0], "-test.run=TestFatal_NewEstimationStats")
    cmd.Env = append(os.Environ(), "WANT_FATAL_NEW_ESTIMATION_STATS=1")
    err := cmd.Run()
    if err == nil {
        t.Fatalf("expected process to exit due to log.Fatalf")
    }
}

func TestEstimationModelJSON_WriteJSON_WriteError(t *testing.T) {
    if runtime.GOOS != "linux" {
        t.Skip("/dev/full is Linux-specific")
    }
    model := NewEstimationModelJSON(mockEventRegistryJSON)
    err := model.WriteJSON("/dev/full")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to convert JSON file")
}

func TestStochastic_NewEstimationStats(t *testing.T) {
	stats := NewEstimationStats(&mockEventRegistryJSON.Contracts)
	assert.Equal(t, int64(0), stats.NumKeys)
	assert.Equal(t, 24.999999991320017, stats.Lambda)
	assert.Equal(t, []float64{}, stats.QueueDistribution)
}
