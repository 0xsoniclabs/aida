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

package clone

import (
	"fmt"

	"go.uber.org/mock/gomock"
	"math/big"
	"os"
	"strconv"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func testClone(t *testing.T, aidaDb db.SubstateDB, cloningType utils.AidaDbType, name string, dbc string) error {
	cfg := &utils.Config{
		First:       0,
		Last:        100,
		Workers:     1,
		Validate:    true, // TODO add substates with code to testDb then validate would produce error as count wouldn't match
		DbComponent: dbc,
		CompactDb:   true,
	}
	cloneDb, err := db.NewDefaultSubstateDB(t.TempDir() + "/clonedb_" + name)
	assert.NoError(t, err)

	err = clone(cfg, aidaDb, cloneDb, cloningType)
	if err != nil {
		//t.Fatalf("Clone failed for %s: %v", testName, err)
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
			aidaDb, err := db.NewDefaultSubstateDB(tmpDir)
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
				t.Fatalf("error putting invalid db key: %v", err)
			}

			err = testClone(t, aidaDb, utils.CustomType, tt.name, tt.dbComponent)
			if err == nil {
				t.Fatalf("Expected error for invalid db key, but got none")
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

	cloneDb, err := db.NewDefaultSubstateDB(t.TempDir() + "/clonedb")
	assert.NoError(t, err)

	err = clone(cfg, aidaDb, cloneDb, utils.CustomType)

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

	cloneDb, err := db.NewDefaultSubstateDB(t.TempDir() + "/clonedb")
	assert.NoError(t, err)

	err = clone(cfg, aidaDb, cloneDb, utils.CloneType)

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
	_, _, err := openCloningDbs(t.TempDir(), t.TempDir())
	assert.Error(t, err)
}

func TestClone_OpenCloningDbs_TargetExists(t *testing.T) {
	tmpFile := t.TempDir()
	_, _, err := openCloningDbs(tmpFile, tmpFile)
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

func TestClone_Commands(t *testing.T) {
	ss, srcDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	tests := []struct {
		cmdName  string
		testName string
		action   cli.ActionFunc
		wantErr  string
		args     []string
	}{
		{
			cmdName:  cloneCustomCommand.Name,
			testName: cloneCustomCommand.Name + "_Success",
			action:   cloneCustomAction,
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
				"0",
				"0",
			},
		},
		{
			cmdName:  cloneDbCommand.Name,
			testName: cloneDbCommand.Name + "_Success",
			action:   cloneDbAction,
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
				"0",
				"0",
			},
		},
		{
			cmdName:  clonePatchCommand.Name,
			testName: clonePatchCommand.Name + "_Success",
			action:   clonePatchAction,
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
				"0",
				"0",
			},
		},
		{
			cmdName:  clonePatchCommand.Name,
			testName: clonePatchCommand.Name + "_Incorrect_Block_Range",
			action:   clonePatchAction,
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				"11",
				"10",
				"0",
				"0",
			},
			wantErr: "first block 11 has larger number than last block 10",
		},
		{
			cmdName:  clonePatchCommand.Name,
			testName: clonePatchCommand.Name + "_Incorrect_Epoch_Range",
			action:   clonePatchAction,
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
				"11",
				"10",
			},
			wantErr: "first block 11 has larger number than last block 10",
		},
		{
			cmdName:  cloneCustomCommand.Name,
			testName: cloneCustomCommand.Name + "_Invalid_NumberOfArgs",
			action:   cloneCustomAction,
			wantErr:  "command requires 2 arguments",
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
			},
		},
		{
			cmdName:  cloneDbCommand.Name,
			testName: cloneDbCommand.Name + "_Invalid_NumberOfArgs",
			action:   cloneDbAction,
			wantErr:  "command requires 2 arguments",
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
			},
		},
		{
			cmdName:  clonePatchCommand.Name,
			testName: clonePatchCommand.Name + "_Invalid_NumberOfArgs",
			action:   clonePatchAction,
			wantErr:  "clone patch command requires exactly 4 arguments",
			args: []string{
				"--aida-db",
				srcDbPath,
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
			},
		},
		{
			cmdName:  cloneCustomCommand.Name,
			testName: cloneCustomCommand.Name + "_SrcDoesNotExist",
			action:   cloneCustomAction,
			args: []string{
				"--aida-db",
				"/some/wrong/src/path",
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
				"0",
				"0",
			},
			wantErr: "specified aida-db /some/wrong/src/path is empty",
		},
		{
			cmdName:  cloneDbCommand.Name,
			testName: cloneDbCommand.Name + "_SrcDoesNotExist",
			action:   cloneDbAction,
			args: []string{
				"--aida-db",
				"/some/wrong/src/path",
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
				"0",
				"0",
			},
			wantErr: "specified aida-db /some/wrong/src/path is empty",
		},
		{
			cmdName:  clonePatchCommand.Name,
			testName: clonePatchCommand.Name + "_SrcDoesNotExist",
			action:   clonePatchAction,
			args: []string{
				"--aida-db",
				"/some/wrong/src/path",
				"--target-db",
				t.TempDir() + "/target.db",
				"--db-component",
				"all",
				"-l",
				"CRITICAL",
				strconv.FormatUint(ss.Block-1, 10),
				strconv.FormatUint(ss.Block+1, 10),
				"0",
				"0",
			},
			wantErr: "specified aida-db /some/wrong/src/path is empty",
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			app := cli.NewApp()
			app.Action = test.action
			app.Flags = []cli.Flag{
				&utils.AidaDbFlag,
				&utils.TargetDbFlag,
				&logger.LogLevelFlag,
				&utils.DbComponentFlag,
			}
			targetDbPath := test.args[3]

			err := app.Run(append([]string{test.cmdName}, test.args...))
			if test.wantErr == "" {
				require.NoError(t, err)
				require.Condition(t, func() bool {
					stat, err := os.Stat(targetDbPath)
					require.NoError(t, err)
					return stat != nil && stat.IsDir() && stat.Size() > 0
				}, "Target database seems to be empty")
			} else {
				require.ErrorContains(t, err, test.wantErr)
			}

		})
	}
}

