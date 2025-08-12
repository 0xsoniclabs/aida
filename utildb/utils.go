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

package utildb

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

// OpenSourceDatabases opens all databases required for merge
func OpenSourceDatabases(sourceDbPaths []string) ([]db.BaseDB, error) {
	if len(sourceDbPaths) < 1 {
		return nil, fmt.Errorf("no source database were specified")
	}

	var sourceDbs []db.BaseDB
	for i := 0; i < len(sourceDbPaths); i++ {
		path := sourceDbPaths[i]
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source database %s; doesn't exist", path)
		}
		db, err := db.NewReadOnlyBaseDB(path)
		if err != nil {
			return nil, fmt.Errorf("source database %s; error: %v", path, err)
		}
		sourceDbs = append(sourceDbs, db)
	}
	return sourceDbs, nil
}

// MustCloseDB close database safely
func MustCloseDB(db db.BaseDB) {
	if db != nil {
		err := db.Close()
		if err != nil {
			if err.Error() != "leveldb: closed" {
				fmt.Printf("could not close database; %s\n", err.Error())
			}
		}
	}
}

// calculateMD5Sum calculates MD5 hash of given file
func calculateMD5Sum(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable open file %s; %v", filePath, err.Error())
	}
	defer file.Close()

	// Create a new MD5 hash instance
	hash := md5.New()

	// Copy the file content into the hash instance
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", fmt.Errorf("unable to calculate md5; %v", err)
	}

	// Calculate the MD5 checksum as a byte slice
	checksum := hash.Sum(nil)

	// Convert the checksum to a hexadecimal string
	md5sum := hex.EncodeToString(checksum)

	return md5sum, nil
}

// GetDbSize retrieves database size
func GetDbSize(db db.BaseDB) uint64 {
	var count uint64
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		count++
	}
	return count
}

// PrintMetadata from given AidaDb
func PrintMetadata(pathToDb string) error {
	log := logger.NewLogger("INFO", "Print-Metadata")
	base, err := db.NewReadOnlyBaseDB(pathToDb)
	if err != nil {
		return err
	}

	md := utils.NewAidaDbMetadata(base, "INFO")

	log.Notice("AIDA-DB INFO:")

	if err = printDbType(md); err != nil {
		return err
	}

	lastBlock := md.GetLastBlock()

	firstBlock := md.GetFirstBlock()

	// CHAIN-ID
	chainID := md.GetChainID()

	if firstBlock == 0 && lastBlock == 0 && chainID == 0 {
		log.Error("your db does not contain metadata; please use metadata generate command")
	} else {
		log.Infof("Chain-ID: %v", chainID)

		// BLOCKS
		log.Infof("First Block: %v", firstBlock)

		log.Infof("Last Block: %v", lastBlock)

		// EPOCHS
		firstEpoch := md.GetFirstEpoch()

		log.Infof("First Epoch: %v", firstEpoch)

		lastEpoch := md.GetLastEpoch()

		log.Infof("Last Epoch: %v", lastEpoch)

		dbHash := md.GetDbHash()

		log.Infof("Db Hash: %v", hex.EncodeToString(dbHash))

		// TIMESTAMP
		timestamp := md.GetTimestamp()

		log.Infof("Created: %v", time.Unix(int64(timestamp), 0))
	}

	// UPDATE-SET
	printUpdateSetInfo(md)

	return nil
}

// printUpdateSetInfo from given AidaDb
func printUpdateSetInfo(m *utils.AidaDbMetadata) {
	log := logger.NewLogger("INFO", "Print-Metadata")

	log.Notice("UPDATE-SET INFO:")

	intervalBytes, err := m.Db.Get([]byte(db.UpdatesetIntervalKey))
	if err != nil {
		log.Warning("Value for update-set interval does not exist in given Dbs metadata")
	} else {
		log.Infof("Interval: %v blocks", bigendian.BytesToUint64(intervalBytes))
	}

	sizeBytes, err := m.Db.Get([]byte(db.UpdatesetSizeKey))
	if err != nil {
		log.Warning("Value for update-set size does not exist in given Dbs metadata")
	} else {
		u := bigendian.BytesToUint64(sizeBytes)

		log.Infof("Size: %.1f MB", float64(u)/float64(1_000_000))
	}
}

