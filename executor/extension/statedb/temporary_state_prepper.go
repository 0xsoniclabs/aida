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
	statedb "github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
)

// MakeTemporaryStatePrepper creates an executor.Extension which Makes a fresh StateDb
// after each txcontext. Default is offTheChainStateDb.
// NOTE: inMemoryStateDb currently does not work for block 67m onwards.
func MakeTemporaryStatePrepper(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	switch cfg.DbImpl {
	case "in-memory", "memory":
		return &temporaryInMemoryStatePrepper{}
	case "off-the-chain":
		fallthrough
	default:
		// offTheChainStateDb is default value
		return &temporaryOffTheChainStatePrepper{cfg: cfg}
	}
}

// temporaryInMemoryStatePrepper is an extension that introduces a fresh in-memory
// StateDB instance before each transaction execution.
type temporaryInMemoryStatePrepper struct {
	extension.NilExtension[txcontext.TxContext]
}

func (p *temporaryInMemoryStatePrepper) PreTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	alloc := state.Data.GetInputState()
	ctx.State = statedb.MakeInMemoryStateDB(alloc, uint64(state.Block))
	return nil
}

// temporaryOffTheChainStatePrepper is an extension that introduces a fresh offTheChain
// StateDB instance before each transaction execution.
type temporaryOffTheChainStatePrepper struct {
	extension.NilExtension[txcontext.TxContext]
	cfg          *utils.Config
	chainConduit *statedb.ChainConduit
}

func (p *temporaryOffTheChainStatePrepper) PreRun(executor.State[txcontext.TxContext], *executor.Context) error {
	chainCfg, err := p.cfg.GetChainConfig("")
	if err != nil {
		return fmt.Errorf("cannot get chain config: %w", err)
	}
	p.chainConduit = statedb.NewChainConduit(utils.IsEthereumNetwork(p.cfg.ChainID), chainCfg)
	return nil
}

func (p *temporaryOffTheChainStatePrepper) PreTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	var err error
	ctx.State, err = statedb.MakeOffTheChainStateDB(state.Data.GetInputState(), uint64(state.Block), p.chainConduit)
	return err
}
