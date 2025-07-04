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

package primer

import (
	"fmt"
	"path/filepath"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/google/martian/log"
)

func MakeStateDbPrimer[T any](cfg *utils.Config) executor.Extension[T] {
	if cfg.SkipPriming {
		return extension.NilExtension[T]{}
	}

	return makeStateDbPrimer[T](cfg, logger.NewLogger(cfg.LogLevel, "StateDb-Primer"))
}

func makeStateDbPrimer[T any](cfg *utils.Config, log logger.Logger) *stateDbPrimer[T] {
	return &stateDbPrimer[T]{
		cfg: cfg,
		log: log,
	}
}

type stateDbPrimer[T any] struct {
	extension.NilExtension[T]
	cfg *utils.Config
	log logger.Logger
	ctx *utils.PrimeContext
}

// PreRun primes StateDb to given block.
func (p *stateDbPrimer[T]) PreRun(_ executor.State[T], ctx *executor.Context) (err error) {
	if p.cfg.SkipPriming {
		p.log.Warning("Skipping priming (disabled by user)...")
		return nil
	}

	// As different substates start on different blocks (either 0 or 1)
	// we must check the starting block with the key word "first" with appropriate chainid
	if p.cfg.First == utils.KeywordBlocks[p.cfg.ChainID]["first"] {
		if utils.IsEthereumNetwork(p.cfg.ChainID) {
			p.log.Noticef("Priming ethereum genesis...")
			p.ctx = utils.NewPrimeContext(p.cfg, ctx.State, utils.KeywordBlocks[p.cfg.ChainID]["first"], p.log)
			return p.primeEthereumGenesis(ctx.State, ctx.AidaDb)
		}
		return nil
	}

	// is used to determine block from which the priming starts
	var primingStartBlock uint64
	if p.cfg.IsExistingStateDb {
		var stateDbInfo utils.StateDbInfo
		var err error
		if p.cfg.ShadowDb {
			stateDbInfo, err = utils.ReadStateDbInfo(filepath.Join(p.cfg.StateDbSrc, utils.PathToPrimaryStateDb))
		} else {
			stateDbInfo, err = utils.ReadStateDbInfo(p.cfg.StateDbSrc)
		}
		if err != nil {
			return fmt.Errorf("cannot read state db info; %w", err)
		}
		primingStartBlock = stateDbInfo.Block + 1
	}

	if primingStartBlock == p.cfg.First {
		p.log.Debugf("skipping priming; first priming block %v; first block %v", primingStartBlock-1, p.cfg.First)
		return nil
	}

	// user incorrectly tries to prime data into database even tho database is already advanced further
	if primingStartBlock > p.cfg.First {
		return fmt.Errorf("undefined behaviour; starting block %v shouldn't lower than block of provided stateDb %v", p.cfg.First, primingStartBlock-1)
	}

	if p.cfg.PrimeRandom {
		p.log.Infof("Randomized Priming enabled; Seed: %v, threshold: %v", p.cfg.RandomSeed, p.cfg.PrimeThreshold)
	}

	p.log.Infof("Update buffer size: %v bytes", p.cfg.UpdateBufferSize)

	p.log.Noticef("Priming from block %v...", primingStartBlock)
	p.log.Noticef("Priming to block %v...", p.cfg.First-1)
	p.ctx = utils.NewPrimeContext(p.cfg, ctx.State, primingStartBlock, p.log)

	return p.prime(ctx.State, ctx.AidaDb)
}

// prime advances the stateDb to given first block.
func (p *stateDbPrimer[T]) prime(stateDb state.StateDB, aidaDb db.BaseDB) error {
	var (
		totalSize uint64             // total size of unprimed update set
		hasPrimed bool               // if true, db has been primed
		block     = p.ctx.GetBlock() // current block priming position
	)

	// load pre-computed update-set from update-set db
	udb := db.MakeDefaultUpdateDBFromBaseDB(aidaDb)
	updateIter := udb.NewUpdateSetIterator(block, p.cfg.First-1)
	update := make(substate.WorldState)

	for updateIter.Next() {
		newSet := updateIter.Value()
		if newSet.Block > p.cfg.First-1 {
			break
		}
		block = newSet.Block

		incrementalSize := update.EstimateIncrementalSize(newSet.WorldState)

		// Prime StateDB
		if totalSize+incrementalSize > p.cfg.UpdateBufferSize {
			p.log.Infof("\tPriming...")
			if err := p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
				return fmt.Errorf("cannot prime state-db; %v", err)
			}

			totalSize = 0
			update = make(substate.WorldState)
			hasPrimed = true
		}

		// Reset accessed storage locations of suicided accounts prior to update-set block.
		// The known accessed storage locations in the update-set range has already been
		// reset when generating the update set database.
		utils.ClearAccountStorage(update, newSet.DeletedAccounts)
		// if exists in DB, suicide
		if hasPrimed {
			p.ctx.SelfDestructAccounts(stateDb, newSet.DeletedAccounts)
		}

		update.Merge(newSet.WorldState)
		totalSize += incrementalSize
		p.log.Infof("\tMerge update set at block %v. New total size %v MB (+%v MB)",
			newSet.Block, totalSize/1_000_000,
			incrementalSize/1_000_000)
		// advance block after merge update set
		block++
	}
	updateIter.Release()

	// if update set is not empty, prime the remaining
	if len(update) > 0 {
		if err := p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
			return fmt.Errorf("cannot prime state-db; %v", err)
		}
		update = make(substate.WorldState)
		hasPrimed = true
	}

	// advance from the latest precomputed update-set to the target block
	// if the first block is 1, target must prime the genesis block
	if block < p.cfg.First || p.cfg.First-1 == 0 {
		log.Infof("\tPriming using substate from %v to %v", block, p.cfg.First-1)
		update, deletedAccounts, err := utils.GenerateUpdateSet(block, p.cfg.First-1, p.cfg, aidaDb)
		if err != nil {
			return fmt.Errorf("cannot generate update-set; %w", err)
		}
		if hasPrimed {
			p.ctx.SelfDestructAccounts(stateDb, deletedAccounts)
		}
		if err = p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
			return fmt.Errorf("cannot prime state-db; %w", err)
		}
	}

	p.log.Noticef("Delete destroyed accounts until block %v", p.cfg.First-1)

	// remove destroyed accounts until one block before the first block
	err := utils.DeleteDestroyedAccountsFromStateDB(stateDb, p.cfg, p.cfg.First-1, aidaDb)
	if err != nil {
		return fmt.Errorf("cannot delete destroyed accounts from state-db; %v", err)
	}

	return nil
}

func (p *stateDbPrimer[T]) primeEthereumGenesis(stateDb state.StateDB, aidaDb db.BaseDB) error {
	// load pre-computed update-set from update-set db
	block := p.ctx.GetBlock()
	udb := db.MakeDefaultUpdateDBFromBaseDB(aidaDb)
	updateIter := udb.NewUpdateSetIterator(block, block+1)
	update := make(substate.WorldState)

	for updateIter.Next() {
		newSet := updateIter.Value()
		block = newSet.Block
		update.Merge(newSet.WorldState)
		// advance block after merge update set
		block++
	}
	updateIter.Release()

	if err := p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
		return fmt.Errorf("cannot prime state-db; %v", err)
	}

	return nil
}
