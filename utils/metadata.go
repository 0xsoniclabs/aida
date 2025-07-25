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

package utils

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	geth_leveldb "github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/syndtr/goleveldb/leveldb"
)

type AidaDbType byte

const (
	NoType AidaDbType = iota
	GenType
	PatchType
	CloneType
	CustomType
)

const (
	FirstBlockPrefix        = db.MetadataPrefix + "fb"
	LastBlockPrefix         = db.MetadataPrefix + "lb"
	FirstEpochPrefix        = db.MetadataPrefix + "fe"
	LastEpochPrefix         = db.MetadataPrefix + "le"
	TypePrefix              = db.MetadataPrefix + "ty"
	ChainIDPrefix           = db.MetadataPrefix + "ci"
	TimestampPrefix         = db.MetadataPrefix + "ti"
	DbHashPrefix            = db.MetadataPrefix + "md"
	HasStateHashPatchPrefix = db.MetadataPrefix + "sh"
)

// merge is determined by what are we merging
// genType + CloneType / CloneType + CloneType / = NOT POSSIBLE
// genType + genType = genType
// genType + PatchType = genType
// CloneType + PatchType = CloneType
// PatchType + PatchType = PatchType

// PatchJson represents struct of JSON file where information about patches is written
type PatchJson struct {
	FileName           string
	FromBlock, ToBlock uint64
	FromEpoch, ToEpoch uint64
	DbHash, TarHash    string
	Nightly            bool
}

// AidaDbMetadata holds any information about AidaDb needed for putting it into the Db
type AidaDbMetadata struct {
	Db                    db.BaseDB
	log                   logger.Logger
	FirstBlock, LastBlock uint64
	FirstEpoch, LastEpoch uint64
	ChainId               ChainID
	DbType                AidaDbType
	timestamp             uint64
}

// todo we need to check block alignment and chainID match before any merging

// NewAidaDbMetadata creates new instance of AidaDbMetadata
func NewAidaDbMetadata(db db.BaseDB, logLevel string) *AidaDbMetadata {
	return &AidaDbMetadata{
		Db:  db,
		log: logger.NewLogger(logLevel, "aida-metadata"),
	}
}

// ProcessPatchLikeMetadata decides whether patch is new or not. If so the DbType is Set to GenType, otherwise its PatchType.
// Then it inserts all given metadata
func ProcessPatchLikeMetadata(aidaDb db.BaseDB, logLevel string, firstBlock, lastBlock, firstEpoch, lastEpoch uint64, chainID ChainID, isNew bool, dbHash []byte) error {
	var (
		dbType AidaDbType
		err    error
	)

	// if this is brand-new patch, it should be treated as a gen type db
	if isNew {
		dbType = GenType
	} else {
		dbType = PatchType
	}

	md := NewAidaDbMetadata(aidaDb, logLevel)

	if err = md.SetFirstBlock(firstBlock); err != nil {
		return err
	}
	if err = md.SetLastBlock(lastBlock); err != nil {
		return err
	}

	if err = md.SetFirstEpoch(firstEpoch); err != nil {
		return err
	}
	if err = md.SetLastEpoch(lastEpoch); err != nil {
		return err
	}

	if err = md.SetChainID(chainID); err != nil {
		return err
	}

	if err = md.SetDbType(dbType); err != nil {
		return err
	}

	if err = md.SetTimestamp(); err != nil {
		return err
	}

	err = md.SetDbHash(dbHash)
	if err != nil {
		return err
	}

	md.log.Notice("Metadata added successfully")

	return nil
}

// ProcessCloneLikeMetadata inserts every metadata from sourceDb, only epochs are excluded.
// We can't be certain if given epoch is whole
func ProcessCloneLikeMetadata(aidaDb db.BaseDB, typ AidaDbType, logLevel string, firstBlock, lastBlock uint64, chainID ChainID) error {
	var err error

	md := NewAidaDbMetadata(aidaDb, logLevel)

	firstBlock, lastBlock = md.compareBlocks(firstBlock, lastBlock)

	if err = md.SetFirstBlock(firstBlock); err != nil {
		return err
	}
	if err = md.SetLastBlock(lastBlock); err != nil {
		return err
	}

	if err = md.SetChainID(chainID); err != nil {
		return err
	}

	if err = md.findEpochs(); err != nil {
		return err
	}

	if err = md.SetFirstEpoch(md.FirstEpoch); err != nil {
		return err
	}

	if err = md.SetLastEpoch(md.LastEpoch); err != nil {
		return err
	}

	if err = md.SetDbType(typ); err != nil {
		return err
	}

	if err = md.SetTimestamp(); err != nil {
		return err
	}

	md.log.Notice("Metadata added successfully")
	return nil
}

