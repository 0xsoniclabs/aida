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

package blockprofile

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"

	"github.com/0xsoniclabs/aida/utils"
	// Your main or test packages require this import so the sql package is properly initialized.
	_ "github.com/mattn/go-sqlite3"
)

const (
	// bufferSize of the in-memory buffer for storing profile data
	bufferSize = 1000

	// SQL statement for inserting a profile record of a new block
	insertBlockSQL = `
INSERT INTO blockProfile (
	block, tBlock, tSequential, tCritical, tCommit, speedup, ubNumProc, numTx, gasBlock
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?
)
`
	// SQL statement for inserting a profile record of a new transaction
	insertTxSQL = `
INSERT INTO txProfile (
block, tx, txType, duration, gas
) VALUES (
?, ?, ?, ?, ?
)
`

	// SQL statement for inserting metadata of the profiling run
	insertMetadataSQL = `
INSERT INTO metadata (
    chainid, processor, memory, disks, os, machine
) VALUES (
    ?, ?, ?, ?, ?, ?
)
`

	// SQL statement for creating profiling tables
	createSQL = `
PRAGMA journal_mode = MEMORY;
CREATE TABLE IF NOT EXISTS blockProfile (
	block INTEGER,
	tBlock INTEGER,
	tSequential INTEGER,
	tCritical INTEGER,
	tCommit INTEGER,
	speedup FLOAT,
	ubNumProc INTEGER,
	numTx INTEGER,
	gasBlock INTEGER
);
CREATE TABLE IF NOT EXISTS txProfile (
	block INTEGER,
	tx    INTEGER, 
	txType INTEGER,
	duration INTEGER,
	gas INTEGER
);
CREATE TABLE IF NOT EXISTS metadata (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    chainid INTEGER,
    processor TEXT,
    memory TEXT,
    disks TEXT,
    os TEXT,
    machine TEXT
);
`
)

// ProfileDB is a profiling database for block processing.
type ProfileDB struct {
	sql       *sql.DB       // Sqlite3 database
	blockStmt *sql.Stmt     // Prepared insert statement for a block
	txStmt    *sql.Stmt     // Prepared insert statement for a transaction
	buffer    []ProfileData // record buffer
}

// NewProfileDB constructs a new profiling database.
func NewProfileDB(dbFile string, chainID utils.ChainID) (*ProfileDB, error) {
	// open SQLITE3 DB
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %v; %v", dbFile, err)
	}
	// create profile schema if not exists
	if _, err = sqlDB.Exec(createSQL); err != nil {
		return nil, fmt.Errorf("sqlDB.Exec, err: %q", err)
	}
	// prepare INSERT statements for subsequent use
	blockStmt, err := sqlDB.Prepare(insertBlockSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare a SQL statement for block profile; %v", err)
	}
	txStmt, err := sqlDB.Prepare(insertTxSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare a SQL statement for tx profile; %v", err)
	}

	err = insertMetadata(sqlDB, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert metadata; %v", err)
	}

	return &ProfileDB{
		sql:       sqlDB,
		blockStmt: blockStmt,
		txStmt:    txStmt,
		buffer:    make([]ProfileData, 0, bufferSize),
	}, nil
}

// Close flushes buffers of profiling database and closes the profiling database.
func (db *ProfileDB) Close() error {
	defer func() {
		db.txStmt.Close()
		db.blockStmt.Close()
		db.sql.Close()
	}()
	if err := db.Flush(); err != nil {
		return err
	}
	return nil
}

// Add a profile data record to the profiling database.
func (db *ProfileDB) Add(ProfileData ProfileData) error {
	db.buffer = append(db.buffer, ProfileData)
	if len(db.buffer) == cap(db.buffer) {
		if err := db.Flush(); err != nil {
			return fmt.Errorf("unable to flush ProfileDatas: %w", err)
		}
	}
	return nil
}

// Flush the profiling records in the database.
func (db *ProfileDB) Flush() error {
	// open new transaction
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	// write profiling records into sqlite3 database
	for _, ProfileData := range db.buffer {
		// write block data
		_, err := tx.Stmt(db.blockStmt).Exec(ProfileData.curBlock, ProfileData.tBlock, ProfileData.tSequential, ProfileData.tCritical,
			ProfileData.tCommit, ProfileData.speedup, ProfileData.ubNumProc, ProfileData.numTx, ProfileData.gasBlock)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		// write transactions
		for i, tTransaction := range ProfileData.tTransactions {
			_, err = tx.Stmt(db.txStmt).Exec(ProfileData.curBlock, i, ProfileData.tTypes[i], tTransaction, ProfileData.gasTransactions[i])
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}
	// clear buffer
	db.buffer = db.buffer[:0]
	// commit transaction
	return tx.Commit()
}

// DeleteByBlockRange deletes information for a block range; used prior insertion
func (db *ProfileDB) DeleteByBlockRange(firstBlock, lastBlock uint64) (int64, error) {
	const (
		blockProfile = "blockProfile"
		txProfile    = "txProfile"
	)
	var totalNumRows int64

	tx, err := db.sql.Begin()
	if err != nil {
		return 0, err
	}

	for _, table := range []string{blockProfile, txProfile} {
		deleteSql := fmt.Sprintf("DELETE FROM %s WHERE block >= %d AND block <= %d;", table, firstBlock, lastBlock)
		res, err := db.sql.Exec(deleteSql)
		if err != nil {
			return 0, err
		}

		numRowsAffected, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}

		totalNumRows += numRowsAffected
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return totalNumRows, nil
}

// insertMetadata inserts metadata of the profiling run
func insertMetadata(sqlDB *sql.DB, chainID utils.ChainID) error {
	metadataStmt, err := sqlDB.Prepare(insertMetadataSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare a SQL statement for metadata; %v", err)
	}

	processor, err := getProcessor()
	if err != nil {
		return fmt.Errorf("failed to get processor information; %v", err)
	}
	memory, err := getMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory information; %v", err)
	}
	disks, err := getDisks()
	if err != nil {
		return fmt.Errorf("failed to get disk information; %v", err)
	}
	os, err := getOS()
	if err != nil {
		return fmt.Errorf("failed to get OS information; %v", err)
	}

	machine, err := getMachine()
	if err != nil {
		return fmt.Errorf("failed to get machine information; %v", err)
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction; %v", err)
	}
	_, err = tx.Stmt(metadataStmt).Exec(chainID, processor, memory, disks, os, machine)
	if err != nil {
		return fmt.Errorf("failed to execute metadata statement; %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction; %v", err)
	}
	return nil
}

func getProcessor() (string, error) {
	cmd := exec.Command("sh", "-c", `cat /proc/cpuinfo | grep "^model name" | head -n 1 | awk -F': ' '{print $2}'`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getMemory() (string, error) {
	cmd := exec.Command("sh", "-c", `free | grep "^Mem:" | awk '{printf("%dGB RAM\n", $2/1024/1024)}'`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getDisks() (string, error) {
	cmd := exec.Command("sh", "-c", `hwinfo --disk | grep Model | awk -F ': \"' '{if (NR > 1) printf(", "); printf("%s", substr($2,1,length($2)-1));}  END {printf("\n")}'`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// check if output contains `hwinfo: not found`
	if strings.Contains(string(output), "hwinfo: not found") {
		return "", fmt.Errorf(string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func getOS() (string, error) {
	cmd := exec.Command("sh", "-c", `lsb_release -d | awk -F"\t" '{print $2}'`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getMachine() (string, error) {
	cmd := exec.Command("sh", "-c", "echo \"`hostname`(`curl -s api.ipify.org`)\"")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
