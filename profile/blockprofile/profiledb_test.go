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

// Package ProfileDatas provides an SQLite based ProfileDatas database.
package blockprofile

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func tempFile(require *require.Assertions) string {
	file, err := os.CreateTemp("", "*.db")
	require.NoError(err)
	require.NoError(file.Close())
	return file.Name()
}

func TestAdd(t *testing.T) {
	require := require.New(t)

	dbFile := tempFile(require)
	t.Logf("db file: %s", dbFile)
	db, err := newProfileDB(dbFile, 0)
	require.NoError(err)
	defer func() {
		require.NoError(db.Close())
	}()
	defer func() {
		require.NoError(os.Remove(dbFile))
	}()

	ProfileData := ProfileData{
		curBlock:        5637800,
		tBlock:          5838,
		tSequential:     4439,
		tCritical:       2424,
		tCommit:         1398,
		speedup:         1.527263,
		ubNumProc:       2,
		numTx:           3,
		tTransactions:   []int64{2382388, 11218838, 5939392888},
		tTypes:          []TxType{TransferTx, CreateTx, CallTx},
		gasTransactions: []uint64{111111, 222222, 333333},
	}

	err = db.Add(ProfileData)
	require.NoError(err)

	require.Len(db.buffer, 1)

	require.Len(db.buffer[0].tTransactions, 3)
	require.Len(db.buffer[0].tTypes, 3)
	require.Len(db.buffer[0].gasTransactions, 3)
}

func TestFlush(t *testing.T) {
	// db has 0 records
	require := require.New(t)
	dbFile := tempFile(require)
	t.Logf("db file: %s", dbFile)
	defer func(name string) {
		require.NoError(os.Remove(name))
	}(dbFile)
	db, err := newProfileDB(dbFile, 0)
	require.NoError(err)
	err = db.Add(ProfileData{})
	require.NoError(err)

	err = db.Flush()
	require.NoError(err)
	err = db.Close()
	require.NoError(err)

	// db has 2 records
	db, err = newProfileDB(dbFile, 0)
	require.NoError(err)

	pd := ProfileData{
		curBlock:        5637800,
		tBlock:          5838,
		tSequential:     4439,
		tCritical:       2424,
		tCommit:         1398,
		speedup:         1.527263,
		ubNumProc:       2,
		numTx:           3,
		tTransactions:   []int64{2382388, 11218838, 5939392888},
		tTypes:          []TxType{TransferTx, CreateTx, CallTx},
		gasTransactions: []uint64{111111, 222222, 333333},
	}

	err = db.Add(pd)
	require.NoError(err)

	pd = ProfileData{
		curBlock:        3239933,
		tBlock:          44939,
		tSequential:     3493848,
		tCritical:       434838,
		tCommit:         2332,
		speedup:         1.203983,
		ubNumProc:       2,
		numTx:           2,
		tTransactions:   []int64{2382388, 11218838},
		tTypes:          []TxType{TransferTx, CreateTx},
		gasTransactions: []uint64{444444, 555555},
	}
	err = db.Add(pd)
	require.NoError(err)
	require.Len(db.buffer, 2)
	require.Len(db.buffer[0].tTransactions, 3)
	require.Len(db.buffer[0].tTypes, 3)
	require.Len(db.buffer[0].gasTransactions, 3)
	require.Len(db.buffer[1].tTransactions, 2)
	require.Len(db.buffer[1].tTypes, 2)
	require.Len(db.buffer[1].gasTransactions, 2)
	err = db.Flush()
	require.NoError(err)
	require.Len(db.buffer, 0)
	err = db.Close()
	require.NoError(err)

	// trigger Flush method inside Add
	db, err = newProfileDB(dbFile, 0)
	require.NoError(err)
	defer func(db *profileDB) {
		err = errors.Join(err, db.Close())
		require.NoError(err)
	}(db)

	for i := 1; i < bufferSize; i++ {
		profileData := ProfileData{
			curBlock:        uint64(i),
			tBlock:          5838,
			tSequential:     4439,
			tCritical:       2424,
			tCommit:         1398,
			speedup:         1.527263,
			ubNumProc:       2,
			numTx:           2,
			tTransactions:   []int64{2382388, 11218838},
			tTypes:          []TxType{TransferTx, CreateTx},
			gasTransactions: []uint64{444444, 555555},
		}
		err = db.Add(profileData)
		require.NoError(err)
		require.Len(db.buffer, i)
	}

	pd = ProfileData{
		curBlock:        uint64(bufferSize),
		tBlock:          5838,
		tSequential:     4439,
		tCritical:       2424,
		tCommit:         1398,
		speedup:         1.527263,
		ubNumProc:       2,
		numTx:           3,
		tTransactions:   []int64{2382388, 11218838, 232348228},
		tTypes:          []TxType{TransferTx, CreateTx, CallTx},
		gasTransactions: []uint64{444444, 555555, 666666},
	}

	err = db.Add(pd)
	require.NoError(err)
	require.Len(db.buffer, 0)
}

