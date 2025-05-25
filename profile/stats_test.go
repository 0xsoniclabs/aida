package profile

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStats_NewStats(t *testing.T) {
	tempDir := os.TempDir()
	ps := NewStats(tempDir + "/test_print.csv")
	assert.NotNil(t, ps)
	assert.Equal(t, tempDir+"/test_print.csv", ps.csv)
	assert.True(t, ps.writeToFile)
	assert.False(t, ps.hasHeader)
	assert.Equal(t, 1, len(ps.opOrder))
	assert.Equal(t, byte(0), ps.opOrder[0])
}
func TestStats_Profile(t *testing.T) {
	tempDir := os.TempDir()
	ps := NewStats(tempDir + "/test_print.csv")

	// Profile a new operation
	ps.Profile(1, 100*time.Millisecond)
	assert.Equal(t, uint64(1), ps.opFrequency[1])
	assert.Equal(t, 100*time.Millisecond, ps.opDuration[1])
	assert.Equal(t, 100*time.Millisecond, ps.opMinDuration[1])
	assert.Equal(t, 100*time.Millisecond, ps.opMaxDuration[1])
	assert.Equal(t, 0.0, ps.opVariance[1])

	// Profile the same operation again
	ps.Profile(1, 200*time.Millisecond)
	assert.Equal(t, uint64(2), ps.opFrequency[1])
	assert.Equal(t, 300*time.Millisecond, ps.opDuration[1])
	assert.Equal(t, 100*time.Millisecond, ps.opMinDuration[1])
	assert.Equal(t, 200*time.Millisecond, ps.opMaxDuration[1])
	// Variance calculation: for two samples 100ms and 200ms, variance = ((100-150)^2 + (200-150)^2)/2 = 2500
	expectedVariance := 1.25e+15
	assert.Equal(t, expectedVariance, ps.opVariance[1])

	// Profile a different operation
	ps.Profile(2, 150*time.Millisecond)
	assert.Equal(t, uint64(1), ps.opFrequency[2])
	assert.Equal(t, 150*time.Millisecond, ps.opDuration[2])
	assert.Equal(t, 150*time.Millisecond, ps.opMinDuration[2])
	assert.Equal(t, 150*time.Millisecond, ps.opMaxDuration[2])
	assert.Equal(t, 0.0, ps.opVariance[2])
}

func TestStats_FillLabels(t *testing.T) {
	tempDir := os.TempDir()
	ps := NewStats(tempDir + "/test_print.csv")
	labels := map[byte]string{
		2: "Operation 2",
		1: "Operation 1",
		3: "Operation 3",
	}
	ps.FillLabels(labels)

	assert.Equal(t, "Operation 1", ps.opLabel[1])
	assert.Equal(t, "Operation 2", ps.opLabel[2])
	assert.Equal(t, "Operation 3", ps.opLabel[3])
	assert.Equal(t, []byte{1, 2, 3}, ps.opOrder)
}

func TestStats_PrintProfiling(t *testing.T) {
	ps := NewStats("")

	// Prepare stats
	ps.Profile(1, 100*time.Millisecond)
	ps.Profile(1, 200*time.Millisecond)
	ps.Profile(2, 150*time.Millisecond)
	ps.FillLabels(map[byte]string{
		1: "Op1",
		2: "Op2",
	})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := ps.PrintProfiling(10, 20)
	assert.NoError(t, err)

	// Restore stdout and read output
	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout
	var buf [1024]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	assert.Contains(t, output, "Op1")
	assert.Contains(t, output, "Op2")
	assert.Contains(t, output, "Total StateDB net execution time")

}

func TestStats_writeCsv(t *testing.T) {
	tempDir := os.TempDir()
	ps := NewStats(tempDir + "/test_print.csv")
	var sb strings.Builder
	sb.WriteString("Hello, ")
	sb.WriteString("world!")
	err := ps.writeCsv(sb)
	assert.NoError(t, err)

	// Check if file exists and has content
	fileInfo, err := os.Stat(tempDir + "/test_print.csv")
	assert.NoError(t, err)
	assert.False(t, fileInfo.IsDir())
	assert.Greater(t, fileInfo.Size(), int64(0))
}

func TestStats_GetOpFrequency(t *testing.T) {
	ps := Stats{
		opOrder: []byte{1, 2, 3},
	}
	ops := ps.GetOpOrder()
	assert.Equal(t, []byte{1, 2, 3}, ops)
}

func TestStats_GetStatByOpId(t *testing.T) {
	tempDir := os.TempDir()
	ps := NewStats(tempDir + "/test_print.csv")
	stat := ps.GetStatByOpId(1)
	assert.NotNil(t, stat)
	assert.Equal(t, uint64(0), stat.Frequency)
	assert.Equal(t, time.Duration(0), stat.Duration)
	assert.Equal(t, time.Duration(0), stat.MinDuration)
	assert.Equal(t, time.Duration(0), stat.MaxDuration)
	assert.Equal(t, 0.0, stat.Variance)
	assert.Equal(t, "", stat.Label)
}

func TestStats_GetTotalOpFreq(t *testing.T) {
	ps := Stats{
		opFrequency: map[byte]uint64{
			1: 10,
			2: 20,
			3: 30,
		},
	}
	total := ps.GetTotalOpFreq()
	assert.Equal(t, int(60), total)
}
