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

package validator

import (
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEthereumPreTransactionUpdater_FixBalanceWhenNewBalanceIsHigher(t *testing.T) {
	cfg := &utils.Config{}
	cfg.ChainID = utils.EthereumChainID

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	data := createTestTransaction()
	ctx := new(executor.Context)
	ctx.State = db
	st := executor.State[txcontext.TxContext]{Block: getEthereumExceptionBlock(), Transaction: 1, Data: data}

	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x1")).Return(true),
		db.EXPECT().GetBalance(common.HexToAddress("0x1")).Return(uint256.NewInt(1)),
		db.EXPECT().SubBalance(common.HexToAddress("0x1"), uint256.NewInt(1), tracing.BalanceChangeUnspecified),
		db.EXPECT().AddBalance(common.HexToAddress("0x1"), uint256.NewInt(1000), tracing.BalanceChangeUnspecified),
	)

	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x2")).Return(true),
		db.EXPECT().GetBalance(common.HexToAddress("0x2")).Return(uint256.NewInt(2000)),
	)

	ext := makeEthereumDbPreTransactionUpdater(cfg, log)
	err := ext.PreTransaction(st, ctx)
	if err != nil {
		t.Fatal("post-transaction unexpected error: ", err)
	}
}

func TestEthereumPreTransactionUpdater_DontFixBalanceIfLower(t *testing.T) {
	cfg := &utils.Config{}
	cfg.ChainID = utils.EthereumChainID

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	data := ethtest.CreateTestTransaction(t)
	ctx := new(executor.Context)
	ctx.State = db
	st := executor.State[txcontext.TxContext]{Block: getEthereumExceptionBlock(), Transaction: 1, Data: data}

	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x1")).Return(true),
		db.EXPECT().GetBalance(common.HexToAddress("0x1")).Return(uint256.NewInt(10000)),
	)

	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x2")).Return(true),
		db.EXPECT().GetBalance(common.HexToAddress("0x2")).Return(uint256.NewInt(2000)),
	)

	ext := makeEthereumDbPreTransactionUpdater(cfg, log)
	err := ext.PreTransaction(st, ctx)
	if err != nil {
		t.Fatal("post-transaction unexpected error: ", err)
	}
}

func TestEthereumPreTransactionUpdater_BeaconRootsAddressStorageException(t *testing.T) {
	testEthereumSystemContractStorageException(t, params.BeaconRootsAddress)
}

func TestEthereumPreTransactionUpdater_WithdrawalQueueAddressStorageException(t *testing.T) {
	testEthereumSystemContractStorageException(t, params.WithdrawalQueueAddress)
}

func TestEthereumPreTransactionUpdater_ConsolidationQueueAddressStorageException(t *testing.T) {
	testEthereumSystemContractStorageException(t, params.ConsolidationQueueAddress)
}

func testEthereumSystemContractStorageException(t *testing.T, address common.Address) {
	cfg := &utils.Config{}
	cfg.ChainID = utils.EthereumChainID

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	data := createEthereumSystemContractTestTransaction(address)

	ctx := new(executor.Context)
	ctx.State = db
	st := executor.State[txcontext.TxContext]{Block: getEthereumExceptionBlock(), Transaction: 1, Data: data}

	gomock.InOrder(
		db.EXPECT().Exist(address).Return(true),
		db.EXPECT().GetBalance(address).Return(uint256.NewInt(1)),
		db.EXPECT().GetState(address, common.HexToHash("0x1")),
		db.EXPECT().SetState(address, common.HexToHash("0x1"), common.HexToHash("0x2")),
		db.EXPECT().EndTransaction().Return(nil),
		db.EXPECT().BeginTransaction(uint32(utils.PseudoTx)),
	)

	ext := makeEthereumDbPreTransactionUpdater(cfg, log)
	err := ext.PreTransaction(st, ctx)
	if err != nil {
		t.Fatal("post-transaction unexpected error: ", err)
	}
}

func TestEthereumPreTransactionUpdater_DaoFork(t *testing.T) {
	cfg := &utils.Config{}
	cfg.ChainID = utils.EthereumChainID

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	data := createDaoForkAddressTestTransaction()

	ctx := new(executor.Context)
	ctx.State = db
	st := executor.State[txcontext.TxContext]{Block: getEthereumExceptionBlock(), Transaction: 1, Data: data}

	gomock.InOrder(
		db.EXPECT().Exist(params.DAODrainList()[0]).Return(true),
		db.EXPECT().GetBalance(params.DAODrainList()[0]).Return(uint256.NewInt(1)),
		db.EXPECT().SubBalance(params.DAODrainList()[0], uint256.NewInt(1), tracing.BalanceChangeUnspecified),
		db.EXPECT().AddBalance(params.DAODrainList()[0], uint256.NewInt(0), tracing.BalanceChangeUnspecified),
	)

	ext := makeEthereumDbPreTransactionUpdater(cfg, log)
	err := ext.PreTransaction(st, ctx)
	if err != nil {
		t.Fatal("post-transaction unexpected error: ", err)
	}
}

func createEthereumSystemContractTestTransaction(address common.Address) txcontext.TxContext {
	return substatecontext.NewTxContext(&substate.Substate{
		InputSubstate: substate.WorldState{
			substatetypes.BytesToAddress(address.Bytes()): &substate.Account{
				Balance: uint256.NewInt(1),
				Storage: map[substatetypes.Hash]substatetypes.Hash{
					substatetypes.BytesToHash([]byte{0x1}): substatetypes.BytesToHash([]byte{0x2})},
			},
		},
	})
}

func createDaoForkAddressTestTransaction() txcontext.TxContext {
	return substatecontext.NewTxContext(&substate.Substate{
		InputSubstate: substate.WorldState{
			substatetypes.BytesToAddress(params.DAODrainList()[0].Bytes()): &substate.Account{
				Balance: uint256.NewInt(0),
			},
		},
	})
}

func TestEthereumDbPreTransactionUpdater_PreRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	cfg := &utils.Config{}
	st := executor.State[txcontext.TxContext]{}
	ctx := new(executor.Context)
	log.EXPECT().Warning(gomock.Any())
	ext := &ethereumDbPreTransactionUpdater{
		cfg: cfg,
		log: log,
	}
	err := ext.PreRun(st, ctx)
	assert.NoError(t, err)
}

func TestEthereumDbPreTransactionUpdater_MakeEthereumDbPreTransactionUpdater(t *testing.T) {
	cfg := &utils.Config{}
	cfg.ChainID = utils.PseudoTx
	ext := MakeEthereumDbPreTransactionUpdater(cfg)
	assert.IsType(t, extension.NilExtension[txcontext.TxContext]{}, ext)
}
