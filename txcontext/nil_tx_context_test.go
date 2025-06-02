// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/nil_tx_context.go
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
