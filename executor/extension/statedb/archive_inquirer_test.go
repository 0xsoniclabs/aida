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

package statedb

import (
	"math"
	"math/big"
	"slices"
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
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestArchiveInquirer_makeArchiveInquirer(t *testing.T) {
	t.Run("no duration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		log := logger.NewMockLogger(ctrl)
		cfg := utils.Config{}
		cfg.ChainID = utils.MainnetChainID
		cfg.ArchiveQueryRate = 100
		ext, err := makeArchiveInquirer(&cfg, log, nil)
		assert.NoError(t, err)
		out, ok := ext.(*archiveInquirer)
		assert.True(t, ok)
		assert.Equal(t, defaultTickerDuration, out.tickerDuration)
	})

	t.Run("valid duration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		log := logger.NewMockLogger(ctrl)
		cfg := utils.Config{}
		cfg.ChainID = utils.MainnetChainID
		cfg.ArchiveQueryRate = 100
		duration := 150 * time.Second
		ext, err := makeArchiveInquirer(&cfg, log, &duration)
		assert.NoError(t, err)
		out, ok := ext.(*archiveInquirer)
		assert.True(t, ok)
		assert.Equal(t, duration, out.tickerDuration)
	})

	t.Run("invalid duration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		log := logger.NewMockLogger(ctrl)
		cfg := utils.Config{}
		cfg.ChainID = utils.MainnetChainID
		cfg.ArchiveQueryRate = 100
		duration := -150 * time.Second
		ext, err := makeArchiveInquirer(&cfg, log, &duration)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must greater than 0")
		assert.Nil(t, ext)
	})

}

func TestArchiveInquirer_DisabledIfNoQueryRateIsGiven(t *testing.T) {
	config := utils.Config{}
	ext, err := MakeArchiveInquirer(&config)
	if err != nil {
		t.Fatalf("failed to create inquirer: %v", err)
	}
	if _, ok := ext.(extension.NilExtension[txcontext.TxContext]); !ok {
		t.Errorf("inquirer should not be active by default")
	}
}

func TestArchiveInquirer_ReportsErrorIfNoArchiveIsPresent(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	cfg := utils.Config{}
	cfg.ChainID = utils.MainnetChainID
	cfg.ArchiveQueryRate = 100
	ext, err := makeArchiveInquirer(&cfg, log, nil)
	if err != nil {
		t.Fatalf("failed to create inquirer: %v", err)
	}
	state := executor.State[txcontext.TxContext]{}
	if err := ext.PreRun(state, nil); err == nil {
		t.Errorf("expected an error, got nothing")
	}
	if err := ext.PostRun(state, nil, nil); err != nil {
		t.Errorf("failed to shut down gracefully, got %v", err)
	}
}

func TestArchiveInquirer_CanStartUpAndShutdownGracefully(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	cfg := utils.Config{}
	cfg.ChainID = utils.MainnetChainID
	cfg.ArchiveMode = true
	cfg.ArchiveQueryRate = 100
	ext, err := makeArchiveInquirer(&cfg, log, nil)
	if err != nil {
		t.Fatalf("failed to create inquirer: %v", err)
	}
	state := executor.State[txcontext.TxContext]{}
	context := executor.Context{State: db}

	if err := ext.PreRun(state, &context); err != nil {
		t.Errorf("failed PreRun, got %v", err)
	}
	if err := ext.PostRun(state, nil, nil); err != nil {
		t.Errorf("failed to shut down gracefully, got %v", err)
	}
}

