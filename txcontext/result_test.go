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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestResult_NewResult(t *testing.T) {
	// Test parameters
	status := uint64(1)
	bloom := types.Bloom{}
	logs := []*types.Log{}
	contractAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	gasUsed := uint64(21000)

	// Create a new result
	receipt := NewResult(status, bloom, logs, contractAddress, gasUsed)

	// Verify the result was created correctly
	assert.Equal(t, status, receipt.GetStatus())
	assert.Equal(t, bloom, receipt.GetBloom())
	assert.Equal(t, logs, receipt.GetLogs())
	assert.Equal(t, contractAddress, receipt.GetContractAddress())
	assert.Equal(t, gasUsed, receipt.GetGasUsed())
}

func TestResult_ReceiptEqual(t *testing.T) {
	// Create common test data
	logs1 := []*types.Log{
		{
			Address: common.HexToAddress("0x1"),
			Topics:  []common.Hash{common.HexToHash("0xa"), common.HexToHash("0xb")},
			Data:    []byte{1, 2, 3},
		},
	}
	logs2 := []*types.Log{
		{
			Address: common.HexToAddress("0x1"),
			Topics:  []common.Hash{common.HexToHash("0xa"), common.HexToHash("0xb")},
			Data:    []byte{1, 2, 3},
		},
	}
	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create two identical receipts
	receipt1 := NewResult(1, types.Bloom{}, logs1, addr, 21000)
	receipt2 := NewResult(1, types.Bloom{}, logs2, addr, 21000)

	// Test equality with identical receipts
	assert.True(t, ReceiptEqual(receipt1, receipt2))

	// Test equality with same receipt reference
	assert.True(t, ReceiptEqual(receipt1, receipt1))

	// Test with nil receipts
	assert.False(t, ReceiptEqual(receipt1, nil))
	assert.False(t, ReceiptEqual(nil, receipt1))
	assert.True(t, ReceiptEqual(nil, nil))

	// Test with different status
	receiptDiffStatus := NewResult(0, types.Bloom{}, logs1, addr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffStatus))

	// Test with different bloom
	differentBloom := types.Bloom{}
	differentBloom[0] = 1 // Set a bit to make it different
	receiptDiffBloom := NewResult(1, differentBloom, logs1, addr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffBloom))

	// Test with different contract address
	diffAddr := common.HexToAddress("0x0987654321098765432109876543210987654321")
	receiptDiffAddr := NewResult(1, types.Bloom{}, logs1, diffAddr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffAddr))

	// Test with different gas used
	receiptDiffGas := NewResult(1, types.Bloom{}, logs1, addr, 30000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffGas))

	// Test with different number of logs
	logsEmpty := []*types.Log{}
	receiptDiffLogsCount := NewResult(1, types.Bloom{}, logsEmpty, addr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffLogsCount))

	// Test with logs that have different address
	logsDiffAddr := []*types.Log{
		{
			Address: common.HexToAddress("0x2"),
			Topics:  []common.Hash{common.HexToHash("0xa"), common.HexToHash("0xb")},
			Data:    []byte{1, 2, 3},
		},
	}
	receiptDiffLogAddr := NewResult(1, types.Bloom{}, logsDiffAddr, addr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffLogAddr))

	// Test with logs that have different topics count
	logsDiffTopicCount := []*types.Log{
		{
			Address: common.HexToAddress("0x1"),
			Topics:  []common.Hash{common.HexToHash("0xa")}, // Only one topic
			Data:    []byte{1, 2, 3},
		},
	}
	receiptDiffTopicCount := NewResult(1, types.Bloom{}, logsDiffTopicCount, addr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffTopicCount))

	// Test with logs that have different topic value
	logsDiffTopicValue := []*types.Log{
		{
			Address: common.HexToAddress("0x1"),
			Topics:  []common.Hash{common.HexToHash("0xa"), common.HexToHash("0xc")}, // Second topic is different
			Data:    []byte{1, 2, 3},
		},
	}
	receiptDiffTopicValue := NewResult(1, types.Bloom{}, logsDiffTopicValue, addr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffTopicValue))

	// Test with logs that have different data
	logsDiffData := []*types.Log{
		{
			Address: common.HexToAddress("0x1"),
			Topics:  []common.Hash{common.HexToHash("0xa"), common.HexToHash("0xb")},
			Data:    []byte{4, 5, 6}, // Different data
		},
	}
	receiptDiffData := NewResult(1, types.Bloom{}, logsDiffData, addr, 21000)
	assert.False(t, ReceiptEqual(receipt1, receiptDiffData))
}

func TestResult_ReceiptEqual2(t *testing.T) {
	// Create a receipt
	status := uint64(1)
	bloom := types.Bloom{}
	logs := []*types.Log{
		{
			Address: common.HexToAddress("0x1"),
			Topics:  []common.Hash{common.HexToHash("0xa")},
			Data:    []byte{1, 2, 3},
		},
	}
	contractAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	gasUsed := uint64(21000)

	receipt := NewResult(status, bloom, logs, contractAddress, gasUsed)

	// Test the Equal method
	identicalReceipt := NewResult(status, bloom, logs, contractAddress, gasUsed)
	assert.True(t, receipt.Equal(identicalReceipt))

	differentReceipt := NewResult(0, bloom, logs, contractAddress, gasUsed)
	assert.False(t, receipt.Equal(differentReceipt))
}
