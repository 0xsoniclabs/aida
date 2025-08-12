// Copyright 2024 Fantom Foundation
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

package main

import (
	"errors"
	"flag"
	"fmt"

	"strconv"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestCmd_RunEthereumTest(t *testing.T) {
	app := cli.NewApp()
	app.Action = RunEthereumTest
	app.Flags = []cli.Flag{
		&utils.ChainIDFlag,
	}

	err := app.Run([]string{RunEthTestsCmd.Name, "--chainid", strconv.Itoa(int(utils.EthTestsChainID)), t.TempDir()})
	require.NoError(t, err)
}

func TestVmSdb_Eth_AllDbEventsAreIssuedInOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := executor.NewMockProvider[txcontext.TxContext](ctrl)
	db := state.NewMockStateDB(ctrl)

	cfg := utils.NewTestConfig(t, utils.EthTestsChainID, 2, 4, false, "Cancun")
	cfg.ContinueOnFailure = true

	data := ethtest.CreateTestTransaction(t)

	provider.EXPECT().
		Run(2, 5, gomock.Any()).
		DoAndReturn(func(_ int, _ int, consumer executor.Consumer[txcontext.TxContext]) error {
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 2, Transaction: 0, Data: data})
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 3, Transaction: 1, Data: data})
			return nil
		})

	gomock.InOrder(
		// Tx 1
		db.EXPECT().BeginBlock(uint64(2)),
		db.EXPECT().BeginTransaction(uint32(0)),
		db.EXPECT().SetTxContext(gomock.Any(), 0),
		db.EXPECT().Snapshot().Return(15),
		db.EXPECT().GetNonce(data.GetMessage().From).Return(uint64(1)),
		db.EXPECT().GetCode(data.GetMessage().From).Return([]byte{}),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1000)),
		db.EXPECT().SubBalance(gomock.Any(), gomock.Any(), tracing.BalanceDecreaseGasBuy),
		db.EXPECT().RevertToSnapshot(15),
		db.EXPECT().GetLogs(common.HexToHash(fmt.Sprintf("0x%016d%016d", 2, 0)), uint64(2), common.HexToHash(fmt.Sprintf("0x%016d", 2)), uint64(10)),
		db.EXPECT().EndTransaction(),
		db.EXPECT().EndBlock(),

		db.EXPECT().BeginBlock(uint64(3)),
		db.EXPECT().BeginTransaction(uint32(0)),
		db.EXPECT().SetTxContext(gomock.Any(), 1),
		db.EXPECT().Snapshot().Return(15),
		db.EXPECT().GetNonce(data.GetMessage().From).Return(uint64(1)),
		db.EXPECT().GetCode(data.GetMessage().From).Return([]byte{}),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1000)),
		db.EXPECT().SubBalance(gomock.Any(), gomock.Any(), tracing.BalanceDecreaseGasBuy),
		db.EXPECT().RevertToSnapshot(15),
		db.EXPECT().GetLogs(common.HexToHash(fmt.Sprintf("0x%016d%016d", 3, 1)), uint64(3), common.HexToHash(fmt.Sprintf("0x%016d", 3)), uint64(10)),
		db.EXPECT().EndTransaction(),
		db.EXPECT().EndBlock(),
	)

	processor, err := executor.MakeEthTestProcessor(cfg)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	err = runEth(cfg, provider, db, processor, nil)
	if err != nil {
		errors.Unwrap(err)
		if strings.Contains(err.Error(), "intrinsic gas too low") {
			return
		}
		t.Fatalf("run failed, %v", err)
	}
}

