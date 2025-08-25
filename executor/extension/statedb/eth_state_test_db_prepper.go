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
	"os"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/log"
)

func MakeEthStateTestDbPrepper(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	// Disable spam from eth logger when creating database
	log.SetDefault(log.NewLogger(log.DiscardHandler()))
	return makeEthStateTestDbPrepper(logger.NewLogger(cfg.LogLevel, "EthStatePrepper"), cfg)
}

func makeEthStateTestDbPrepper(log logger.Logger, cfg *utils.Config) *ethStateTestDbPrepper {
	return &ethStateTestDbPrepper{
		cfg: cfg,
		log: log,
	}
}

type ethStateTestDbPrepper struct {
	extension.NilExtension[txcontext.TxContext]
	cfg *utils.Config
	log logger.Logger
}

func (e ethStateTestDbPrepper) PreBlock(s executor.State[txcontext.TxContext], ctx *executor.Context) error {
	var err error
	cfg := e.cfg
	// We reduce the node cache size to be used by Carmen to 1 MB
	// reduce the cache creation and flush time, and thus to improve
	// the processing time of each transaction.
	cfg.CarmenStateCacheSize = 1000
	cfg.CarmenNodeCacheSize = (16 << 20) // = 16 MiB
	// Init the chain config with current tests fork
	cfg.ChainCfg, err = cfg.GetChainConfig(s.Data.GetBlockEnvironment().GetFork())
	if err != nil {
		return fmt.Errorf("cannot init chain config: %w", err)
	}
	ctx.State, ctx.StateDbPath, err = utils.PrepareStateDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to prepare statedb; %v", err)
	}

	return nil
}

func (e ethStateTestDbPrepper) PostBlock(_ executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if ctx.State != nil {
		err := ctx.State.Close()
		if err != nil {
			return fmt.Errorf("cannot close db %v; %v", ctx.StateDbPath, err)
		}
	}

	err := os.RemoveAll(ctx.StateDbPath)
	if err != nil {
		return fmt.Errorf("cannot remove db %v; %v", ctx.StateDbPath, err)
	}

	return nil
}
