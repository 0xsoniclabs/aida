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
