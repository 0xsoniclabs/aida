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

package profiler

import (
	"fmt"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/profile/blockprofile"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
)

func MakeBlockRuntimeAndGasCollector(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	if !cfg.ProfileBlocks {
		return extension.NilExtension[txcontext.TxContext]{}
	}
	return &BlockRuntimeAndGasCollector{
		cfg: cfg,
		log: logger.NewLogger(cfg.LogLevel, "Block-Profile"),
	}
}

type BlockRuntimeAndGasCollector struct {
	extension.NilExtension[txcontext.TxContext]
	log        logger.Logger
	cfg        *utils.Config
	profileDb  *blockprofile.ProfileDB
	ctx        *blockprofile.Context
	blockTimer time.Time
	txTimer    time.Time
}

// PreRun prepares the ProfileDB
func (b *BlockRuntimeAndGasCollector) PreRun(executor.State[txcontext.TxContext], *executor.Context) error {
	var err error
	b.profileDb, err = blockprofile.NewProfileDB(b.cfg.ProfileDB)
	if err != nil {
		return fmt.Errorf("cannot create profile-db; %v", err)
	}

	b.log.Notice("Deleting old data from ProfileDB")
	_, err = b.profileDb.DeleteByBlockRange(b.cfg.First, b.cfg.Last)
	if err != nil {
		return fmt.Errorf("cannot delete old data from profile-db; %v", err)
	}

	return nil
}

// PreTransaction resets the transaction timer.
func (b *BlockRuntimeAndGasCollector) PreTransaction(executor.State[txcontext.TxContext], *executor.Context) error {
	b.txTimer = time.Now()
	return nil
}

// PostTransaction records tx into profile context.
func (b *BlockRuntimeAndGasCollector) PostTransaction(state executor.State[txcontext.TxContext], _ *executor.Context) error {
	err := b.ctx.RecordTransaction(state, time.Since(b.txTimer))
	if err != nil {
		return fmt.Errorf("cannot record transaction; %v", err)
	}
	return nil
}

// PreBlock resets the block times and profile context.
func (b *BlockRuntimeAndGasCollector) PreBlock(executor.State[txcontext.TxContext], *executor.Context) error {
	b.ctx = blockprofile.NewContext()
	b.blockTimer = time.Now()
	return nil
}

// PostBlock extracts data from profile context and writes them to ProfileDB.
func (b *BlockRuntimeAndGasCollector) PostBlock(state executor.State[txcontext.TxContext], _ *executor.Context) error {
	data, err := b.ctx.GetProfileData(uint64(state.Block), time.Since(b.blockTimer))
	if err != nil {
		return fmt.Errorf("cannot get profile data from context; %v", err)
	}

	err = b.profileDb.Add(*data)
	if err != nil {
		return fmt.Errorf("cannot add data to profile-db; %v", err)
	}

	return nil
}

// PostRun closes ProfileDB
func (b *BlockRuntimeAndGasCollector) PostRun(executor.State[txcontext.TxContext], *executor.Context, error) error {
	defer func() {
		if r := recover(); r != nil {
			b.log.Errorf("recovered panic in block-profiler; %v", r)
		}
	}()

	err := b.profileDb.Close()
	if err != nil {
		return fmt.Errorf("cannot close profile-db; %v", err)
	}

	return nil
}
