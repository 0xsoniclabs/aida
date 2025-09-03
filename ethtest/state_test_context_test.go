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

package ethtest

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEthTest_NewMockStateTestContext(t *testing.T) {
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
		ChainID: big.NewInt(int64(config.SepoliaChainID)),
		V:       common.Big0,
		R:       common.Big1,
		S:       common.Big1,
	})
	mockBytes := utils.Must(mockTx.MarshalBinary())
	mockTxContext := NewStateTestContext(message, mockBlockEnv, mockBytes)
	assert.Equal(t, mockBlockEnv, mockTxContext.env)
	assert.Equal(t, message, mockTxContext.msg)
	assert.Len(t, mockTxContext.txBytes, len(mockBytes))
	for i, txBytes := range mockTxContext.txBytes {
		assert.Equal(t, mockBytes[i], txBytes)
	}
}

func newTestStateTestContext() *StateTestContext {
	msg := &core.Message{}
	alloc := types.GenesisAlloc{}
	stJson := &stJSON{
		path: "testpath.json",
		Pre:  alloc,
	}
	post := stPost{
		RootHash:        common.HexToHash("0x1234"),
		ExpectException: "error",
		TxBytes:         hexutil.Bytes{0x01, 0x02},
		LogsHash:        common.HexToHash("0xabcd"),
	}
	chainCfg := &params.ChainConfig{}
	ctx := newStateTestTxContext(stJson, msg, post, chainCfg, "label", "fork", 7)
	return ctx.(*StateTestContext)
}

func TestStateTestContext_GetTxBytes(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Equal(t, hexutil.Bytes{0x01, 0x02}, stCtx.GetTxBytes())
}

func TestStateTestContext_GetLogsHash(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Equal(t, common.HexToHash("0xabcd"), stCtx.GetLogsHash())
}

func TestStateTestContext_GetStateHash(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Equal(t, common.HexToHash("0x1234"), stCtx.GetStateHash())
}

func TestStateTestContext_GetOutputState(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Nil(t, stCtx.GetOutputState())
}

func TestStateTestContext_GetInputState(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.NotNil(t, stCtx.GetInputState())
}

func TestStateTestContext_GetBlockEnvironment(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.NotNil(t, stCtx.GetBlockEnvironment())
}

func TestStateTestContext_GetMessage(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.NotNil(t, stCtx.GetMessage())
}

func TestStateTestContext_GetResult(t *testing.T) {
	stCtx := newTestStateTestContext()
	res := stCtx.GetResult()
	assert.Equal(t, "error", res.(stateTestResult).expectedErr)
}

func TestStateTestContext_String(t *testing.T) {
	stCtx := newTestStateTestContext()
	str := stCtx.String()
	assert.Contains(t, str, "testpath.json")
	assert.Contains(t, str, "label")
	assert.Contains(t, str, "fork")
	assert.Contains(t, str, fmt.Sprint(7))
}