func TestCloner_CloneCodes_ClonesCodesFromInputAndOutputSubstate(t *testing.T) {
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	ss := createTestSubstate(t, 1, []byte{1}, []byte{2})
	err = srcDb.PutSubstate(ss)
	require.NoError(t, err, "failed to put substate")

	dstPath := t.TempDir()
	dstDb, err := db.NewDefaultSubstateDB(dstPath)
	require.NoError(t, err, "failed to create destination db")
	err = dstDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	clnr := cloner{
		cfg: &utils.Config{
			First:            1,
			Last:             10,
			Workers:          1,
			SubstateEncoding: "protobuf",
		},
		aidaDb:  srcDb,
		cloneDb: dstDb,
		log:     logger.NewLogger("warn", "CloneCodesTest"),
	}

	err = clnr.cloneCodes()
	require.NoError(t, err, "failed to clone codes")

	codeDb := db.MakeDefaultCodeDBFromBaseDB(dstDb)
	ok, err := codeDb.HasCode(hash.Keccak256Hash([]byte{1}))
	require.NoError(t, err, "failed to check if code exists")
	require.True(t, ok, "code does not exist")
	ok, err = codeDb.HasCode(hash.Keccak256Hash([]byte{2}))
	require.NoError(t, err, "failed to check if code exists")
	require.True(t, ok, "code does not exist")
}

func TestCloner_PutCode_DoesNotPutNilCode(t *testing.T) {
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	// Create one substate with nil code and one with empty code
	ss1 := createTestSubstate(t, 1, nil, []byte{123})
	err = srcDb.PutSubstate(ss1)
	require.NoError(t, err, "failed to put substate")

	// PutCode must be called only once for each code
	ctrl := gomock.NewController(t)
	dstDb := db.NewMockSubstateDB(ctrl)
	// only one code should be put
	dstDb.EXPECT().PutCode([]byte{123}).Times(1)

	clnr := cloner{
		cfg: &utils.Config{
			First:            1,
			Last:             10,
			Workers:          1,
			SubstateEncoding: "protobuf",
		},
		aidaDb:  srcDb,
		cloneDb: dstDb,
		log:     logger.NewLogger("warn", "CloneCodesTest"),
	}

	err = clnr.cloneCodes()
	require.NoError(t, err, "failed to clone codes")
}

