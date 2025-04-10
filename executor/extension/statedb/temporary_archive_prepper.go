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
	"github.com/0xsoniclabs/aida/rpc"
)

// MakeTemporaryArchivePrepper creates an extension for retrieving temporary archive before every txcontext.
// Archive is assigned to context.Archive. Archive is released after transaction.
func MakeTemporaryArchivePrepper() executor.Extension[*rpc.RequestAndResults] {
	return &temporaryArchivePrepper{}
}

type temporaryArchivePrepper struct {
	extension.NilExtension[*rpc.RequestAndResults]
}

// PreTransaction creates temporary archive that is released after transaction is executed.
func (r *temporaryArchivePrepper) PreTransaction(state executor.State[*rpc.RequestAndResults], ctx *executor.Context) error {
	var err error
	ctx.Archive, err = ctx.State.GetArchiveState(uint64(state.Data.RequestedBlock))
	if err != nil {
		return err
	}

	return ctx.Archive.BeginTransaction(uint32(state.Transaction))
}

// PostTransaction releases temporary Archive.
func (r *temporaryArchivePrepper) PostTransaction(_ executor.State[*rpc.RequestAndResults], ctx *executor.Context) error {
	if err := ctx.Archive.EndTransaction(); err != nil {
		return fmt.Errorf("cannot end transaction; %w", err)
	}
	return ctx.Archive.Release()
}
