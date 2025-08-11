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

	if p.cfg.PrimeRandom {
		p.log.Infof("Randomized Priming enabled; Seed: %v, threshold: %v", p.cfg.RandomSeed, p.cfg.PrimeThreshold)
	}

	p.log.Infof("Update buffer size: %v bytes", p.cfg.UpdateBufferSize)

	p.ctx = utils.NewPrimeContext(p.cfg, ctx.State, p.log)
	return p.prime(ctx.State, ctx.AidaDb)
}

// getFirstPrimableBlock calculates the first block to prime the stateDb.
func (p *stateDbPrimer[T]) getFirstPrimableBlock(udb db.UpdateDB, sdb db.SubstateDB) (uint64, error) {
	primable := uint64(0) // default value; start priming from block 0

	if p.cfg.IsExistingStateDb {
		stateDbInfo, err := utils.ReadStateDbInfo(p.cfg.StateDbSrc)
		if err != nil {
			return 0, fmt.Errorf("cannot read state db info; %w", err)
		}
		primable = stateDbInfo.Block + 1
	} else {
		substate := sdb.GetFirstSubstate()
		if substate == nil {
			return 0, fmt.Errorf("cannot get first substate; substate db is empty")
		}
		substateFirst := substate.Block
		updateSetFirst, err := udb.GetFirstKey()
		// Update-set may or may not exist. If it does not exist, we set the first block to
		// the largest block in the stateDb.
		if err != nil {
			updateSetFirst = p.cfg.Last // if update-set does not exist, set to largest block
		}
		primable = min(substateFirst, updateSetFirst)
	}
	return primable, nil
}

// mayPrimeFromUpdateSet primes the stateDb from the update-set database if data is available.
func (p *stateDbPrimer[T]) mayPrimeFromUpdateSet(stateDb state.StateDB, block uint64, udb db.UpdateDB) (uint64, error) {
	var (
		totalSize uint64 // total size of unprimed update set
		hasPrimed bool   // if true, db has been primed
	)

	// Primable block is already ahead of the first target block. No priming is needed.
	if block >= p.ctx.GetFirst() {
		return block, nil
	}
	// create iterator starting from the first primable block.
	updateIter := udb.NewUpdateSetIterator(block, p.ctx.GetFirst()-1)
	update := make(substate.WorldState)

	for updateIter.Next() {
		newSet := updateIter.Value()
		if newSet.Block >= p.ctx.GetFirst() {
			break
		}
		block = newSet.Block
		incrementalSize := update.EstimateIncrementalSize(newSet.WorldState)

		// Prime StateDB
		if totalSize+incrementalSize > p.cfg.UpdateBufferSize {
			p.log.Infof("\tPriming...")
			if err := p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
				return block, fmt.Errorf("cannot prime state-db; %v", err)
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
			hasPrimed = false
		}

		update.Merge(newSet.WorldState)
		totalSize += incrementalSize
		p.log.Infof("\tMerge update set at block %v. New total size %v MB (+%v MB)",
			newSet.Block, totalSize/1_000_000,
			incrementalSize/1_000_000)
		// advance next primable block after merge update set
		block++
	}
	updateIter.Release()

	// if update set is not empty, prime the remaining
	if len(update) > 0 {
		if err := p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
			return block, fmt.Errorf("cannot prime state-db; %v", err)
		}
		update = make(substate.WorldState)
	}

	return block, nil
}

// mayPrimeFromSubstate prime from current block to the runnable first block.
func (p *stateDbPrimer[T]) mayPrimeFromSubstate(stateDb state.StateDB, block uint64, aidaDb db.BaseDB) error {
	if block >= p.ctx.GetFirst() {
		return nil
	}
	log.Infof("\tPriming using substate from %v to %v", block, p.ctx.GetFirst()-1)
	update, deletedAccounts, err := utils.GenerateUpdateSet(block, p.ctx.GetFirst()-1, p.cfg, aidaDb)
	if err != nil {
		return fmt.Errorf("cannot generate update-set; %w", err)
	}
	// remove deleted accounts from statedb before priming only if statedb is not empty
	if p.ctx.HasPrimed() {
		p.ctx.SelfDestructAccounts(stateDb, deletedAccounts)
	}
	if err = p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
		return fmt.Errorf("cannot prime state-db; %w", err)
	}
	return nil
}

// prime advances the stateDb to given first block.
// A--B--C, If A is the First block in passed by user, B is the first
// primmable block and C is the first substate (true first block).
// Primming should be able to prime from B to C.
func (p *stateDbPrimer[T]) prime(stateDb state.StateDB, aidaDb db.BaseDB) error {
	var primeBlock uint64

	// load pre-computed update-set from update-set db
	udb := db.MakeDefaultUpdateDBFromBaseDB(aidaDb)
	sdb := db.MakeDefaultSubstateDBFromBaseDB(aidaDb)
	sdb.SetSubstateEncoding(p.cfg.SubstateEncoding)

	// calculate key blocks
	primeBlock, err := p.getFirstPrimableBlock(udb, sdb)
	if err != nil {
		return fmt.Errorf("cannot get first primable block; %w", err)
	}
	substate := sdb.GetFirstSubstate()
	if substate == nil {
		return fmt.Errorf("cannot get first substate; substate db is empty")
	}
	p.ctx.SetFirst(max(substate.Block, p.cfg.First))
	// skip priming
	if primeBlock >= p.ctx.GetFirst() {
		return nil
	}
	p.log.Noticef("Priming from block %v...", primeBlock)
	p.log.Noticef("Priming to block %v...", p.ctx.GetFirst()-1)

	// try advance from update-set
	primeBlock, err = p.mayPrimeFromUpdateSet(stateDb, primeBlock, udb)
	if err != nil {
		return fmt.Errorf("cannot prime from update-set; %w", err)
	}

	// advance from the latest precomputed update-set to the target block
	err = p.mayPrimeFromSubstate(stateDb, primeBlock, aidaDb)

	p.log.Noticef("Delete destroyed accounts until block %v", p.ctx.GetFirst()-1)
	// remove destroyed accounts until one block before the first block
	err = utils.MayDeleteDestroyedAccountsFromStateDB(stateDb, p.cfg, p.ctx.GetFirst()-1, aidaDb)
	if err != nil {
		return fmt.Errorf("cannot delete destroyed accounts from state-db; %v", err)
	}

	return nil
}
