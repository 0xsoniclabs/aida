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
	sdb                     db.SubstateDB // substate db pointer
}

func (v *stateHashValidator[T]) PreRun(_ executor.State[T], ctx *executor.Context) error {
	if v.cfg.DbImpl == "carmen" {
		if v.cfg.CarmenSchema != 5 {
			return errors.New("state-hash-validation only works with carmen schema 5")
		}

		if v.cfg.ArchiveMode && v.cfg.ArchiveVariant != "s5" {
			return errors.New("archive state-hash-validation only works with archive variant s5")
		}
	} else if v.cfg.DbImpl != "geth" {
		return errors.New("state-hash-validation only works with db-impl carmen or geth")
	}

	// adjust first block to the earliest substate block if available
	// this condition is added for setting sdb in seting.
	if ctx.AidaDb != nil && v.sdb == nil {
		v.sdb = db.MakeDefaultSubstateDBFromBaseDB(ctx.AidaDb)
	}
	if v.sdb != nil {
		sub := v.sdb.GetFirstSubstate()
		if sub != nil {
			block := int(sub.Block)
			if block > v.nextArchiveBlockToCheck {
				v.nextArchiveBlockToCheck = block
			}
		}
	}

	v.hashProvider = utils.MakeHashProvider(ctx.AidaDb)
	return nil
}

func (v *stateHashValidator[T]) PostBlock(state executor.State[T], ctx *executor.Context) error {
	if ctx.State == nil {
		return nil
	}

	want, err := v.getStateHash(state.Block)
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
	if v.cfg.ArchiveMode {
		v.lastProcessedBlock = state.Block
		if err = v.checkArchiveHashes(ctx.State, ctx.AidaDb); err != nil {
			return err
		}
	}

	return nil
}

func (v *stateHashValidator[T]) PostRun(_ executor.State[T], ctx *executor.Context, err error) error {
	// Skip processing if run is aborted due to an error.
	if err != nil {
		return nil
	}
	// Complete processing remaining archive blocks.
	if v.cfg.ArchiveMode {
		for v.nextArchiveBlockToCheck < v.lastProcessedBlock {
			if err = v.checkArchiveHashes(ctx.State, ctx.AidaDb); err != nil {
				return err
			}
			if v.nextArchiveBlockToCheck < v.lastProcessedBlock {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
	return nil
}

func (v *stateHashValidator[T]) checkArchiveHashes(state state.StateDB, aidaDb db.BaseDB) error {
	// Note: the archive may be lagging behind the life DB, so block hashes need
	// to be checked as they become available.

	height, empty, err := state.GetArchiveBlockHeight()
	if err != nil {
		return fmt.Errorf("failed to get archive block height: %v", err)
	}

	cur := uint64(v.nextArchiveBlockToCheck)
	for !empty && cur <= height {
		want, err := v.getStateHash(int(cur))
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

			block, blockErr := v.sdb.GetBlockSubstates(cur)
			if blockErr != nil {
				return fmt.Errorf("cannot get substates for block %d; %v; archive-hash-error: %w", cur, blockErr, unexpectedHashErr)
			}
			// skip check if block is empty, because it could have been trailing an exception block
			if len(block) > 0 {
				return unexpectedHashErr
			}
			v.log.Warningf("Empty block %d has mismatch hash; %v", cur, unexpectedHashErr)

		}

		cur++
	}
	v.nextArchiveBlockToCheck = int(cur)
	return nil
}

func (v *stateHashValidator[T]) getStateHash(blockNumber int) (common.Hash, error) {
	want, err := v.hashProvider.GetStateRootHash(blockNumber)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return common.Hash{}, fmt.Errorf("state hash for block %v is not present in the db", blockNumber)
		}
		return common.Hash{}, fmt.Errorf("cannot get state hash for block %v; %v", blockNumber, err)
	}

	return want, nil

}
