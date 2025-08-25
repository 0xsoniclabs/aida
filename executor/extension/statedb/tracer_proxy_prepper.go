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
	"encoding/gob"
	"fmt"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// MakeTracerProxyPrepper creates an extension which
// creates a temporary RecorderProxy before each txcontext
func MakeTracerProxyPrepper[T any](cfg *utils.Config) executor.Extension[T] {
	return makeTracerProxyPrepper[T](cfg)
}

func makeTracerProxyPrepper[T any](cfg *utils.Config) *tracerProxyPrepper[T] {
	return &tracerProxyPrepper[T]{
		cfg: cfg,
	}
}

type tracerProxyPrepper[T any] struct {
	extension.NilExtension[T]
	cfg        *utils.Config
	ctx        tracer.Context
	syncPeriod uint64
}

func (p *tracerProxyPrepper[T]) PreRun(executor.State[T], *executor.Context) error {
	var err error
	writer, err := tracer.NewFileWriter(p.cfg.TraceFile)
	if err != nil {
		return err
	}
	p.ctx, err = tracer.NewContext(writer, p.cfg.First, p.cfg.Last)
	if err != nil {
		return fmt.Errorf("cannot create argument context: %w", err)
	}
	// Register all necessary types to gob
	gob.Register(txcontext.NewNilAccount())
	gob.Register(params.Rules{})
	gob.Register(types.AccessList{})
	gob.Register(types.Log{})
	return nil
}

func (p *tracerProxyPrepper[T]) PreBlock(st executor.State[T], _ *executor.Context) error {
	// Underlying offTheChainStateDb is not yet initialized, so we need to record Block operations manually
	// (State must be initialized in PreTransaction)
	if err := p.ctx.WriteOp(tracer.BeginBlockID, bigendian.Uint64ToBytes(uint64(st.Block))); err != nil {
		return fmt.Errorf("cannot write BeginBlockID: %w", err)
	}
	return nil
}

func (p *tracerProxyPrepper[T]) PreTransaction(_ executor.State[T], ctx *executor.Context) error {
	ctx.State = proxy.NewTracerProxy(ctx.State, p.ctx)
	return nil
}
func (p *tracerProxyPrepper[T]) PostTransaction(_ executor.State[T], ctx *executor.Context) error {
	// Check if the proxy produced any errors
	return ctx.State.Error()
}

func (p *tracerProxyPrepper[T]) PostBlock(executor.State[T], *executor.Context) error {
	if err := p.ctx.WriteOp(tracer.EndBlockID, []byte{}); err != nil {
		return fmt.Errorf("cannot write EndBlockID: %w", err)
	}
	return nil
}

func (p *tracerProxyPrepper[T]) PostRun(_ executor.State[T], _ *executor.Context, _ error) error {
	if p.ctx == nil {
		return nil
	}
	return p.ctx.Close()
}
