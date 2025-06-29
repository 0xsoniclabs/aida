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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	gc "github.com/ethereum/go-ethereum/common"
)

// MakeStateDbManager creates a executor.Extension that commits state of StateDb if keep-db is enabled
func MakeStateDbManager[T any](cfg *utils.Config, knownDbPath string) executor.Extension[T] {
	return &stateDbManager[T]{
		cfg:    cfg,
		log:    logger.NewLogger(cfg.LogLevel, "Db manager"),
		dbPath: knownDbPath,
	}
}

type stateDbManager[T any] struct {
	extension.NilExtension[T]
	cfg    *utils.Config
	log    logger.Logger
	dbPath string // state db path if the  db is created out side of this extension
}

func (m *stateDbManager[T]) PreRun(_ executor.State[T], ctx *executor.Context) error {
	var err error
	if ctx.State == nil {
		ctx.State, ctx.StateDbPath, err = utils.PrepareStateDB(m.cfg)
		if err != nil {
			return err
		}
	} else {
		ctx.StateDbPath = m.dbPath
	}

	if !m.cfg.ShadowDb {
		m.logDbMode("Db Implementation", m.cfg.DbImpl, m.cfg.DbVariant)
	} else {
		m.logDbMode("Prime Db Implementation", m.cfg.DbImpl, m.cfg.DbVariant)
		m.logDbMode("Shadow Db Implementation", m.cfg.ShadowImpl, m.cfg.ShadowVariant)
	}

	if m.cfg.StateDbSrc != "" {
		// if using pre-existing StateDb and running in read-only mode, we must report both source db and working tmp dir
		m.log.Infof("Source storage directory: %v", m.cfg.StateDbSrc)
		if m.cfg.StateDbSrcDirectAccess {
			m.log.Infof("Working storage directory: %v", m.cfg.DbTmp)
		}

	} else {
		// otherwise only working directory is reported
		m.log.Infof("Working storage directory: %v", m.cfg.DbTmp)
	}

	if m.cfg.ArchiveMode {
		var archiveVariant string
		if m.cfg.ArchiveVariant == "" {
			archiveVariant = "<implementation-default>"
		} else {
			archiveVariant = m.cfg.ArchiveVariant
		}

		m.log.Noticef("Archive mode enabled; Variant: %v", archiveVariant)

	} else {
		m.log.Infof("Archive mode disabled")
	}

	if !m.cfg.KeepDb && !m.cfg.StateDbSrcDirectAccess {
		m.log.Warningf("--keep-db is not used. Directory %v with DB will be removed at the end of this run.", ctx.StateDbPath)
	}

	// Set state-db info to incomplete state at the beginning of the run.
	// If state-db info exists, read block number and hash from it.
	var blockNum uint64
	var blockHash gc.Hash
	if m.cfg.IsExistingStateDb {
		dbinfo, err := utils.ReadStateDbInfo(ctx.StateDbPath)
		if err != nil {
			return fmt.Errorf("failed to read state-db info file; %v", err)
		}
		blockNum = dbinfo.Block
		m.log.Infof("Resuming from block %v", blockNum)
	}

	// Mark state-db info with incomplete state.
	if err := utils.WriteStateDbInfo(ctx.StateDbPath, m.cfg, blockNum, blockHash, false); err != nil {
		return fmt.Errorf("failed to create state-db info file; %v", err)
	}
	return nil
}

func (m *stateDbManager[T]) PostRun(state executor.State[T], ctx *executor.Context, _ error) error {
	//  if state was not correctly initialized remove the stateDbPath and abort
	if ctx.State == nil {
		var err = fmt.Errorf("state-db is nil")
		if !m.cfg.StateDbSrcDirectAccess {
			err = errors.Join(err, os.RemoveAll(ctx.StateDbPath))
		}
		return err
	}

	// get root hash before closing db
	rootHash, err := ctx.State.GetHash()
	if err != nil {
		return fmt.Errorf("cannot get state hash; %w", err)
	}

	start := time.Now()
	if err := ctx.State.Close(); err != nil {
		return fmt.Errorf("failed to close state-db; %v", err)
	}
	m.log.Noticef("DB close time: %v seconds", time.Since(start).Round(time.Second))

	// db was not modified, then close db without chnaging state-db info and keep db folder as-is.
	if m.cfg.StateDbSrcReadOnly {
		m.log.Noticef("State-db directory was read-only %v. No updates to state-db info", ctx.StateDbPath)
		return nil
	}

	// if db isn't kept and db was not modified in-place, then close and delete temporary state-db.
	if !m.cfg.KeepDb && !m.cfg.StateDbSrcDirectAccess {
		return os.RemoveAll(ctx.StateDbPath)
	}

	// Db is kept after run. Rename db folder and write meta information to a file.
	// lastProcessedBlock contains number of last successfully processed block
	// - processing finished successfully to the end, but then state.Block is set to params.To
	// - error occurred therefore previous block is last successful
	lastProcessedBlock := uint64(state.Block)
	if lastProcessedBlock > 0 {
		lastProcessedBlock -= 1
	}

	// write state db info
	if err := utils.WriteStateDbInfo(ctx.StateDbPath, m.cfg, lastProcessedBlock, rootHash, true); err != nil {
		return fmt.Errorf("failed to create state-db info file; %v", err)
	}
	// if db was modified in-place, no need to rename state-db folder.
	if !m.cfg.StateDbSrcDirectAccess {
		newName := utils.RenameTempStateDbDirectory(m.cfg, ctx.StateDbPath, lastProcessedBlock)
		m.log.Noticef("State-db directory: %v", newName)
	}
	return nil
}

func (m *stateDbManager[T]) logDbMode(prefix, impl, variant string) {
	if m.cfg.DbImpl == "carmen" {
		m.log.Noticef("%s: %v; Variant: %v, Carmen Schema: %d", prefix, impl, variant, m.cfg.CarmenSchema)
	} else {
		m.log.Noticef("%s: %v; Variant: %v", prefix, impl, variant)
	}
}
