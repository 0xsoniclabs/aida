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
	"errors"
	"fmt"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
)

func MakeStateHashValidator[T any](cfg *utils.Config) executor.Extension[T] {
	// todo make true when --validate is chosen (validate should enable all validations)

	if !cfg.ValidateStateHashes {
		return extension.NilExtension[T]{}
	}

	log := logger.NewLogger("INFO", "state-hash-validator")
	return makeStateHashValidator[T](cfg, log)
}

func makeStateHashValidator[T any](cfg *utils.Config, log logger.Logger) *stateHashValidator[T] {
	return &stateHashValidator[T]{cfg: cfg, log: log, nextArchiveBlockToCheck: int(cfg.First)}
}

type stateHashValidator[T any] struct {
	extension.NilExtension[T]
	cfg                     *utils.Config
	log                     logger.Logger
	nextArchiveBlockToCheck int
	lastProcessedBlock      int
	hashProvider            utils.HashProvider
	excDb                   db.ExceptionDB
}

func (e *stateHashValidator[T]) PreRun(_ executor.State[T], ctx *executor.Context) error {
	if e.cfg.DbImpl == "carmen" {
		if e.cfg.CarmenSchema != 5 {
			return errors.New("state-hash-validation only works with carmen schema 5")
		}

		if e.cfg.ArchiveMode && e.cfg.ArchiveVariant != "s5" {
			return errors.New("archive state-hash-validation only works with archive variant s5")
		}
	} else if e.cfg.DbImpl != "geth" {
		return errors.New("state-hash-validation only works with db-impl carmen or geth")
	}

	e.hashProvider = utils.MakeHashProvider(ctx.AidaDb)
	e.excDb = db.MakeDefaultExceptionDBFromBaseDB(ctx.AidaDb)
	return nil
}

func (e *stateHashValidator[T]) PostBlock(state executor.State[T], ctx *executor.Context) error {
	if ctx.State == nil {
		return nil
	}

	want, err := e.getStateHash(state.Block)
	if err != nil {
		return err
	}

	// NOTE: ContinueOnFailure does not make sense here, if hash does not
	// match every block after this block would have different hash
	got, err := ctx.State.GetHash()
	if err != nil {
		return fmt.Errorf("cannot get state hash; %w", err)
	}
	if want != got {
		return fmt.Errorf("unexpected hash for Live block %d\nwanted %v\n   got %v", state.Block, want, got)
	}

	// Check the ArchiveDB
	if e.cfg.ArchiveMode {
		e.lastProcessedBlock = state.Block
		if err = e.checkArchiveHashes(ctx.State); err != nil {
			return err
		}
	}

	return nil
}

func (e *stateHashValidator[T]) PostRun(_ executor.State[T], ctx *executor.Context, err error) error {
	// Skip processing if run is aborted due to an error.
	if err != nil {
		return nil
	}
	// Complete processing remaining archive blocks.
	if e.cfg.ArchiveMode {
		for e.nextArchiveBlockToCheck < e.lastProcessedBlock {
			if err = e.checkArchiveHashes(ctx.State); err != nil {
				return err
			}
			if e.nextArchiveBlockToCheck < e.lastProcessedBlock {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
	return nil
}

func (e *stateHashValidator[T]) checkArchiveHashes(state state.StateDB) error {
	// Note: the archive may be lagging behind the life DB, so block hashes need
	// to be checked as they become available.
	height, empty, err := state.GetArchiveBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get archive block height: %v", err)
	}

	cur := uint64(e.nextArchiveBlockToCheck)
	for !empty && cur <= height {
		want, err := e.getStateHash(int(cur))
		if err != nil {
			return err
		}

		archive, err := state.GetArchiveState(cur)
		if err != nil {
			return err
		}

		// NOTE: ContinueOnFailure does not make sense here, if hash does not
		// match every block after this block would have different hash
		got, err := archive.GetHash()
		archive.Release()
		if err != nil {
			return fmt.Errorf("cannot GetHash; %w", err)
		}
		if want != got {
			unexpectedHashErr := fmt.Errorf("unexpected hash for archive block %d\nwanted %v\n   got %v", cur, want, got)
			// verify that this block is not an exception
			exc, errExc := e.excDb.GetException(cur)
			if errExc != nil {
				return fmt.Errorf("cannot get exception for archive block %d; %v; archive-hash-error: %w", cur, errExc, unexpectedHashErr)
			}
			if exc == nil {
				// this is not an exception, so we return the error
				return unexpectedHashErr
			}
			// we need to skip the whole range of checking loop, because the exception could have following blank blocks,
			// which would not have the exception in them, but would still contain wrong state hash
			// cur still needs to be set to height-1 to check the currently processed block
			cur = height - 1
		}

		cur++
	}
	e.nextArchiveBlockToCheck = int(cur)
	return nil
}

func (e *stateHashValidator[T]) getStateHash(blockNumber int) (common.Hash, error) {
	want, err := e.hashProvider.GetStateRootHash(blockNumber)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return common.Hash{}, fmt.Errorf("state hash for block %v is not present in the db", blockNumber)
		}
		return common.Hash{}, fmt.Errorf("cannot get state hash for block %v; %v", blockNumber, err)
	}

	return want, nil

}
