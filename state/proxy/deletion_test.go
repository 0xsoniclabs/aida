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
