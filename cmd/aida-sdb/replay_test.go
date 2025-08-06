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
	"github.com/0xsoniclabs/aida/utils"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCmd_RunReplay(t *testing.T) {
	app := cli.NewApp()
	app.Action = RunReplay
	app.Flags = []cli.Flag{
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

	err = app.Run([]string{RunReplayCmd.Name, "--trace-file", traceFile, "first", "last"})
	require.NoError(t, err)
}

func TestSdbReplay_AllDbEventsAreIssuedInOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := executor.NewMockProvider[tracer.Operation](ctrl)
	processor := executor.NewMockProcessor[tracer.Operation](ctrl)
	ext := executor.NewMockExtension[tracer.Operation](ctrl)

	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.KeepDb = false

	cfg.First = 0
	cfg.Last = 0

	provider.EXPECT().
		Run(0, 1, gomock.Any()).
		DoAndReturn(func(from int, to int, consumer executor.Consumer[tracer.Operation]) error {
			for i := from; i < to; i++ {
				err := consumer(executor.TransactionInfo[tracer.Operation]{Block: 0, Transaction: 0, Data: tracer.Operation{
					Op: tracer.BeginBlockID,
					Data: []any{
						uint64(123),
					},
				}})
				require.NoError(t, err)
			}
			return nil
		})

	// All transactions are processed in order
	gomock.InOrder(
		ext.EXPECT().PreRun(executor.AtBlock[tracer.Operation](0), gomock.Any()),

		// tx 0
		ext.EXPECT().PreBlock(executor.AtBlock[tracer.Operation](0), gomock.Any()),
		ext.EXPECT().PreTransaction(executor.AtTransaction[tracer.Operation](0, 0), gomock.Any()),
		processor.EXPECT().Process(executor.AtTransaction[tracer.Operation](0, 0), gomock.Any()),
		ext.EXPECT().PostTransaction(executor.AtTransaction[tracer.Operation](0, 0), gomock.Any()),
		ext.EXPECT().PostBlock(executor.AtBlock[tracer.Operation](0), gomock.Any()),

		ext.EXPECT().PostRun(executor.AtBlock[tracer.Operation](1), gomock.Any(), nil),
	)

	if err := replay(cfg, provider, nil, processor, []executor.Extension[tracer.Operation]{ext}, nil); err != nil {
		t.Errorf("record failed: %v", err)
	}
}
