// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/txgenerator/tx_generator.go
package txgenerator

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewTxContext(t *testing.T) {
	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock block environment
	mockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	// Setup expected calls
	blockNumber := uint64(12345)
	timestamp := uint64(1621234567)
	fork := "shanghai"

	mockEnv.EXPECT().GetNumber().Return(blockNumber).AnyTimes()
	mockEnv.EXPECT().GetTimestamp().Return(timestamp).AnyTimes()
	mockEnv.EXPECT().GetFork().Return(fork).AnyTimes()

	// Create a core.Message
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	msg := &core.Message{
		From:      from,
		To:        &to,
		Nonce:     1,
		Value:     big.NewInt(1000),
		GasLimit:  21000,
		GasPrice:  big.NewInt(1),
		GasFeeCap: big.NewInt(1),
		GasTipCap: big.NewInt(1),
		Data:      []byte{1, 2, 3, 4},
	}

	// Create a new TxContext
	ctx := NewTxContext(mockEnv, msg)
	assert.NotNil(t, ctx)

	// Test GetBlockEnvironment
	env := ctx.GetBlockEnvironment()
	assert.Equal(t, mockEnv, env)
	assert.Equal(t, blockNumber, env.GetNumber())
	assert.Equal(t, timestamp, env.GetTimestamp())
	assert.Equal(t, fork, env.GetFork())

	// Test GetMessage
	retrievedMsg := ctx.GetMessage()
	assert.Equal(t, msg, retrievedMsg)
	assert.Equal(t, from, retrievedMsg.From)
	assert.Equal(t, &to, retrievedMsg.To)
	assert.Equal(t, uint64(1), retrievedMsg.Nonce)
	assert.Equal(t, big.NewInt(1000), retrievedMsg.Value)
	assert.Equal(t, uint64(21000), retrievedMsg.GasLimit)
	assert.Equal(t, big.NewInt(1), retrievedMsg.GasPrice)
	assert.Equal(t, []byte{1, 2, 3, 4}, retrievedMsg.Data)

	// Test GetLogsHash returns empty hash
	logsHash := ctx.(*txData).GetLogsHash()
	assert.Equal(t, common.Hash{}, logsHash)

	// Test GetStateHash returns empty hash
	stateHash := ctx.(*txData).GetStateHash()
	assert.Equal(t, common.Hash{}, stateHash)

	// Test NilTxContext methods are properly inherited
	assert.Nil(t, ctx.GetInputState())
	assert.Nil(t, ctx.GetOutputState())
	assert.Nil(t, ctx.GetResult())
}

func TestTxDataStructure(t *testing.T) {
	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock block environment
	mockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	// Setup expected calls
	blockNumber := uint64(12345)
	timestamp := uint64(1621234567)
	fork := "shanghai"

	mockEnv.EXPECT().GetNumber().Return(blockNumber).AnyTimes()
	mockEnv.EXPECT().GetTimestamp().Return(timestamp).AnyTimes()
	mockEnv.EXPECT().GetFork().Return(fork).AnyTimes()

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	msg := &core.Message{
		From:  from,
		To:    &to,
		Value: big.NewInt(1000),
		Data:  []byte{1, 2, 3, 4},
	}

	txDataInstance := &txData{
		Env:     mockEnv,
		Message: msg,
	}

	// Verify fields are set correctly
	assert.Equal(t, mockEnv, txDataInstance.Env)
	assert.Equal(t, msg, txDataInstance.Message)

	// Verify methods return expected values
	assert.Equal(t, mockEnv, txDataInstance.GetBlockEnvironment())
	assert.Equal(t, msg, txDataInstance.GetMessage())
	assert.Equal(t, common.Hash{}, txDataInstance.GetLogsHash())
	assert.Equal(t, common.Hash{}, txDataInstance.GetStateHash())

	// Verify inherited NilTxContext methods
	assert.Nil(t, txDataInstance.GetInputState())
	assert.Nil(t, txDataInstance.GetOutputState())
	assert.Nil(t, txDataInstance.GetResult())
}
