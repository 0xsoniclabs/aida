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
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/utils"
)

func TestTemporaryProxyRecorderPrepper_PreTransaction_CreatesProxy(t *testing.T) {
	cfg := &utils.Config{
		TraceFile: t.TempDir() + "test_trace",
	}

	p := MakeTracerProxyPrepper[any](cfg)

	ctx := &executor.Context{}

	err := p.PreRun(executor.State[any]{}, ctx)
	require.NoError(t, err)

	err = p.PreTransaction(executor.State[any]{}, ctx)
	require.NoError(t, err)

	_, ok := ctx.State.(*proxy.TracerProxy)
	assert.True(t, ok, "Proxy was not created in PreRun")
}

func TestTemporaryProxyRecorderPrepper_PostTransaction_ChecksForErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)
	db.EXPECT().Error()

	cfg := &utils.Config{
		TraceFile: t.TempDir() + "test_trace",
	}

	p := MakeTracerProxyPrepper[any](cfg)

	ctx := &executor.Context{
		State: db,
	}
	err := p.PostTransaction(executor.State[any]{}, ctx)
	require.NoError(t, err)
}

func TestTemporaryProxyRecorderPrepper_PostRun_ClosesCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	argCtx := tracer.NewMockArgumentContext(ctrl)
	argCtx.EXPECT().Close()

	cfg := &utils.Config{
		TraceFile: t.TempDir() + "test_trace",
	}

	p := makeTracerProxyPrepper[any](cfg)
	p.ctx = argCtx

	ctx := &executor.Context{}
	err := p.PostRun(executor.State[any]{}, ctx, nil)
	require.NoError(t, err)
}
