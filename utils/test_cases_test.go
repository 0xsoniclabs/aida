package utils

import (
	"github.com/0xsoniclabs/aida/config/chainid"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCases_GetStateDbTestCases(t *testing.T) {
	testCases := GetStateDbTestCases()
	assert.Equal(t, 5, len(testCases))
}

func TestCases_MakeRandomByteSlice(t *testing.T) {
	bufferLength := 32
	buffer := MakeRandomByteSlice(t, bufferLength)
	assert.Equal(t, bufferLength, len(buffer))
}

func TestCases_GetRandom(t *testing.T) {
	rangeLower := 1
	rangeUpper := 10
	randomValue := GetRandom(rangeLower, rangeUpper)
	assert.GreaterOrEqual(t, randomValue, rangeLower)
	assert.LessOrEqual(t, randomValue, rangeUpper)
}

func TestCases_MakeAccountStorage(t *testing.T) {
	storage := MakeAccountStorage(t)
	assert.Equal(t, 10, len(storage))
	for _, value := range storage {
		assert.NotEmpty(t, value)
	}
}

func TestCases_MakeTestConfig(t *testing.T) {
	testCases := GetStateDbTestCases()
	cfg := MakeTestConfig(testCases[0])
	assert.NotNil(t, cfg)
	assert.Equal(t, chainid.MainnetChainID, cfg.ChainID)
	assert.Equal(t, "", cfg.VmImpl)
	assert.Equal(t, "geth", cfg.DbImpl)
	assert.Equal(t, "", cfg.ArchiveVariant)
	assert.Equal(t, true, cfg.ArchiveMode)
}

func TestCases_MakeWorldState(t *testing.T) {
	ws, addr := MakeWorldState(t)
	assert.NotNil(t, ws)
	assert.Equal(t, 100, len(ws))
	assert.Equal(t, 100, len(addr))
}
