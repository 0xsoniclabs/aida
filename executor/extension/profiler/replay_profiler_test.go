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

package profiler

import (
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/profile"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
)

func TestReplayProfiler_MakeReplayProfiler(t *testing.T) {
	// case profile true
	cfg := &utils.Config{
		Profile: true,
	}
	ext := MakeReplayProfiler[int](cfg, nil)
	assert.NotNil(t, ext)

	// case profile false
	cfg = &utils.Config{
		Profile: false,
	}
	ext = MakeReplayProfiler[int](cfg, nil)
	_, ok := ext.(extension.NilExtension[int])
	assert.True(t, ok)
}

func TestReplayProfiler_PostRun(t *testing.T) {
	cfg := &utils.Config{
		Profile: true,
		First:   0,
		Last:    100,
	}
	rCtx := &context.Replay{
		Stats: &profile.Stats{},
	}
	profiler := MakeReplayProfiler[int](cfg, rCtx)

	state := executor.State[int]{}
	ctx := &executor.Context{}

	err := profiler.PostRun(state, ctx, nil)
	assert.Nil(t, err)
	assert.NotNil(t, rCtx.Stats)
}
