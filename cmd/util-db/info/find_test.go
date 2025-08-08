package info

import (
	"testing"

	"github.com/0xsoniclabs/aida/cmd/util-db/dbutils"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
)

func TestFind_Info_FindBlockRangeInStateHash_Success(t *testing.T) {
	testDb, _ := dbutils.GenerateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	_, _, err := findBlockRangeInStateHash(testDb, log)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
	assert.Equal(t, "cannot get first state hash; not implemented", err.Error())
}

func TestFind_FindBlockRangeInStateHash_FirstError(t *testing.T) {
	testDb, _ := dbutils.GenerateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	first, last, err := findBlockRangeInStateHash(testDb, log)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), first)
	assert.Equal(t, uint64(0), last)
	assert.Contains(t, err.Error(), "cannot get first state hash")
}

func TestFind_FindBlockRangeInBlockHash_Success(t *testing.T) {
	testDb, _ := dbutils.GenerateTestAidaDb(t)
	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	first, last, err := findBlockRangeOfBlockHashes(testDb, log)
	assert.NoError(t, err)
	assert.Equal(t, uint64(21), first)
	assert.Equal(t, uint64(30), last)
}

func TestFind_FindBlockRangeInBlockHash_FirstError(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashDbFirstError"
	testDb, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}

	log := logger.NewLogger("Warning", "TestFindBlockRangeInStateHash")
	_, _, err = findBlockRangeOfBlockHashes(testDb, log)
	assert.Error(t, err)
	assert.Equal(t, "cannot get first block hash; no block hash found", err.Error())
}
