package profiler

import (
	"github.com/0xsoniclabs/aida/config"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/profile"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/stretchr/testify/assert"
)

func TestReplayProfiler_MakeReplayProfiler(t *testing.T) {
	// case profile true
	cfg := &config.Config{
		Profile: true,
	}
	ext := MakeReplayProfiler[int](cfg, nil)
	_, ok := ext.(executor.Extension[int])
	assert.True(t, ok)

	// case profile false
	cfg = &config.Config{
		Profile: false,
	}
	ext = MakeReplayProfiler[int](cfg, nil)
	_, ok = ext.(extension.NilExtension[int])
	assert.True(t, ok)
}

func TestReplayProfiler_PostRun(t *testing.T) {
	cfg := &config.Config{
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
