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
	"path/filepath"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/utils"
)

type liveDbBlockChecker[T any] struct {
	extension.NilExtension[T]
	cfg *utils.Config
}

// MakeLiveDbBlockChecker creates an executor.Extension which checks block alignment of given Live StateDb
func MakeLiveDbBlockChecker[T any](cfg *utils.Config) executor.Extension[T] {
	// this extension is only necessary for existing LiveDb
	if cfg.StateDbSrc == "" {
		return extension.NilExtension[T]{}
	}

	return &liveDbBlockChecker[T]{
		cfg: cfg,
	}
}

// PreRun checks existing LiveDbs block alignment
func (c *liveDbBlockChecker[T]) PreRun(executor.State[T], *executor.Context) error {
	var (
		primeDbInfo  utils.StateDbInfo
		shadowDbInfo utils.StateDbInfo
		lastBlock    uint64
		err          error
	)

	if c.cfg.ShadowDb {
		primeDbInfo, err = utils.ReadStateDbInfo(filepath.Join(c.cfg.StateDbSrc, utils.PathToPrimaryStateDb))
		if err != nil {
			return fmt.Errorf("cannot read state db info for primary db; %v", err)
		}

		shadowDbInfo, err = utils.ReadStateDbInfo(filepath.Join(c.cfg.StateDbSrc, utils.PathToShadowStateDb))
		if err != nil {
			return fmt.Errorf("cannot read state db info for shadow db; %v", err)
		}

		// both shadow and prime dbs have to contain same last block
		if shadowDbInfo.Block != primeDbInfo.Block {
			return fmt.Errorf("shadow (%v) and prime (%v) state dbs have different last block", primeDbInfo.Block, shadowDbInfo.Block)
		}

		lastBlock = primeDbInfo.Block

	} else {
		primeDbInfo, err = utils.ReadStateDbInfo(c.cfg.StateDbSrc)
		if err != nil {
			return fmt.Errorf("cannot read state db info; %v", err)
		}

		lastBlock = primeDbInfo.Block
	}

	// ethereum doesn't support priming
	if c.cfg.ChainID == 1 {
		// first block must be exactly +1 so data aligns correctly
		if lastBlock+1 != c.cfg.First {
			return fmt.Errorf("if using existing live-db with vm-sdb first block needs to be last block of live-db + 1, in your case %v", lastBlock+1)
		}
	} else if lastBlock >= c.cfg.First {
		// Ignore if run was not finished - db might have been healed so block alignment would not work
		if !primeDbInfo.HasFinished {
			return nil
		}

		// user incorrectly tries to prime data into database even tho database is already advanced further
		return fmt.Errorf("if using existing live-db with vm-sdb first block needs to be higher than last block of live-db, in your case %v", lastBlock+1)
	}

	return nil
}
