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
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/syndtr/goleveldb/leveldb"
)

// FindBlockRangeInUpdate finds the first and last block in the update set
func FindBlockRangeInUpdate(udb db.UpdateDB) (uint64, uint64, error) {
	firstBlock, err := udb.GetFirstKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get first updateset; %w", err)
	}

	// get last updateset
	lastBlock, err := udb.GetLastKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get last updateset; %w", err)
	}
	return firstBlock, lastBlock, nil
}

// FindBlockRangeInDeleted finds the first and last block in the deleted accounts
func FindBlockRangeInDeleted(ddb db.DestroyedAccountDB) (uint64, uint64, error) {
	firstBlock, err := ddb.GetFirstKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get first deleted accounts; %w", err)
	}

	// get last updateset
	lastBlock, err := ddb.GetLastKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get last deleted accounts; %w", err)
	}
	return firstBlock, lastBlock, nil
}

// FindBlockRangeInStateHash finds the first and last block of block hashes within given AidaDb
func FindBlockRangeInStateHash(shdb db.StateHashDB) (uint64, uint64, error) {
	var firstStateHashBlock, lastStateHashBlock uint64
	var err error
	firstStateHashBlock, err = shdb.GetFirstKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get first state hash; %w", err)
	}

	lastStateHashBlock, err = shdb.GetLastKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get last state hash; %w", err)
	}
	return firstStateHashBlock, lastStateHashBlock, nil
}

// FindBlockRangeOfBlockHashes finds the first and last block in the block hash
func FindBlockRangeOfBlockHashes(bdb db.BlockHashDB) (uint64, uint64, error) {
	firstBlock, err := bdb.GetFirstKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get first blockHash; %w", err)
	}

	// get last blockHash
	lastBlock, err := bdb.GetLastKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get last blockHash; %w", err)
	}
	return firstBlock, lastBlock, nil
}

// FindBlockRangeInException finds the first and last block in the exception
func FindBlockRangeInException(edb db.ExceptionDB) (uint64, uint64, error) {
	firstBlock, err := edb.GetFirstKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get first exception; %w", err)
	}

	// get last exception
	lastBlock, err := edb.GetLastKey()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get last exception; %w", err)
	}
	return firstBlock, lastBlock, nil
}

// GetSubstateCount in given AidaDb
func GetSubstateCount(cfg *utils.Config, sdb db.SubstateDB) uint64 {
	var count uint64

	iter := sdb.NewSubstateIterator(int(cfg.First), 10)
	defer iter.Release()
	for iter.Next() {
		if iter.Value().Block > cfg.Last {
			break
		}
		count++
	}

	return count
}

// GetDeletedCount in given AidaDb
func GetDeletedCount(cfg *utils.Config, database db.BaseDB) (int, error) {
	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, cfg.First)

	iter := database.NewIterator([]byte(db.DestroyedAccountPrefix), startingBlockBytes)
	defer iter.Release()

	count := 0
	for iter.Next() {
		block, _, err := db.DecodeDestroyedAccountKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot Get all destroyed accounts; %w", err)
		}
		if block > cfg.Last {
			break
		}
		count++
	}

	return count, nil
}

// GetUpdateCount in given AidaDb
func GetUpdateCount(cfg *utils.Config, database db.BaseDB) (uint64, error) {
	var count uint64

	start := db.SubstateDBBlockPrefix(cfg.First)[2:]
	iter := database.NewIterator([]byte(db.UpdateDBPrefix), start)
	defer iter.Release()
	for iter.Next() {
		block, err := db.DecodeUpdateSetKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot decode updateset key; %w", err)
		}
		if block > cfg.Last {
			break
		}
		count++
	}

	return count, nil
}

// GetStateHashCount in given AidaDb
func GetStateHashCount(cfg *utils.Config, database db.BaseDB) (uint64, error) {
	var count uint64

	shdb := db.MakeDefaultStateHashDBFromBaseDB(database)
	for i := cfg.First; i <= cfg.Last; i++ {
		_, err := shdb.GetStateHash(int(i))
		if err != nil {
			if errors.Is(err, leveldb.ErrNotFound) {
				continue
			}
			return 0, err
		}
		count++
	}

	return count, nil
}

// GetBlockHashCount in given AidaDb
func GetBlockHashCount(cfg *utils.Config, database db.BaseDB) (uint64, error) {
	var count uint64

	bhdb := db.MakeDefaultBlockHashDBFromBaseDB(database)
	for i := cfg.First; i <= cfg.Last; i++ {
		_, err := bhdb.GetBlockHash(int(i))
		if err != nil {
			if errors.Is(err, leveldb.ErrNotFound) {
				continue
			}
			return 0, err
		}
		count++
	}

	return count, nil
}

// GetExceptionCount in given AidaDb
func GetExceptionCount(cfg *utils.Config, database db.BaseDB) (int, error) {
	startingBlockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startingBlockBytes, cfg.First)

	iter := database.NewIterator([]byte(db.ExceptionDBPrefix), startingBlockBytes)
	defer iter.Release()

	count := 0
	for iter.Next() {
		block, err := db.DecodeExceptionDBKey(iter.Key())
		if err != nil {
			return 0, fmt.Errorf("cannot get exception count; %w", err)
		}
		if block > cfg.Last {
			break
		}
		count++
	}

	return count, nil
}
