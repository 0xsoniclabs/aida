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
)

// txGeneratorBlockEventEmitter is an extension to call BeginBlock() and EndBlock()
// for tx generator.
type txGeneratorBlockEventEmitter[T any] struct {
	extension.NilExtension[T]
	lastBlock *uint64
}

// MakeTxGeneratorBlockEventEmitter creates a executor.Extension to call BeginBlock() and EndBlock()
// for tx generator
func MakeTxGeneratorBlockEventEmitter[T any]() executor.Extension[T] {
	return &txGeneratorBlockEventEmitter[T]{}
}

func (l *txGeneratorBlockEventEmitter[T]) PreTransaction(state executor.State[T], ctx *executor.Context) error {
	// if last block is nil, begin block for the current block
	// this is to ensure that the block is started before the first transaction
	if l.lastBlock == nil {
		err := ctx.State.BeginBlock(uint64(state.Block))
		if err != nil {
			return fmt.Errorf("cannot begin block; %w", err)
		}
		blk := uint64(state.Block)
		l.lastBlock = &blk
	} else if *l.lastBlock != uint64(state.Block) {
		// if the last block is not equal to the current block, end the last block
		// and begin the current block
		err := ctx.State.EndBlock()
		if err != nil {
			return fmt.Errorf("cannot begin block; %w", err)
		}
		err = ctx.State.BeginBlock(uint64(state.Block))
		if err != nil {
			return fmt.Errorf("cannot begin block; %w", err)
		}
		blk := uint64(state.Block)
		l.lastBlock = &blk
	}
	err := ctx.State.BeginTransaction(uint32(state.Transaction))
	if err != nil {
		return fmt.Errorf("cannot begin transaction; %w", err)
	}
	return nil
}

func (l *txGeneratorBlockEventEmitter[T]) PostTransaction(_ executor.State[T], ctx *executor.Context) error {
	err := ctx.State.EndTransaction()
	if err != nil {
		return fmt.Errorf("cannot end transaction; %w", err)
	}
	return nil
}
func (l *txGeneratorBlockEventEmitter[T]) PostRun(state executor.State[T], ctx *executor.Context, _ error) error {
	// end the last block
	if l.lastBlock != nil {
		ctx.State.EndBlock()
	}
	return nil
}
