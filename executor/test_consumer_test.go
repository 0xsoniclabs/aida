package executor

import (
	"testing"

	"github.com/0xsoniclabs/aida/rpc"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestExecutor_ToSubstateConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	consumer := NewMockTxConsumer(ctrl)
	mockTxContext := txcontext.NewMockTxContext(ctrl)

	// Create a mock TransactionInfo
	info := TransactionInfo[txcontext.TxContext]{
		Block:       1,
		Transaction: 2,
		Data:        mockTxContext,
	}

	// Expect the Consume method to be called with the correct parameters
	consumer.EXPECT().Consume(info.Block, info.Transaction, info.Data).Return(nil)

	// Convert to Consumer and call it
	c := toSubstateConsumer(consumer)
	err := c(info)
	assert.NoError(t, err)
}

func TestExecutor_ToRPCConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	consumer := NewMockRPCReqConsumer(ctrl)
	mockRequest := &rpc.RequestAndResults{}

	// Create a mock TransactionInfo
	info := TransactionInfo[*rpc.RequestAndResults]{
		Block:       1,
		Transaction: 2,
		Data:        mockRequest,
	}

	// Expect the Consume method to be called with the correct parameters
	consumer.EXPECT().Consume(info.Block, info.Transaction, info.Data).Return(nil)

	// Convert to Consumer and call it
	c := toRPCConsumer(consumer)
	err := c(info)
	assert.NoError(t, err)
}