func TestVmSdb_Eth_AllTransactionsAreProcessedInOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := executor.NewMockProvider[txcontext.TxContext](ctrl)
	db := state.NewMockStateDB(ctrl)
	ext := executor.NewMockExtension[txcontext.TxContext](ctrl)
	processor := executor.NewMockProcessor[txcontext.TxContext](ctrl)

	cfg := utils.NewTestConfig(t, utils.EthTestsChainID, 2, 4, false, "Cancun")
	data := ethtest.CreateTestTransaction(t)

	// Simulate the execution of 4 transactions
	provider.EXPECT().
		Run(2, 5, gomock.Any()).
		DoAndReturn(func(_ int, _ int, consumer executor.Consumer[txcontext.TxContext]) error {
			// Tx 1
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 2, Transaction: 1, Data: data})
			// Tx 2
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 3, Transaction: 2, Data: data})
			// Tx 3
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 4, Transaction: 3, Data: data})
			// Tx 4
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 5, Transaction: utils.PseudoTx, Data: data})
			return nil
		})

	// The expectation is that all of those blocks and transactions
	// are properly opened, prepared, executed, and closed.
	// Since we are running sequential mode with 1 worker,
	// all block and transactions need to be in order.
	gomock.InOrder(
		ext.EXPECT().PreRun(executor.AtBlock[txcontext.TxContext](2), gomock.Any()),

		// Tx 1
		db.EXPECT().BeginBlock(uint64(2)),
		db.EXPECT().BeginTransaction(uint32(0)),
		ext.EXPECT().PreBlock(executor.AtTransaction[txcontext.TxContext](2, 0), gomock.Any()),
		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](2, 1), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](2, 1), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](2, 1), gomock.Any()),
		ext.EXPECT().PostBlock(executor.AtTransaction[txcontext.TxContext](2, 1), gomock.Any()),
		db.EXPECT().EndTransaction(),
		db.EXPECT().EndBlock(),
		// Tx 2
		db.EXPECT().BeginBlock(uint64(3)),
		db.EXPECT().BeginTransaction(uint32(1)),
		ext.EXPECT().PreBlock(executor.AtTransaction[txcontext.TxContext](3, 1), gomock.Any()),
		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](3, 2), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](3, 2), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](3, 2), gomock.Any()),
		ext.EXPECT().PostBlock(executor.AtTransaction[txcontext.TxContext](3, 2), gomock.Any()),
		db.EXPECT().EndTransaction(),
		db.EXPECT().EndBlock(),
		// Tx 3
		db.EXPECT().BeginBlock(uint64(4)),
		db.EXPECT().BeginTransaction(uint32(2)),
		ext.EXPECT().PreBlock(executor.AtTransaction[txcontext.TxContext](4, 2), gomock.Any()),
		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](4, 3), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](4, 3), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](4, 3), gomock.Any()),
		ext.EXPECT().PostBlock(executor.AtTransaction[txcontext.TxContext](4, 3), gomock.Any()),
		db.EXPECT().EndTransaction(),
		db.EXPECT().EndBlock(),
		// Tx 4
		db.EXPECT().BeginBlock(uint64(5)),
		db.EXPECT().BeginTransaction(uint32(3)),
		ext.EXPECT().PreBlock(executor.AtTransaction[txcontext.TxContext](5, 3), gomock.Any()),
		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](5, utils.PseudoTx), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](5, utils.PseudoTx), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](5, utils.PseudoTx), gomock.Any()),
		ext.EXPECT().PostBlock(executor.AtTransaction[txcontext.TxContext](5, utils.PseudoTx), gomock.Any()),
		db.EXPECT().EndTransaction(),
		db.EXPECT().EndBlock(),
		ext.EXPECT().PostRun(executor.AtBlock[txcontext.TxContext](5), gomock.Any(), nil),
	)

	if err := runEth(cfg, provider, db, processor, []executor.Extension[txcontext.TxContext]{ext}); err != nil {
		t.Errorf("run failed: %v", err)
	}
}

