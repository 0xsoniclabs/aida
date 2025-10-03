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

package txcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilTxContext(t *testing.T) {
	// Create a new NilTxContext instance
	ctx := NilTxContext{}

	// Test GetInputState
	inputState := ctx.GetInputState()
	assert.Nil(t, inputState)

	// Test GetBlockEnvironment
	blockEnv := ctx.GetBlockEnvironment()
	assert.Nil(t, blockEnv)

	// Test GetMessage
	message := ctx.GetMessage()
	assert.NotNil(t, message)

	// Test GetOutputState
	outputState := ctx.GetOutputState()
	assert.Nil(t, outputState)

	// Test GetResult
	result := ctx.GetResult()
	assert.Nil(t, result)
}
