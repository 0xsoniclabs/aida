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

package logger

import (
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"go.uber.org/mock/gomock"
)

func TestEthStateTestLogger_PreBlockLogsProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	ext := makeEthStateTestLogger(log, 2)
	s := executor.State[txcontext.TxContext]{Data: ethtest.CreateTestTransaction(t)}

	gomock.InOrder(
		log.EXPECT().Infof("Currently running:\n%s", s.Data),
		log.EXPECT().Infof("Currently running:\n%s", s.Data),
		log.EXPECT().Noticef("%v tests has been processed so far...", 2),
	)

	err := ext.PreBlock(s, &executor.Context{})
	if err != nil {
		t.Fatalf("pre-tx failed: %v", err)
	}

	err = ext.PreBlock(s, &executor.Context{})
	if err != nil {
		t.Fatalf("pre-tx failed: %v", err)
	}
}

func TestEthStateTestLogger_PostRunLogsOverall(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	ext := makeEthStateTestLogger(log, 0)
	ext.overall = 2
	s := executor.State[txcontext.TxContext]{Data: ethtest.CreateTestTransaction(t)}

	gomock.InOrder(
		log.EXPECT().Noticef("Total %v tests processed.", 2),
	)

	err := ext.PostRun(s, &executor.Context{}, nil)
	if err != nil {
		t.Fatalf("post-run failed: %v", err)
	}

}

func TestEthStateTestLogger_MakeEthStateTestLogger(t *testing.T) {
	cfg := &utils.Config{}
	ext := MakeEthStateTestLogger(cfg, 0)

	if _, ok := ext.(*ethStateTestLogger); !ok {
		t.Fatal("unexpected extension type")
	}
	if ext.(*ethStateTestLogger).reportFrequency != defaultReportFrequency {
		t.Fatal("default report frequency is not set")
	}
}