func TestArchiveInquirer_RunsRandomTransactionsInBackground(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	archive := state.NewMockNonCommittableStateDB(ctrl)

	cfg := utils.NewTestConfig(t, utils.TestnetChainID, 0, 0, false, "")
	cfg.ArchiveMode = true
	cfg.ArchiveQueryRate = 100
	cfg.ArchiveMaxQueryAge = 100

	state := executor.State[txcontext.TxContext]{}
	context := executor.Context{State: db}

	substate1 := makeValidSubstate()
	substate2 := makeValidSubstate()

	db.EXPECT().GetArchiveBlockHeight().AnyTimes().Return(uint64(14), false, nil)
	db.EXPECT().GetArchiveState(uint64(12)).MinTimes(1).Return(archive, nil)
	db.EXPECT().GetArchiveState(uint64(14)).MinTimes(1).Return(archive, nil)

	archive.EXPECT().BeginTransaction(gomock.Any()).MinTimes(1)
	archive.EXPECT().SetTxContext(gomock.Any(), gomock.Any()).AnyTimes()
	archive.EXPECT().Snapshot().AnyTimes()
	archive.EXPECT().GetBalance(gomock.Any()).AnyTimes().Return(uint256.NewInt(1000))
	archive.EXPECT().GetNonce(gomock.Any()).AnyTimes().Return(uint64(0))
	archive.EXPECT().SetNonce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return()
	archive.EXPECT().GetCode(gomock.Any()).AnyTimes().Return([]byte{})
	archive.EXPECT().GetCodeHash(gomock.Any()).AnyTimes().Return(common.Hash{})
	archive.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	archive.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	archive.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	archive.EXPECT().SetCode(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	archive.EXPECT().GetRefund().AnyTimes()
	archive.EXPECT().RevertToSnapshot(gomock.Any()).AnyTimes()
	archive.EXPECT().GetLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	archive.EXPECT().EndTransaction().AnyTimes()
	archive.EXPECT().Release().MinTimes(1)
	archive.EXPECT().GetStorageRoot(gomock.Any()).AnyTimes()
	archive.EXPECT().Exist(gomock.Any()).AnyTimes()
	archive.EXPECT().CreateContract(gomock.Any()).AnyTimes()
	archive.EXPECT().Prepare(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	ext, err := makeArchiveInquirer(cfg, log, nil)
	if err != nil {
		t.Fatalf("failed to create inquirer: %v", err)
	}

	if err := ext.PreRun(state, &context); err != nil {
		t.Errorf("failed PreRun, got %v", err)
	}

	// Add two transaction to the pool
	state.Block = 13
	state.Transaction = 0
	state.Data = substate1
	if err := ext.PostTransaction(state, &context); err != nil {
		t.Fatalf("failed to add transaction to pool: %v", err)
	}

	state.Block = 15
	state.Transaction = 0
	state.Data = substate2
	if err := ext.PostTransaction(state, &context); err != nil {
		t.Fatalf("failed to add transaction to pool: %v", err)
	}

	time.Sleep(time.Second)

	if err := ext.PostRun(state, nil, nil); err != nil {
		t.Errorf("failed to shut down gracefully, got %v", err)
	}
}

func makeValidSubstate() txcontext.TxContext {
	// This Substate is a minimal data that can be successfully processed.
	sub := &substate.Substate{
		Env: &substate.Env{
			GasLimit: 100_000_000,
		},
		Message: &substate.Message{
			Gas:       100_000,
			GasPrice:  big.NewInt(0),
			Value:     big.NewInt(0),
			GasFeeCap: big.NewInt(math.MaxInt64),
			GasTipCap: big.NewInt(math.MaxInt64),
		},
		Result: &substate.Result{
			GasUsed: 1,
		},
	}
	return substatecontext.NewTxContext(sub)
}

func TestCircularBuffer_EnforcesSize(t *testing.T) {
	for _, size := range []int{0, 1, 2, 10, 50} {
		buffer := newBuffer[int](size)
		for i := 0; i < 100; i++ {
			want := i
			if i > size {
				want = size
			}
			if got := buffer.Size(); want != got {
				t.Errorf("expected size, wanted %d, got %d", want, got)
			}
			buffer.Add(0)
		}
	}
}

func TestCircularBuffer_GetReturnsValueAtPosition(t *testing.T) {
	buffer := newBuffer[int](3)
	buffer.Add(1)
	buffer.Add(2)
	buffer.Add(3)
	for i := 0; i < buffer.Size(); i++ {
		if want, got := i+1, buffer.Get(i); want != got {
			t.Errorf("unexpected element at position %d: want %d, got %d", i, want, got)
		}
	}
}

func TestCircularBuffer_CyclesThroughContent(t *testing.T) {
	buffer := newBuffer[int](3)
	if want, got := []int{}, buffer.data; !slices.Equal(want, got) {
		t.Errorf("unexpected content, wanted %v, got %v", want, got)
	}

	buffer.Add(1)
	if want, got := []int{1}, buffer.data; !slices.Equal(want, got) {
		t.Errorf("unexpected content, wanted %v, got %v", want, got)
	}
	buffer.Add(2)
	if want, got := []int{1, 2}, buffer.data; !slices.Equal(want, got) {
		t.Errorf("unexpected content, wanted %v, got %v", want, got)
	}
	buffer.Add(3)
	if want, got := []int{1, 2, 3}, buffer.data; !slices.Equal(want, got) {
		t.Errorf("unexpected content, wanted %v, got %v", want, got)
	}
	buffer.Add(4)
	if want, got := []int{4, 2, 3}, buffer.data; !slices.Equal(want, got) {
		t.Errorf("unexpected content, wanted %v, got %v", want, got)
	}
	buffer.Add(5)
	if want, got := []int{4, 5, 3}, buffer.data; !slices.Equal(want, got) {
		t.Errorf("unexpected content, wanted %v, got %v", want, got)
	}
}

func TestThrottler_ProducesEventsInExpectedRate(t *testing.T) {
	const testPeriod = 500 * time.Millisecond
	for _, rate := range []int{5, 10, 100, 1000} {
		throttler := *newThrottler(rate)

		count := 0
		start := time.Now()
		for time.Since(start) < testPeriod {
			if throttler.shouldRunNow() {
				count++
			}
		}

		expected := float64(rate) * float64(testPeriod) / float64(time.Second)
		diff := float64(count) - expected
		if diff > 5 || diff < -5 {
			t.Errorf("failed to reproduce rate %d, did %d events in %v", rate, count, testPeriod)
		}
	}
}

func TestArchiveInquirer_RunProgressReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLog := logger.NewMockLogger(ctrl)

	duration := 1 * time.Second
	inquirer := &archiveInquirer{
		log:            mockLog,
		finished:       utils.MakeEvent(),
		tickerDuration: duration,
	}

	initialTxCount := uint64(20)
	initialGasCount := uint64(300_000_000) // 300 M Gas
	initialDurationMs := uint64(1000)      // 1000 ms (1 second)

	inquirer.transactionCounter.Store(initialTxCount)
	inquirer.gasCounter.Store(initialGasCount)
	inquirer.totalQueryTimeMilliseconds.Store(initialDurationMs)

	inquirer.done.Add(1) // For the runProgressReport goroutine

	formatString := "Archive throughput: t=%ds, %.2f Tx/s, %.2f MGas/s, average duration %.2f ms"

	// Expect at least one call to Infof.
	// We capture the arguments to verify them after the goroutine finishes.
	mockLog.EXPECT().Infof(formatString, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(format string, args ...interface{}) {
			if len(args) == 4 {
				var ok bool
				_, ok = args[0].(int)
				if !ok {
					t.Logf("Failed to cast loggedTotalTime: %v", args[0])
				}
				_, ok = args[1].(float64)
				if !ok {
					t.Logf("Failed to cast loggedTPS: %v", args[1])
				}
				_, ok = args[2].(float64)
				if !ok {
					t.Logf("Failed to cast loggedMGPS: %v", args[2])
				}
				_, ok = args[3].(float64)
				if !ok {
					t.Logf("Failed to cast loggedAvgDuration: %v", args[3])
				}
			} else {
				t.Logf("Infof called with unexpected number of arguments: %d", len(args))
			}
		}).MinTimes(1)

	go inquirer.runProgressReport()

	time.Sleep(duration)

	inquirer.finished.Signal()
	inquirer.done.Wait() // Wait for runProgressReport to complete
}
