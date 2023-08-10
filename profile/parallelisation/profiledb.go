package parallelisation

import (
	"database/sql"
	"fmt"

	// Your main or test packages require this import so the sql package is properly initialized.
	_ "github.com/mattn/go-sqlite3"
)

const (
	// bufferSize is the buffer size of the in-memory buffer for storing profile data
	bufferSize = 1000

	// SQL for inserting new block
	insertBlockSQL = `
INSERT INTO parallelprofile (
	block, tBlock, tSequential, tCritical, tCommit, speedup, ubNumProc, numTx
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?
)
`
	// SQL for inserting new transaction
	insertTxSQL = `
INSERT INTO txProfile (
block, tx, duration
) VALUES (
?, ?, ?
)
`

	// SQL for creating a new profiling table
	createSQL = `
	PRAGMA journal_mode = MEMORY;
	CREATE TABLE IF NOT EXISTS parallelprofile (
    block INTEGER,
	tBlock INTEGER,
	tSequential INTEGER,
	tCritical INTEGER,
	tCommit INTEGER,
	speedup FLOAT,
	ubNumProc INTEGER,
	numTx INTEGER);
	CREATE TABLE IF NOT EXISTS txProfile (
    block INTEGER,
	tx    INTEGER, 
	duration INTEGER
);
`
)

// ProfileDB is a database of ProfileData
type ProfileDB struct {
	sql       *sql.DB       // Sqlite3 database
	blockStmt *sql.Stmt     // Prepared insert statement for a block
	txStmt    *sql.Stmt     // Prepared insert statement for a transaction
	buffer    []ProfileData // record buffer
}

// NewProfileDB constructs a ProfileDatas value for managing stock ProfileDatas in a
// SQLite database. This API is not thread safe.
func NewProfileDB(dbFile string) (*ProfileDB, error) {
	// open SQLITE3 DB
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	// create profile schema if not exists
	if _, err = sqlDB.Exec(createSQL); err != nil {
		return nil, fmt.Errorf("sqlDB.Exec, err: %q", err)
	}
	// prepare the INSERT statement for subsequent use
	blockStmt, err := sqlDB.Prepare(insertBlockSQL)
	if err != nil {
		return nil, err
	}

	txStmt, err := sqlDB.Prepare(insertTxSQL)
	if err != nil {
		return nil, err
	}
	db := ProfileDB{
		sql:       sqlDB,
		blockStmt: blockStmt,
		txStmt:    txStmt,
		buffer:    make([]ProfileData, 0, bufferSize),
	}
	return &db, nil
}

// Close flushes all ProfileDatas to the database and prevents any future trading.
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

// Add stores a profile data record into a buffer. Once the buffer is full, the
// records are flushed into the database.
func (db *ProfileDB) Add(ProfileData ProfileData) error {
	db.buffer = append(db.buffer, ProfileData)
	if len(db.buffer) == cap(db.buffer) {
		if err := db.Flush(); err != nil {
			return fmt.Errorf("unable to flush ProfileDatas: %w", err)
		}
	}
	return nil
}

// Flush inserts pending ProfileDatas into the database inside DB transaction.
func (db *ProfileDB) Flush() error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	for _, ProfileData := range db.buffer {
		_, err := tx.Stmt(db.blockStmt).Exec(ProfileData.curBlock, ProfileData.tBlock, ProfileData.tSequential, ProfileData.tCritical,
			ProfileData.tCommit, ProfileData.speedup, ProfileData.ubNumProc, ProfileData.numTx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		// write into new txProfile table here the transaction durations
		for i, tTransaction := range ProfileData.tTransactions {
			_, err = tx.Stmt(db.txStmt).Exec(ProfileData.curBlock, i, tTransaction)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}
	db.buffer = db.buffer[:0]
	return tx.Commit()
}

// DeleteByBlockRange deletes rows in a given block range

func (db *ProfileDB) DeleteByBlockRange(firstBlock, lastBlock uint64) (int64, error) {
	const (
		parallelProfile = "parallelprofile"
		txProfile       = "txProfile"
	)
	var totalNumRows int64

	tx, err := db.sql.Begin()
	if err != nil {
		return 0, err
	}

	for _, table := range []string{parallelProfile, txProfile} {
		deleteSql := fmt.Sprintf("DELETE FROM %s WHERE block >= %d AND block <= %d;", table, firstBlock, lastBlock)
		res, err := db.sql.Exec(deleteSql)
		if err != nil {
			panic(err.Error())
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