func ProcessGenLikeMetadata(aidaDb db.BaseDB, firstBlock uint64, lastBlock uint64, firstEpoch uint64, lastEpoch uint64, chainID ChainID, logLevel string, dbHash []byte) error {
	md := NewAidaDbMetadata(aidaDb, logLevel)
	return md.genMetadata(firstBlock, lastBlock, firstEpoch, lastEpoch, chainID, dbHash)
}

// genMetadata inserts metadata into newly generated AidaDb.
// If generate is used onto an existing AidaDb it updates last block, last epoch and timestamp.
func (md *AidaDbMetadata) genMetadata(firstBlock uint64, lastBlock uint64, firstEpoch uint64, lastEpoch uint64, chainID ChainID, dbHash []byte) error {
	var err error

	firstBlock, lastBlock = md.compareBlocks(firstBlock, lastBlock)

	if err = md.SetFirstBlock(firstBlock); err != nil {
		return err
	}
	if err = md.SetLastBlock(lastBlock); err != nil {
		return err
	}

	firstEpoch, lastEpoch = md.compareEpochs(firstEpoch, lastEpoch)

	if err = md.SetFirstEpoch(firstEpoch); err != nil {
		return err
	}
	if err = md.SetLastEpoch(lastEpoch); err != nil {
		return err
	}

	if err = md.SetChainID(chainID); err != nil {
		return err
	}

	if err = md.SetDbType(GenType); err != nil {
		return err
	}

	if err = md.SetTimestamp(); err != nil {
		return err
	}

	if err = md.SetDbHash(dbHash); err != nil {
		return err
	}

	return nil
}

