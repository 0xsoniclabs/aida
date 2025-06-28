package ethtest

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewMockStateTestContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test message
	sender := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recipient := common.HexToAddress("0x0987654321098765432109876543210987654321")
	message := &core.Message{
		From:     sender,
		To:       &recipient,
		Nonce:    10,
		Value:    big.NewInt(1000),
		GasLimit: 21000,
		GasPrice: big.NewInt(50),
		Data:     []byte{1, 2, 3},
	}
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	// Create mocks
	mockTx := types.NewTx(&types.DynamicFeeTx{
		ChainID: big.NewInt(int64(utils.SepoliaChainID)),
		V:       common.Big0,
		R:       common.Big1,
		S:       common.Big1,
	})
	mockBytes := utils.Must(mockTx.MarshalBinary())
	mockTxContext := NewMockStateTestContext(message, mockBlockEnv, mockBytes)
	assert.Equal(t, mockBlockEnv, mockTxContext.env)
	assert.Equal(t, message, mockTxContext.msg)
	assert.Len(t, mockTxContext.txBytes, len(mockBytes))
	for i, txBytes := range mockTxContext.txBytes {
		assert.Equal(t, mockBytes[i], txBytes)
	}
}
