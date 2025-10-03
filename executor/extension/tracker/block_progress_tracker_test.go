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

package tracker

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	"go.uber.org/mock/gomock"
)

const testStateDbInfoFrequency = 2

func TestSubstateProgressTrackerExtension_NoLoggerIsCreatedIfDisabled(t *testing.T) {
	cfg := &utils.Config{}
	cfg.TrackProgress = false
	ext := MakeBlockProgressTracker(cfg, testStateDbInfoFrequency)
	if _, ok := ext.(extension.NilExtension[txcontext.TxContext]); !ok {
		t.Errorf("Logger is enabled although not set in configuration")
	}

}

func TestSubstateProgressTrackerExtension_LoggingHappens(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.First = 4
	dummyStateDbPath := t.TempDir()

	if err := os.WriteFile(dummyStateDbPath+"/dummy.txt", []byte("hello world"), 0x600); err != nil {
		t.Fatalf("failed to prepare disk content")
	}

	ext := makeBlockProgressTracker(cfg, testStateDbInfoFrequency, log)

	ctx := &executor.Context{
		State:           db,
		StateDbPath:     dummyStateDbPath,
		ExecutionResult: substatecontext.NewReceipt(&substate.Result{GasUsed: 100}),
	}

	s := substatecontext.NewTxContext(&substate.Substate{
		Result: &substate.Result{
			Status: 0,
		},
	})

	gomock.InOrder(
		db.EXPECT().GetMemoryUsage().Return(&state.MemoryUsage{UsedBytes: 1234}),
		log.EXPECT().Noticef(substateProgressTrackerReportFormat,
			6, uint64(1234), int64(11),
			executor.MatchRate(gomock.All(executor.Gt(3), executor.Lt(5)), "blkRate"),
			executor.MatchRate(gomock.All(executor.Gt(7), executor.Lt(9)), "txRate"),
			executor.MatchRate(gomock.All(executor.Gt(700), executor.Lt(900)), "gasRate"),
			executor.MatchRate(gomock.All(executor.Gt(3), executor.Lt(5)), "blkRate"),
			executor.MatchRate(gomock.All(executor.Gt(7), executor.Lt(9)), "txRate"),
			executor.MatchRate(gomock.All(executor.Gt(700), executor.Lt(900)), "gasRate"),
		),
		db.EXPECT().GetMemoryUsage().Return(&state.MemoryUsage{UsedBytes: 4321}),
		log.EXPECT().Noticef(substateProgressTrackerReportFormat,
			8, uint64(4321), int64(11),
			executor.MatchRate(gomock.All(executor.Gt(3), executor.Lt(5)), "blkRate"),
			executor.MatchRate(gomock.All(executor.Gt(1), executor.Lt(2)), "txRate"),
			executor.MatchRate(gomock.All(executor.Gt(180), executor.Lt(220)), "gasRate"),
			executor.MatchRate(gomock.All(executor.Gt(3), executor.Lt(5)), "blkRate"),
			executor.MatchRate(gomock.All(executor.Gt(4), executor.Lt(6)), "txRate"),
			executor.MatchRate(gomock.All(executor.Gt(400), executor.Lt(600)), "gasRate"),
		),
	)

	if err := ext.PreRun(executor.State[txcontext.TxContext]{}, ctx); err != nil {
		t.Fatalf("PreRun failed: %v", err)
	}

	// first processed block
	if err := ext.PostTransaction(executor.State[txcontext.TxContext]{Data: s}, ctx); err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if err := ext.PostTransaction(executor.State[txcontext.TxContext]{Data: s}, ctx); err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if err := ext.PostBlock(executor.State[txcontext.TxContext]{
		Block: 5,
	}, ctx); err != nil {
		t.Fatalf("PostBlock failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// second processed block
	if err := ext.PostTransaction(executor.State[txcontext.TxContext]{Data: s}, ctx); err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if err := ext.PostTransaction(executor.State[txcontext.TxContext]{Data: s}, ctx); err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if err := ext.PostBlock(executor.State[txcontext.TxContext]{
		Block: 6,
	}, ctx); err != nil {
		t.Fatalf("PostBlock failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	if err := ext.PostTransaction(executor.State[txcontext.TxContext]{Data: s}, ctx); err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if err := ext.PostBlock(executor.State[txcontext.TxContext]{
		Block: 8,
	}, ctx); err != nil {
		t.Fatalf("PostBlock failed: %v", err)
	}
}

func TestSubstateProgressTrackerExtension_FirstLoggingIsIgnored(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.First = 4

	ext := makeBlockProgressTracker(cfg, testStateDbInfoFrequency, log)

	ctx := &executor.Context{
		State:           db,
		ExecutionResult: substatecontext.NewReceipt(&substate.Result{GasUsed: 10}),
	}

	s := substatecontext.NewTxContext(&substate.Substate{
		Result: &substate.Result{
			Status: 0,
		},
	})

	if err := ext.PreRun(executor.State[txcontext.TxContext]{
		Block:       4,
		Transaction: 0,
		Data:        s,
	}, ctx); err != nil {
		t.Fatalf("PreRun failed: %v", err)
	}

	if err := ext.PostTransaction(executor.State[txcontext.TxContext]{
		Block:       4,
		Transaction: 0,
		Data:        s,
	}, ctx); err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if err := ext.PostTransaction(executor.State[txcontext.TxContext]{
		Block:       4,
		Transaction: 1,
		Data:        s,
	}, ctx); err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if err := ext.PostBlock(executor.State[txcontext.TxContext]{
		Block:       5,
		Transaction: 0,
		Data:        s,
	}, ctx); err != nil {
		t.Fatalf("PostBlock failed: %v", err)
	}
}

func Test_LoggingFormatMatchesRubyScript(t *testing.T) {
	// NOTE: keep this in sync with the pattern used by scripts/run_throughput_eval.rb
	pattern := `Track: block \d+, memory \d+, disk \d+, interval_blk_rate \d+.\d*, interval_tx_rate \d+.\d*, interval_gas_rate \d+.\d*, overall_blk_rate \d+.\d*, overall_tx_rate \d+.\d*, overall_gas_rate \d+.\d*`
	example := fmt.Sprintf(substateProgressTrackerReportFormat, 1, 2, 3, 4.5, 6.7, 8.9, 0.1, 2.3, 4.5)
	if match, err := regexp.Match(pattern, []byte(example)); !match || err != nil {
		t.Errorf("Logging format '%v' does not match required format '%v'; err %v", example, pattern, err)
	}
}
