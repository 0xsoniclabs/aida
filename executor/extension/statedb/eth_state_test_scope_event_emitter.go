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

package statedb

import (
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/txcontext"
)

func MakeEthStateScopeTestEventEmitter() executor.Extension[txcontext.TxContext] {
	return ethStateScopeEventEmitter{}
}

type ethStateScopeEventEmitter struct {
	extension.NilExtension[txcontext.TxContext]
}

func (e ethStateScopeEventEmitter) PreBlock(s executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if err := ctx.State.BeginBlock(uint64(s.Block)); err != nil {
		return err
	}
	return ctx.State.BeginTransaction(uint32(s.Transaction))
}

func (e ethStateScopeEventEmitter) PostBlock(_ executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if err := ctx.State.EndTransaction(); err != nil {
		return err
	}
	return ctx.State.EndBlock()
}
