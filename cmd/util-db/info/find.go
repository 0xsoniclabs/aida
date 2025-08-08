package info

import (
	"encoding/binary"
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// findBlockRangeInUpdate finds the first and last block in the update set
func findBlockRangeInUpdate(udb db.UpdateDB) (uint64, uint64, error) {
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

// findBlockRangeInDeleted finds the first and last block in the deleted accounts
func findBlockRangeInDeleted(ddb *db.DestroyedAccountDB) (uint64, uint64, error) {
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

// findBlockRangeInStateHash finds the first and last block of block hashes within given AidaDb
func findBlockRangeInStateHash(db db.BaseDB, log logger.Logger) (uint64, uint64, error) {
	var firstStateHashBlock, lastStateHashBlock uint64
	var err error
	firstStateHashBlock, err = utils.GetFirstStateHash(db)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get first state hash; %w", err)
	}

	lastStateHashBlock, err = utils.GetLastStateHash(db)
	if err != nil {
		log.Infof("Found first state hash at %v", firstStateHashBlock)
		return 0, 0, fmt.Errorf("cannot get last state hash; %w", err)
	}
	return firstStateHashBlock, lastStateHashBlock, nil
}

// findBlockRangeOfBlockHashes finds the first and last block in the block hash
func findBlockRangeOfBlockHashes(db db.BaseDB, log logger.Logger) (uint64, uint64, error) {
	var firstBlock, lastBlock uint64
	var err error

	firstBlock, err = utils.GetFirstBlockHash(db)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get first block hash; %w", err)
	}
	lastBlock, err = utils.GetLastBlockHash(db)
	if err != nil {
		log.Infof("Found first block hash at %v", firstBlock)
		return 0, 0, fmt.Errorf("cannot get last block hash; %w", err)
	}
	return firstBlock, lastBlock, nil
}

// findBlockRangeInException finds the first and last block in the exception
func findBlockRangeInException(edb db.ExceptionDB) (uint64, uint64, error) {
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

// getDeletedCount in given AidaDb
func getDeletedCount(cfg *utils.Config, database db.BaseDB) (int, error) {
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

// getUpdateCount in given AidaDb
func getUpdateCount(cfg *utils.Config, database db.BaseDB) (uint64, error) {
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

// getStateHashCount in given AidaDb
func getStateHashCount(cfg *utils.Config, database db.BaseDB) (uint64, error) {
	var count uint64

	hashProvider := utils.MakeHashProvider(database)
	for i := cfg.First; i <= cfg.Last; i++ {
		_, err := hashProvider.GetStateRootHash(int(i))
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

// getBlockHashCount in given AidaDb
func getBlockHashCount(cfg *utils.Config, database db.BaseDB) (uint64, error) {
	var count uint64

	hashProvider := utils.MakeHashProvider(database)
	for i := cfg.First; i <= cfg.Last; i++ {
		_, err := hashProvider.GetBlockHash(int(i))
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

// getExceptionCount in given AidaDb
func getExceptionCount(cfg *utils.Config, database db.BaseDB) (int, error) {
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

// GetSubstateCount in given AidaDb
func getSubstateCount(cfg *utils.Config, sdb db.SubstateDB) uint64 {
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
