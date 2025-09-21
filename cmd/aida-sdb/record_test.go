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

package main

import (
	"math/big"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestSdbRecord_AllDbEventsAreIssuedInOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := executor.NewMockProvider[txcontext.TxContext](ctrl)
	processor := executor.NewMockProcessor[txcontext.TxContext](ctrl)
	ext := executor.NewMockExtension[txcontext.TxContext](ctrl)
	path := t.TempDir() + "test_trace"

	cfg := utils.NewTestConfig(t, utils.MainnetChainID, 10, 11, false, "")
	cfg.TraceFile = path
	cfg.SyncPeriodLength = 1
	provider.EXPECT().
		Run(10, 12, gomock.Any()).
		DoAndReturn(func(from int, to int, consumer executor.Consumer[txcontext.TxContext]) error {
			for i := from; i < to; i++ {
				err := consumer(executor.TransactionInfo[txcontext.TxContext]{Block: i, Transaction: 3, Data: substatecontext.NewTxContext(emptyTx)})
				require.NoError(t, err)
				err = consumer(executor.TransactionInfo[txcontext.TxContext]{Block: i, Transaction: utils.PseudoTx, Data: substatecontext.NewTxContext(emptyTx)})
				require.NoError(t, err)
			}
			return nil
		})

	// All transactions are processed in order
	gomock.InOrder(
		ext.EXPECT().PreRun(executor.AtBlock[txcontext.TxContext](10), gomock.Any()),

		// block 10
		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](10, 3), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](10, 3), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](10, 3), gomock.Any()),

		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](10, utils.PseudoTx), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](10, utils.PseudoTx), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](10, utils.PseudoTx), gomock.Any()),

		// block 11
		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](11, 3), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](11, 3), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](11, 3), gomock.Any()),

		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](11, utils.PseudoTx), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](11, utils.PseudoTx), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](11, utils.PseudoTx), gomock.Any()),

		ext.EXPECT().PostRun(executor.AtBlock[txcontext.TxContext](12), gomock.Any(), nil),
	)

	if err := record(cfg, provider, processor, []executor.Extension[txcontext.TxContext]{ext}); err != nil {
		t.Errorf("record failed: %v", err)
	}
}

// emptyTx is a dummy substate that will be processed without crashing.
var emptyTx = &substate.Substate{
	Env: &substate.Env{},
	Message: &substate.Message{
		GasPrice: big.NewInt(12),
	},
	Result: &substate.Result{
		GasUsed: 1,
	},
}

func TestCmd_RunRecordCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	traceFile := filepath.Join(tempDir, "trace.bin")
	require.NoError(t, utils.CopyDir(path.Join(testDataDir, "sample-pb-db"), aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&RecordCommand}
	args := utils.NewArgs("test").
		Arg(RecordCommand.Name).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.TraceFileFlag.Name, traceFile).
		Flag(utils.WorkersFlag.Name, 1).
		Arg("1").
		Arg("1000").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	stat, err := os.Stat(traceFile)
	require.NoError(t, err)
	assert.NotZero(t, stat.Size())
}
