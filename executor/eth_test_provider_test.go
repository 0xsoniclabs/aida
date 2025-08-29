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

package executor

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_ethTestProvider_Run(t *testing.T) {
	pathFile := createTestDataFile(t)

	cfg := &utils.Config{
		ArgPath: pathFile,
		Fork:    "all",
	}

	provider := NewEthStateTestProvider(cfg)

	ctrl := gomock.NewController(t)

	var consumer = NewMockTxConsumer(ctrl)

	gomock.InOrder(
		consumer.EXPECT().Consume(2, 0, gomock.Any()),
		consumer.EXPECT().Consume(3, 1, gomock.Any()),
		consumer.EXPECT().Consume(4, 2, gomock.Any()),
		consumer.EXPECT().Consume(5, 3, gomock.Any()),
	)

	err := provider.Run(0, 0, toSubstateConsumer(consumer))
	if err != nil {
		t.Errorf("Run() error = %v, wantErr %v", err, nil)
	}
}

func createTestDataFile(t *testing.T) string {
	path := t.TempDir()
	pathFile := path + "/test.json"
	stData := ethtest.CreateTestStJson(t)

	jsonData, err := json.Marshal(stData)
	if err != nil {
		t.Errorf("Marshal() error = %v, wantErr %v", err, nil)
	}

	jsonStr := "{ \"test\" : " + string(jsonData) + "}"

	jsonData = []byte(jsonStr)
	// Initialize pathFile
	err = os.WriteFile(pathFile, jsonData, 0644)
	if err != nil {
		t.Errorf("WriteFile() error = %v, wantErr %v", err, nil)
	}
	return pathFile
}

func TestExecutor_NewEthStateTestProvider(t *testing.T) {
	cfg := &utils.Config{ArgPath: "somepath"}
	provider := NewEthStateTestProvider(cfg)
	require.NotNil(t, provider)

	_, ok := provider.(ethTestProvider)
	assert.True(t, ok, "NewEthStateTestProvider should return an ethTestProvider")

}

func TestEthTestProvider_Close(t *testing.T) {
	provider := NewEthStateTestProvider(&utils.Config{})

	assert.NotPanics(t, func() { provider.Close() })
}

func TestEthTestProvider_Run_NewSplitterFails_NonExistentFile(t *testing.T) {
	cfg := &utils.Config{
		ArgPath: filepath.Join(t.TempDir(), "non_existent_test_file.json"),
		Fork:    "all",
	}
	provider := NewEthStateTestProvider(cfg)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConsumer := NewMockTxConsumer(ctrl)

	err := provider.Run(0, 0, toSubstateConsumer(mockConsumer))
	require.Error(t, err)

	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestEthTestProvider_Run_ConsumerError(t *testing.T) {
	pathFile := createTestDataFile(t)
	cfg := &utils.Config{
		ArgPath: pathFile,
		Fork:    "all",
	}
	provider := NewEthStateTestProvider(cfg)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConsumer := NewMockTxConsumer(ctrl)

	expectedErr := errors.New("consumer failed")

	mockConsumer.EXPECT().Consume(2, 0, gomock.Any()).Return(expectedErr)

	runErr := provider.Run(0, 0, toSubstateConsumer(mockConsumer))
	require.Error(t, runErr)
	assert.True(t, errors.Is(runErr, expectedErr))
	assert.Contains(t, runErr.Error(), "transaction failed")
}

func TestEthTestProvider_Run_NoTestsFromSplitter(t *testing.T) {
	emptyTestFilePath := filepath.Join(t.TempDir(), "empty_tests.json")

	emptyTestData := `{"test": {}}`
	err := os.WriteFile(emptyTestFilePath, []byte(emptyTestData), 0644)
	require.NoError(t, err)

	cfg := &utils.Config{
		ArgPath: emptyTestFilePath,
		Fork:    "all",
	}
	provider := NewEthStateTestProvider(cfg)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConsumer := NewMockTxConsumer(ctrl)

	runErr := provider.Run(0, 0, toSubstateConsumer(mockConsumer))
	assert.NoError(t, runErr)
}