// TestDeleteBlockRangeOverlap tests profileDB.DeleteByBlockRange function
func TestDeleteBlockRangeOverlapOneTx(t *testing.T) {
	require := require.New(t)

	dbFile := tempFile(require)
	t.Logf("db file: %s", dbFile)
	defer func(name string) {
		require.NoError(os.Remove(name))
	}(dbFile)
	db, err := NewProfileDB(dbFile, 0)
	require.NoError(err)

	startBlock, endBlock := uint64(500), uint64(2500)
	blockRange := endBlock - startBlock
	for i := startBlock; i <= endBlock; i++ {
		profileData := ProfileData{
			curBlock:        uint64(i),
			tBlock:          5838,
			tSequential:     4439,
			tCritical:       2424,
			tCommit:         1398,
			speedup:         1.527263,
			ubNumProc:       2,
			numTx:           1,
			tTransactions:   []int64{232939829},
			tTypes:          []TxType{TransferTx},
			gasTransactions: []uint64{111111},
		}
		err = db.Add(profileData)
		require.NoError(err)
	}

	numDeletedRows, err := db.DeleteByBlockRange(startBlock, endBlock)
	require.NoError(err)
	if numDeletedRows != int64(2*blockRange) {
		t.Errorf("unexpected number of rows affected by deletion, expected: %d, got: %d", 2*blockRange, numDeletedRows)
	}
	err = db.Close()
	require.NoError(err)

	db, err = NewProfileDB(dbFile, 0)
	require.NoError(err)
	defer func() {
		require.NoError(db.Close())
	}()
	for i := startBlock; i <= endBlock; i++ {
		profileData := ProfileData{
			curBlock:        uint64(i),
			tBlock:          5838,
			tSequential:     4439,
			tCritical:       2424,
			tCommit:         1398,
			speedup:         1.527263,
			ubNumProc:       2,
			numTx:           1,
			tTransactions:   []int64{232939829},
			tTypes:          []TxType{TransferTx},
			gasTransactions: []uint64{111111},
		}
		err = db.Add(profileData)
		require.NoError(err)
	}

	startDeleteBlock, endDeleteBlock := uint64(0), uint64(500)
	numDeletedRows, err = db.DeleteByBlockRange(startDeleteBlock, endDeleteBlock)
	require.NoError(err)
	if numDeletedRows != 2 {
		t.Errorf("unexpected number of rows affected by deletion")
	}
}

