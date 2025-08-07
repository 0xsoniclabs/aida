// Copyright 2024 Fantom Foundation
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

package clone

import (
	"fmt"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	tests := []struct {
		name        string
		cloningType utils.AidaDbType
		dbc         string
		wantErr     string
	}{
		{"NoType", utils.NoType, "", "clone failed for NoType: incorrect clone type: 0"},
		{"GenType", utils.GenType, "", "clone failed for GenType: incorrect clone type: 1"},
		{"PatchType", utils.PatchType, "", ""},
		{"CloneType", utils.CloneType, "", ""},
		{"CustomTypeAll", utils.CustomType, "all", ""},
		{"CustomTypeSubstate", utils.CustomType, "substate", ""},
		{"CustomTypeDelete", utils.CustomType, "delete", ""},
		{"CustomTypeUpdate", utils.CustomType, "update", ""},
		{"CustomTypeStateHash", utils.CustomType, "state-hash", ""},
		{"CustomTypeBlockHash", utils.CustomType, "block-hash", ""},
		{"CustomTypeException", utils.CustomType, "exception", ""},
		{"CustomTypeInvalid", utils.CustomType, "invalid", "invalid db component: invalid. Usage: (\"all\", \"substate\", \"delete\", \"update\", \"state-hash\", \"block-hash\", \"exception\")"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aidaDb := utildb.GenerateTestAidaDb(t)
			err := testClone(t, aidaDb, tt.cloningType, tt.name, tt.dbc)
			if tt.wantErr != "" {
				assert.Error(t, err, "Expected error but got none")
				assert.Contains(t, err.Error(), tt.wantErr, "Error message does not match")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}

func testClone(t *testing.T, aidaDb db.BaseDB, cloningType utils.AidaDbType, name string, dbc string) error {
	cfg := &utils.Config{
		First:       0,
		Last:        100,
		Validate:    true, // TODO add substates with code to testDb then validate would produce error as count wouldn't match
		DbComponent: dbc,
	}
	cloneDb, err := db.NewDefaultBaseDB(t.TempDir() + "/clonedb_" + name)
	assert.NoError(t, err)

	err = clone(cfg, aidaDb, cloneDb, cloningType, false)
	if err != nil {
		//t.Fatalf("Clone failed for %s: %v", name, err)
		return fmt.Errorf("clone failed for %s: %v", name, err)
	}

	if dbc == "" || dbc == "all" || dbc == "substate" {
		t.Run("Substates", func(t *testing.T) {
			substateCount := 0
			substateDb := db.MakeDefaultSubstateDBFromBaseDB(cloneDb)
			substateIter := substateDb.NewIterator([]byte(db.SubstateDBPrefix), nil)
			for substateIter.Next() {
				substateCount++
			}
			assert.Equal(t, 10, substateCount, "Expected 10 substates in the cloned database")
		})
	}

	if dbc == "" || dbc == "all" || dbc == "update" {
		t.Run("UpdateSets", func(t *testing.T) {
			udb := db.MakeDefaultUpdateDBFromBaseDB(cloneDb)
			updateSetCount := 0
			updateSetIter := udb.NewUpdateSetIterator(cfg.First, cfg.Last)
			for updateSetIter.Next() {
				updateSetCount++
			}
			assert.Equal(t, 10, updateSetCount, "Expected 10 update sets in the cloned database")
		})
	}

	if dbc == "" || dbc == "all" || dbc == "delete" {
		t.Run("DeleteAccounts", func(t *testing.T) {
			deleteAccountCount := 0
			deleteAccountIter := cloneDb.NewIterator([]byte(db.DestroyedAccountPrefix), nil)
			for deleteAccountIter.Next() {
				deleteAccountCount++
			}
			assert.Equal(t, 10, deleteAccountCount, "Expected 10 deleted accounts in the cloned database")
		})
	}

	if dbc == "" || dbc == "all" || dbc == "state-hash" {
		t.Run("StateHashes", func(t *testing.T) {
			stateHashCount := 0
			stateHashIter := cloneDb.NewIterator([]byte(utils.StateRootHashPrefix), nil)
			for stateHashIter.Next() {
				stateHashCount++
			}
			assert.Equal(t, 10, stateHashCount, "Expected 10 state hashes in the cloned database")
		})
	}

	if dbc == "" || dbc == "all" || dbc == "block-hash" {
		t.Run("BlockHashes", func(t *testing.T) {
			blockHashCount := 0
			blockHashIter := cloneDb.NewIterator([]byte(utils.BlockHashPrefix), nil)
			for blockHashIter.Next() {
				blockHashCount++
			}
			assert.Equal(t, 10, blockHashCount, "Expected 10 block hashes in the cloned database")
		})
	}

	if dbc == "" || dbc == "all" || dbc == "exception" {
		t.Run("Exception", func(t *testing.T) {
			exceptionCount := 0
			exceptionIter := cloneDb.NewIterator([]byte(db.ExceptionDBPrefix), nil)
			for exceptionIter.Next() {
				exceptionCount++
			}
			assert.Equal(t, 10, exceptionCount, "Expected 10 exceptions in the cloned database")
		})
	}

	return nil
}

func TestClone_InvalidDbKeys(t *testing.T) {
	tests := []struct {
		name        string
		keyPrefix   string
		dbComponent string
		expectedErr string
	}{
		{
			name:        "SubstateInvalidDbKey",
			keyPrefix:   db.SubstateDBPrefix,
			dbComponent: "substate",
			expectedErr: "clone failed for SubstateInvalidDbKey: condition emit error; invalid length of substate db key: 5",
		},
		{
			name:        "UpdateSetsInvalidDbKey",
			keyPrefix:   db.UpdateDBPrefix,
			dbComponent: "update",
			expectedErr: "clone failed for UpdateSetsInvalidDbKey: condition emit error; invalid length of updateset key: 5",
		},
		{
			name:        "DestroyedAccountsInvalidDbKey",
			keyPrefix:   db.DestroyedAccountPrefix,
			dbComponent: "delete",
			expectedErr: "clone failed for DestroyedAccountsInvalidDbKey: condition emit error; invalid length of destroyed account key, expected 14, got 5",
		},
		{
			name:        "BlockHashInvalidDbKey",
			keyPrefix:   utils.BlockHashPrefix,
			dbComponent: "block-hash",
			expectedErr: "clone failed for BlockHashInvalidDbKey: condition emit error; invalid length of block hash key, expected at least 10, got 5",
		},
		{
			name:        "ExceptionInvalidDbKey",
			keyPrefix:   db.ExceptionDBPrefix,
			dbComponent: "exception",
			expectedErr: "clone failed for ExceptionInvalidDbKey: condition emit error; invalid length of exception key: 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir() + "/testAidaDb"
			aidaDb, err := db.NewDefaultBaseDB(tmpDir)
			if err != nil {
				t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
			}
			defer func() {
				err = aidaDb.Close()
				if err != nil {
					t.Fatalf("error closing aidaDb %s: %v", tmpDir, err)
				}
			}()

			err = aidaDb.Put([]byte(tt.keyPrefix+"inv"), []byte("test"))
			if err != nil {
				t.Fatalf("error putting invalid cloneDbCommand key: %v", err)
			}

			err = testClone(t, aidaDb, utils.CustomType, tt.name, tt.dbComponent)
			if err == nil {
				t.Fatalf("Expected error for invalid cloneDbCommand key, but got none")
			} else {
				assert.Equal(t, tt.expectedErr, err.Error())
			}
		})
	}
}

func TestClone_BlockHashes(t *testing.T) {
	cfg := &utils.Config{
		First:       0,
		Last:        100,
		Validate:    false,
		DbComponent: "block-hash",
	}
	aidaDb := utildb.GenerateTestAidaDb(t)

	cloneDb, err := db.NewDefaultBaseDB(t.TempDir() + "/clonedb")
	assert.NoError(t, err)

	err = clone(cfg, aidaDb, cloneDb, utils.CustomType, false)

	assert.NoError(t, err)

	// Verify that the cloned database has the expected block hashes
	blockHashCount := 0
	blockHashIter := cloneDb.NewIterator([]byte(utils.BlockHashPrefix), nil)
	for blockHashIter.Next() {
		blockHashCount++
	}

	assert.Equal(t, 10, blockHashCount, "Expected 10 block hashes in the cloned database")
}

func TestClone_LastUpdateBeforeRange(t *testing.T) {
	cfg := &utils.Config{
		First:       1000,
		Last:        1001,
		Validate:    false,
		DbComponent: "block-hash",
	}
	aidaDb := utildb.GenerateTestAidaDb(t)

	cloneDb, err := db.NewDefaultBaseDB(t.TempDir() + "/clonedb")
	assert.NoError(t, err)

	err = clone(cfg, aidaDb, cloneDb, utils.CloneType, false)

	assert.NoError(t, err)

	// Verify that the cloned database has the expected block hashes
	blockHashCount := 0
	blockHashIter := cloneDb.NewIterator([]byte(utils.BlockHashPrefix), nil)
	for blockHashIter.Next() {
		blockHashCount++
	}

	assert.Equal(t, 10, blockHashCount, "Expected 10 block hashes in the cloned database")
}

func TestClone_OpenCloningDbs_SourceDbNotExist(t *testing.T) {
	_, _, err := openCloningDbs("/not/exist/source", "/tmp/target")
	assert.Error(t, err)
}

func TestClose_OpenCloningDbs_SourceDbInvalid(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "sourcedb")
	_, _, err := openCloningDbs(tmpFile.Name(), "/tmp/target")
	assert.Error(t, err)
}

func TestClone_OpenCloningDbs_TargetExists(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "targetdb")
	defer os.Remove(tmpFile.Name())
	_, _, err := openCloningDbs(tmpFile.Name(), tmpFile.Name())
	assert.Error(t, err)
}

func TestClone_OpenCloningDbs_Success(t *testing.T) {
	sourceDir := t.TempDir() + "/source"
	targetDir := t.TempDir() + "/target"

	// Create a source database
	sourceDb, err := db.NewDefaultBaseDB(sourceDir)
	assert.NoError(t, err)

	err = sourceDb.Close()
	assert.NoError(t, err)

	// Open cloning databases
	openedSourceDb, openedTargetDb, err := openCloningDbs(sourceDir, targetDir)
	assert.NoError(t, err)

	err = openedSourceDb.Close()
	assert.NoError(t, err)
	err = openedTargetDb.Close()
	assert.NoError(t, err)
}
