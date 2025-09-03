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

package rpc

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/sonic/ethapi"
	"github.com/0xsoniclabs/sonic/opera"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRpc_newEvmExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockArchive := state.NewMockNonCommittableStateDB(ctrl)

	t.Run("success", func(t *testing.T) {
		cfg := &config.Config{
			ChainID: config.MainnetChainID,
		}
		p := map[string]interface{}{"from": "0x0000000000000000000000000000000000000001"}
		exec, err := newEvmExecutor(1, mockArchive, cfg, p, 123)
		assert.Nil(t, err)
		assert.NotNil(t, exec)
	})

	t.Run("no factory", func(t *testing.T) {
		cfg := &config.Config{
			ChainID: config.MainnetChainID,
			VmImpl:  "1234",
		}
		p := map[string]interface{}{"from": "0x0000000000000000000000000000000000000001"}
		exec, err := newEvmExecutor(1, mockArchive, cfg, p, 123)
		assert.Nil(t, exec)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "cannot get interpreter factory")
	})

	t.Run("no chain config", func(t *testing.T) {
		cfg := &config.Config{
			ChainID: config.PseudoTx,
		}
		p := map[string]interface{}{"from": "0x0000000000000000000000000000000000000001"}
		exec, err := newEvmExecutor(1, mockArchive, cfg, p, 123)
		assert.Nil(t, exec)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "cannot get chain config")
	})

}

func TestRpc_newTxArgs(t *testing.T) {
	params := map[string]interface{}{
		"from":     "0x0000000000000000000000000000000000000001",
		"to":       "0x0000000000000000000000000000000000000002",
		"value":    "0x10",
		"gas":      "0x5208",
		"gasPrice": "0x1",
		"data":     "0x00",
	}
	args := newTxArgs(params)

	expectedFrom := common.HexToAddress("0x0000000000000000000000000000000000000001")
	expectedTo := common.HexToAddress("0x0000000000000000000000000000000000000002")
	expectedValue := (*hexutil.Big)(big.NewInt(16))
	expectedGas := hexutil.Uint64(21000)
	expectedGasPrice := (*hexutil.Big)(big.NewInt(1))

	assert.Equal(t, expectedFrom, *args.From)
	assert.Equal(t, expectedTo, *args.To)
	assert.Equal(t, expectedValue.String(), args.Value.String())
	assert.Equal(t, expectedGas, *args.Gas)
	assert.Equal(t, expectedGasPrice.String(), args.GasPrice.String())
	assert.Equal(t, hexutil.Bytes{0x0}, *args.Data)
}

func TestEvmExecutor_newEVM(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockArchive := state.NewMockNonCommittableStateDB(ctrl)
	mockArchive.EXPECT().GetHash().Return(common.Hash{}, nil).AnyTimes()
	e := &EvmExecutor{
		archive:  mockArchive,
		chainCfg: params.MainnetChainConfig,
		blockId:  big.NewInt(1),
		rules:    opera.DefaultEconomyRules(),
	}
	msg := &core.Message{}
	var hashErr error
	evm := e.newEVM(msg, &hashErr)
	assert.NotNil(t, evm)
	assert.NotNil(t, evm.Context.GetHash(1))
}

func TestEvmExecutor_sendCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchive := state.NewMockNonCommittableStateDB(ctrl)
	mockArchive.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1000000000)).AnyTimes()
	mockArchive.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockArchive.EXPECT().Prepare(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockArchive.EXPECT().GetNonce(gomock.Any()).Return(uint64(1234)).AnyTimes()
	mockArchive.EXPECT().SetNonce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockArchive.EXPECT().GetRefund().Return(uint64(0)).AnyTimes()
	mockArchive.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	mockArchive.EXPECT().GetCode(gomock.Any()).Return([]uint8{}).AnyTimes()
	mockArchive.EXPECT().Snapshot().Return(0).AnyTimes()
	mockArchive.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()

	t.Run("success", func(t *testing.T) {
		e := &EvmExecutor{
			args:     newTxArgs(map[string]interface{}{"from": "0x0000000000000000000000000000000000000001", "to": "0x0000000000000000000000000000000000000002"}),
			archive:  mockArchive,
			chainCfg: params.MainnetChainConfig,
			blockId:  big.NewInt(1),
			rules:    opera.DefaultEconomyRules(),
		}
		out, err := e.sendCall()
		assert.Nil(t, err)
		assert.NotNil(t, out)
	})

	t.Run("error to message", func(t *testing.T) {
		jsonData := []byte(`{"from": "0x0000000000000000000000000000000000000001", "gasPrice": "0x1234", "maxFeePerGas": "0x5678"}`)
		var args ethapi.TransactionArgs
		err := json.Unmarshal(jsonData, &args)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		e := &EvmExecutor{
			args:     args,
			archive:  mockArchive,
			chainCfg: params.MainnetChainConfig,
			blockId:  big.NewInt(1),
			rules:    opera.DefaultEconomyRules(),
		}
		out, err := e.sendCall()
		assert.NotNil(t, err)
		assert.Nil(t, out)
	})

	t.Run("error apply message", func(t *testing.T) {
		e := &EvmExecutor{
			args:     newTxArgs(map[string]interface{}{"from": "0x0000000000000000000000000000000000000001", "gasPrice": "1234"}),
			archive:  mockArchive,
			chainCfg: params.MainnetChainConfig,
			blockId:  big.NewInt(1),
			rules:    opera.DefaultEconomyRules(),
		}
		out, err := e.sendCall()
		assert.NotNil(t, err)
		assert.Nil(t, out)
	})

}

