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
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCmd_RunReplaySubstate(t *testing.T) {
	_, _, path := utils.CreateTestSubstateDb(t)
	app := cli.NewApp()
	app.Action = RunReplaySubstate
	app.Flags = []cli.Flag{
		&utils.AidaDbFlag,
		&utils.SubstateEncodingFlag,
		&utils.TraceFileFlag,
	}
	traceFile := t.TempDir() + "/trace-file"
	writer, err := tracer.NewFileWriter(traceFile)
	require.NoError(t, err)
	err = writer.WriteData(append(bigendian.Uint64ToBytes(0), bigendian.Uint64ToBytes(2)...))
	require.NoError(t, err)
	op, err := tracer.EncodeArgOp(tracer.BeginBlockID, tracer.NoArgID, tracer.NoArgID, tracer.NoArgID)
	require.NoError(t, err)
	err = writer.WriteUint16(op)
	require.NoError(t, err)
	err = writer.WriteData(bigendian.Uint64ToBytes(123))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	err = app.Run([]string{RunReplaySubstateCmd.Name, "--trace-file", traceFile, "--substate-encoding", "pb", "--aida-db", path, "first", "last"})
	require.NoError(t, err)
}

func TestSdbReplaySubstate_AllDbEventsAreIssuedInOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := executor.NewMockProvider[txcontext.TxContext](ctrl)
	processor := executor.NewMockProcessor[txcontext.TxContext](ctrl)
	ext := executor.NewMockExtension[txcontext.TxContext](ctrl)

	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.KeepDb = false

	cfg.First = 0
	cfg.Last = 0

	provider.EXPECT().
		Run(0, 1, gomock.Any()).
		DoAndReturn(func(from int, to int, consumer executor.Consumer[txcontext.TxContext]) error {
			for i := from; i < to; i++ {
				err := consumer(executor.TransactionInfo[txcontext.TxContext]{Block: 0, Transaction: 0, Data: substatecontext.NewTxContext(emptyTx)})
				require.NoError(t, err)
			}
			return nil
		})

	// All transactions are processed in order
	gomock.InOrder(
		ext.EXPECT().PreRun(executor.AtBlock[txcontext.TxContext](0), gomock.Any()),

		// tx 0
		ext.EXPECT().PreBlock(executor.AtBlock[txcontext.TxContext](0), gomock.Any()),
		ext.EXPECT().PreTransaction(executor.AtTransaction[txcontext.TxContext](0, 0), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[txcontext.TxContext](0, 0), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[txcontext.TxContext](0, 0), gomock.Any()),
		ext.EXPECT().PostBlock(executor.AtBlock[txcontext.TxContext](0), gomock.Any()),

		ext.EXPECT().PostRun(executor.AtBlock[txcontext.TxContext](1), gomock.Any(), nil),
	)

	if err := replaySubstate(cfg, provider, processor, nil, []executor.Extension[txcontext.TxContext]{ext}); err != nil {
		t.Errorf("record failed: %v", err)
	}
}
