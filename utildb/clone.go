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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb/dbcomponent"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/syndtr/goleveldb/leveldb"
)

const cloneWriteChanSize = 1

type cloner struct {
	cfg             *utils.Config
	log             logger.Logger
	aidaDb, cloneDb db.BaseDB
	cloneComponent  dbcomponent.DbComponent
	count           uint64
	typ             utils.AidaDbType
	writeCh         chan rawEntry
	errCh           chan error
	stopCh          chan any
}

// rawEntry representation of database entry
type rawEntry struct {
	Key   []byte
	Value []byte
}

// CreatePatchClone creates aida-db patch
func CreatePatchClone(cfg *utils.Config, aidaDb, targetDb db.BaseDB, firstEpoch, lastEpoch uint64, isNewOpera bool) error {
	var isFirstGenerationFromGenesis = false

	var cloneType = utils.PatchType

	// if the patch is first, we need to make some exceptions hence cloner needs to know
	if isNewOpera {
		if firstEpoch == 5577 && cfg.ChainID == utils.MainnetChainID {
			isFirstGenerationFromGenesis = true
		} else if firstEpoch == 2458 && cfg.ChainID == utils.TestnetChainID {
			isFirstGenerationFromGenesis = true
		}
	}

	err := Clone(cfg, aidaDb, targetDb, cloneType, isFirstGenerationFromGenesis)
	if err != nil {
		return err
	}

	md := utils.NewAidaDbMetadata(targetDb, cfg.LogLevel)
	err = md.SetFirstEpoch(firstEpoch)
	if err != nil {
		return err
	}

	err = md.SetLastEpoch(lastEpoch)
	if err != nil {
		return err
	}

	return nil
}

// Clone creates aida-db copy or subset - either clone(standalone - containing all necessary data for given range) or patch(containing data only for given range)
func Clone(cfg *utils.Config, aidaDb, cloneDb db.BaseDB, cloneType utils.AidaDbType, isFirstGenerationFromGenesis bool) error {
	var err error
	log := logger.NewLogger(cfg.LogLevel, "AidaDb Clone")

	var dbComponent dbcomponent.DbComponent

	if cloneType == utils.CustomType {
		dbComponent, err = dbcomponent.ParseDbComponent(cfg.DbComponent)
		if err != nil {
			return err
		}
	}

	start := time.Now()
	c := cloner{
		cfg:            cfg,
		cloneDb:        cloneDb,
		aidaDb:         aidaDb,
		log:            log,
		typ:            cloneType,
		cloneComponent: dbComponent,
		writeCh:        make(chan rawEntry, cloneWriteChanSize),
		errCh:          make(chan error, 1),
		stopCh:         make(chan any),
	}

	if err = c.clone(isFirstGenerationFromGenesis); err != nil {
		return err
	}

	c.log.Noticef("Cloning finished. Db saved to %v. Total elapsed time: %v", cfg.TargetDb, time.Since(start).Round(1*time.Second))
	return nil
}

