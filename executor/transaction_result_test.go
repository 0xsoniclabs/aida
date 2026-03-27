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
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestTransactionResult_GetReceipt(t *testing.T) {
	obj := transactionResult{}
	out := obj.GetReceipt()
	assert.Equal(t, obj, out)
}

func TestTransactionResult_GetRawResult(t *testing.T) {
	r := []byte("result")
	e := errors.New("mock error")
	obj := transactionResult{
		result: r,
		err:    e,
	}
	out, err := obj.GetRawResult()
	assert.Equal(t, r, out)
	assert.Equal(t, e, err)
}

func TestTransactionResult_GetGasUsed(t *testing.T) {
	obj := transactionResult{
		gasUsed: 100,
	}
	out := obj.GetGasUsed()
	assert.Equal(t, uint64(100), out)
}

func TestTransactionResult_GetStatus(t *testing.T) {
	obj := transactionResult{
		status: 1,
	}
	out := obj.GetStatus()
	assert.Equal(t, uint64(1), out)
}

func TestTransactionResult_GetBloom(t *testing.T) {
	obj := transactionResult{
		bloom: types.Bloom{1, 2, 3},
	}
	out := obj.GetBloom()
	assert.Equal(t, types.Bloom{1, 2, 3}, out)
}

func TestTransactionResult_GetLogs(t *testing.T) {
	obj := transactionResult{
		logs: []*types.Log{
			{Address: common.HexToAddress("0x123")},
			{Address: common.HexToAddress("0x456")},
		},
	}
	out := obj.GetLogs()
	assert.Equal(t, 2, len(out))
	assert.Equal(t, common.HexToAddress("0x123"), out[0].Address)
	assert.Equal(t, common.HexToAddress("0x456"), out[1].Address)
}

func TestTransactionResult_GetContractAddress(t *testing.T) {
	obj := transactionResult{
		contractAddress: common.HexToAddress("0x789"),
	}
	out := obj.GetContractAddress()
	assert.Equal(t, common.HexToAddress("0x789"), out)
}

func TestTransactionResult_Equal(t *testing.T) {
	obj1 := transactionResult{
		status:          1,
		bloom:           types.Bloom{1, 2, 3},
		contractAddress: common.HexToAddress("0x789"),
		gasUsed:         100,
		logs: []*types.Log{
			{Address: common.HexToAddress("0x123")},
			{Address: common.HexToAddress("0x456")},
		},
	}

	obj2 := transactionResult{
		status:          1,
		bloom:           types.Bloom{1, 2, 3},
		contractAddress: common.HexToAddress("0x789"),
		gasUsed:         100,
		logs: []*types.Log{
			{Address: common.HexToAddress("0x123")},
			{Address: common.HexToAddress("0x456")},
		},
	}

	assert.True(t, obj1.Equal(obj2))
	assert.False(t, obj1.Equal(transactionResult{}))
}

func TestTransactionResult_String(t *testing.T) {
	obj := transactionResult{
		status:          1,
		bloom:           types.Bloom{1, 2, 3},
		contractAddress: common.HexToAddress("0x789"),
		gasUsed:         100,
		logs: []*types.Log{
			{Address: common.HexToAddress("0x123")},
			{Address: common.HexToAddress("0x456")},
		},
	}
	out := obj.String()
	assert.Equal(t, 383, len(out))
}

func TestTransactionResult_newTransactionResult(t *testing.T) {
	logs := []*types.Log{
		{Address: common.HexToAddress("0x123")},
		{Address: common.HexToAddress("0x456")},
	}
	msg := &core.Message{
		To:    nil,
		Nonce: 1,
	}
	msgResult := &messageResult{
		gasUsed:    200,
		failed:     false,
		returnData: []byte("result data"),
	}
	err := errors.New("mock error")
	origin := common.HexToAddress("0x789")

	result := newTransactionResult(logs, msg, msgResult, err, origin)

	assert.Equal(t, logs, result.logs)
	assert.Equal(t, common.HexToAddress("0x0Ef6E08F95ac9263d5159A1e57050FE7454b3b9f"), result.contractAddress)
	assert.Equal(t, uint64(200), result.gasUsed)
	assert.Equal(t, types.ReceiptStatusSuccessful, result.status)
	assert.Equal(t, types.CreateBloom(&types.Receipt{Logs: logs}), result.bloom)
	assert.Equal(t, []byte("result data"), result.result)
	assert.Equal(t, err, result.err)
}

func TestTransactionResult_newPseudoExecutionResult(t *testing.T) {
	result := newPseudoExecutionResult()

	r, e := result.GetRawResult()
	re := result.GetReceipt()
	assert.Equal(t, []byte{}, r)
	assert.Equal(t, nil, e)
	assert.Equal(t, types.ReceiptStatusSuccessful, re.GetStatus())
	assert.Equal(t, types.Bloom{}, re.GetBloom())
	assert.Equal(t, common.Address{}, re.GetContractAddress())
	assert.Equal(t, uint64(0), result.GetGasUsed())
	assert.Equal(t, []*types.Log(nil), re.GetLogs())
	assert.False(t, re.Equal(transactionResult{}))
}
