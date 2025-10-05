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
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/sonic/opera"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRpc_Execute(t *testing.T) {
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
	mockArchive.EXPECT().GetState(gomock.Any(), gomock.Any()).Return(common.HexToHash("0x1234"))

	t.Run("getBalance", func(t *testing.T) {
		rec := &RequestAndResults{
			Query: &Body{
				MethodBase: "getBalance",
				Params:     []interface{}{"1234567890abcdef"},
			},
		}
		cfg := &utils.Config{}

		out, err := Execute(uint64(0), rec, mockArchive, cfg)
		assert.NotNil(t, out)
		assert.Nil(t, err)
	})

	t.Run("getTransactionCount", func(t *testing.T) {
		rec := &RequestAndResults{
			Query: &Body{
				MethodBase: "getTransactionCount",
				Params:     []interface{}{"1234567890abcdef"},
			},
		}
		cfg := &utils.Config{}

		out, err := Execute(uint64(0), rec, mockArchive, cfg)
		assert.NotNil(t, out)
		assert.Nil(t, err)
	})

	t.Run("call", func(t *testing.T) {
		p := map[string]interface{}{"from": "0x0000000000000000000000000000000000000001", "to": "0x1"}
		rec := &RequestAndResults{
			Query: &Body{
				MethodBase: "call",
				Params:     []interface{}{p},
			},
			Timestamp: uint64(42),
		}
		cfg := &utils.Config{
			ChainID: utils.MainnetChainID,
		}
		out, err := Execute(uint64(0), rec, mockArchive, cfg)
		assert.NotNil(t, out)
		assert.Nil(t, err)
	})

	t.Run("estimateGas", func(t *testing.T) {
		rec := &RequestAndResults{
			Query: &Body{
				MethodBase: "estimateGas",
				Params:     []interface{}{"1234567890abcdef"},
			},
		}
		cfg := &utils.Config{}

		out, err := Execute(uint64(0), rec, mockArchive, cfg)
		assert.Nil(t, out)
		assert.Nil(t, err)
	})

	t.Run("getCode", func(t *testing.T) {
		rec := &RequestAndResults{
			Query: &Body{
				MethodBase: "getCode",
				Params:     []interface{}{"1234567890abcdef"},
			},
		}
		cfg := &utils.Config{}

		out, err := Execute(uint64(0), rec, mockArchive, cfg)
		assert.NotNil(t, out)
		assert.Nil(t, err)
	})

	t.Run("getStorageAt", func(t *testing.T) {
		rec := &RequestAndResults{
			Query: &Body{
				MethodBase: "getStorageAt",
				Params:     []interface{}{"1234567890abcdef", "0x0"},
			},
		}
		cfg := &utils.Config{}

		out, err := Execute(uint64(0), rec, mockArchive, cfg)
		assert.NotNil(t, out)
		assert.Nil(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		rec := &RequestAndResults{
			Query: &Body{
				MethodBase: "invalid",
				Params:     []interface{}{"1234567890abcdef"},
			},
		}
		cfg := &utils.Config{}

		out, err := Execute(uint64(0), rec, mockArchive, cfg)
		assert.Nil(t, out)
		assert.Nil(t, err)
	})

}

func TestRpc_executeGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchive := state.NewMockVmStateDB(ctrl)
	mockArchive.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(42))
	out := executeGetBalance("0x1234567890abcdef", mockArchive)
	assert.NotNil(t, out)
}

func TestRpc_executeGetTransactionCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchive := state.NewMockVmStateDB(ctrl)
	mockArchive.EXPECT().GetNonce(gomock.Any()).Return(uint64(42))
	out := executeGetTransactionCount("0x1234567890abcdef", mockArchive)
	assert.NotNil(t, out)
}

func TestRpc_executeCall(t *testing.T) {
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

	e := &EvmExecutor{
		args:     newTxArgs(map[string]interface{}{"from": "0x0000000000000000000000000000000000000001", "to": "0x0000000000000000000000000000000000000002"}),
		archive:  mockArchive,
		chainCfg: params.MainnetChainConfig,
		blockId:  big.NewInt(1),
		rules:    opera.DefaultEconomyRules(),
	}
	out := executeCall(e)
	assert.NotNil(t, out)
	assert.Nil(t, out.err)
}

func TestRpc_executeGetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchive := state.NewMockVmStateDB(ctrl)
	mockArchive.EXPECT().GetCode(gomock.Any()).Return([]byte{0x10})
	out := executeGetCode("0x1234567890abcdef", mockArchive)
	assert.NotNil(t, out)
}

func TestRpc_executeGetStorageAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArchive := state.NewMockVmStateDB(ctrl)
	mockArchive.EXPECT().GetState(gomock.Any(), gomock.Any()).Return(common.HexToHash("0x1234"))
	out := executeGetStorageAt([]interface{}{"0x1234567890abcdef", "0x0"}, mockArchive)
	assert.NotNil(t, out)
}
