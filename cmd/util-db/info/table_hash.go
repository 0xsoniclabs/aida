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

package info

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/0xsoniclabs/aida/cmd/util-db/dbutils/dbcomponent"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/syndtr/goleveldb/leveldb"
)

// tableHash generates a hash for given dbcomponent
func tableHash(cfg *utils.Config, base db.BaseDB, log logger.Logger) error {
	dbComponent, err := dbcomponent.ParseDbComponent(cfg.DbComponent)
	if err != nil {
		return err
	}

	reportFrequency := time.Minute

	if dbComponent == dbcomponent.Substate || dbComponent == dbcomponent.All {
		log.Info("Generating Substate hash...")
		aidaDbSubstateHash, count, err := getSubstateHash(cfg, base, reportFrequency, log)
		if err != nil {
			return err
		}
		log.Infof("Substate hash: %x; count %v", aidaDbSubstateHash, count)
	}

	if dbComponent == dbcomponent.Delete || dbComponent == dbcomponent.All {
		log.Info("Generating Deletion hash...")
		aidaDbDeletionHash, count, err := GetDeletionHash(cfg, base, reportFrequency, log)
		if err != nil {
			return err
		}
		log.Infof("Deletion hash: %x; count %v", aidaDbDeletionHash, count)
	}

	if dbComponent == dbcomponent.Update || dbComponent == dbcomponent.All {
		log.Info("Generating Updateset hash...")
		aidaDbUpdateDbHash, count, err := GetUpdateDbHash(cfg, base, log)
		if err != nil {
			return err
		}
		log.Infof("Updateset hash: %x; count %v", aidaDbUpdateDbHash, count)
	}

	if dbComponent == dbcomponent.StateHash || dbComponent == dbcomponent.All {
		log.Info("Generating State-Hashes hash...")
		aidaDbStateHashesHash, count, err := GetStateRootHashesHash(cfg, base, reportFrequency, log)
		if err != nil {
			return err
		}
		log.Infof("State-Hashes hash: %x; count %v", aidaDbStateHashesHash, count)
	}

	if dbComponent == dbcomponent.BlockHash || dbComponent == dbcomponent.All {
		log.Info("Generating Block-Hashes hash...")
		aidaDbBlockHashesHash, count, err := GetBlockHashesHash(cfg, base, reportFrequency, log)
		if err != nil {
			return err
		}
		log.Infof("Block-Hashes hash: %x; count %v", aidaDbBlockHashesHash, count)
	}

	if dbComponent == dbcomponent.Exception || dbComponent == dbcomponent.All {
		log.Info("Generating Exception hash...")
		exceptionHash, count, err := GetExceptionDbHash(cfg, base, reportFrequency, log)
		if err != nil {
			return err
		}
		log.Infof("Exception hash: %x; count %v", exceptionHash, count)
	}

	return nil
}

// combineJson reads objects from in channel, encodes their []byte representation and writes to out channel
func combineJson(in chan any, out chan []byte, errChan chan error) {
	for {
		select {
		case value, ok := <-in:
			if !ok {
				return
			}
			jsonData, err := json.Marshal(value)
			if err != nil {
				errChan <- err
				return
			}
			out <- jsonData
		}
	}
}

func getSubstateHash(
	cfg *utils.Config,
	base db.BaseDB,
	progressLoggerFrequency time.Duration,
	log logger.Logger,
) ([]byte, uint64, error) {
	ticker := time.NewTicker(progressLoggerFrequency)
	defer ticker.Stop()

	feeder := func(feederChan chan any, errChan chan error) {
		defer close(feederChan)

		sdb := db.MakeDefaultSubstateDBFromBaseDB(base)
		err := sdb.SetSubstateEncoding(cfg.SubstateEncoding)
		if err != nil {
			errChan <- err
			return
		}
		it := sdb.NewSubstateIterator(int(cfg.First), 10)
		defer it.Release()

		for it.Next() {
			if it.Value().Block > cfg.Last {
				break
			}

			select {
			case <-ticker.C:
				log.Infof("SubstateDb hash progress: %v/%v", it.Value().Block, cfg.Last)
			default:
			}

			select {
			case err := <-errChan:
				errChan <- err
				return
			case feederChan <- it.Value():
			}
		}
	}

	return parallelHashComputing(feeder)
}

func GetDeletionHash(
	cfg *utils.Config,
	aidaDb db.BaseDB,
	progressLoggerFrequency time.Duration,
	log logger.Logger,
) ([]byte, uint64, error) {
	ticker := time.NewTicker(progressLoggerFrequency)
	defer ticker.Stop()

	feeder := func(feederChan chan any, errChan chan error) {
		defer close(feederChan)

		startingBlockBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(startingBlockBytes, cfg.First)

		iter := aidaDb.NewIterator([]byte(db.DestroyedAccountPrefix), startingBlockBytes)
		defer iter.Release()

		for iter.Next() {
			block, _, err := db.DecodeDestroyedAccountKey(iter.Key())
			if err != nil {
				errChan <- err
				return
			}
			if block > cfg.Last {
				break
			}

			list, err := db.DecodeAddressList(iter.Value())
			if err != nil {
				errChan <- err
				return
			}

			combined := append(list.DestroyedAccounts, list.ResurrectedAccounts...)

			sort.Slice(combined, func(i, j int) bool {
				return bytes.Compare(combined[i].Bytes(), combined[j].Bytes()) < 0
			})

			for _, address := range combined {
				select {
				case <-ticker.C:
					log.Infof("DeletionDb hash progress: %v/%v", block, cfg.Last)
				default:
				}

				select {
				case err = <-errChan:
					errChan <- err
					return
				case feederChan <- address.String():
				}
			}
		}
	}
	return parallelHashComputing(feeder)
}