func TestVmSdb_Eth_ValidationDoesNotFailOnValidTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := executor.NewMockProvider[txcontext.TxContext](ctrl)
	db := state.NewMockStateDB(ctrl)

	cfg := utils.NewTestConfig(t, utils.EthTestsChainID, 2, 4, true, "Cancun")
	data := ethtest.CreateTestTransaction(t)

	provider.EXPECT().
		Run(2, 5, gomock.Any()).
		DoAndReturn(func(_ int, _ int, consumer executor.Consumer[txcontext.TxContext]) error {
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 2, Transaction: 1, Data: data})
			return nil
		})

	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x1")).Return(true),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1000)),
		db.EXPECT().GetNonce(common.HexToAddress("0x1")).Return(uint64(1)),
		db.EXPECT().GetCode(common.HexToAddress("0x1")).Return([]byte{}),
	)
	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x2")).Return(true),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(2000)),
		db.EXPECT().GetNonce(common.HexToAddress("0x2")).Return(uint64(2)),
		db.EXPECT().GetCode(common.HexToAddress("0x2")).Return([]byte{}),
	)

	gomock.InOrder(
		// Tx execution
		db.EXPECT().BeginBlock(uint64(2)),
		db.EXPECT().BeginTransaction(uint32(0)),
		db.EXPECT().SetTxContext(gomock.Any(), 1),
		db.EXPECT().Snapshot().Return(15),
		db.EXPECT().GetNonce(data.GetMessage().From).Return(uint64(1)),
		db.EXPECT().GetCode(data.GetMessage().From).Return(nil),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1000)),
		db.EXPECT().SubBalance(gomock.Any(), gomock.Any(), tracing.BalanceDecreaseGasBuy),
		db.EXPECT().RevertToSnapshot(15),
		db.EXPECT().GetLogs(common.HexToHash(fmt.Sprintf("0x%016d%016d", 2, 1)), uint64(2), common.HexToHash(fmt.Sprintf("0x%016d", 2)), uint64(10)),
		db.EXPECT().EndTransaction(),
		db.EXPECT().EndBlock(),
		db.EXPECT().GetHash(),
		// EndTransaction is not called because execution fails
	)

	processor, err := executor.MakeEthTestProcessor(cfg)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	err = runEth(cfg, provider, db, processor, nil)
	if err != nil {
		errors.Unwrap(err)
		if strings.Contains(err.Error(), "intrinsic gas too low") {
			return
		}
		t.Fatal("run failed")
	}
}

func TestVmSdb_Eth_ValidationDoesFailOnInvalidTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := executor.NewMockProvider[txcontext.TxContext](ctrl)
	db := state.NewMockStateDB(ctrl)

	cfg := utils.NewTestConfig(t, utils.EthTestsChainID, 2, 4, true, "Cancun")
	data := ethtest.CreateTestTransaction(t)

	provider.EXPECT().
		Run(2, 5, gomock.Any()).
		DoAndReturn(func(_ int, _ int, consumer executor.Consumer[txcontext.TxContext]) error {
			consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 2, Transaction: 1, Data: data})
			return nil
		})

	// state map contains two accounts, but the validation order of map is not guaranteed
	gomock.InOrder(
		// Tx 1
		// Validation fails on incorrect input so no db events are expected
		// first account has correct data
		db.EXPECT().Exist(common.HexToAddress("0x1")).Return(true),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1000)),
		db.EXPECT().GetNonce(common.HexToAddress("0x1")).Return(uint64(1)),
		db.EXPECT().GetCode(common.HexToAddress("0x1")).Return([]byte{}),
	)
	gomock.InOrder(
		// second account has incorrect balance
		db.EXPECT().Exist(common.HexToAddress("0x2")).Return(true),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(9999)), // incorrect balance
		db.EXPECT().GetNonce(common.HexToAddress("0x2")).Return(uint64(2)),
		db.EXPECT().GetCode(common.HexToAddress("0x2")).Return([]byte{}),
	)
	db.EXPECT().BeginBlock(uint64(2))
	db.EXPECT().BeginTransaction(uint32(0))

	processor, err := executor.MakeEthTestProcessor(cfg)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	err = runEth(cfg, provider, db, processor, nil)
	if err == nil {
		t.Fatal("run must fail")
	}

	errors.Unwrap(err)
	if !strings.Contains(err.Error(), "pre alloc validation failed") {
		t.Fatalf("unexpected error\ngot: %v\n want: %v", err, "pre alloc validation failed")
	}
}

func TestVmSdb_EthTest_FailsWhenChainIDIsNotSet(t *testing.T) {
	flagSet := flag.NewFlagSet("utils_config_test", 0)
	flagSet.Int(utils.ChainIDFlag.Name, int(utils.SonicMainnetChainID), "Chain ID.")
	err := flagSet.Parse([]string{t.TempDir()})
	require.NoError(t, err)

	ctx := cli.NewContext(cli.NewApp(), flagSet, nil)
	ctx.Command = &cli.Command{Name: "test_command"}

	err = RunEthereumTest(ctx)
	require.ErrorContains(t, err, fmt.Sprintf("please specify chain ID using --%s flag (1337 for most cases for this tool)", utils.ChainIDFlag.Name))
}
