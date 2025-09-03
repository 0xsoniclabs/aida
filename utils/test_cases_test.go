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

package utils

import (
	"testing"

	"github.com/0xsoniclabs/aida/config"
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
	assert.Equal(t, config.MainnetChainID, cfg.ChainID)
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