// ProcessMergeMetadata decides the type according to the types of merged Dbs and inserts every metadata
func ProcessMergeMetadata(cfg *Config, aidaDb db.BaseDB, sourceDbs []db.BaseDB, paths []string) (*AidaDbMetadata, error) {
	var (
		err error
		ok  bool
	)

	targetMD := NewAidaDbMetadata(aidaDb, cfg.LogLevel)

	for i, database := range sourceDbs {
		md := NewAidaDbMetadata(database, cfg.LogLevel)
		md.GetMetadata()

		// todo do we need to check whether blocks align?

		// Get chainID of first source database
		if targetMD.ChainId == 0 {
			targetMD.ChainId = md.ChainId
		}

		// if chain ids doesn't match, we should not be merging
		if md.ChainId != targetMD.ChainId {
			md.log.Critical("ChainIDs in Dbs metadata does not match!")
		}

		hasNoBlockRangeInMetadata := md.FirstBlock == 0 && md.LastBlock == 0

		// if database had no metadata we will look for blocks in substate
		if hasNoBlockRangeInMetadata {
			// we need to close database before opening substate
			if err = database.Close(); err != nil {
				return nil, fmt.Errorf("cannot close database; %v", err)
			}

			sdb := db.MakeDefaultSubstateDBFromBaseDB(database)
			err = sdb.SetSubstateEncoding(cfg.SubstateEncoding)
			if err != nil {
				return nil, err
			}
			md.FirstBlock, md.LastBlock, ok = FindBlockRangeInSubstate(sdb)
			if !ok {
				md.log.Warningf("Cannot find blocks in substate; is substate present in given database? %v", paths[i])
			} else {
				md.log.Noticef("Found block range inside substate of %v (%v-%v)", paths[i], md.FirstBlock, md.LastBlock)
			}
		} else {
			ok = true
		}

		// only check blocks when merged database has metadata or substate
		if ok {
			if md.FirstBlock < targetMD.FirstBlock || hasNoBlockRangeInMetadata {
				targetMD.FirstBlock = md.FirstBlock
			}

			if md.LastBlock > targetMD.LastBlock {
				targetMD.LastBlock = md.LastBlock
			}
		}

		// set first
		if targetMD.DbType == NoType {
			targetMD.DbType = md.DbType
			continue
		}

		if targetMD.DbType == GenType && (md.DbType == PatchType || md.DbType == GenType) {
			targetMD.DbType = GenType
			continue
		}

		if targetMD.DbType == PatchType {
			switch md.DbType {
			case GenType:
				targetMD.DbType = GenType
				continue
			case PatchType:
				targetMD.DbType = PatchType
				continue
			case CloneType:
				targetMD.DbType = CloneType
				// we cannot merge patch with smaller first block onto clone because... todo explain + error
				if targetMD.FirstBlock < md.FirstBlock {
					return nil, errors.New("cannot prepend patch on clone")
				}
				continue
			default:
				targetMD.DbType = GenType
			}
		}

		if targetMD.DbType == CloneType && md.DbType == PatchType {
			targetMD.DbType = CloneType
			// we cannot merge patch with smaller first block onto clone because... todo explain + error
			if md.FirstBlock < targetMD.FirstBlock {
				return nil, errors.New("cannot prepend patch on clone")
			}
			continue
		}

		return nil, fmt.Errorf("cannot merge %v with %v", targetMD.getVerboseDbType(), md.getVerboseDbType())
	}

	// if source dbs had neither metadata nor substate, we try to find the block range inside substate of targetDb
	if targetMD.FirstBlock == 0 && targetMD.LastBlock == 0 {
		// we must close database before accessing substate
		if err = targetMD.Db.Close(); err != nil {
			return nil, fmt.Errorf("cannot close targetDb; %v", err)
		}
		sdb := db.MakeDefaultSubstateDBFromBaseDB(targetMD.Db)
		err = sdb.SetSubstateEncoding(cfg.SubstateEncoding)
		if err != nil {
			return nil, err
		}
		targetMD.FirstBlock, targetMD.LastBlock, ok = FindBlockRangeInSubstate(sdb)
		if !ok {
			targetMD.log.Warningf("Cannot find block range in substate of AidaDb (%v); this will in corrupted metadata but will not affect data itself", cfg.AidaDb)
		} else {
			targetMD.log.Noticef("Found block range inside substate of AidaDb %v (%v-%v)", cfg.AidaDb, targetMD.FirstBlock, targetMD.LastBlock)
		}
	}

	if targetMD.ChainId == 0 {
		targetMD.log.Warningf("your dbs does not have chain-id, Setting value from config (%v)", cfg.ChainID)
		targetMD.ChainId = cfg.ChainID
	}

	if err = targetMD.findEpochs(); err != nil {
		return nil, err
	}

	return targetMD, nil
}

// GetMetadata from given Db and save it
func (md *AidaDbMetadata) GetMetadata() {
	md.FirstBlock = md.GetFirstBlock()

	md.LastBlock = md.GetLastBlock()

	md.FirstEpoch = md.GetFirstEpoch()

	md.LastEpoch = md.GetLastEpoch()

	md.DbType = md.GetDbType()

	md.timestamp = md.GetTimestamp()

	md.ChainId = md.GetChainID()
}

// compareBlocks from given Db and return them
func (md *AidaDbMetadata) compareBlocks(firstBlock uint64, lastBlock uint64) (uint64, uint64) {
	var (
		dbFirst, dbLast uint64
	)

	dbFirst = md.GetFirstBlock()
	if (dbFirst != 0 && dbFirst < firstBlock) || firstBlock == 0 {
		firstBlock = dbFirst
	}

	dbLast = md.GetLastBlock()

	if dbLast > lastBlock || lastBlock == 0 {
		lastBlock = dbLast
	}

	return firstBlock, lastBlock
}

// compareEpochs from given Db and return them
func (md *AidaDbMetadata) compareEpochs(firstEpoch uint64, lastEpoch uint64) (uint64, uint64) {
	var (
		dbFirst, dbLast uint64
	)

	dbFirst = md.GetFirstEpoch()
	if (dbFirst != 0 && dbFirst < firstEpoch) || firstEpoch == 0 {
		firstEpoch = dbFirst
	}

	dbLast = md.GetLastEpoch()
	if dbLast > lastEpoch || lastEpoch == 0 {
		lastEpoch = dbLast
	}

	return firstEpoch, lastEpoch
}

// GetFirstBlock and return it
func (md *AidaDbMetadata) GetFirstBlock() uint64 {
	firstBlockBytes, err := md.Db.Get([]byte(FirstBlockPrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0
		}
		md.log.Criticalf("cannot get first block from metadata; %v", err)
		return 0
	}

	return bigendian.BytesToUint64(firstBlockBytes)
}

