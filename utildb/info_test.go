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

func TestFindBlockRangeInStateHash_Success(t *testing.T) {
	testDb := generateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	_, _, err := FindBlockRangeInStateHash(testDb, log)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	assert.Equal(t, "cannot get first state hash; not implemented", err.Error())
}

func TestFindBlockRangeInStateHash_FirstError(t *testing.T) {
	testDb := generateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	first, last, err := FindBlockRangeInStateHash(testDb, log)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), first)
	assert.Equal(t, uint64(0), last)
	assert.Contains(t, err.Error(), "cannot get first state hash")
}

func TestFindBlockRangeInBlockHash_Success(t *testing.T) {
	testDb := generateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	first, last, err := FindBlockRangeOfBlockHashes(testDb, log)
	assert.NoError(t, err)
	assert.Equal(t, uint64(21), first)
	assert.Equal(t, uint64(30), last)
}

func TestFindBlockRangeInBlockHash_FirstError(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashDbFirstError"
	testDb, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}

	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	_, _, err = FindBlockRangeOfBlockHashes(testDb, log)
	assert.Error(t, err)
	assert.Equal(t, "cannot get first block hash; no block hash found", err.Error())
}

func TestGetStateHashCount_Success(t *testing.T) {
	testDb := generateTestAidaDb(t)
	cfg := &utils.Config{
		First: 11,
		Last:  20,
	}
	count, err := GetStateHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), count)
}

func TestGetStateHashCount_Empty(t *testing.T) {
	testDb := generateTestAidaDb(t)
	cfg := &utils.Config{
		First: 1, // Intentionally outside of state hash range
		Last:  1,
	}
	count, err := GetStateHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestGetBlockHashCount_Success(t *testing.T) {
	testDb := generateTestAidaDb(t)
	cfg := &utils.Config{
		First: 21,
		Last:  30,
	}
	count, err := GetBlockHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), count)
}

func TestGetBlockHashCount_Empty(t *testing.T) {
	testDb := generateTestAidaDb(t)
	cfg := &utils.Config{
		First: 1, // Intentionally outside of block hash range
		Last:  1,
	}
	count, err := GetBlockHashCount(cfg, testDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestGetStateHashCount_Error(t *testing.T) {
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

func TestGetBlockHashCount_Error(t *testing.T) {
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