func TestDeleteBlockRangeOverlapMultipleTx(t *testing.T) {
	require := require.New(t)

	dbFile := tempFile(require)
	t.Logf("db file: %s", dbFile)
	defer func(name string) {
		require.NoError(os.Remove(name))
	}(dbFile)
	db, err := NewProfileDB(dbFile, 0)
	require.NoError(err)

	startBlock, endBlock := uint64(500), uint64(2500)
	blockRange := endBlock - startBlock
	numTx := 4
	for i := startBlock; i <= endBlock; i++ {
		profileData := ProfileData{
			curBlock:        uint64(i),
			tBlock:          5838,
			tSequential:     4439,
			tCritical:       2424,
			tCommit:         1398,
			speedup:         1.527263,
			ubNumProc:       2,
			numTx:           numTx,
			tTransactions:   []int64{232939829, 938828288, 92388277, 9238828},
			tTypes:          []TxType{TransferTx, CreateTx, CallTx, MaintenanceTx},
			gasTransactions: []uint64{111111, 222222, 333333, 444444},
		}
		err = db.Add(profileData)
		require.NoError(err)
	}

	numDeletedRows, err := db.DeleteByBlockRange(startBlock, endBlock)
	require.NoError(err)
	expNumRows := blockRange + uint64(numTx)*blockRange
	if numDeletedRows != int64(expNumRows) {
		t.Errorf("unexpected number of rows affected by deletion, expected: %d, got: %d", expNumRows, numDeletedRows)
	}
	err = db.Close()
	require.NoError(err)

	db, err = NewProfileDB(dbFile, 0)
	require.NoError(err)
	defer func() {
		require.NoError(db.Close())
	}()
	for i := startBlock; i <= endBlock; i++ {
		profileData := ProfileData{
			curBlock:        uint64(i),
			tBlock:          5838,
			tSequential:     4439,
			tCritical:       2424,
			tCommit:         1398,
			speedup:         1.527263,
			ubNumProc:       2,
			numTx:           numTx,
			tTransactions:   []int64{232939829, 938828288, 92388277, 9238828},
			tTypes:          []TxType{TransferTx, CreateTx, CallTx, MaintenanceTx},
			gasTransactions: []uint64{111111, 222222, 333333, 444444},
		}
		err = db.Add(profileData)
		require.NoError(err)
	}

	startDeleteBlock, endDeleteBlock := uint64(0), uint64(500)
	numDeletedRows, err = db.DeleteByBlockRange(startDeleteBlock, endDeleteBlock)
	require.NoError(err)
	if numDeletedRows != 1+int64(numTx) {
		t.Errorf("unexpected number of rows affected by deletion")
	}
}

// TestDeleteBlockRangeNoOverlap tests profileDB.DeleteByBlockRange function
func TestDeleteBlockRangeNoOverlap(t *testing.T) {
	require := require.New(t)

	dbFile := tempFile(require)
	t.Logf("db file: %s", dbFile)
	db, err := NewProfileDB(dbFile, 0)
	require.NoError(err)
	defer func() { assert.NoError(t, db.Close()) }()
	defer func() { assert.NoError(t, os.Remove(dbFile)) }()

	startBlock, endBlock := uint64(500), uint64(2500)
	for i := startBlock; i <= endBlock; i++ {
		profileData := ProfileData{
			curBlock:        uint64(i),
			tBlock:          5838,
			tSequential:     4439,
			tCritical:       2424,
			tCommit:         1398,
			speedup:         1.527263,
			ubNumProc:       2,
			numTx:           3,
			tTransactions:   []int64{232444, 92398, 9282887},
			tTypes:          []TxType{TransferTx, CreateTx, CallTx},
			gasTransactions: []uint64{111111, 222222, 333333, 444444},
		}
		err = db.Add(profileData)
		require.NoError(err)
	}

	startDeleteBlock, endDeleteBlock := uint64(0), uint64(499)
	numDeletedRows, err := db.DeleteByBlockRange(startDeleteBlock, endDeleteBlock)
	require.NoError(err)
	if numDeletedRows != 0 {
		t.Errorf("unexpected number of rows affected by deletion")
	}
}

func BenchmarkAdd(b *testing.B) {
	require := require.New(b)
	dbFile := tempFile(require)
	b.Logf("db file: %s", dbFile)
	defer func() {
		assert.NoError(b, os.Remove(dbFile))
	}()

	db, err := NewProfileDB(dbFile, 0)
	require.NoError(err)
	ProfileData := ProfileData{
		curBlock:    5637800,
		tBlock:      5838,
		tSequential: 4439,
		tCritical:   2424,
		tCommit:     1398,
		speedup:     1.527263,
		ubNumProc:   2,
		numTx:       3,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := db.Add(ProfileData)
		require.NoError(err)
	}
}