func GetUpdateDbHash(cfg *utils.Config, base db.BaseDB, log logger.Logger) ([]byte, uint64, error) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	feeder := func(feederChan chan any, errChan chan error) {
		defer close(feederChan)

		udb := db.MakeDefaultUpdateDBFromBaseDB(base)
		iter := udb.NewUpdateSetIterator(cfg.First, cfg.Last)
		defer iter.Release()

		for iter.Next() {
			select {
			case <-ticker.C:
				log.Infof("UpdateDb hash progress: %v/%v", iter.Value().Block, cfg.Last)
			default:
			}

			value := iter.Value()
			select {
			case err := <-errChan:
				errChan <- err
				return
			case feederChan <- value:
			}
		}
	}
	return parallelHashComputing(feeder)
}

func GetStateRootHashesHash(
	cfg *utils.Config,
	base db.BaseDB,
	progressLoggerFrequency time.Duration,
	log logger.Logger,
) ([]byte, uint64, error) {
	ticker := time.NewTicker(progressLoggerFrequency)
	defer ticker.Stop()

	feeder := func(feederChan chan any, errChan chan error) {
		defer close(feederChan)

		provider := utils.MakeHashProvider(base)

		for i := cfg.First; i <= cfg.Last; i++ {
			select {
			case <-ticker.C:
				log.Infof("State-Hashes hash progress: %v/%v", i, cfg.Last)
			default:
			}

			h, err := provider.GetStateRootHash(int(i))
			if err != nil {
				if errors.Is(err, leveldb.ErrNotFound) {
					continue
				}
				errChan <- err
				return
			}

			select {
			case err = <-errChan:
				errChan <- err
				return
			case feederChan <- h:
			}
		}
	}

	return parallelHashComputing(feeder)
}

func GetBlockHashesHash(
	cfg *utils.Config,
	base db.BaseDB,
	progressLoggerFrequency time.Duration,
	log logger.Logger,
) ([]byte, uint64, error) {
	ticker := time.NewTicker(progressLoggerFrequency)
	defer ticker.Stop()

	feeder := func(feederChan chan any, errChan chan error) {
		defer close(feederChan)

		provider := utils.MakeHashProvider(base)

		for i := cfg.First; i <= cfg.Last; i++ {
			select {
			case <-ticker.C:
				log.Infof("Block-Hashes hash progress: %v/%v", i, cfg.Last)
			default:
			}

			h, err := provider.GetBlockHash(int(i))
			if err != nil {
				if errors.Is(err, leveldb.ErrNotFound) {
					continue
				}
				errChan <- err
				return
			}

			select {
			case err = <-errChan:
				errChan <- err
				return
			case feederChan <- h:
			}
		}
	}

	return parallelHashComputing(feeder)
}

func GetExceptionDbHash(
	cfg *utils.Config,
	base db.BaseDB,
	progressLoggerFrequency time.Duration,
	log logger.Logger,
) ([]byte, uint64, error) {
	ticker := time.NewTicker(progressLoggerFrequency)
	defer ticker.Stop()

	feeder := func(feederChan chan any, errChan chan error) {
		defer close(feederChan)

		startingBlockBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(startingBlockBytes, cfg.First)

		edb := db.MakeDefaultExceptionDBFromBaseDB(base)
		iter := edb.NewIterator([]byte(db.ExceptionDBPrefix), startingBlockBytes)
		defer iter.Release()

		i := 0
		for iter.Next() {
			i++
			select {
			case <-ticker.C:
				log.Infof("Exception hash progress: %v/%v", i, cfg.Last)
			default:
			}

			block, err := db.DecodeExceptionDBKey(iter.Key())
			if err != nil {
				errChan <- err
				return
			}
			if block > cfg.Last {
				break
			}

			// make copy of value before sending to channel
			value := bytes.Clone(iter.Value())
			select {
			case err = <-errChan:
				errChan <- err
				return
			case feederChan <- value:
			}
		}
	}
	return parallelHashComputing(feeder)
}

func parallelHashComputing(feeder func(chan any, chan error)) ([]byte, uint64, error) {
	var wg sync.WaitGroup
	feederChan := make(chan any, 1)
	processedChan := make(chan []byte, 1)

	errChan := make(chan error)

	var counter uint64 = 0

	countingFeeder := make(chan any)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case err := <-errChan:
				errChan <- err
				return
			case item, ok := <-countingFeeder:
				if !ok {
					close(feederChan)
					return
				}
				counter++

				select {
				case err := <-errChan:
					errChan <- err
					return
				case feederChan <- item:
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		feeder(countingFeeder, errChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		combineJson(feederChan, processedChan, errChan)
	}()

	// Start a goroutine to close hashChan when all workers finish
	go func() {
		wg.Wait()
		close(errChan)
		close(processedChan)
	}()

	hasher := md5.New()

	for {
		select {
		case err, ok := <-errChan:
			if ok {
				if err != nil {
					return nil, counter, err
				}
			}
		case value, ok := <-processedChan:
			if !ok {
				return hasher.Sum(nil), counter, nil
			}
			hasher.Write(value)
		}
	}
}
