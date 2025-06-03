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

package validator

import (
	"fmt"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
)

type exceptionScope int

const (
	preBlock        exceptionScope = iota // pre-block exception scope
	postBlock                             // post-block exception scope
	preTransaction                        // pre-transaction exception scope
	postTransaction                       // post-transaction exception scope
)

// MakeExceptionUpdater creates an extension, which fixes Exception in LiveDB
func MakeExceptionUpdater(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	log := logger.NewLogger(cfg.LogLevel, "Exception-Updater")

	return makeExceptionUpdater(cfg, log)
}

func makeExceptionUpdater(cfg *utils.Config, log logger.Logger) executor.Extension[txcontext.TxContext] {
	return &exceptionUpdater{
		cfg: cfg,
		log: log,
	}
}

type exceptionUpdater struct {
	extension.NilExtension[txcontext.TxContext]
	cfg                   *utils.Config
	log                   logger.Logger
	db                    db.ExceptionDB
	currentBlockException *substate.Exception
	lastFixedBlock        int
}

func (e *exceptionUpdater) PreTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	// Sonic: fixing exceptions caused by skipped transactions during the block
	if e.currentBlockException != nil {
		err := e.fixExceptionAt(ctx.State, preTransaction, state.Block, state.Transaction, false)
		if err != nil {
			return fmt.Errorf("failed to fix exception at block %d, tx %d: %w", state.Block, state.Transaction, err)
		}
	}
	return nil
}

func (e *exceptionUpdater) PreRun(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	e.log.Warning("Exception updater is enabled.")

	e.db = db.MakeDefaultExceptionDBFromBaseDB(ctx.AidaDb)

	return nil
}

func (e *exceptionUpdater) PreBlock(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if e.lastFixedBlock == 0 {
		// initialization of lastFixedBlock
		e.lastFixedBlock = state.Block
	} else {
		// incrementing lastFixedBlock to the current block
		e.lastFixedBlock++
	}

	// searching for all blocks that didn't have transactions, if there were any exceptions then fix everything
	for ; e.lastFixedBlock < state.Block; e.lastFixedBlock++ {
		err := e.loadCurrentException(e.lastFixedBlock)
		if err != nil {
			return fmt.Errorf("failed to load exception for pre block %d: %w", e.lastFixedBlock, err)
		}
		if e.currentBlockException != nil {
			err = e.fixExceptionAt(ctx.State, preBlock, e.lastFixedBlock, 0, true)
			if err != nil {
				return fmt.Errorf("failed to fix exception at pre block %d, %w", e.lastFixedBlock, err)
			}
		}
	}

	if e.lastFixedBlock != state.Block {
		return fmt.Errorf("internal error: last fixed block %d does not match state block %d", e.lastFixedBlock, state.Block)
	}

	// load exception for the current block
	err := e.loadCurrentException(state.Block)
	if err != nil {
		return err
	}

	return nil
}

func (e *exceptionUpdater) PostBlock(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	// Sonic: search for skippedTx at the end of the block
	// if block has trailing skipped transactions causing exception this has to be fixed

	// Ethereum: search for pseudoTx containing miner rewards, withdrawals, etc.
	if e.currentBlockException != nil {
		err := e.fixExceptionAt(ctx.State, postBlock, state.Block, 0, true)
		if err != nil {
			return fmt.Errorf("failed to fix exception at post block %d, %w", state.Block, err)
		}
	}

	return nil
}

// loadCurrentException loads the exception for the current block
func (e *exceptionUpdater) loadCurrentException(block int) error {
	e.lastFixedBlock = block
	var err error
	if e.currentBlockException == nil || e.currentBlockException.Block != uint64(block) {
		e.currentBlockException, err = e.db.GetException(uint64(block))
	}
	return err
}

// fixExceptionAt applies the exception fix for the given transaction index
func (e *exceptionUpdater) fixExceptionAt(db state.StateDB, scope exceptionScope, block int, tx int, wrapInTx bool) error {
	if e.currentBlockException.Block != uint64(block) {
		return fmt.Errorf("current exception block %d does not match state block %d", e.currentBlockException.Block, block)
	}

	alloc, err := e.loadAllocFromException(scope, tx)
	if err != nil {
		return fmt.Errorf("failed to load allocation from exception at block %d, tx %d: %w", e.lastFixedBlock, tx, err)
	}
	if alloc == nil || len(alloc) == 0 {
		return nil
	}

	if wrapInTx {
		err := db.BeginTransaction(uint32(0))
		if err != nil {
			return fmt.Errorf("cannot begin transaction; %w", err)
		}
	}
	err = overwriteWorldState(e.cfg, substatecontext.NewWorldState(alloc), db)
	if err != nil {
		return fmt.Errorf("failed to overwrite world state at block %d, tx %d: %w", e.lastFixedBlock, tx, err)
	}
	if wrapInTx {
		err = db.EndTransaction()
		if err != nil {
			return fmt.Errorf("failed to end transaction at block %d, tx %d: %w", e.lastFixedBlock, tx, err)
		}
	}

	return nil
}

// loadAllocFromException retrieves the allocation from the exception data based on the scope and transaction index
func (e *exceptionUpdater) loadAllocFromException(scope exceptionScope, tx int) (substate.WorldState, error) {
	var alloc substate.WorldState
	switch {
	case scope == preBlock:
		alloc = *e.currentBlockException.Data.PreBlock
	case scope == postBlock:
		alloc = *e.currentBlockException.Data.PostBlock
	case scope == preTransaction:
		if _, ok := e.currentBlockException.Data.Transactions[tx]; !ok {
			return nil, nil
		}
		alloc = *e.currentBlockException.Data.Transactions[tx].PreTransaction
	case scope == postTransaction:
		if _, ok := e.currentBlockException.Data.Transactions[tx]; !ok {
			return nil, nil
		}
		alloc = *e.currentBlockException.Data.Transactions[tx].PostTransaction
	default:
		return nil, fmt.Errorf("unknown exception scope: %v", scope)
	}
	if alloc == nil {
		return nil, fmt.Errorf("no allocation found for exception scope %v and transaction %d in block %d", scope, tx, e.lastFixedBlock)
	}

	return alloc, nil
}