// GetLastBlock and return it
func (md *AidaDbMetadata) GetLastBlock() uint64 {
	lastBlockBytes, err := md.Db.Get([]byte(LastBlockPrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0
		}
		md.log.Criticalf("cannot get last block from metadata; %v", err)
		return 0
	}

	return bigendian.BytesToUint64(lastBlockBytes)
}

// GetFirstEpoch and return it
func (md *AidaDbMetadata) GetFirstEpoch() uint64 {
	firstEpochBytes, err := md.Db.Get([]byte(FirstEpochPrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0
		}
		md.log.Criticalf("cannot get first epoch from metadata; %v", err)
		return 0
	}

	return bigendian.BytesToUint64(firstEpochBytes)
}

// GetLastEpoch and return it
func (md *AidaDbMetadata) GetLastEpoch() uint64 {
	lastEpochBytes, err := md.Db.Get([]byte(LastEpochPrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0
		}
		md.log.Criticalf("cannot get last epoch from metadata; %v", err)
		return 0
	}

	return bigendian.BytesToUint64(lastEpochBytes)
}

// GetChainID and return it
func (md *AidaDbMetadata) GetChainID() ChainID {
	chainIDBytes, err := md.Db.Get([]byte(ChainIDPrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0
		}
		md.log.Criticalf("cannot get chain id from metadata; %v", err)
		return 0
	}

	// chainID used to be 2 bytes long, now it is 8 bytes long
	if len(chainIDBytes) == 2 {
		return ChainID(bigendian.BytesToUint16(chainIDBytes))
	}

	return ChainID(bigendian.BytesToUint64(chainIDBytes))
}

// GetTimestamp and return it
func (md *AidaDbMetadata) GetTimestamp() uint64 {
	byteTimestamp, err := md.Db.Get([]byte(TimestampPrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return 0
		}
		md.log.Criticalf("cannot get timestamp from metadata; %v", err)
		return 0
	}

	return bigendian.BytesToUint64(byteTimestamp)
}

// GetDbType and return it
func (md *AidaDbMetadata) GetDbType() AidaDbType {
	byteDbType, err := md.Db.Get([]byte(TypePrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return NoType
		}
		md.log.Criticalf("cannot get db type from metadata; %v", err)
		return 0
	}

	return AidaDbType(byteDbType[0])
}

// SetFirstBlock in given Db
func (md *AidaDbMetadata) SetFirstBlock(firstBlock uint64) error {
	firstBlockBytes := db.BlockToBytes(firstBlock)

	if err := md.Db.Put([]byte(FirstBlockPrefix), firstBlockBytes); err != nil {
		return fmt.Errorf("cannot put first block; %v", err)
	}

	md.FirstBlock = firstBlock

	md.log.Info("METADATA: First block saved successfully")

	return nil
}

// SetLastBlock in given Db
func (md *AidaDbMetadata) SetLastBlock(lastBlock uint64) error {
	lastBlockBytes := db.BlockToBytes(lastBlock)

	if err := md.Db.Put([]byte(LastBlockPrefix), lastBlockBytes); err != nil {
		return fmt.Errorf("cannot put last block; %v", err)
	}

	md.LastBlock = lastBlock

	md.log.Info("METADATA: Last block saved successfully")

	return nil
}

// SetFirstEpoch in given Db
func (md *AidaDbMetadata) SetFirstEpoch(firstEpoch uint64) error {
	firstEpochBytes := db.BlockToBytes(firstEpoch)

	if err := md.Db.Put([]byte(FirstEpochPrefix), firstEpochBytes); err != nil {
		return fmt.Errorf("cannot put first epoch; %v", err)
	}

	md.log.Info("METADATA: First epoch saved successfully")

	return nil
}

// SetLastEpoch in given Db
func (md *AidaDbMetadata) SetLastEpoch(lastEpoch uint64) error {
	lastEpochBytes := db.BlockToBytes(lastEpoch)

	if err := md.Db.Put([]byte(LastEpochPrefix), lastEpochBytes); err != nil {
		return fmt.Errorf("cannot put last epoch; %v", err)
	}

	md.log.Info("METADATA: Last epoch saved successfully")

	return nil
}