// printDbType from given AidaDb
func printDbType(m *utils.AidaDbMetadata) error {
	t := m.GetDbType()

	var typePrint string
	switch t {
	case utils.GenType:
		typePrint = "Generate"
	case utils.CloneType:
		typePrint = "Clone"
	case utils.PatchType:
		typePrint = "Patch"
	case utils.NoType:
		typePrint = "NoType"

	default:
		return errors.New("unknown db type")
	}

	logger.NewLogger("INFO", "Print-Metadata").Noticef("DB-Type: %v", typePrint)

	return nil
}

func GenerateTestAidaDb(t *testing.T) db.BaseDB {
	tmpDir := t.TempDir() + "/testAidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	md := utils.NewAidaDbMetadata(database, "ERROR")
	err = md.SetAllMetadata(1, 50, 1, 50, 250, []byte("0x0"), 1)
	assert.NoError(t, err)

	// write substates to the database
	substateDb := db.MakeDefaultSubstateDBFromBaseDB(database)
	state := substate.Substate{
		Block:       10,
		Transaction: 7,
		Env: &substate.Env{
			Number:     11,
			Difficulty: big.NewInt(1),
			GasLimit:   uint64(15),
		},
		Message: &substate.Message{
			Value:    big.NewInt(12),
			GasPrice: big.NewInt(14),
		},
		InputSubstate:  substate.WorldState{},
		OutputSubstate: substate.WorldState{},
		Result:         &substate.Result{},
	}

	for i := 0; i < 10; i++ {
		state.Block = uint64(10 + i)
		err = substateDb.PutSubstate(&state)
		require.NoError(t, err)
	}

	udb := db.MakeDefaultUpdateDBFromBaseDB(database)
	// write update sets to the database
	for i := 1; i <= 10; i++ {
		updateSet := &updateset.UpdateSet{
			WorldState: substate.WorldState{
				types.Address{1}: &substate.Account{
					Nonce:   1,
					Balance: new(uint256.Int).SetUint64(1),
					Code:    []byte{0x01, 0x02},
				},
			},
			Block: uint64(i),
		}
		err = udb.PutUpdateSet(updateSet, []types.Address{})
		require.NoError(t, err)
	}

	// write delete accounts to the database
	for i := 1; i <= 10; i++ {
		err = database.Put(db.EncodeDestroyedAccountKey(uint64(i), i), []byte("0x1234567812345678123456781234567812345678123456781234567812345678"))
		require.NoError(t, err)
	}

	// write state hashes to the database
	for i := 11; i <= 20; i++ {
		key := "0x" + strconv.FormatInt(int64(i), 16)
		err = utils.SaveStateRoot(database, key, "0x1234567812345678123456781234567812345678123456781234567812345678")
		require.NoError(t, err)
	}

	// write block hashes to the database
	for i := 21; i <= 30; i++ {
		key := "0x" + strconv.FormatInt(int64(i), 16)
		err = utils.SaveBlockHash(database, key, "0x1234567812345678123456781234567812345678123456781234567812345678")
		require.NoError(t, err)
	}

	// write exceptions to the database
	for i := 31; i <= 40; i++ {
		exception := &substate.Exception{
			Block: uint64(i),
			Data: substate.ExceptionBlock{
				PreBlock:  &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 1, Balance: uint256.NewInt(100)}},
				PostBlock: &substate.WorldState{types.Address{0x02}: &substate.Account{Nonce: 2, Balance: uint256.NewInt(200)}},
			},
		}
		eDb := db.MakeDefaultExceptionDBFromBaseDB(database)
		err = eDb.PutException(exception)
		require.NoError(t, err)
	}

	return database
}
