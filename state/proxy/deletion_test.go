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

package proxy

import (
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProxy_NewDeletionProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")
	assert.NotNil(t, proxy)
	assert.Equal(t, mockDb, proxy.db)
	assert.Equal(t, mockChan, proxy.ch)
}
func TestDeletionProxy_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	mockDb.EXPECT().CreateAccount(address).Times(1)

	proxy.CreateAccount(address)
}

func TestDeletionProxy_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	proxy := NewDeletionProxy(mockDb, mockChan, "info")

	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedRefund := uint256.NewInt(100)

	mockDb.EXPECT().SelfDestruct(address).Return(*expectedRefund).Times(1)

	refund := proxy.SelfDestruct(address)
	assert.Equal(t, *expectedRefund, refund)
}

func TestDeletionProxy_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockChan := make(chan ContractLiveliness, 10)
	mockLogger := logger.NewMockLogger(ctrl)
	proxy := &DeletionProxy{
		db:  mockDb,
		ch:  mockChan,
		log: mockLogger,
	}

	expectedBlock := uint64(100)

	mockLogger.EXPECT().Fatal(gomock.Any()).Times(1)
	bulkLoad, err := proxy.StartBulkLoad(expectedBlock)
	assert.Nil(t, err)
	assert.Nil(t, bulkLoad)
}
