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

package utildb

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestInfo_Info_FindBlockRangeInStateHash_Success(t *testing.T) {
	testDb := GenerateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	_, _, err := FindBlockRangeInStateHash(testDb, log)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	assert.Equal(t, "cannot get first state hash; not implemented", err.Error())
}

func TestInfo_FindBlockRangeInStateHash_FirstError(t *testing.T) {
	testDb := GenerateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	first, last, err := FindBlockRangeInStateHash(testDb, log)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), first)
	assert.Equal(t, uint64(0), last)
	assert.Contains(t, err.Error(), "cannot get first state hash")
}

func TestInfo_FindBlockRangeInBlockHash_Success(t *testing.T) {
	testDb := GenerateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	first, last, err := FindBlockRangeOfBlockHashes(testDb, log)
	assert.NoError(t, err)
	assert.Equal(t, uint64(21), first)
	assert.Equal(t, uint64(30), last)
}

func TestInfo_FindBlockRangeInBlockHash_FirstError(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashDbFirstError"
	testDb, err := db.NewDefaultSubstateDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}

	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	_, _, err = FindBlockRangeOfBlockHashes(testDb, log)
	assert.Error(t, err)
	assert.Equal(t, "cannot get first block hash; no block hash found", err.Error())
}

func TestInfo_GetStateHashCount_Success(t *testing.T) {
	testDb := GenerateTestAidaDb(t)
	cfg := &utils.Config{
		First: 11,
		Last:  20,
	}
	count, err := GetStateHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), count)
}

func TestInfo_GetStateHashCount_Empty(t *testing.T) {
	testDb := GenerateTestAidaDb(t)
	cfg := &utils.Config{
		First: 1, // Intentionally outside of state hash range
		Last:  1,
	}
	count, err := GetStateHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestInfo_GetBlockHashCount_Success(t *testing.T) {
	testDb := GenerateTestAidaDb(t)
	cfg := &utils.Config{
		First: 21,
		Last:  30,
	}
	count, err := GetBlockHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), count)
}

func TestInfo_GetBlockHashCount_Empty(t *testing.T) {
	testDb := GenerateTestAidaDb(t)
	cfg := &utils.Config{
		First: 1, // Intentionally outside of block hash range
		Last:  1,
	}
	count, err := GetBlockHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestInfo_GetStateHashCount_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	testDb := db.NewMockBaseDB(ctrl)

	errWant := errors.New("test error")
	testDb.EXPECT().Get(gomock.Any()).Return(nil, errWant).AnyTimes()

	cfg := &utils.Config{
		First: 100, // Intentionally outside of state hash range
		Last:  100,
	}
	_, err := GetStateHashCount(cfg, testDb)
	assert.Equal(t, errWant, err, "expected error to match")
}

func TestInfo_GetBlockHashCount_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	testDb := db.NewMockBaseDB(ctrl)

	errWant := errors.New("test error")
	testDb.EXPECT().Get(gomock.Any()).Return(nil, errWant).AnyTimes()

	cfg := &utils.Config{
		First: 100, // Intentionally outside of block hash range
		Last:  100,
	}
	_, err := GetBlockHashCount(cfg, testDb)
	assert.Equal(t, errWant, err, "expected error to match")
}