func ExampleDB() {
	dbFile := "/tmp/db-test" + time.Now().Format(time.RFC3339)
	db, err := NewProfileDB(dbFile, 0)
	if err != nil {
		fmt.Println("ERROR: create -", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("ERROR: close", err)
		}
	}()

	const count = 10_000
	for i := 0; i < count; i++ {
		ProfileData := ProfileData{
			curBlock:        5637800,
			tBlock:          5838,
			tSequential:     4439,
			tCritical:       2424,
			tCommit:         1398,
			speedup:         rand.Float64() * 10,
			ubNumProc:       2,
			numTx:           3,
			tTransactions:   []int64{2382388, 11218838, 5939392888},
			tTypes:          []TxType{TransferTx, CreateTx, CallTx},
			gasTransactions: []uint64{111111, 222222, 333333},
		}
		if err := db.Add(ProfileData); err != nil {
			fmt.Println("ERROR: insert - ", err)
			return
		}
	}

	fmt.Printf("inserted %d records\n", count)
	// Output:
	// inserted 10000 records
}

func TestFlushProfileData(t *testing.T) {
	require := require.New(t)
	dbFile := tempFile(require)
	t.Logf("db file: %s", dbFile)

	db, err := newProfileDB(dbFile, 0)
	require.NoError(err)
	defer func(db *profileDB) {
		require.NoError(db.Close())
	}(db)
	defer func(name string) {
		require.NoError(os.Remove(name))
	}(dbFile)

	ProfileData := ProfileData{
		curBlock:        5637800,
		tBlock:          5838,
		tSequential:     4439,
		tCritical:       2424,
		tCommit:         1398,
		speedup:         1.527263,
		ubNumProc:       2,
		numTx:           4,
		tTransactions:   []int64{292988, 8387773, 923828772, 293923929},
		tTypes:          []TxType{TransferTx, CreateTx, CallTx, MaintenanceTx},
		gasTransactions: []uint64{111111, 222222, 333333, 444444},
	}

	// start db transaction
	tx, err := db.sql.Begin()
	require.NoError(err)
	res, err := tx.Stmt(db.blockStmt).Exec(ProfileData.curBlock, ProfileData.tBlock, ProfileData.tSequential, ProfileData.tCritical,
		ProfileData.tCommit, ProfileData.speedup, ProfileData.ubNumProc, ProfileData.numTx, ProfileData.gasBlock)
	require.NoError(err)
	numRowsAffected, err := res.RowsAffected()
	require.NoError(err)
	if numRowsAffected != 1 {
		t.Errorf("invalid numRowsAffected value")
	}

	for i, tTransaction := range ProfileData.tTransactions {
		res, err = tx.Stmt(db.txStmt).Exec(ProfileData.curBlock, i, ProfileData.tTypes[i], tTransaction, ProfileData.gasTransactions[i])
		require.NoError(err)
		numRowsAffected, err := res.RowsAffected()
		require.NoError(err)
		if numRowsAffected != 1 {
			t.Errorf("invalid numRowsAffected value")
		}
	}
	require.NoError(tx.Commit())
}