// SetChainID in given Db
func (md *AidaDbMetadata) SetChainID(chainID ChainID) error {
	chainIDBytes := bigendian.Uint64ToBytes(uint64(chainID))

	if err := md.Db.Put([]byte(ChainIDPrefix), chainIDBytes); err != nil {
		return fmt.Errorf("cannot put chain-id; %v", err)
	}

	md.ChainId = chainID

	md.log.Info("METADATA: ChainID saved successfully")

	return nil
}

// SetTimestamp in given Db
func (md *AidaDbMetadata) SetTimestamp() error {
	createTime := make([]byte, 8)

	binary.BigEndian.PutUint64(createTime, uint64(time.Now().Unix()))
	if err := md.Db.Put([]byte(TimestampPrefix), createTime); err != nil {
		return fmt.Errorf("cannot put timestamp into db metadata; %v", err)
	}

	md.log.Info("METADATA: Creation timestamp saved successfully")

	return nil
}

// SetDbType in given Db
func (md *AidaDbMetadata) SetDbType(dbType AidaDbType) error {
	dbTypeBytes := make([]byte, 1)
	dbTypeBytes[0] = byte(dbType)

	if err := md.Db.Put([]byte(TypePrefix), dbTypeBytes); err != nil {
		return fmt.Errorf("cannot put db-type into aida-db; %v", err)
	}
	md.DbType = dbType

	md.log.Info("METADATA: DB Type saved successfully")

	return nil
}

// SetAll in given Db
func (md *AidaDbMetadata) SetAll() error {
	var err error

	if err = md.SetFirstBlock(md.FirstBlock); err != nil {
		return err
	}

	if err = md.SetLastBlock(md.LastBlock); err != nil {
		return err
	}

	if err = md.SetFirstEpoch(md.FirstEpoch); err != nil {
		return err
	}

	if err = md.SetLastEpoch(md.LastEpoch); err != nil {
		return err
	}

	if err = md.SetChainID(md.ChainId); err != nil {
		return err
	}

	if err = md.SetDbType(md.DbType); err != nil {
		return err
	}

	if err = md.SetTimestamp(); err != nil {
		return err
	}
	return nil
}

// SetDbHash in given Db
func (md *AidaDbMetadata) SetDbHash(dbHash []byte) error {
	if err := md.Db.Put([]byte(DbHashPrefix), dbHash); err != nil {
		return fmt.Errorf("cannot put metadata; %v", err)
	}

	md.log.Info("METADATA: Db hash saved successfully")

	return nil
}

// GetDbHash and return it
func (md *AidaDbMetadata) GetDbHash() []byte {
	dbHash, err := md.Db.Get([]byte(DbHashPrefix))
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil
		}
		md.log.Criticalf("cannot get Db hash from metadata; %v", err)
		return nil
	}

	return dbHash
}

// SetAllMetadata in given Db
func (md *AidaDbMetadata) SetAllMetadata(firstBlock uint64, lastBlock uint64, firstEpoch uint64, lastEpoch uint64, chainID ChainID, dbHash []byte, dbType AidaDbType) error {
	var err error

	if err = md.SetFirstBlock(firstBlock); err != nil {
		return err
	}

	if err = md.SetLastBlock(lastBlock); err != nil {
		return err
	}

	if err = md.SetFirstEpoch(firstEpoch); err != nil {
		return err
	}

	if err = md.SetLastEpoch(lastEpoch); err != nil {
		return err
	}

	if err = md.SetChainID(chainID); err != nil {
		return err
	}

	if err = md.SetDbType(dbType); err != nil {
		return err
	}

	if err = md.SetTimestamp(); err != nil {
		return err
	}

	if err = md.SetDbHash(dbHash); err != nil {
		return err
	}

	return nil
}

