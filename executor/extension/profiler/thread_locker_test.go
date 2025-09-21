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
	"github.com/stretchr/testify/assert"
)

func TestThreadLocker_MakeThreadLocker(t *testing.T) {
	ext := MakeThreadLocker[int]()
	_, ok := ext.(executor.Extension[int])
	assert.True(t, ok)
}

func TestThreadLocker_PreRun(t *testing.T) {
	locker := threadLocker[int]{}
	var state executor.State[int] // Can be nil or a mock if needed for other tests
	var ctx *executor.Context     // Can be nil or a mock

	err := locker.PreRun(state, ctx)
	assert.Nil(t, err, "PreRun should not return an error")
}

func TestThreadLocker_PostRun(t *testing.T) {
	locker := threadLocker[int]{}
	var state executor.State[int]
	var ctx *executor.Context
	var runError error // Can be nil or an actual error

	err := locker.PostRun(state, ctx, runError)
	assert.Nil(t, err, "PostRun should not return an error")
}

func TestThreadLocker_PreBlock(t *testing.T) {
	locker := threadLocker[int]{}
	var state executor.State[int]
	var ctx *executor.Context

	err := locker.PreBlock(state, ctx)
	assert.Nil(t, err, "PreBlock should not return an error")
}

func TestThreadLocker_PostBlock(t *testing.T) {
	locker := threadLocker[int]{}
	var state executor.State[int]
	var ctx *executor.Context

	err := locker.PostBlock(state, ctx)
	assert.Nil(t, err, "PostBlock should not return an error")
}

func TestThreadLocker_PreTransaction(t *testing.T) {
	locker := threadLocker[int]{}
	var state executor.State[int]
	var ctx *executor.Context

	err := locker.PreTransaction(state, ctx)
	assert.Nil(t, err, "PreTransaction should not return an error")
}

func TestThreadLocker_PostTransaction(t *testing.T) {
	locker := threadLocker[int]{}
	var state executor.State[int]
	var ctx *executor.Context

	err := locker.PostTransaction(state, ctx)
	assert.Nil(t, err, "PostTransaction should not return an error")
}
