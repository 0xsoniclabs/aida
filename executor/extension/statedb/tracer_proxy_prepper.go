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
	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/utils"
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
	ctx        tracer.ArgumentContext
	syncPeriod uint64
}

func (p *tracerProxyPrepper[T]) PreRun(_ executor.State[T], ctx *executor.Context) error {
	var err error
	fh, err := tracer.NewFileWriter(p.cfg.TraceFile)
	if err != nil {
		return err
	}
	p.ctx = tracer.NewArgumentContext(fh)
	// Register all necessary types to gob
	gob.Register(txcontext.NewNilAccount())
	gob.Register(params.Rules{})
	gob.Register(types.AccessList{})
	gob.Register(types.Log{})
	return nil
}
func (p *tracerProxyPrepper[T]) PreTransaction(_ executor.State[T], ctx *executor.Context) error {
	ctx.State = proxy.NewTracerProxy(ctx.State, p.ctx)
	return nil
}
