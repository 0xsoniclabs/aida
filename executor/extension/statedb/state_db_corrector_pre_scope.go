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

package statedb

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

// correctorScope defines the scope of exceptions that can be fixed by the updater.
type correctorScope int

// Constants representing different scopes for the state-db corrector.
const (
	preBlock        correctorScope = iota // fix state-db in pre-block scope
	postBlock                             // fix state-db in post-block scope
	preTransaction                        // fix state-db in pre-transaction scope
	postTransaction                       // fix state-db in post-transaction scope
)

// MakeStateDbCorrector creates an extension, which fixes LiveDB with data in Exception database.
func MakeStateDbCorrector(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	log := logger.NewLogger(cfg.LogLevel, "State-DB-Corrector")

	return makeStateDbCorrector(cfg, log)
}

// makeStateDbCorrector creates an exception updater with the given configuration and logger.
func makeStateDbCorrector(cfg *utils.Config, log logger.Logger) executor.Extension[txcontext.TxContext] {
	return &stateDbCorrector{
		cfg: cfg,
		log: log,
	}
}

type stateDbCorrector struct {
	extension.NilExtension[txcontext.TxContext]
	cfg              *utils.Config       // configuration for the updater
	log              logger.Logger       // logger for the updater
	db               db.ExceptionDB      // contains a list of database exceptions
	currentException *substate.Exception // current exception for the block being processed
	nextBlock        int                 // last fixed block, used to track the progress of the updater
}

func (e *stateDbCorrector) PreRun(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if ctx.AidaDb == nil {
		// AidaDb is not set, only occurs during tests
		return nil
	}
	// Initialize the last fixed block to 0, meaning no blocks have been fixed yet.
	e.db = db.MakeDefaultExceptionDBFromBaseDB(ctx.AidaDb)

	return nil
}

func (e *stateDbCorrector) PreBlock(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if ctx.AidaDb == nil {
		return nil
	}

	// initialization of lastFixedBlock
	if e.nextBlock == 0 {
		e.nextBlock = state.Block
	}

	// searching for all blocks that didn't have transactions, if there were any exceptions then fix everything
	for ; e.nextBlock <= state.Block; e.nextBlock++ {
		err := e.loadCurrentException(e.nextBlock)
		if err != nil {
			return fmt.Errorf("failed to load exception for pre block %d: %w", e.nextBlock, err)
		}
		if e.currentException != nil {
			err = e.fixExceptionAt(ctx.State, preBlock, e.nextBlock, utils.PseudoTx)
			if err != nil {
				return fmt.Errorf("failed to fix exception at pre block %d, %w", e.nextBlock, err)
			}
		}
	}

	return nil
}

func (e *stateDbCorrector) PreTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if ctx.AidaDb == nil || e.currentException == nil {
		return nil
	}
	err := e.fixExceptionAt(ctx.State, preTransaction, state.Block, state.Transaction)
	if err != nil {
		return fmt.Errorf("failed to fix exception at block %d, tx %d; %w", state.Block, state.Transaction, err)
	}
	return nil
}

func (e *stateDbCorrector) PostBlock(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	// Sonic: search for skippedTx at the end of the block
	// if block has trailing skipped transactions causing exception this has to be fixed

	// Ethereum: search for pseudoTx containing miner rewards, withdrawals, etc.
	if ctx.AidaDb == nil || e.currentException == nil {
		return nil
	}
	err := e.fixExceptionAt(ctx.State, postBlock, state.Block, utils.PseudoTx)
	if err != nil {
		return fmt.Errorf("failed to fix exception at post block %d; %w", state.Block, err)
	}

	e.currentException = nil // reset current exception after processing the block

	return nil
}

// loadCurrentException loads the current exception for the given block
func (e *stateDbCorrector) loadCurrentException(block int) error {
	exception, err := e.db.GetException(uint64(block))
	if err != nil {
		return fmt.Errorf("failed to get exception for block %d; %w", block, err)
	}

	if exception == nil {
		e.currentException = nil
		return nil
	}

	e.currentException = exception
	return nil
}

// fixExceptionAt applies the exception fix for the given transaction index
func (e *stateDbCorrector) fixExceptionAt(db state.StateDB, scope correctorScope, block int, tx int) error {
	if e.currentException == nil {
		return nil // no exception to fix
	}
	if e.currentException.Block != uint64(block) {
		return fmt.Errorf("current exception block %d does not match state block %d", e.currentException.Block, block)
	}

	ws, err := e.loadStateFromException(scope, tx)
	if err != nil {
		return fmt.Errorf("failed to load state from exception at block %d, tx %d; %w", block, tx, err)
	}
	if ws == nil {
		return nil
	}

	// In Block scope, we need to start a transaction to overwrite the world state.
	if scope == preBlock || scope == postBlock {
		err := db.BeginTransaction(uint32(tx))
		if err != nil {
			return fmt.Errorf("cannot begin transaction; %w", err)
		}
	}

	// Perform the overwrite of state db with the world state from the exception.
	utils.OverwriteStateDB(substatecontext.NewWorldState(ws), db)

	// In Block scope, we need to end the transaction after overwriting the world state.
	if scope == preBlock || scope == postBlock {
		err = db.EndTransaction()
		if err != nil {
			return fmt.Errorf("failed to end transaction at block %d, tx %d; %w", block, tx, err)
		}
	}

	return nil
}

// loadStateFromException retrieves the state from the exception data based on the scope and transaction index
func (e *stateDbCorrector) loadStateFromException(scope correctorScope, tx int) (substate.WorldState, error) {
	switch scope {
	case preBlock:
		if e.currentException.Data.PreBlock != nil {
			return *e.currentException.Data.PreBlock, nil
		}
	case postBlock:
		if e.currentException.Data.PostBlock != nil {
			return *e.currentException.Data.PostBlock, nil
		}
	case preTransaction:
		if _, ok := e.currentException.Data.Transactions[tx]; ok {
			if e.currentException.Data.Transactions[tx].PreTransaction != nil {
				return *e.currentException.Data.Transactions[tx].PreTransaction, nil
			}
		}
	case postTransaction:
		if _, ok := e.currentException.Data.Transactions[tx]; ok {
			if e.currentException.Data.Transactions[tx].PostTransaction != nil {
				return *e.currentException.Data.Transactions[tx].PostTransaction, nil
			}
		}
	default:
		return nil, fmt.Errorf("unknown exception scope: %v", scope)
	}
	return nil, nil
}
