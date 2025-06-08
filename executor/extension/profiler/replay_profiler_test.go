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

func TestMakeReplayProfiler(t *testing.T) {
	// case profile true
	cfg := &utils.Config{
		Profile: true,
	}
	ext := MakeReplayProfiler[int](cfg, nil)
	_, ok := ext.(executor.Extension[int])
	assert.True(t, ok)

	// case profile false
	cfg = &utils.Config{
		Profile: false,
	}
	ext = MakeReplayProfiler[int](cfg, nil)
	_, ok = ext.(extension.NilExtension[int])
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