func TestEvmExecutor_sendEstimateGas(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockArchive := state.NewMockNonCommittableStateDB(ctrl)
	e := &EvmExecutor{
		args:     newTxArgs(map[string]interface{}{"from": "0x0000000000000000000000000000000000000001"}),
		archive:  mockArchive,
		chainCfg: params.MainnetChainConfig,
		blockId:  big.NewInt(1),
		rules:    opera.DefaultEconomyRules(),
	}
	assert.Panics(t, func() {
		_, _ = e.sendEstimateGas()
	})
}

func TestEvmExecutor_executable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockArchive := state.NewMockNonCommittableStateDB(ctrl)
	mockArchive.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1000000000)).AnyTimes()
	mockArchive.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockArchive.EXPECT().Prepare(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockArchive.EXPECT().GetNonce(gomock.Any()).Return(uint64(1234)).AnyTimes()
	mockArchive.EXPECT().SetNonce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockArchive.EXPECT().GetRefund().Return(uint64(0)).AnyTimes()
	mockArchive.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	mockArchive.EXPECT().GetCode(gomock.Any()).Return([]uint8{}).AnyTimes()
	mockArchive.EXPECT().Snapshot().Return(0).AnyTimes()
	mockArchive.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
	mockArchive.EXPECT().GetCodeHash(gomock.Any()).Return(common.Hash{}).AnyTimes()
	mockArchive.EXPECT().GetStorageRoot(gomock.Any()).Return(common.Hash{}).AnyTimes()
	mockArchive.EXPECT().GetRefund().Return(uint64(0)).AnyTimes()
	mockArchive.EXPECT().RevertToSnapshot(gomock.Any()).AnyTimes()

	t.Run("success", func(t *testing.T) {
		jsonData := []byte(`{"from": "0x0000000000000000000000000000000000000001","to":"0x0000000000000000000000000000000000000002"}`)
		var args ethapi.TransactionArgs
		err := json.Unmarshal(jsonData, &args)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		e := &EvmExecutor{
			args:     args,
			archive:  mockArchive,
			chainCfg: params.MainnetChainConfig,
			blockId:  big.NewInt(1),
			rules:    opera.DefaultEconomyRules(),
		}
		failed, result, err := e.executable(210000)
		assert.False(t, failed)
		assert.NotNil(t, result)
		assert.Nil(t, err)
	})

	t.Run("success gas limit", func(t *testing.T) {
		jsonData := []byte(`{"from": "0x0000000000000000000000000000000000000001","to":"0x0000000000000000000000000000000000000002"}`)
		var args ethapi.TransactionArgs
		err := json.Unmarshal(jsonData, &args)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		e := &EvmExecutor{
			args:     args,
			archive:  mockArchive,
			chainCfg: params.MainnetChainConfig,
			blockId:  big.NewInt(1),
			rules:    opera.DefaultEconomyRules(),
		}
		failed, result, err := e.executable(0)
		assert.True(t, failed)
		assert.Nil(t, result)
		assert.Nil(t, err)
	})

	t.Run("success bail out", func(t *testing.T) {
		jsonData := []byte(`{"from": "0x0000000000000000000000000000000000000001","to":"0x0000000000000000000000000000000000000002"}`)
		var args ethapi.TransactionArgs
		err := json.Unmarshal(jsonData, &args)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		e := &EvmExecutor{
			args:     args,
			archive:  mockArchive,
			chainCfg: params.MainnetChainConfig,
			blockId:  big.NewInt(1),
			rules:    opera.DefaultEconomyRules(),
		}
		failed, result, err := e.executable(21000)
		assert.True(t, failed)
		assert.Nil(t, result)
		assert.NotNil(t, err)
	})
}