// createDbClone AidaDb in given block range
func (c *cloner) clone(isFirstGenerationFromGenesis bool) error {
	go c.write()

	err := c.readData(isFirstGenerationFromGenesis)
	if err != nil {
		return err
	}

	// wait for writer result
	err, ok := <-c.errCh
	if ok {
		return err
	}

	if c.cfg.Validate {
		err = c.validateDbSize()
		if err != nil {
			return err
		}
	}

	if c.typ != utils.CustomType {
		sourceMD := utils.NewAidaDbMetadata(c.aidaDb, c.cfg.LogLevel)
		chainID := sourceMD.GetChainID()

		if err = utils.ProcessCloneLikeMetadata(c.cloneDb, c.typ, c.cfg.LogLevel, c.cfg.First, c.cfg.Last, chainID); err != nil {
			return err
		}
	}

	//  compact written data
	if c.cfg.CompactDb {
		c.log.Noticef("Starting compaction")
		err = c.cloneDb.Compact(nil, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// readData from source AidaDb
func (c *cloner) readData(isFirstGenerationFromGenesis bool) error {
	// notify writer that all data was read
	defer close(c.writeCh)

	if c.typ == utils.CustomType {
		return c.readDataCustom()
	}

	c.read([]byte(db.CodeDBPrefix), 0, nil)

	firstDeletionBlock := c.cfg.First

	// update c.cfg.First block before loading deletions and substates, because for utils.CloneType those are necessary to be from last updateset onward
	// lastUpdateBeforeRange contains block number at which is first updateset preceding the given block range,
	// it is only required in CloneType db
	lastUpdateBeforeRange := c.readUpdateSet(isFirstGenerationFromGenesis)
	if c.typ == utils.CloneType {
		// check whether updateset before interval exists
		if lastUpdateBeforeRange < c.cfg.First && lastUpdateBeforeRange != 0 {
			c.log.Noticef("Last updateset found at block %v, changing first block to %v", lastUpdateBeforeRange, lastUpdateBeforeRange+1)
			c.cfg.First = lastUpdateBeforeRange + 1
		}

		// if database type is going to be CloneType, we need to load all deletion data, because some commands need to load deletionDb from block 0
		firstDeletionBlock = 0
	}

	c.readDeletions(firstDeletionBlock)

	c.readSubstate()

	err := c.readStateHashes()
	if err != nil {
		return fmt.Errorf("cannot read state hashes; %v", err)
	}

	c.readBlockHashes()

	err = c.readExceptions()
	if err != nil {
		return fmt.Errorf("cannot read exceptions; %v", err)
	}

	return nil
}

// write data read from func read() into new createDbClone
func (c *cloner) write() {
	defer close(c.errCh)

	var (
		err         error
		data        rawEntry
		ok          bool
		batchWriter db.Batch
	)

	batchWriter = c.cloneDb.NewBatch()

	for {
		select {
		case data, ok = <-c.writeCh:
			if !ok {
				// iteration completed - read rest of the pending data
				if batchWriter.ValueSize() > 0 {
					err = batchWriter.Write()
					if err != nil {
						c.errCh <- fmt.Errorf("cannot read rest of the data into createDbClone; %v", err)
						return
					}
				}
				return
			}

			err = batchWriter.Put(data.Key, data.Value)
			if err != nil {
				c.errCh <- fmt.Errorf("cannot put data into createDbClone %v", err)
				return
			}

			// writing data in batches
			if batchWriter.ValueSize() > kvdb.IdealBatchSize {
				err = batchWriter.Write()
				if err != nil {
					c.errCh <- fmt.Errorf("cannot write batch; %v", err)
					return
				}

				// reset writer after writing batch
				batchWriter.Reset()
			}
		case <-c.stopCh:
			return
		}
	}
}

// read data with given prefix until given condition is fulfilled from source AidaDb
func (c *cloner) read(prefix []byte, start uint64, condition func(key []byte) (bool, error)) {
	c.log.Noticef("Copying data with prefix %v", string(prefix))

	iter := c.aidaDb.NewIterator(prefix, db.BlockToBytes(start))
	defer iter.Release()

	for iter.Next() {
		if condition != nil {
			finished, err := condition(iter.Key())
			if err != nil {
				c.errCh <- fmt.Errorf("condition emit error; %v", err)
				return
			}
			if finished {
				break
			}
		}

		c.count++
		ok := c.sendToWriteChan(iter.Key(), iter.Value())
		if !ok {
			return
		}

	}
	c.log.Noticef("Prefix %v done", string(prefix))

	return
}

// readUpdateSet from UpdateDb
func (c *cloner) readUpdateSet(isFirstGenerationFromGenesis bool) uint64 {
	// labeling last updateSet before interval - need to export substate for that range as well
	var lastUpdateBeforeRange uint64
	endCond := func(key []byte) (bool, error) {
		block, err := db.DecodeUpdateSetKey(key)
		if err != nil {
			return false, err
		}
		if block > c.cfg.Last {
			return true, nil
		}
		if block < c.cfg.First {
			lastUpdateBeforeRange = block
		}
		return false, nil
	}

	if c.typ == utils.CloneType {
		c.read([]byte(db.UpdateDBPrefix), 0, endCond)

		// if there is no updateset before interval (first 1M blocks) then 0 is returned
		return lastUpdateBeforeRange
	} else if c.typ == utils.PatchType || c.typ == utils.CustomType {
		var wantedBlock uint64

		// if we are working with first patch that was created from genesis we need to move the start of the iterator minus one block
		// so first update-set from worldstate gets inserted
		if isFirstGenerationFromGenesis {
			wantedBlock = c.cfg.First - 1
		} else {
			wantedBlock = c.cfg.First
		}

		c.read([]byte(db.UpdateDBPrefix), wantedBlock, endCond)
		return 0
	} else {
		c.errCh <- fmt.Errorf("incorrect clone type: %v", c.typ)
		return 0
	}
}

// readSubstate from last updateSet before cfg.First until cfg.Last
func (c *cloner) readSubstate() {
	endCond := func(key []byte) (bool, error) {
		block, _, err := db.DecodeSubstateDBKey(key)
		if err != nil {
			return false, err
		}
		if block > c.cfg.Last {
			return true, nil
		}
		return false, nil
	}

	c.read([]byte(db.SubstateDBPrefix), c.cfg.First, endCond)
}

func (c *cloner) readStateHashes() error {
	c.log.Noticef("Copying state hashes")

	var errCounter uint64

	for i := c.cfg.First; i <= c.cfg.Last; i++ {
		key := []byte(utils.StateRootHashPrefix + hexutil.EncodeUint64(i))
		value, err := c.aidaDb.Get(key)
		if err != nil {
			if errors.Is(err, leveldb.ErrNotFound) {
				errCounter++
				continue
			} else {
				return err
			}
		}
		c.count++
		ok := c.sendToWriteChan(key, value)
		if !ok {
			return nil
		}
	}

	if errCounter > 0 {
		c.log.Warningf("State hashes were missing for %v blocks", errCounter)
	}

	c.log.Noticef("State hashes done")

	return nil
}

// readBlockHashes from last updateSet before cfg.First until cfg.Last
func (c *cloner) readBlockHashes() {
	endCond := func(key []byte) (bool, error) {
		block, err := utils.DecodeBlockHashDBKey(key)
		if err != nil {
			return false, err
		}
		if block > c.cfg.Last {
			return true, nil
		}
		return false, nil
	}

	c.read([]byte(utils.BlockHashPrefix), c.cfg.First, endCond)
}

// readExceptions reading exceptions from AidaDb
func (c *cloner) readExceptions() error {
	endCond := func(key []byte) (bool, error) {
		block, err := db.DecodeExceptionDBKey(key)
		if err != nil {
			return false, err
		}
		if block > c.cfg.Last {
			return true, nil
		}
		return false, nil
	}

	c.read([]byte(db.ExceptionDBPrefix), c.cfg.First, endCond)

	return nil
}

func (c *cloner) sendToWriteChan(k, v []byte) bool {
	// make deep read key and value
	// need to pass deep read of values into the channel
	// golang channels were using pointers and values read from channel were incorrect
	key := make([]byte, len(k))
	copy(key, k)
	value := make([]byte, len(v))
	copy(value, v)

	select {
	case <-c.stopCh:
		return false
	case c.writeCh <- rawEntry{Key: key, Value: value}:
		return true
	}
}

// readDeletions from last updateSet before cfg.First until cfg.Last
func (c *cloner) readDeletions(firstDeletionBlock uint64) {
	endCond := func(key []byte) (bool, error) {
		block, _, err := db.DecodeDestroyedAccountKey(key)
		if err != nil {
			return false, err
		}
		if block > c.cfg.Last {
			return true, nil
		}
		return false, nil
	}

	c.read([]byte(db.DestroyedAccountPrefix), firstDeletionBlock, endCond)
}

// validateDbSize compares size of database and expectedWritten
func (c *cloner) validateDbSize() error {
	actualWritten := GetDbSize(c.cloneDb)
	if actualWritten != c.count {
		return fmt.Errorf("TargetDb has %v records; expected: %v", actualWritten, c.count)
	}
	return nil
}

// closeDbs when cloning is done
func (c *cloner) closeDbs() {
	var err error

	if err = c.aidaDb.Close(); err != nil {
		c.log.Errorf("cannot close aida-db")
	}

	if err = c.cloneDb.Close(); err != nil {
		c.log.Errorf("cannot close aida-db")
	}
}

// stop all cloner threads
func (c *cloner) stop() {
	select {
	case <-c.stopCh:
		return
	default:
		close(c.stopCh)
		c.closeDbs()
	}
}

// readDataCustom retrieves data from source AidaDb based on given dbComponent
func (c *cloner) readDataCustom() error {
	if c.cloneComponent == dbcomponent.Substate || c.cloneComponent == dbcomponent.All {
		c.read([]byte(db.CodeDBPrefix), 0, nil)
		c.readSubstate()
	}

	if c.cloneComponent == dbcomponent.Delete || c.cloneComponent == dbcomponent.All {
		c.readDeletions(c.cfg.First)
	}

	if c.cloneComponent == dbcomponent.Update || c.cloneComponent == dbcomponent.All {
		lastUpdateBeforeRange := c.readUpdateSet(false)
		c.log.Noticef("Last updateset found at block %v", lastUpdateBeforeRange)
	}

	if c.cloneComponent == dbcomponent.StateHash || c.cloneComponent == dbcomponent.All {
		err := c.readStateHashes()
		if err != nil {
			return err
		}
	}

	if c.cloneComponent == dbcomponent.BlockHash || c.cloneComponent == dbcomponent.All {
		c.readBlockHashes()
	}

	if c.cloneComponent == dbcomponent.Exception || c.cloneComponent == dbcomponent.All {
		err := c.readExceptions()
		if err != nil {
			return fmt.Errorf("cannot read exceptions; %v", err)
		}
	}

	return nil
}

// OpenCloningDbs prepares aida and target databases
func OpenCloningDbs(aidaDbPath, targetDbPath string) (db.BaseDB, db.BaseDB, error) {
	var err error

	// if source db doesn't exist raise error
	_, err = os.Stat(aidaDbPath)
	if os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("specified aida-db %v is empty\n", aidaDbPath)
	}

	// if target db exists raise error
	_, err = os.Stat(targetDbPath)
	if !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("specified target-db %v already exists\n", targetDbPath)
	}

	var aidaDb, cloneDb db.BaseDB

	// open db
	aidaDb, err = db.NewReadOnlyBaseDB(aidaDbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("aidaDb %v; %v", aidaDbPath, err)
	}

	// open createDbClone
	cloneDb, err = db.NewDefaultBaseDB(targetDbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("targetDb %v; %v", targetDbPath, err)
	}

	return aidaDb, cloneDb, nil
}
