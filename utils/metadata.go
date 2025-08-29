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

func (t AidaDbType) String() string {
	switch t {
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

type Metadata interface {
	// GenerateMetadata generates new or updates metadata in AidaDb.
	GenerateMetadata() error
	// Merge merges source metadata into target metadata if its possible.
	Merge(AidaDbMetadata) error
	// Delete deletes metadata from AidaDb.
	Delete()

	// Getters
	GetFirstBlock() uint64
	GetLastBlock() uint64
	GetFirstEpoch() uint64
	GetLastEpoch() uint64
	GetChainID() ChainID
	GetTimestamp() uint64
	GetDbType() AidaDbType
	GetDbHash() []byte

	// Setters
	SetFirstBlock(uint64) error
	SetLastBlock(uint64) error
	SetFirstEpoch(uint64) error
	SetLastEpoch(uint64) error
	SetChainID(ChainID) error
	SetTimestamp() error
	SetDbType(AidaDbType) error
	SetDbHash([]byte) error
	SetHasHashPatch() error
	SetUpdatesetInterval(uint64) error
	SetUpdatesetSize(uint64) error
}

// AidaDbMetadata holds any information about AidaDb needed for putting it into the Db
type AidaDbMetadata struct {
	Db                    db.SubstateDB
	log                   logger.Logger
	FirstBlock, LastBlock uint64
	FirstEpoch, LastEpoch uint64
	ChainId               ChainID
	DbType                AidaDbType
	timestamp             uint64
}

func (md *AidaDbMetadata) GenerateMetadata() error {
	chainId := md.GetChainID()
	if chainId == 0 {
		md.log.Warningf("Your AidaDb does not contain chain-id, metadata will be incomplete")
	}

	fss := md.Db.GetFirstSubstate()
	// if there is no substate, we cannot find blocks and epochs
	if fss == nil {
		md.log.Warningf("Your AidaDb does not contain any substate, metadata will be incomplete")
	} else {
		lss, err := md.Db.GetLastSubstate()
		if err != nil {
			return fmt.Errorf("cannot get last substate; %v", err)
		}
		err = md.SetFirstBlock(fss.Block)
		if err != nil {
			return fmt.Errorf("cannot set first block; %v", err)
		}
		err = md.SetLastBlock(lss.Block)
		if err != nil {
			return fmt.Errorf("cannot set last block; %v", err)
		}

		// Epoch numbers can only be found if chainID is known and AidaDB has substates
		if chainId != 0 {
			err = md.findEpochs()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// NewAidaDbMetadata creates new instance of AidaDbMetadata
func NewAidaDbMetadata(db db.SubstateDB, logLevel string) Metadata {
	return &AidaDbMetadata{
		Db:  db,
		log: logger.NewLogger(logLevel, "aida-metadata"),
	}
}

func (md *AidaDbMetadata) Merge(src AidaDbMetadata) error {
	targetChainID := md.GetChainID()
	srcChainID := src.GetChainID()
	if targetChainID != 0 {
		if targetChainID != srcChainID {
			return fmt.Errorf("cannot merge dbs with different chainIDs; target db chainID %v, source db chainID %v", targetChainID, srcChainID)
		}
	} else {
		if srcChainID == 0 {
			return errors.New("cannot merge dbs with no chainIDs in metadata; you can set chainID manually using the util-db insert cmd")
		}
		err := md.SetChainID(srcChainID)
		if err != nil {
			return fmt.Errorf("cannot set chainID while merging dbs; %v", err)
		}
	}

	// Set DbType
	targetDbType := md.GetDbType()
	srcDbType := src.GetDbType()
	switch targetDbType {
	case NoType:
		targetDbType = srcDbType
	case GenType:
		switch srcDbType {
		// GetType and PatchType can be merged onto GenType
		case GenType:
			break
		case PatchType:
			break
		default:
			targetDbType = CustomType
		}
	default:
		targetDbType = CustomType
	}
	err := md.SetDbType(targetDbType)
	if err != nil {
		return fmt.Errorf("cannot merge db type: %v", err)
	}
	// Find block range
	targetFirstBlock := md.GetFirstBlock()
	srcFirstBlock := src.GetFirstBlock()
	targetLastBlock := md.GetLastBlock()
	srcLastBlock := src.GetLastBlock()

	// Source is a subset of target
	if targetFirstBlock < srcFirstBlock && targetLastBlock > srcLastBlock {
		return fmt.Errorf("source db (%v-%v) is subset of target db (%v-%v)", srcFirstBlock, srcLastBlock, targetFirstBlock, targetLastBlock)
	}
	// Target is a subset of source
	if targetFirstBlock > srcFirstBlock && targetLastBlock < srcLastBlock {
		return fmt.Errorf("target db (%v-%v) is subset of source db (%v-%v)", targetFirstBlock, targetLastBlock, srcFirstBlock, srcLastBlock)
	}

	blocksOk := false
	// Check alignment - dbs can overlap but cannot have gaps
	// Target is before source
	if targetLastBlock+1 >= srcFirstBlock {
		err = md.SetLastBlock(srcLastBlock)
		if err != nil {
			return fmt.Errorf("cannot merge last block: %v", err)
		}
		blocksOk = true
	}

	// Target is after source
	if srcLastBlock+1 >= targetFirstBlock {
		err = md.SetFirstBlock(srcFirstBlock)
		if err != nil {
			return fmt.Errorf("cannot merge first block: %v", err)
		}
		blocksOk = true
	}

	if !blocksOk {
		return fmt.Errorf("blocks does not align; target db (%v-%v), source db (%v-%v)", targetFirstBlock, targetLastBlock, srcFirstBlock, srcLastBlock)
	}
	return md.findEpochs()
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
	if md.ChainId != 0 {
		return md.ChainId
	}
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
		md.ChainId = ChainID(bigendian.BytesToUint16(chainIDBytes))
	} else {
		md.ChainId = ChainID(bigendian.BytesToUint64(chainIDBytes))
	}
	return md.ChainId
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

	firstEpoch, err := FindEpochNumber(md.FirstBlock, md.ChainId)
	if err != nil {
		return fmt.Errorf("cannot find first epoch; %v", err)
	}
	// if first block is 0 we can be sure the block begins an epoch so no need to check that
	if md.FirstBlock != 0 {
		// we need to check if block is really first block of an epoch
		firstEpochMinus, err = FindEpochNumber(md.FirstBlock-1, md.ChainId)
		if err != nil {
			return err
		}

		if firstEpochMinus >= md.FirstEpoch {
			md.log.Warningf("first block of db is not beginning of an epoch")
		} else {
			md.log.Noticef("Found first epoch #%v", md.FirstEpoch)
			err = md.SetFirstEpoch(firstEpoch)
			if err != nil {
				return fmt.Errorf("cannot set first epoch; %v", err)
			}
		}
	}

	lastEpoch, err := FindEpochNumber(md.LastBlock, md.ChainId)
	if err != nil {
		return fmt.Errorf("cannot find last epoch; %v", err)
	}
	// we need to check if block is really last block of an epoch
	lastEpochPlus, err = FindEpochNumber(md.LastBlock+1, md.ChainId)
	if err != nil {
		return err
	}

	if lastEpochPlus <= md.LastEpoch {
		md.log.Warningf("last block block of db is not end of an epoch")
	} else {
		md.log.Noticef("Found last epoch #%v", md.LastEpoch)
		err = md.SetLastEpoch(lastEpoch)
		if err != nil {
			return fmt.Errorf("cannot set last epoch; %v", err)
		}
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

func (md *AidaDbMetadata) Delete() {
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
