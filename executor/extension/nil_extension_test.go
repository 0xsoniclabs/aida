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

package extension

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/stretchr/testify/assert"
)

func TestNilExtensionIsExtension(t *testing.T) {
	var _ executor.Extension[any] = NilExtension[any]{}
}

func TestNilExtension_PreRun(t *testing.T) {
	ext := NilExtension[string]{}
	state := executor.State[string]{Data: "test"}
	ctx := &executor.Context{}
	err := ext.PreRun(state, ctx)
	assert.NoError(t, err)
}

func TestNilExtension_PostRun(t *testing.T) {
	ext := NilExtension[int]{}
	state := executor.State[int]{Data: 123}
	ctx := &executor.Context{}
	originalErr := errors.New("some error")

	err := ext.PostRun(state, ctx, nil)
	assert.NoError(t, err)

	errWithOriginal := ext.PostRun(state, ctx, originalErr)
	assert.NoError(t, errWithOriginal, "PostRun should return nil even if an error is passed in")
}

func TestNilExtension_PreBlock(t *testing.T) {
	ext := NilExtension[any]{}
	state := executor.State[any]{}
	ctx := &executor.Context{}
	err := ext.PreBlock(state, ctx)
	assert.NoError(t, err)
}

func TestNilExtension_PostBlock(t *testing.T) {
	ext := NilExtension[float64]{}
	state := executor.State[float64]{Data: 1.23}
	ctx := &executor.Context{}
	err := ext.PostBlock(state, ctx)
	assert.NoError(t, err)
}

func TestNilExtension_PreTransaction(t *testing.T) {
	ext := NilExtension[bool]{}
	state := executor.State[bool]{Data: true}
	ctx := &executor.Context{}
	err := ext.PreTransaction(state, ctx)
	assert.NoError(t, err)
}

func TestNilExtension_PostTransaction(t *testing.T) {
	ext := NilExtension[*string]{}
	val := "test"
	state := executor.State[*string]{Data: &val}
	ctx := &executor.Context{}
	err := ext.PostTransaction(state, ctx)
	assert.NoError(t, err)
}
