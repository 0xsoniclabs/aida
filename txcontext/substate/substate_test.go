// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/substate/substate.go
package substate

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestSubstateData_GetLogsHash(t *testing.T) {
	ss := &substateData{
		Substate: &substate.Substate{},
	}

	logsHash := ss.GetLogsHash()
	assert.Equal(t, common.Hash{}, logsHash)
}

func TestSubstateData_GetStateHash(t *testing.T) {
	ss := &substateData{
		Substate: &substate.Substate{},
	}

	stateHash := ss.GetStateHash()
	assert.Equal(t, common.Hash{}, stateHash)
}

func TestSubstateData_GetInputState(t *testing.T) {
	ss := &substateData{
		Substate: &substate.Substate{
			InputSubstate: substate.WorldState{},
		},
	}

	inputState := ss.GetInputState()
	assert.NotNil(t, inputState)
	assert.Equal(t, 0, inputState.Len())
}

func TestSubstateData_GetOutputState(t *testing.T) {
	ss := &substateData{
		Substate: &substate.Substate{
			OutputSubstate: substate.WorldState{},
		},
	}

	outputState := ss.GetOutputState()
	assert.NotNil(t, outputState)
	assert.Equal(t, 0, outputState.Len())
}

func TestSubstateData_GetBlockEnvironment(t *testing.T) {
	ss := &substateData{
		Substate: &substate.Substate{
			Env: &substate.Env{},
		},
	}

	blockEnv := ss.GetBlockEnvironment()
	assert.NotNil(t, blockEnv)
	assert.Equal(t, common.Address{}, blockEnv.GetCoinbase())
	assert.Equal(t, uint64(0), blockEnv.GetGasLimit())
	assert.Equal(t, uint64(0), blockEnv.GetNumber())
	assert.Equal(t, uint64(0), blockEnv.GetTimestamp())
}

