// Copyright 2024 Fantom Foundation
// Unit tests for txcontext/substate/receipt.go
package substate

import (
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestReceipt_NewReceipt(t *testing.T) {
	// Create a substate.Result with test values
	status := uint64(1)
	bloom := types.Bloom{1, 2, 3}
	contractAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	gasUsed := uint64(21000)

	// Create logs
	logs := []*substatetypes.Log{
		{
			Address:     substatetypes.Address(common.HexToAddress("0x1111111111111111111111111111111111111111")),
			Topics:      []substatetypes.Hash{substatetypes.Hash(common.HexToHash("0xabc")), substatetypes.Hash(common.HexToHash("0xdef"))},
			Data:        []byte{1, 2, 3, 4},
			BlockNumber: 12345,
			TxHash:      substatetypes.Hash(common.HexToHash("0xabcd")),
			TxIndex:     0,
			BlockHash:   substatetypes.Hash(common.HexToHash("0x1234")),
			Index:       0,
			Removed:     false,
		},
	}

	substateResult := &substate.Result{
		Status:          status,
		Bloom:           substatetypes.Bloom(bloom),
		Logs:            logs,
		ContractAddress: substatetypes.Address(contractAddress),
		GasUsed:         gasUsed,
	}

	// Create result from substate.Result
	result := NewReceipt(substateResult)

	// Test GetReceipt
	receipt := result.GetReceipt()
	assert.Equal(t, result, receipt)

	// Test GetRawResult
	rawData, err := result.GetRawResult()
	assert.Nil(t, rawData)
	assert.Nil(t, err)

	// Test GetStatus
	assert.Equal(t, status, receipt.GetStatus())

	// Test GetBloom
	assert.Equal(t, bloom, receipt.GetBloom())

	// Test GetLogs
	ethLogs := receipt.GetLogs()
	assert.Equal(t, 1, len(ethLogs))

	// Verify log conversion
	ethLog := ethLogs[0]
	assert.Equal(t, common.HexToAddress("0x1111111111111111111111111111111111111111"), ethLog.Address)
	assert.Equal(t, []common.Hash{common.HexToHash("0xabc"), common.HexToHash("0xdef")}, ethLog.Topics)
	assert.Equal(t, []byte{1, 2, 3, 4}, ethLog.Data)
	assert.Equal(t, uint64(12345), ethLog.BlockNumber)
	assert.Equal(t, common.HexToHash("0xabcd"), ethLog.TxHash)
	assert.Equal(t, uint(0), ethLog.TxIndex)
	assert.Equal(t, common.HexToHash("0x1234"), ethLog.BlockHash)
	assert.Equal(t, uint(0), ethLog.Index)
	assert.Equal(t, false, ethLog.Removed)

	// Test GetContractAddress
	assert.Equal(t, contractAddress, receipt.GetContractAddress())

	// Test GetGasUsed
	assert.Equal(t, gasUsed, receipt.GetGasUsed())

	// Test Equal with same receipt
	assert.True(t, receipt.Equal(receipt))

	// Test String method
	str := result.String()
	assert.Contains(t, str, "Status: 1")
	assert.Contains(t, str, "Gas Used: 21000")
}

func TestReceipt_WithEmptyLogs(t *testing.T) {
	// Create a substate.Result with empty logs
	substateResult := &substate.Result{
		Status:  1,
		Logs:    []*substatetypes.Log{},
		GasUsed: 21000,
	}

	// Create result from substate.Result
	result := NewReceipt(substateResult)

	// Test GetLogs returns empty slice
	logs := result.GetLogs()
	assert.Equal(t, 0, len(logs))
}