// findEpochs for block range in metadata
func (md *AidaDbMetadata) findEpochs() error {
	var (
		err                            error
		firstEpochMinus, lastEpochPlus uint64
	)

	// Finding epoch number calls rpc method eth_getBlockByNumber.
	// Ethereum does not provide information about epoch number in their RPC interface.
	if IsEthereumNetwork(md.ChainId) {
		return nil
	}

	md.FirstEpoch, err = FindEpochNumber(md.FirstBlock, md.ChainId)
	if err != nil {
		return err
	}

	// if first block is 0 we can be sure the block begins an epoch so no need to check that
	if md.FirstBlock != 0 {
		// we need to check if block is really first block of an epoch
		firstEpochMinus, err = FindEpochNumber(md.FirstBlock-1, md.ChainId)
		if err != nil {
			return err
		}

		if firstEpochMinus >= md.FirstEpoch {
			md.log.Warningf("first block of db is not beginning of an epoch; setting first epoch to 0")
			md.FirstEpoch = 0
		} else {
			md.log.Noticef("Found first epoch #%v", md.FirstEpoch)
		}
	}

	md.LastEpoch, err = FindEpochNumber(md.LastBlock, md.ChainId)
	if err != nil {
		return err
	}

	// we need to check if block is really last block of an epoch
	lastEpochPlus, err = FindEpochNumber(md.LastBlock+1, md.ChainId)
	if err != nil {
		return err
	}

	if lastEpochPlus <= md.LastEpoch {
		md.log.Warningf("last block block of db is not end of an epoch; setting last epoch to 0")
		md.LastEpoch = 0
	} else {
		md.log.Noticef("Found last epoch #%v", md.LastEpoch)
	}

	return nil
}

// CheckUpdateMetadata goes through metadata of updated AidaDb and its patch,
// looks if blocks and epoch align and if chainIDs are same for both Dbs
func (md *AidaDbMetadata) CheckUpdateMetadata(cfg *Config, patchDb db.BaseDB) error {
	var (
		err                                   error
		ignoreBlockAlignment, isLachesisPatch bool
	)

	patchMD := NewAidaDbMetadata(patchDb, cfg.LogLevel)

	patchMD.GetMetadata()

	// if we are updating existing AidaDb and this Db does not have metadata, we go through substate to find
	// blocks and epochs, chainID is Set from user via chain-id flag and db type in this case will always be genType
	md.GetMetadata()
	if md.LastBlock == 0 {
		if err = md.SetFreshMetadata(cfg.ChainID); err != nil {
			return fmt.Errorf("cannot set fresh metadata for existing AidaDb; %v", err)
		}
	}
	// we check if patch is lachesis with first condition
	// we also need to check that metadata were set with second condition
	if patchMD.FirstBlock == 0 {
		if patchMD.LastBlock == 0 {
			var ok bool

			sdb := db.MakeDefaultSubstateDBFromBaseDB(patchDb)
			patchMD.FirstBlock, patchMD.LastBlock, ok = FindBlockRangeInSubstate(sdb)
			if !ok {
				return errors.New("patch does not contain metadata and block range was not found in substate")
			}

			md.FirstEpoch, err = FindEpochNumber(md.FirstBlock, md.ChainId)
			if err != nil {
				return err
			}
			md.LastEpoch, err = FindEpochNumber(md.LastBlock, md.ChainId)
			if err != nil {
				return err
			}
		}
		// we need to check again whether first block is still 0 after substate search
		if patchMD.FirstBlock == 0 {
			ignoreBlockAlignment = true
			isLachesisPatch = true
		}
	}

	// we ignore block alignment also for first patch - this exception is for a situation when user has first patch
	// and lachesis is being installed, so first patch is getting replaced
	if patchMD.FirstBlock == 4564026 {
		ignoreBlockAlignment = true
	}

	// the patch is usable only if its FirstBlock is within targetDbs block range
	// and if its last block is bigger than tarGetDBs last block
	if patchMD.FirstBlock > md.LastBlock+1 || patchMD.FirstBlock < md.FirstBlock || patchMD.LastBlock <= md.LastBlock {
		// if patch is lachesis patch, we continue with merge

		if !ignoreBlockAlignment {
			return fmt.Errorf("metadata blocks does not align; aida-db %v-%v, patch %v-%v", md.FirstBlock, md.LastBlock, patchMD.FirstBlock, patchMD.LastBlock)
		}

	}

	// if chainIDs doesn't match, we can't patch the DB
	if md.ChainId != patchMD.ChainId {
		return fmt.Errorf("metadata chain-ids does not match; aida-db: %v, patch: %v", md.ChainId, patchMD.ChainId)
	}

	if isLachesisPatch {
		// we set the first block and epoch to 0
		// last block and epoch stays
		md.FirstBlock = 0
		md.FirstEpoch = 0
	} else if md.LastBlock < patchMD.LastBlock {
		// this condition is needed when we try to overwrite the first patch, then we dont want to overwrite the metadata
		// if patch is not lachesis hence is being appended, we take last block and epoch from it
		// first block and epoch stays
		md.LastBlock = patchMD.LastBlock
		md.LastEpoch = patchMD.LastEpoch
	}

	return nil
}