func TestCloner_CloneCodes_DoesNotCloneDuplicates(t *testing.T) {
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	// Create two identical substates with different tx numbers
	ss1 := createTestSubstate(t, 1, []byte{1}, []byte{1})
	err = srcDb.PutSubstate(ss1)
	require.NoError(t, err, "failed to put substate")

	// PutCode must be called only once for each code
	ctrl := gomock.NewController(t)
	dstDb := db.NewMockSubstateDB(ctrl)
	dstDb.EXPECT().PutCode([]byte{1}).Times(1)

	clnr := cloner{
		cfg: &utils.Config{
			First:            1,
			Last:             10,
			Workers:          1,
			SubstateEncoding: "protobuf",
		},
		aidaDb:  srcDb,
		cloneDb: dstDb,
		log:     logger.NewLogger("warn", "CloneCodesTest"),
	}

	err = clnr.cloneCodes()
	require.NoError(t, err, "failed to clone codes")
}

func TestOpenCloningDbs_OpensDbsCorrectly(t *testing.T) {
	tmp := t.TempDir()
	srcPath := tmp + "/src"
	dstPath := tmp + "/dst"
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	ss1 := createTestSubstate(t, 1, []byte{1}, []byte{1})
	err = srcDb.PutSubstate(ss1)
	require.NoError(t, err, "failed to put substate")

	// Close the db to test opening
	require.NoError(t, srcDb.Close())

	srcDb, dstDb, err := OpenCloningDbs(srcPath, dstPath, "protobuf")
	require.NoError(t, err, "failed to open cloning dbs")

	// check correct opening of source db
	srcDbSs, err := srcDb.GetSubstate(1, 1)
	require.NoError(t, err, "failed to get substate")
	require.NoError(t, srcDbSs.Equal(ss1))
	// Make sure destination db is empty
	iter := dstDb.NewSubstateIterator(0, 1)
	require.False(t, iter.Next())
}

func createTestSubstate(t *testing.T, tx int, codeA, codeB []byte) *substate.Substate {
	t.Helper()
	random := types.Hash{1}
	to := types.Address{1}
	return &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{1}: &substate.Account{
				Code:    codeA,
				Balance: uint256.NewInt(10),
				Storage: make(map[types.Hash]types.Hash),
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{2}: &substate.Account{
				Code:    codeB,
				Balance: uint256.NewInt(10),
				Storage: make(map[types.Hash]types.Hash),
			},
		},
		Env: &substate.Env{
			Difficulty:  big.NewInt(10),
			BaseFee:     big.NewInt(10),
			BlobBaseFee: big.NewInt(10),
			BlockHashes: make(map[uint64]types.Hash),
			Random:      &random,
		},
		Message: &substate.Message{
			CheckNonce:            true,
			GasPrice:              big.NewInt(10),
			To:                    &to,
			Value:                 big.NewInt(10),
			AccessList:            make(types.AccessList, 0),
			GasFeeCap:             big.NewInt(10),
			GasTipCap:             big.NewInt(10),
			BlobHashes:            make([]types.Hash, 0),
			SetCodeAuthorizations: make([]types.SetCodeAuthorization, 0),
			Data:                  []byte{0x1, 0x2, 0x3, 0x4, 0x5},
		},
		Result:      &substate.Result{},
		Block:       uint64(1),
		Transaction: tx,
	}
}

func TestClone_CorrectlyClonesData(t *testing.T) {
	// prepare the source db
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	md := utils.NewAidaDbMetadata(srcDb, "INFO")
	err = md.SetChainID(utils.MainnetChainID)
	require.NoError(t, err, "failed to set chain id")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")
	ss := createTestSubstate(t, 1, []byte{1}, []byte{1})
	err = srcDb.PutSubstate(ss)

	targetPath := t.TempDir()
	targetDb, err := db.NewDefaultSubstateDB(targetPath)
	require.NoError(t, err, "failed to create target db")

	cfg := &utils.Config{First: 0, Last: 1, ChainID: utils.MainnetChainID, Workers: 1}
	err = CreatePatchClone(cfg, srcDb, targetDb, 5577, 5578, true)
	require.NoError(t, err, "failed to clone codes")

	gotSs, err := targetDb.GetSubstate(1, 1)
	require.NoError(t, err, "failed to get substate")

	err = ss.Equal(gotSs)
	require.NoError(t, err)
}