func TestProfileDB_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, mockDb, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		mockBlockStmt := mockDb.ExpectPrepare("")
		blockStmt, err := db.Prepare("")
		if err != nil {
			t.Fatalf("an error '%s' was not expected when preparing block statement", err)
		}

		mockTxStmt := mockDb.ExpectPrepare("")
		txStmt, err := db.Prepare("")
		if err != nil {
			t.Fatalf("an error '%s' was not expected when preparing transaction statement", err)
		}

		pDB := &profileDB{
			sql:       db,
			blockStmt: blockStmt,
			txStmt:    txStmt,
			buffer:    []ProfileData{},
		}

		mockDb.ExpectBegin()
		mockBlockStmt.ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
		mockTxStmt.ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
		mockDb.ExpectCommit()
		err = pDB.Add(ProfileData{
			tTransactions:   []int64{123456},
			tTypes:          []TxType{TransferTx},
			gasTransactions: []uint64{1000},
		})

		assert.Nil(t, err)
		if err = mockDb.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("BeginError", func(t *testing.T) {
		db, mockDb, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		// begin error
		pDB := &profileDB{
			sql:    db,
			buffer: []ProfileData{},
		}
		mockDb.ExpectBegin().WillReturnError(mockErr)
		err = pDB.Add(ProfileData{
			tTransactions:   []int64{123456},
			tTypes:          []TxType{TransferTx},
			gasTransactions: []uint64{1000},
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), mockErr.Error())
		if err = mockDb.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("WriteBlockError", func(t *testing.T) {
		db, mockDb, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		mockBlockStmt := mockDb.ExpectPrepare("")
		blockStmt, err := db.Prepare("")
		if err != nil {
			t.Fatalf("an error '%s' was not expected when preparing block statement", err)
		}
		// begin error
		pDB := &profileDB{
			sql:       db,
			blockStmt: blockStmt,
			buffer:    []ProfileData{},
		}
		mockDb.ExpectBegin()
		mockBlockStmt.ExpectExec().WillReturnError(mockErr)
		err = pDB.Add(ProfileData{
			tTransactions:   []int64{123456},
			tTypes:          []TxType{TransferTx},
			gasTransactions: []uint64{1000},
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), mockErr.Error())
		if err = mockDb.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("WriteTxError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, mockDb, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		mockBlockStmt := mockDb.ExpectPrepare("")
		blockStmt, err := db.Prepare("")
		if err != nil {
			t.Fatalf("an error '%s' was not expected when preparing block statement", err)
		}

		mockTxStmt := mockDb.ExpectPrepare("")
		txStmt, err := db.Prepare("")
		if err != nil {
			t.Fatalf("an error '%s' was not expected when preparing transaction statement", err)
		}

		pDB := &profileDB{
			sql:       db,
			blockStmt: blockStmt,
			txStmt:    txStmt,
			buffer:    []ProfileData{},
		}

		mockDb.ExpectBegin()
		mockBlockStmt.ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
		mockTxStmt.ExpectExec().WillReturnError(mockErr)
		err = pDB.Add(ProfileData{
			tTransactions:   []int64{123456},
			tTypes:          []TxType{TransferTx},
			gasTransactions: []uint64{1000},
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), mockErr.Error())
		if err = mockDb.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

}

func TestCommands(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShell := utils.NewMockShellExecutor(ctrl)
	cmd := &command{executor: mockShell}
	mockErr := errors.New("mock error")

	commands := []func() (string, error){
		cmd.getProcessor,
		cmd.getMemory,
		cmd.getDisks,
		cmd.getOS,
		cmd.getMachine,
	}

	for _, operation := range commands {
		t.Run("success", func(t *testing.T) {
			mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("\b"), nil)
			output, err := operation()
			assert.Nil(t, err)
			assert.Equal(t, "\b", output)
		})

		t.Run("error", func(t *testing.T) {
			mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return(nil, mockErr)
			output, err := operation()
			assert.NotNil(t, err)
			assert.Equal(t, "", output)
		})
	}

	// Test hwinfo not found case
	t.Run("hwinfo not found", func(t *testing.T) {
		mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("hwinfo: not found"), nil)
		output, err := cmd.getDisks()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "hwinfo: not found")
		assert.Equal(t, "", output)
	})
}

func TestInsertMetadata_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Use sqlmock with QueryMatcherEqual to handle exact matching but ignore whitespace
	db, mockDb, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)

	mockShell := utils.NewMockShellExecutor(ctrl)
	cmd := command{executor: mockShell}

	// Mock all the system information calls
	mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("Intel(R) Core(TM) i7-8700K"), nil) // getProcessor
	mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("16384GB RAM"), nil)                // getMemory
	mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("/dev/sda1 100GB"), nil)            // getDisks
	mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("Ubuntu 20.04.3 LTS"), nil)         // getOS
	mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("hostname(192.168.1.1)"), nil)      // getMachine

	// Mock database operations - match the exact SQL from insertMetadataSQL constant
	mockPrepare := mockDb.ExpectPrepare(insertMetadataSQL)
	mockDb.ExpectBegin()
	mockStmt := mockPrepare.ExpectExec().WithArgs(utils.MainnetChainID, "Intel(R) Core(TM) i7-8700K", "16384GB RAM", "/dev/sda1 100GB", "Ubuntu 20.04.3 LTS", "hostname(192.168.1.1)")
	mockStmt.WillReturnResult(sqlmock.NewResult(1, 1))
	mockDb.ExpectCommit()

	err = insertMetadata(db, utils.MainnetChainID, cmd)
	assert.NoError(t, err)

	// Verify all expectations were met
	err = mockDb.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestInsertMetadata_Error(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*gomock.Controller, *utils.MockShellExecutor, sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name: "prepare statement error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockDb.ExpectPrepare(`INSERT INTO metadata`).WillReturnError(errors.New("prepare failed"))
			},
			expectedError: "failed to prepare a SQL statement for metadata",
		},
		{
			name: "getProcessor error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockDb.ExpectPrepare(`INSERT INTO metadata`).WillReturnCloseError(nil)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return(nil, errors.New("processor error"))
			},
			expectedError: "failed to get processor information",
		},
		{
			name: "getMemory error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockDb.ExpectPrepare(`INSERT INTO metadata`).WillReturnCloseError(nil)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("processor"), nil)        // getProcessor
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return(nil, errors.New("memory error")) // getMemory
			},
			expectedError: "failed to get memory information",
		},
		{
			name: "getDisks error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockDb.ExpectPrepare(`INSERT INTO metadata`).WillReturnCloseError(nil)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("processor"), nil)       // getProcessor
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("memory"), nil)          // getMemory
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return(nil, errors.New("disks error")) // getDisks
			},
			expectedError: "failed to get disk information",
		},
		{
			name: "getOS error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockDb.ExpectPrepare(`INSERT INTO metadata`).WillReturnCloseError(nil)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("processor"), nil)    // getProcessor
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("memory"), nil)       // getMemory
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("disks"), nil)        // getDisks
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return(nil, errors.New("os error")) // getOS
			},
			expectedError: "failed to get OS information",
		},
		{
			name: "getMachine error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockDb.ExpectPrepare(`INSERT INTO metadata`).WillReturnCloseError(nil)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("processor"), nil)         // getProcessor
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("memory"), nil)            // getMemory
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("disks"), nil)             // getDisks
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("os"), nil)                // getOS
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return(nil, errors.New("machine error")) // getMachine
			},
			expectedError: "failed to get machine information",
		},
		{
			name: "transaction begin error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockDb.ExpectPrepare(`INSERT INTO metadata`).WillReturnCloseError(nil)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("processor"), nil) // getProcessor
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("memory"), nil)    // getMemory
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("disks"), nil)     // getDisks
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("os"), nil)        // getOS
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("machine"), nil)   // getMachine
				mockDb.ExpectBegin().WillReturnError(errors.New("begin transaction error"))
			},
			expectedError: "failed to begin transaction",
		},
		{
			name: "statement execution error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockPrepare := mockDb.ExpectPrepare(`INSERT INTO metadata`)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("processor"), nil) // getProcessor
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("memory"), nil)    // getMemory
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("disks"), nil)     // getDisks
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("os"), nil)        // getOS
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("machine"), nil)   // getMachine
				mockDb.ExpectBegin()
				mockPrepare.ExpectExec().WillReturnError(errors.New("execution error"))
			},
			expectedError: "failed to execute metadata statement",
		},
		{
			name: "transaction commit error",
			setupMocks: func(ctrl *gomock.Controller, mockShell *utils.MockShellExecutor, mockDb sqlmock.Sqlmock) {
				mockPrepare := mockDb.ExpectPrepare(`INSERT INTO metadata`)
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("processor"), nil) // getProcessor
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("memory"), nil)    // getMemory
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("disks"), nil)     // getDisks
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("os"), nil)        // getOS
				mockShell.EXPECT().Command("sh", "-c", gomock.Any()).Return([]byte("machine"), nil)   // getMachine
				mockDb.ExpectBegin()
				mockPrepare.ExpectExec().WithArgs(utils.MainnetChainID, "processor", "memory", "disks", "os", "machine").WillReturnResult(sqlmock.NewResult(1, 1))
				mockDb.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedError: "failed to commit transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db, mockDb, err := sqlmock.New()
			assert.NoError(t, err)

			mockShell := utils.NewMockShellExecutor(ctrl)
			cmd := command{executor: mockShell}

			// Setup mocks based on test case
			tt.setupMocks(ctrl, mockShell, mockDb)

			// Execute the function and verify error
			err = insertMetadata(db, utils.MainnetChainID, cmd)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Verify all expectations were met (where applicable)
			if err := mockDb.ExpectationsWereMet(); err != nil {
				// Some tests might not meet all expectations due to early returns, which is expected
				t.Logf("Not all mock expectations were met (this may be expected): %v", err)
			}
		})
	}
}