// SetFreshMetadata for an existing AidaDb without metadata
func (md *AidaDbMetadata) SetFreshMetadata(chainID ChainID) error {
	var err error

	if chainID == 0 {
		return fmt.Errorf("since you have aida-db without metadata you need to specify chain-id (--%v)", ChainIDFlag.Name)
	}

	// ChainID is Set by user in
	if err = md.SetChainID(chainID); err != nil {
		return err
	}

	if err = md.findEpochs(); err != nil {
		return err
	}

	if err = md.SetTimestamp(); err != nil {
		return err
	}

	_, err = getPatchFirstBlock(md.LastBlock)
	if err != nil {
		md.log.Warning("Uncertain AidaDbType.")
		if err = md.SetDbType(NoType); err != nil {
			return err
		}
	} else {
		if err = md.SetDbType(GenType); err != nil {
			return err
		}
	}

	return nil
}

func (md *AidaDbMetadata) SetBlockRange(firstBlock uint64, lastBlock uint64) error {
	var err error

	if err = md.SetFirstBlock(firstBlock); err != nil {
		return err
	}
	if err = md.SetLastBlock(lastBlock); err != nil {
		return err
	}

	return nil
}

func (md *AidaDbMetadata) DeleteMetadata() {
	var err error

	if err = md.Db.Delete([]byte(ChainIDPrefix)); err != nil {
		md.log.Criticalf("cannot delete chain-id; %v", err)
	} else {
		md.log.Debugf("ChainID deleted successfully")
	}

	if err = md.Db.Delete([]byte(FirstBlockPrefix)); err != nil {
		md.log.Criticalf("cannot delete first block; %v", err)
	} else {
		md.log.Debugf("First block deleted successfully")
	}

	if err = md.Db.Delete([]byte(LastBlockPrefix)); err != nil {
		md.log.Criticalf("cannot delete last block; %v", err)
	} else {
		md.log.Debugf("Last block deleted successfully")
	}

	if err = md.Db.Delete([]byte(FirstEpochPrefix)); err != nil {
		md.log.Criticalf("cannot delete first epoch; %v", err)
	} else {
		md.log.Debugf("First epoch deleted successfully")
	}

	if err = md.Db.Delete([]byte(LastEpochPrefix)); err != nil {
		md.log.Criticalf("cannot delete last epoch; %v", err)
	} else {
		md.log.Debugf("Last epoch deleted successfully")
	}

	if err = md.Db.Delete([]byte(TypePrefix)); err != nil {
		md.log.Criticalf("cannot delete db type; %v", err)
	} else {
		md.log.Debugf("Timestamp deleted successfully")
	}

	if err = md.Db.Delete([]byte(TimestampPrefix)); err != nil {
		md.log.Criticalf("cannot delete creation timestamp; %v", err)
	} else {
		md.log.Debugf("Timestamp deleted successfully")
	}
}

// UpdateMetadataInOldAidaDb Sets metadata necessary for update in old aida-db, which doesn't have any metadata
func (md *AidaDbMetadata) UpdateMetadataInOldAidaDb(chainId ChainID, firstAidaDbBlock uint64, lastAidaDbBlock uint64) error {
	var err error

	// Set chainid if it doesn't exist
	inCID := md.GetChainID()
	if inCID == 0 {
		err = md.SetChainID(chainId)
		if err != nil {
			return err
		}
	}

	// Set first block if it doesn't exist
	inFB := md.GetFirstBlock()
	if inFB == 0 {
		err = md.SetFirstBlock(firstAidaDbBlock)
		if err != nil {
			return err
		}
	}

	// Set last block if it doesn't exist
	inLB := md.GetLastBlock()
	if inLB == 0 {
		err = md.SetLastBlock(lastAidaDbBlock)
		if err != nil {
			return err
		}
	}

	// anything apart from clone db is always gentype db
	inType := md.GetDbType()
	if inType != CloneType {
		inType = GenType
	}

	err = md.SetDbType(inType)
	if err != nil {
		return err
	}

	return nil
}