func TestSubstateData_GetMessage(t *testing.T) {
	// Test with a full Message including all fields
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	accessList := []substatetypes.AccessTuple{
		{
			Address:     substatetypes.Address(common.HexToAddress("0x3333333333333333333333333333333333333333")),
			StorageKeys: []substatetypes.Hash{substatetypes.Hash(common.HexToHash("0x1"))},
		},
	}

	blobHashes := []substatetypes.Hash{
		substatetypes.Hash(common.HexToHash("0xblobhash")),
	}

	setCodeAuthorizations := []substatetypes.SetCodeAuthorization{
		{
			ChainID: *uint256.NewInt(1),
			Address: substatetypes.Address(common.HexToAddress("0x4444444444444444444444444444444444444444")),
			Nonce:   1,
			V:       uint8(27),
			R:       *uint256.NewInt(1),
			S:       *uint256.NewInt(2),
		},
	}

	ss := &substateData{
		Substate: &substate.Substate{
			Message: &substate.Message{
				From:                  substatetypes.Address(from),
				To:                    (*substatetypes.Address)(&to),
				Nonce:                 1,
				Value:                 big.NewInt(50),
				Gas:                   21000,
				GasPrice:              big.NewInt(2),
				GasFeeCap:             big.NewInt(3),
				GasTipCap:             big.NewInt(1),
				Data:                  []byte{1, 2, 3, 4},
				AccessList:            accessList,
				BlobGasFeeCap:         big.NewInt(1000),
				BlobHashes:            blobHashes,
				CheckNonce:            true,
				SetCodeAuthorizations: setCodeAuthorizations,
			},
		},
	}

	message := ss.GetMessage()
	assert.NotNil(t, message)

	// Test basic fields
	assert.Equal(t, from, message.From)
	assert.Equal(t, &to, message.To)
	assert.Equal(t, uint64(1), message.Nonce)
	assert.Equal(t, big.NewInt(50), message.Value)
	assert.Equal(t, uint64(21000), message.GasLimit)
	assert.Equal(t, big.NewInt(2), message.GasPrice)
	assert.Equal(t, big.NewInt(3), message.GasFeeCap)
	assert.Equal(t, big.NewInt(1), message.GasTipCap)
	assert.Equal(t, []byte{1, 2, 3, 4}, message.Data)

	// Test AccessList conversion
	assert.Equal(t, 1, len(message.AccessList))
	assert.Equal(t, common.HexToAddress("0x3333333333333333333333333333333333333333"), message.AccessList[0].Address)
	assert.Equal(t, []common.Hash{common.HexToHash("0x1")}, message.AccessList[0].StorageKeys)

	// Test BlobHashes conversion
	assert.Equal(t, 1, len(message.BlobHashes))
	assert.Equal(t, common.HexToHash("0xblobhash"), message.BlobHashes[0])

	// Test BlobGasFeeCap
	assert.Equal(t, big.NewInt(1000), message.BlobGasFeeCap)

	// Test SkipNonceChecks (should be false when CheckNonce is true)
	assert.False(t, message.SkipNonceChecks)

	// Test SetCodeAuthorizations conversion
	assert.Equal(t, 1, len(message.SetCodeAuthorizations))
	assert.Equal(t, *uint256.NewInt(1), message.SetCodeAuthorizations[0].ChainID)
	assert.Equal(t, common.HexToAddress("0x4444444444444444444444444444444444444444"), message.SetCodeAuthorizations[0].Address)
	assert.Equal(t, uint64(1), message.SetCodeAuthorizations[0].Nonce)
	assert.Equal(t, uint8(27), message.SetCodeAuthorizations[0].V)
	assert.Equal(t, *uint256.NewInt(1), message.SetCodeAuthorizations[0].R)
	assert.Equal(t, *uint256.NewInt(2), message.SetCodeAuthorizations[0].S)

	// Test with nil fields
	ss2 := &substateData{
		Substate: &substate.Substate{
			Message: &substate.Message{
				From:       substatetypes.Address(from),
				To:         nil,
				Value:      nil,
				GasPrice:   nil,
				GasFeeCap:  nil,
				GasTipCap:  nil,
				AccessList: nil,
				BlobHashes: nil,
				CheckNonce: false,
			},
		},
	}

	message2 := ss2.GetMessage()
	assert.NotNil(t, message2)
	assert.Equal(t, from, message2.From)
	assert.Nil(t, message2.To)
	assert.Nil(t, message2.Value)
	assert.Nil(t, message2.GasPrice)
	assert.Nil(t, message2.GasFeeCap)
	assert.Nil(t, message2.GasTipCap)
	assert.Nil(t, message2.AccessList)
	assert.Nil(t, message2.BlobHashes)
	assert.Nil(t, message2.BlobGasFeeCap)
	assert.True(t, message2.SkipNonceChecks) // Should be true when CheckNonce is false
	assert.Empty(t, message2.SetCodeAuthorizations)

	// Test with empty arrays
	ss3 := &substateData{
		Substate: &substate.Substate{
			Message: &substate.Message{
				From:                  substatetypes.Address(from),
				AccessList:            []substatetypes.AccessTuple{},
				BlobHashes:            []substatetypes.Hash{},
				SetCodeAuthorizations: []substatetypes.SetCodeAuthorization{},
			},
		},
	}

	message3 := ss3.GetMessage()
	assert.NotNil(t, message3)
	assert.Equal(t, from, message3.From)
	assert.Nil(t, message3.AccessList)
	assert.Nil(t, message3.BlobHashes)
	assert.Empty(t, message3.SetCodeAuthorizations)
}

func TestSubstateData_NewTxContext(t *testing.T) {
	ss := &substateData{
		Substate: &substate.Substate{
			InputSubstate:  substate.WorldState{},
			OutputSubstate: substate.WorldState{},
			Env:            &substate.Env{},
			Message:        &substate.Message{},
		},
	}

	txContext := NewTxContext(ss.Substate)
	assert.NotNil(t, txContext)
	assert.Equal(t, ss.GetInputState(), txContext.GetInputState())
	assert.Equal(t, ss.GetOutputState(), txContext.GetOutputState())
	assert.Equal(t, ss.GetBlockEnvironment(), txContext.GetBlockEnvironment())
	assert.Equal(t, ss.GetMessage(), txContext.GetMessage())
}
