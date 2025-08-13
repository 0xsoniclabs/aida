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

package prime

import (
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/martian/log"
)

func MakePrimer(cfg *utils.Config, state state.StateDB, log logger.Logger) *primer {
	return &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, state, log),
	}
}

type primer struct {
	cfg   *utils.Config
	log   logger.Logger
	ctx   *PrimeContext
	block uint64 // current block number used for priming
	first uint64 // first block that can be processed
}

func (p *primer) setFirstPrimableBlock(udb db.UpdateDB, sdb db.SubstateDB) error {

	if p.cfg.IsExistingStateDb {
		stateDbInfo, err := utils.ReadStateDbInfo(p.cfg.StateDbSrc)
		if err != nil {
			return fmt.Errorf("cannot read state db info; %w", err)
		}
		p.block = stateDbInfo.Block + 1
	} else {
		substate := sdb.GetFirstSubstate()
		if substate == nil {
			return fmt.Errorf("cannot get first substate; substate db is empty")
		}
		substateFirst := substate.Block
		updateSetFirst, err := udb.GetFirstKey()
		// Update-set may or may not exist. If it does not exist, use the first substate block to avoid error.
		if err != nil {
			updateSetFirst = substateFirst
		}
		// Choose the minimum of substateFirst and updateSetFirst to ensure priming starts from the earliest available block,
		// as both sources may have different starting points and we want to cover all possible data.
		p.block = min(substateFirst, updateSetFirst)
	}
	return nil
}

// mayPrimeFromUpdateSet primes the stateDb from the update-set database if data is available.
func (p *primer) mayPrimeFromUpdateSet(stateDb state.StateDB, udb db.UpdateDB) error {
	var (
		totalSize uint64 // total size of unprimed update set
		hasPrimed bool   // if true, db has been primed
	)

	// Primable block is already ahead of the first target block. No priming is needed.
	if p.block >= p.first {
		return nil
	}
	// create iterator starting from the first primable block.
	updateIter := udb.NewUpdateSetIterator(p.block, p.first-1)
	defer updateIter.Release()
	update := make(substate.WorldState)

	for updateIter.Next() {
		newSet := updateIter.Value()
		if newSet.Block >= p.first {
			break
		}
		p.block = newSet.Block
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
		ClearAccountStorage(update, newSet.DeletedAccounts)
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
		p.block++
	}

	if len(update) > 0 {
		if err := p.ctx.PrimeStateDB(substatecontext.NewWorldState(update), stateDb); err != nil {
			return fmt.Errorf("cannot prime state-db; %v", err)
		}
	}

	return nil
}

// mayPrimeFromSubstate prime from current block to the runnable first block.
func (p *primer) mayPrimeFromSubstate(stateDb state.StateDB, aidaDb db.BaseDB) error {
	if p.block >= p.first {
		return nil
	}
	log.Infof("\tPriming using substate from %v to %v", p.block, p.first-1)
	update, deletedAccounts, err := generateUpdateSet(p.block, p.first-1, p.cfg, aidaDb)
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
func (p *primer) Prime(stateDb state.StateDB, aidaDb db.BaseDB) error {
	// load pre-computed update-set from update-set db
	udb := db.MakeDefaultUpdateDBFromBaseDB(aidaDb)
	sdb := db.MakeDefaultSubstateDBFromBaseDB(aidaDb)
	sdb.SetSubstateEncoding(p.cfg.SubstateEncoding)

	err := p.setFirstPrimableBlock(udb, sdb)
	if err != nil {
		return fmt.Errorf("cannot get first primable block; %w", err)
	}
	substate := sdb.GetFirstSubstate()
	if substate == nil {
		return fmt.Errorf("cannot get first substate; substate db is empty")
	}
	p.first = max(substate.Block, p.cfg.First)
	// skip priming
	if p.block >= p.first {
		return nil
	}
	p.log.Noticef("Priming from block %v...", p.block)
	p.log.Noticef("Priming to block %v...", p.first-1)

	// try advance from update-set
	err = p.mayPrimeFromUpdateSet(stateDb, udb)
	if err != nil {
		return fmt.Errorf("cannot prime from update-set; %w", err)
	}

	// advance from the latest precomputed update-set to the target block
	err = p.mayPrimeFromSubstate(stateDb, aidaDb)
	if err != nil {
		return fmt.Errorf("cannot prime from substate; %w", err)
	}

	p.log.Noticef("Delete destroyed accounts until block %v", p.first-1)
	err = p.mayDeleteDestroyedAccountsFromStateDB(p.first-1, aidaDb)
	if err != nil {
		return fmt.Errorf("cannot delete destroyed accounts from state-db; %v", err)
	}
	return nil
}

// mayDeleteDestroyedAccountsFromStateDB performs suicide operations on previously
// self-destructed accounts.
func (p *primer) mayDeleteDestroyedAccountsFromStateDB(target uint64, aidaDb db.BaseDB) error {
	log := logger.NewLogger(p.cfg.LogLevel, "DelDestAcc")

	src := db.MakeDefaultDestroyedAccountDBFromBaseDB(aidaDb)
	accounts, err := src.GetAccountsDestroyedInRange(0, target)
	if err != nil {
		return err
	}
	log.Noticef("Deleting %d accounts ...", len(accounts))
	if len(accounts) == 0 {
		// nothing to delete, skip
		return nil
	}
	sdb := p.ctx.db
	sdb.BeginSyncPeriod(0)
	err = sdb.BeginBlock(target)
	if err != nil {
		return err
	}
	err = sdb.BeginTransaction(0)
	if err != nil {
		return err
	}
	for _, addr := range accounts {
		sdb.SelfDestruct(common.Address(addr))
		log.Debugf("Perform suicide on %v", addr)
	}
	err = sdb.EndTransaction()
	if err != nil {
		return err
	}
	err = sdb.EndBlock()
	if err != nil {
		return err
	}
	sdb.EndSyncPeriod()
	return nil
}