// FindBlockRangeInSubstate if AidaDb does not yet have metadata
func FindBlockRangeInSubstate(db db.SubstateDB) (uint64, uint64, bool) {
	firstSubstate := db.GetFirstSubstate()
	if firstSubstate == nil {
		return 0, 0, false
	}
	firstBlock := firstSubstate.Env.Number

	lastSubstate, err := db.GetLastSubstate()
	if err != nil {
		return 0, 0, false
	}
	if lastSubstate == nil {
		return 0, 0, false
	}
	lastBlock := lastSubstate.Env.Number

	return firstBlock, lastBlock, true
}

func (md *AidaDbMetadata) getVerboseDbType() string {
	switch md.DbType {
	case GenType:
		return "Generate"
	case CloneType:
		return "Clone"
	case PatchType:
		return "Patch"
	case NoType:
		return "NoType"

	default:
		return "unknown db type"
	}
}

// DownloadPatchesJson downloads list of available patches from aida-db generation server.
func DownloadPatchesJson() (data []PatchJson, err error) {
	// Make the HTTP GET request
	patchesUrl := AidaDbRepositoryUrl + "/patches.json"
	response, err := http.Get(patchesUrl)
	if err != nil {
		return nil, fmt.Errorf("error making GET request for %s: %v", patchesUrl, err)
	}
	defer func(Body io.ReadCloser) {
		err = errors.Join(err, Body.Close())
	}(response.Body)

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response body: %s ; %v", string(body), err)
	}

	// Access the JSON data
	return data, nil
}

// getPatchFirstBlock finds first block of patch for given lastPatchBlock.
// given lastPatchBlock needs to end an epoch, otherwise an error is raised
func getPatchFirstBlock(lastPatchBlock uint64) (uint64, error) {
	var availableLastBlocks string

	patches, err := DownloadPatchesJson()
	if err != nil {
		return 0, fmt.Errorf("cannot download patches json; %v", err)
	}

	for _, p := range patches {
		if p.ToBlock == lastPatchBlock {
			return p.FromBlock, nil
		}
		availableLastBlocks += fmt.Sprintf("%v ", p.ToBlock)
	}

	return 0, fmt.Errorf("cannot find find first block for requested last block; requested: %v; available: [%v]", lastPatchBlock, availableLastBlocks)

}

// getBlockRange returns first and last block inside metadata.
// If last block is zero, it looks for block range in substate, and tries to get even the epoch range
func (md *AidaDbMetadata) getBlockRange() error {
	md.FirstBlock = md.GetFirstBlock()
	md.LastBlock = md.GetLastBlock()

	// check if AidaDb has block range
	if md.LastBlock == 0 {
		return errors.New("given aida-db does not contain metadata; please generate them using util-db metadata generate")
	}

	return nil
}

// HasStateHashPatch checks whether given db has already acquired patch with StateHashes.
func HasStateHashPatch(path string) (bool, error) {
	db, err := geth_leveldb.New(path, 1024, 100, "profiling", true)
	if err != nil {
		// if AidaDb does not exist force downloading the state hash patch
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("cannot open aida-db to check if it already has state hash patch; %v", err)
	}

	_, getErr := db.Get([]byte(HasStateHashPatchPrefix))

	err = db.Close()
	if err != nil {
		return false, fmt.Errorf("cannot close aida-db after checking if it already has state hash patch; %v", err)
	}

	if getErr != nil {
		if errors.Is(getErr, leveldb.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// SetHasHashPatch marks AidaDb that it already has HashPatch merged so it will not get downloaded next update.
func (md *AidaDbMetadata) SetHasHashPatch() error {
	return md.Db.Put([]byte(HasStateHashPatchPrefix), []byte{1})
}

func (md *AidaDbMetadata) SetUpdatesetInterval(val uint64) error {
	byteInterval := make([]byte, 8)
	binary.BigEndian.PutUint64(byteInterval, val)

	if err := md.Db.Put([]byte(db.UpdatesetIntervalKey), byteInterval); err != nil {
		return err
	}
	md.log.Info("METADATA: Updateset interval saved successfully")

	return nil
}

func (md *AidaDbMetadata) SetUpdatesetSize(val uint64) error {
	sizeInterval := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeInterval, val)

	if err := md.Db.Put([]byte(db.UpdatesetSizeKey), sizeInterval); err != nil {
		return err
	}
	md.log.Info("METADATA: Updateset size saved successfully")
	return nil
}
