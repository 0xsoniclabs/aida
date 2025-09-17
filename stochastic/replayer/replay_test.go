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

package replayer

import (
	"math/rand"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	logmock "github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	recArgs "github.com/0xsoniclabs/aida/stochastic/recorder/arguments"
	repArgs "github.com/0xsoniclabs/aida/stochastic/replayer/arguments"
	"github.com/0xsoniclabs/aida/stochastic/statistics/markov"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

// stubSnapshots implements arguments.SnapshotSet for testing
type stubSnapshots struct{ ret int }

func (s *stubSnapshots) SampleSnapshot(n int) int { return s.ret }

func TestReplay_ExecuteRevertSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := state.NewMockStateDB(ctrl)
	contracts := repArgs.NewMockSet(ctrl)
	keys := repArgs.NewMockSet(ctrl)
	values := repArgs.NewMockSet(ctrl)
	snapshots := &stubSnapshots{ret: 1}

	// Expect a single RevertToSnapshot call
	db.EXPECT().RevertToSnapshot(gomock.Any()).Times(1)

	rg := rand.New(rand.NewSource(999))
	log := logger.NewLogger("INFO", "test")
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, log)
	ss.activeSnapshots = []int{1, 2, 3, 4, 5}
	snapshotSize := len(ss.activeSnapshots)

	if err := ss.execute(operations.RevertToSnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	assert.GreaterOrEqual(t, len(ss.activeSnapshots), 1)         // must have at least one snapshot
	assert.LessOrEqual(t, len(ss.activeSnapshots), snapshotSize) // must not have more than initial snapshots
}

// TestExecute_AllOps covers all operation branches in execute.
func TestExecute_AllOps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := state.NewMockStateDB(ctrl)
	contracts := repArgs.NewMockSet(ctrl)
	keys := repArgs.NewMockSet(ctrl)
	values := repArgs.NewMockSet(ctrl)
	snapshots := &stubSnapshots{ret: 1}

	// Argument sets always return index 1 and no error
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil).AnyTimes()
	keys.EXPECT().Choose(gomock.Any()).Return(int64(1), nil).AnyTimes()
	values.EXPECT().Choose(gomock.Any()).Return(int64(1), nil).AnyTimes()
	contracts.EXPECT().Remove(gomock.Any()).Return(nil).AnyTimes()

	// DB expectations (AnyTimes to avoid fragile counts)
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().CreateContract(gomock.Any()).AnyTimes()
	db.EXPECT().Empty(gomock.Any()).Return(false).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().Exist(gomock.Any()).Return(false).AnyTimes()
	db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(10)).AnyTimes()
	db.EXPECT().GetCode(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().GetCodeHash(gomock.Any()).Return(common.Hash{}).AnyTimes()
	db.EXPECT().GetCodeSize(gomock.Any()).Return(0).AnyTimes()
	db.EXPECT().GetCommittedState(gomock.Any(), gomock.Any()).Return(common.Hash{}).AnyTimes()
	db.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	db.EXPECT().GetState(gomock.Any(), gomock.Any()).Return(common.Hash{}).AnyTimes()
	db.EXPECT().GetStorageRoot(gomock.Any()).Return(common.Hash{}).AnyTimes()
	db.EXPECT().GetTransientState(gomock.Any(), gomock.Any()).Return(common.Hash{}).AnyTimes()
	db.EXPECT().HasSelfDestructed(gomock.Any()).Return(false).AnyTimes()
	db.EXPECT().RevertToSnapshot(gomock.Any()).AnyTimes()
	db.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().SelfDestruct6780(gomock.Any()).Return(*uint256.NewInt(0), false).AnyTimes()
	db.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().SetNonce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	db.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Hash{}).AnyTimes()
	db.EXPECT().SetTransientState(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	db.EXPECT().Snapshot().Return(123).AnyTimes()
	db.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()

	rg := rand.New(rand.NewSource(123))
	log := logger.NewLogger("INFO", "test")
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, log)
	ss.enableDebug() // cover debug branches

	// Ensure we have snapshots for RevertToSnapshot path
	ss.activeSnapshots = []int{1, 2, 3, 4, 5}

	// Build some hashes for two/three-arg ops
	addrCl := stochastic.RandArgID
	keyCl := stochastic.RandArgID
	valCl := stochastic.RandArgID

	// One-arg ops
	_ = ss.execute(operations.AddBalanceID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.CreateAccountID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.CreateContractID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.EmptyID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.ExistID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.GetBalanceID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.GetCodeHashID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.GetCodeID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.GetCodeSizeID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.GetNonceID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.GetStorageRootID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.HasSelfDestructedID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.SetCodeID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.SetNonceID, addrCl, stochastic.NoArgID, stochastic.NoArgID)

	// Two-arg ops
	_ = ss.execute(operations.GetCommittedStateID, addrCl, keyCl, stochastic.NoArgID)
	_ = ss.execute(operations.GetStateID, addrCl, keyCl, stochastic.NoArgID)
	_ = ss.execute(operations.GetTransientStateID, addrCl, keyCl, stochastic.NoArgID)

	// Three-arg ops
	_ = ss.execute(operations.SetStateID, addrCl, keyCl, valCl)
	_ = ss.execute(operations.SetTransientStateID, addrCl, keyCl, valCl)

	// Zero-arg ops
	assert.NoError(t, ss.execute(operations.BeginSyncPeriodID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))
	assert.NoError(t, ss.execute(operations.BeginBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))
	assert.NoError(t, ss.execute(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))
	assert.NoError(t, ss.execute(operations.EndTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))
	assert.NoError(t, ss.execute(operations.EndSyncPeriodID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))

	// Self-destruct path and removal during EndBlock
	_ = ss.execute(operations.SelfDestructID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	_ = ss.execute(operations.SelfDestruct6780ID, addrCl, stochastic.NoArgID, stochastic.NoArgID)
	assert.NoError(t, ss.execute(operations.EndBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))

	// Snapshot and revert
	assert.NoError(t, ss.execute(operations.SnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))
	ss.activeSnapshots = []int{11, 22, 33}
	assert.NoError(t, ss.execute(operations.RevertToSnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID))

	// SubBalance (with non-zero balance branch)
	assert.NoError(t, ss.execute(operations.SubBalanceID, addrCl, stochastic.NoArgID, stochastic.NoArgID))
}

// TestExecute_InvalidOperation ensures an invalid op ID errors out.
func TestExecute_InvalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	ss := newReplayContext(rand.New(rand.NewSource(1)), db, nil, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.execute(99999, 0, 0, 0); err == nil {
		t.Fatalf("expected error for invalid op")
	}
}

// TestPrime_Success validates priming happy path and call sequence.
func TestPrime_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	contracts := repArgs.NewMockSet(ctrl)

	// 3 contracts => 4 accounts including zero
	contracts.EXPECT().Size().Return(int64(3)).AnyTimes()

	db.EXPECT().BeginSyncPeriod(uint64(0))
	db.EXPECT().BeginBlock(uint64(0)).Return(nil)
	db.EXPECT().BeginTransaction(uint32(0)).Return(nil)
	db.EXPECT().CreateAccount(gomock.Any()).Times(4)
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).Times(4)
	db.EXPECT().EndTransaction().Return(nil)
	db.EXPECT().EndBlock().Return(nil)
	db.EXPECT().EndSyncPeriod()

	ss := newReplayContext(rand.New(rand.NewSource(7)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.prime(); err != nil {
		t.Fatalf("unexpected error priming: %v", err)
	}
}

func TestPrime_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	contracts := repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Size().Return(int64(1)).AnyTimes()

	t.Run("begin block fails", func(t *testing.T) {
		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(uint64(0))
		db.EXPECT().BeginBlock(uint64(0)).Return(assert.AnError)
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
		if err := ss.prime(); err == nil {
			t.Fatalf("expected error from BeginBlock")
		}
	})

	t.Run("begin tx fails", func(t *testing.T) {
		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(uint64(0))
		db.EXPECT().BeginBlock(uint64(0)).Return(nil)
		db.EXPECT().BeginTransaction(uint32(0)).Return(assert.AnError)
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
		if err := ss.prime(); err == nil {
			t.Fatalf("expected error from BeginTransaction")
		}
	})

	t.Run("end tx fails", func(t *testing.T) {
		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(uint64(0))
		db.EXPECT().BeginBlock(uint64(0)).Return(nil)
		db.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
		db.EXPECT().EndTransaction().Return(assert.AnError)
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
		if err := ss.prime(); err == nil {
			t.Fatalf("expected error from EndTransaction")
		}
	})

	t.Run("end block fails", func(t *testing.T) {
		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(uint64(0))
		db.EXPECT().BeginBlock(uint64(0)).Return(nil)
		db.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
		db.EXPECT().EndTransaction().Return(nil)
		db.EXPECT().EndBlock().Return(assert.AnError)
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
		if err := ss.prime(); err == nil {
			t.Fatalf("expected error from EndBlock")
		}
	})
}

// TestGetStochasticMatrix_Success ensures MC and initial state are built.
func TestGetStochasticMatrix_Success(t *testing.T) {
	// Labels include BeginSyncPeriod to find initial state
	labels := []string{
		operations.OpMnemo(operations.BeginSyncPeriodID),
		operations.OpMnemo(operations.BeginBlockID),
	}
	A := [][]float64{
		{0, 1},
		{1, 0},
	}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.2
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.8 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{
		Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}},
		Queuing:  recArgs.QueueStatsJSON{Distribution: qpdf},
	}
	e := &recorder.StateJSON{
		Operations:       labels,
		StochasticMatrix: A,
		Contracts:        cls,
		Keys:             cls,
		Values:           cls,
		SnapshotECDF:     [][2]float64{{0, 0}, {1, 1}},
	}
	mc, stateIdx, err := getStochasticMatrix(e)
	if err != nil || mc == nil || stateIdx != 0 {
		t.Fatalf("unexpected MC creation failure: %v", err)
	}
}

// TestRunStochasticReplay_Success runs a tiny simulation and validates no error.
func TestRunStochasticReplay_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := state.NewMockStateDB(ctrl)

	// Prime path expectations
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().Error().Return(nil).AnyTimes()
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	db.EXPECT().Error().Return(nil).AnyTimes()

	// Minimal eCDFs and distributions
	labels := []string{
		operations.OpMnemo(operations.BeginSyncPeriodID),
		operations.OpMnemo(operations.BeginBlockID),
		operations.OpMnemo(operations.EndBlockID),
	}
	A := [][]float64{
		{0, 1, 0}, // BS -> BB
		{0, 0, 1}, // BB -> EB
		{1, 0, 0}, // EB -> BS
	}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.3
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.7 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{
		Operations:       labels,
		StochasticMatrix: A,
		Contracts:        cls,
		Keys:             cls,
		Values:           cls,
		SnapshotECDF:     [][2]float64{{0, 0}, {1, 1}},
	}

	cfg := &utils.Config{BalanceRange: 100, NonceRange: 100, RandomSeed: 1, Debug: true, DebugFrom: 1}
	log := logger.NewLogger("INFO", "test")
	if err := RunStochasticReplay(db, e, 2, cfg, log); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunStochasticReplay_ErrorBreaks exercises error handling and stop condition.
func TestRunStochasticReplay_ErrorBreaks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	// cause a single error to be observed by the loop and break immediately
	db.EXPECT().Error().Return(assert.AnError).Times(1)

	labels := []string{
		operations.OpMnemo(operations.BeginSyncPeriodID),
		operations.OpMnemo(operations.BeginBlockID),
		operations.OpMnemo(operations.EndBlockID),
	}
	A := [][]float64{
		{0, 1, 0},
		{0, 0, 1},
		{1, 0, 0},
	}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.25
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.75 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{
		Operations:       labels,
		StochasticMatrix: A,
		Contracts:        cls,
		Keys:             cls,
		Values:           cls,
		SnapshotECDF:     [][2]float64{{0, 0}, {1, 1}},
	}
	cfg := &utils.Config{BalanceRange: 10, NonceRange: 10, RandomSeed: 2, ContinueOnFailure: false}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err == nil {
		t.Fatalf("expected error due to db.Error()")
	}
}

// TestRunStochasticReplay_ErrorContinue ensures ContinueOnFailure lets the loop proceed and finish.
func TestRunStochasticReplay_ErrorContinue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().Error().Return(assert.AnError)
	db.EXPECT().Error().Return(nil).AnyTimes()

	labels := []string{
		operations.OpMnemo(operations.BeginSyncPeriodID),
		operations.OpMnemo(operations.EndBlockID),
	}
	A := [][]float64{{0, 1}, {1, 0}}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.5
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.5 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{BalanceRange: 10, NonceRange: 10, RandomSeed: 2, ContinueOnFailure: true}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err == nil {
		t.Fatalf("expected aggregated error even when continuing on failure")
	}
}

// TestEnableDebug covers the trivial enableDebug method.
func TestEnableDebug(t *testing.T) {
	ss := newReplayContext(rand.New(rand.NewSource(1)), nil, nil, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if ss.traceDebug {
		t.Fatalf("debug should be disabled by default")
	}
	ss.enableDebug()
	if !ss.traceDebug {
		t.Fatalf("debug should be enabled after enableDebug()")
	}
}

// TestGetStochasticMatrix_BadShape ensures error from MC creation bubbles up.
func TestGetStochasticMatrix_BadShape(t *testing.T) {
	e := &recorder.StateJSON{
		Operations:       []string{"BB", "EB"},
		StochasticMatrix: [][]float64{{1.0, 0.0}}, // bad: rows != len(labels)
		Contracts:        recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: make([]float64, stochastic.QueueLen)}},
		Keys:             recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: make([]float64, stochastic.QueueLen)}},
		Values:           recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: make([]float64, stochastic.QueueLen)}},
		SnapshotECDF:     [][2]float64{{0, 0}, {1, 1}},
	}
	if _, _, err := getStochasticMatrix(e); err == nil {
		t.Fatalf("expected error due to bad matrix shape")
	}
}

// TestRunStochasticReplay_InvalidInitialState exercises error when initial state label is invalid.
func TestRunStochasticReplay_InvalidInitialState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	// allow priming
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().Error().Return(nil).AnyTimes()

	cfg := &utils.Config{BalanceRange: 10, NonceRange: 10, RandomSeed: 3}

	// MC with only an unknown label; BeginSyncPeriod not present -> initial state index -1
	labels := []string{"XX"}
	A := [][]float64{{1.0}}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.5
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.5 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err == nil {
		t.Fatalf("expected error due to invalid initial state")
	}
}

// TestRunStochasticReplay_CannotDecodeOpcode returns error when encountering an unknown label.
func TestRunStochasticReplay_CannotDecodeOpcode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	// priming expectations
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().Error().Return(nil).AnyTimes()

	// labels: valid initial BS, then unknown "ZZ"
	labels := []string{operations.OpMnemo(operations.BeginSyncPeriodID), "ZZ"}
	A := [][]float64{{0, 1}, {1, 1}} // second row not used (we error before)
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.4
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.6 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{RandomSeed: 1, BalanceRange: 10, NonceRange: 10}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err == nil {
		t.Fatalf("expected decode opcode error during run")
	}
}

// TestRunStochasticReplay_DecodeOpcodeError triggers a decode failure mid-run.
// (intentionally no test for mid-run decode error to avoid fragile Error() expectations)

// TestExecute_ChooseErrors covers error paths when argument selection fails.
func TestExecute_ChooseErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	contracts := repArgs.NewMockSet(ctrl)
	keys := repArgs.NewMockSet(ctrl)
	values := repArgs.NewMockSet(ctrl)
	// cause failures on selection
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(0), assert.AnError)
	rg := rand.New(rand.NewSource(9))
	ss := newReplayContext(rg, db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.execute(operations.GetBalanceID, stochastic.RandArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected error for contract Choose failure")
	}

	// key choose failure
	contracts = repArgs.NewMockSet(ctrl)
	keys = repArgs.NewMockSet(ctrl)
	values = repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	keys.EXPECT().Choose(gomock.Any()).Return(int64(0), assert.AnError)
	ss = newReplayContext(rg, db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.execute(operations.GetStateID, stochastic.RandArgID, stochastic.RandArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected error for key Choose failure")
	}

	// value choose failure
	contracts = repArgs.NewMockSet(ctrl)
	keys = repArgs.NewMockSet(ctrl)
	values = repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	keys.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	values.EXPECT().Choose(gomock.Any()).Return(int64(0), assert.AnError)
	ss = newReplayContext(rg, db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.execute(operations.SetStateID, stochastic.RandArgID, stochastic.RandArgID, stochastic.RandArgID); err == nil {
		t.Fatalf("expected error for value Choose failure")
	}
}

// TestExecute_RevertToSnapshot_Clamp branches for snapshot index clamping.
func TestExecute_RevertToSnapshot_Clamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	// Expect RevertToSnapshot to be called for both cases
	db.EXPECT().RevertToSnapshot(gomock.Any()).Times(2)
	rg := rand.New(rand.NewSource(1))
	// Stub that returns too large value (forces snapshotIdx<0)
	ss := newReplayContext(rg, db, nil, nil, nil, &stubSnapshots{ret: 100}, logger.NewLogger("INFO", "test"))
	ss.activeSnapshots = []int{10, 20, 30}
	_ = ss.execute(operations.RevertToSnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	// Stub that returns negative value (forces snapshotIdx>=snapshotNum)
	ss.snapshots = &stubSnapshots{ret: -100}
	ss.activeSnapshots = []int{11, 22, 33}
	_ = ss.execute(operations.RevertToSnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
}

// TestExecute_EndBlock_RemoveError covers error path when removing destroyed accounts fails.
func TestExecute_EndBlock_RemoveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	contracts := repArgs.NewMockSet(ctrl)
	// Self-destructed will contain 1, expect Remove to return error
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	contracts.EXPECT().Remove(int64(1)).Return(assert.AnError)
	db.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(0))
	db.EXPECT().EndBlock().Return(nil)
	rg := rand.New(rand.NewSource(2))
	ss := newReplayContext(rg, db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	// mark self-destructed, then attempt EndBlock
	_ = ss.execute(operations.SelfDestructID, stochastic.RandArgID, stochastic.NoArgID, stochastic.NoArgID)
	if err := ss.execute(operations.EndBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected error from contracts.Remove during EndBlock")
	}
}

// TestExecute_ConvertIndexErrors covers failures converting indices to addresses/hashes.
func TestExecute_ConvertIndexErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Address conversion error
	contracts := repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(-1), nil)
	ss := newReplayContext(rand.New(rand.NewSource(3)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.execute(operations.GetBalanceID, stochastic.RandArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected address conversion error")
	}

	// Key conversion error
	contracts = repArgs.NewMockSet(ctrl)
	keys := repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	keys.EXPECT().Choose(gomock.Any()).Return(int64(-1), nil)
	ss = newReplayContext(rand.New(rand.NewSource(4)), db, contracts, keys, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.execute(operations.GetStateID, stochastic.RandArgID, stochastic.RandArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected key conversion error")
	}

	// Value conversion error
	contracts = repArgs.NewMockSet(ctrl)
	keys = repArgs.NewMockSet(ctrl)
	values := repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	keys.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	values.EXPECT().Choose(gomock.Any()).Return(int64(-1), nil)
	ss = newReplayContext(rand.New(rand.NewSource(5)), db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	if err := ss.execute(operations.SetStateID, stochastic.RandArgID, stochastic.RandArgID, stochastic.RandArgID); err == nil {
		t.Fatalf("expected value conversion error")
	}
}

// TestExecute_FatalBranches covers logger.Fatal branches without exiting the process using a mock logger.
func TestExecute_FatalBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	lg := logmock.NewMockLogger(ctrl)
	// BeginBlock error path
	db.EXPECT().BeginBlock(gomock.Any()).Return(assert.AnError)
	lg.EXPECT().Fatal(gomock.Any()).Times(1)
	ss := newReplayContext(rand.New(rand.NewSource(1)), db, nil, nil, nil, &stubSnapshots{ret: 0}, lg)
	_ = ss.execute(operations.BeginBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)

	// BeginTransaction error path
	db = state.NewMockStateDB(ctrl)
	lg = logmock.NewMockLogger(ctrl)
	db.EXPECT().BeginTransaction(gomock.Any()).Return(assert.AnError)
	lg.EXPECT().Fatal(gomock.Any()).Times(1)
	ss = newReplayContext(rand.New(rand.NewSource(2)), db, nil, nil, nil, &stubSnapshots{ret: 0}, lg)
	_ = ss.execute(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)

	// EndTransaction error path
	db = state.NewMockStateDB(ctrl)
	lg = logmock.NewMockLogger(ctrl)
	db.EXPECT().EndTransaction().Return(assert.AnError)
	lg.EXPECT().Fatal(gomock.Any()).Times(1)
	ss = newReplayContext(rand.New(rand.NewSource(3)), db, nil, nil, nil, &stubSnapshots{ret: 0}, lg)
	_ = ss.execute(operations.EndTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)

	// EndBlock error path
	db = state.NewMockStateDB(ctrl)
	lg = logmock.NewMockLogger(ctrl)
	db.EXPECT().EndBlock().Return(assert.AnError)
	lg.EXPECT().Fatal(gomock.Any()).Times(1)
	ss = newReplayContext(rand.New(rand.NewSource(4)), db, nil, nil, nil, &stubSnapshots{ret: 0}, lg)
	_ = ss.execute(operations.EndBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
}

// TestPopulateReplayContext_PrimeError ensures errors from prime() propagate.
func TestPopulateReplayContext_PrimeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	rg := rand.New(rand.NewSource(7))
	lg := logger.NewLogger("INFO", "test")
	cfg := &utils.Config{BalanceRange: 10, NonceRange: 10}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.2
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.8 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: []string{operations.OpMnemo(operations.BeginSyncPeriodID)}, StochasticMatrix: [][]float64{{1.0}}, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}

	// Cause prime to fail at BeginBlock
	db.EXPECT().BeginSyncPeriod(uint64(0))
	db.EXPECT().BeginBlock(uint64(0)).Return(assert.AnError)
	if _, err := populateReplayContext(cfg, e, db, rg, lg); err == nil {
		t.Fatalf("expected prime error to propagate from populateReplayContext")
	}
}

// TestRunStochasticReplay_ProgressLogBranch forces progress log condition.
func TestRunStochasticReplay_ProgressLogBranch(t *testing.T) {
	// override interval to 0 so condition sec-lastSec >= 0 holds
	old := progressLogIntervalSec
	progressLogIntervalSec = 0
	defer func() { progressLogIntervalSec = old }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	// priming expectations
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().Error().Return(nil).AnyTimes()

	labels := []string{
		operations.OpMnemo(operations.BeginSyncPeriodID),
		operations.OpMnemo(operations.EndBlockID),
	}
	A := [][]float64{{0, 1}, {1, 0}}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.3
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.7 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{RandomSeed: 1, BalanceRange: 10, NonceRange: 10}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunStochasticReplay_LabelError forces label retrieval error via seam.
func TestRunStochasticReplay_LabelError(t *testing.T) {
	// Override mcLabel to error once, then restore
	old := mcLabel
	mcLabel = func(_ *markov.Chain, _ int) (string, error) { return "", assert.AnError }
	defer func() { mcLabel = old }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()

	// Valid MC
	labels := []string{operations.OpMnemo(operations.BeginSyncPeriodID)}
	A := [][]float64{{1.0}}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.5
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.5 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{RandomSeed: 1, BalanceRange: 10, NonceRange: 10}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err == nil {
		t.Fatalf("expected label retrieval error")
	}
}

// TestRunStochasticReplay_SampleError forces sample error via seam.
func TestRunStochasticReplay_SampleError(t *testing.T) {
	old := mcSample
	mcSample = func(_ *markov.Chain, _ int, _ float64) (int, error) { return 0, assert.AnError }
	defer func() { mcSample = old }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().Error().Return(nil).AnyTimes()

	labels := []string{operations.OpMnemo(operations.BeginSyncPeriodID), operations.OpMnemo(operations.EndBlockID)}
	A := [][]float64{{0, 1}, {1, 0}}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.5
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.5 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{RandomSeed: 1, BalanceRange: 10, NonceRange: 10}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err == nil {
		t.Fatalf("expected sample error")
	}
}

// TestRunStochasticReplay_EnableDebugInLoop covers the in-loop debug enabling branch.
func TestRunStochasticReplay_EnableDebugInLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	db.EXPECT().GetShadowDB().Return(nil).AnyTimes()
	db.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil).AnyTimes()
	db.EXPECT().EndBlock().Return(nil).AnyTimes()
	db.EXPECT().EndSyncPeriod().AnyTimes()
	db.EXPECT().Error().Return(nil).AnyTimes()

	labels := []string{
		operations.OpMnemo(operations.BeginSyncPeriodID),
		operations.OpMnemo(operations.EndBlockID),
	}
	A := [][]float64{{0, 1}, {1, 0}}
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.5
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.5 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StateJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{RandomSeed: 1, BalanceRange: 10, NonceRange: 10, Debug: true, DebugFrom: 2}
	if err := RunStochasticReplay(db, e, 2, cfg, logger.NewLogger("INFO", "test")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestExecute_SetCodeRandomReadError forces randReadBytes error and expects Fatalf.
func TestExecute_SetCodeRandomReadError(t *testing.T) {
	old := randReadBytes
	randReadBytes = func(_ *rand.Rand, _ []byte) (int, error) { return 0, assert.AnError }
	defer func() { randReadBytes = old }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	lg := logmock.NewMockLogger(ctrl)
	// contracts choose to supply addr
	contracts := repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	lg.EXPECT().Fatalf(gomock.Any(), gomock.Any()).Times(1)
	ss := newReplayContext(rand.New(rand.NewSource(10)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, lg)
	// Expect no SetCode call since we error before it
	_ = ss.execute(operations.SetCodeID, stochastic.RandArgID, stochastic.NoArgID, stochastic.NoArgID)
}

// TestPopulateReplayContext_Errors covers randomizer ctor error branches.
func TestPopulateReplayContext_Errors(t *testing.T) {
	rg := rand.New(rand.NewSource(123))
	log := logger.NewLogger("INFO", "test")
	db := state.NewMockStateDB(gomock.NewController(t))
	cfg := &utils.Config{}

	badDist := make([]float64, stochastic.QueueLen-1) // invalid length causes error
	goodDist := make([]float64, stochastic.QueueLen)
	goodDist[0] = 0.1
	for i := 1; i < len(goodDist); i++ {
		goodDist[i] = 0.9 / float64(stochastic.QueueLen-1)
	}

	ecdf := [][2]float64{{0, 0}, {1, 1}}
	badE := func() *recorder.StateJSON {
		return &recorder.StateJSON{
			Contracts:    recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: badDist}},
			Keys:         recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Values:       recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			SnapshotECDF: ecdf,
		}
	}
	if _, err := populateReplayContext(cfg, badE(), db, rg, log); err == nil {
		t.Fatalf("expected error for bad contracts distribution")
	}

	badE2 := func() *recorder.StateJSON {
		return &recorder.StateJSON{
			Contracts:    recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Keys:         recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: badDist}},
			Values:       recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			SnapshotECDF: ecdf,
		}
	}
	if _, err := populateReplayContext(cfg, badE2(), db, rg, log); err == nil {
		t.Fatalf("expected error for bad keys distribution")
	}

	badE3 := func() *recorder.StateJSON {
		return &recorder.StateJSON{
			Contracts:    recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Keys:         recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Values:       recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: badDist}},
			SnapshotECDF: ecdf,
		}
	}
	if _, err := populateReplayContext(cfg, badE3(), db, rg, log); err == nil {
		t.Fatalf("expected error for bad values distribution")
	}
}

// TestExecute_DebugEncodeOpcodeError triggers encode error in debug printing.
func TestExecute_DebugEncodeOpcodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	ss := newReplayContext(rand.New(rand.NewSource(1)), db, nil, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"))
	ss.enableDebug()
	// Use a two-arg op with no args to make EncodeOpcode fail without triggering Choose
	if err := ss.execute(operations.GetStateID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected encode opcode error in debug path")
	}
}

//
////bfs: revisit this test case after changing the json format
////func TestReplay_RunStochasticReplay(t *testing.T) {
////	ctrl := gomock.NewController(t)
////	defer ctrl.Finish()
////	tmpDir := t.TempDir()
////	cfg := &utils.Config{
////		ContractNumber:    1000,
////		KeysNumber:        1000,
////		ValuesNumber:      1000,
////		SnapshotDepth:     100,
////		BlockLength:       3,
////		SyncPeriodLength:  10,
////		TransactionLength: 2,
////		BalanceRange:      100,
////		NonceRange:        100,
////		Debug:             true,
////		ShadowImpl:        "geth",
////		DbTmp:             tmpDir,
////		DbImpl:            "carmen",
////		DbVariant:         "go-file",
////	}
////	db := state.NewMockStateDB(ctrl)
////	log := logger.NewLogger("INFO", "test")
////	e, err := recorder.ReadSimulation("data/test_replay.json")
////	if err != nil {
////		t.Fatalf("Failed to read simulation: %v", err)
////	}
////	counter := 0
////	db.EXPECT().CreateAccount(gomock.Any()).Times(1001)
////	db.EXPECT().GetShadowDB().Return(db)
////	db.EXPECT().BeginSyncPeriod(gomock.Any()).Return().Times(2)
////	db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
////	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
////	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(1)).Times(1001)
////	db.EXPECT().EndTransaction().Return(nil)
////	db.EXPECT().EndBlock().Return(nil).Times(2)
////	db.EXPECT().EndSyncPeriod().Return()
////	db.EXPECT().Error().Return(nil)
////	err = RunStochasticReplay(db, e, 0, cfg, log)
////	fmt.Printf("Counter: %d\n", counter)
////	assert.NoError(t, err)
////}
//
//func TestStochasticState_execute(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	// enumerate whole operation space with arguments
//	// and check encoding/decoding whether it is symmetric.
//	var codes = map[int]int{}
//	for op := 0; op < operations.NumOps; op++ {
//		for addr := 0; addr < classifier.NumArgKinds; addr++ {
//			for key := 0; key < classifier.NumArgKinds; key++ {
//				for value := 0; value < classifier.NumArgKinds; value++ {
//					// check legality of argument/op combination
//					if (operations.OpNumArgs[op] == 0 && addr == classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
//						(operations.OpNumArgs[op] == 1 && addr != classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
//						(operations.OpNumArgs[op] == 2 && addr != classifier.NoArgID && key != classifier.NoArgID && value == classifier.NoArgID) ||
//						(operations.OpNumArgs[op] == 3 && addr != classifier.NoArgID && key != classifier.NoArgID && value != classifier.NoArgID) {
//
//						// encode to an argument-encoded operation
//						var err error
//						codes[op], err = operations.EncodeArgOp(op, addr, key, value)
//						if err != nil {
//							t.Fatalf("Encoding failed for %v", codes[op])
//						}
//					}
//				}
//			}
//		}
//	}
//
//	contracts := generator.NewMockArgumentSet(ctrl)
//	keys := generator.NewMockArgumentSet(ctrl)
//	values := generator.NewMockArgumentSet(ctrl)
//	snapshots := generator.NewMockSnapshotSet(ctrl)
//	db := state.NewMockStateDB(ctrl)
//	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0))
//	db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
//	db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
//	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
//	db.EXPECT().CreateAccount(gomock.Any()).Return()
//	db.EXPECT().CreateContract(gomock.Any()).Return()
//	db.EXPECT().Empty(gomock.Any()).Return(false)
//	db.EXPECT().EndBlock().Return(nil)
//	db.EXPECT().EndSyncPeriod().Return()
//	db.EXPECT().EndTransaction().Return(nil)
//	db.EXPECT().Exist(gomock.Any()).Return(false)
//	db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(0))
//	db.EXPECT().GetCode(gomock.Any()).Return(nil)
//	db.EXPECT().GetCodeHash(gomock.Any()).Return(common.Hash{})
//	db.EXPECT().GetCodeSize(gomock.Any()).Return(0)
//	db.EXPECT().GetCommittedState(gomock.Any(), gomock.Any()).Return(common.Hash{})
//	db.EXPECT().GetNonce(gomock.Any()).Return(uint64(0))
//	db.EXPECT().GetState(gomock.Any(), gomock.Any()).Return(common.Hash{})
//	db.EXPECT().GetStorageRoot(gomock.Any()).Return(common.Hash{})
//	db.EXPECT().GetTransientState(gomock.Any(), gomock.Any()).Return(common.Hash{})
//	db.EXPECT().HasSelfDestructed(gomock.Any()).Return(false)
//	db.EXPECT().RevertToSnapshot(gomock.Any()).Return()
//	db.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(0))
//	db.EXPECT().SelfDestruct6780(gomock.Any()).Return(*uint256.NewInt(0), false)
//	db.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return(nil)
//	db.EXPECT().SetNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return()
//	db.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Hash{})
//	db.EXPECT().SetTransientState(gomock.Any(), gomock.Any(), gomock.Any()).Return()
//	db.EXPECT().Snapshot().Return(0)
//	db.EXPECT().GetShadowDB().Return(db)
//	db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(10))
//	db.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0))
//
//	rg := rand.New(rand.NewSource(999))
//	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
//	ss.enableDebug()
//	for i := 0; i < len(codes); i++ {
//		ss.activeSnapshots = []int{1, 2, 3, 4, 5}
//		op, addr, key, value, err := operations.DecodeArgOp(codes[i])
//		if err != nil {
//			t.Fatalf("Decoding failed for %v", codes[i])
//		}
//		ss.execute(op, addr, key, value)
//	}
//}
//
//func TestStochasticState_prime(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	t.Run("success", func(t *testing.T) {
//		rg := rand.New(rand.NewSource(999))
//		qpdf := make([]float64, 2)
//		contracts := generator.NewSingleUseArgumentSet(
//			generator.NewReusableArgumentSet(
//				1000,
//				generator.NewProxyRandomizer(
//					generator.NewExponentialArgRandomizer(rg, 5.0),
//					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//				)))
//
//		keys := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//
//		values := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//
//		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
//
//		db := state.NewMockStateDB(ctrl)
//		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
//		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
//		db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
//		db.EXPECT().CreateAccount(gomock.Any()).Return().Times(1002)
//		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).Times(1002)
//		db.EXPECT().EndTransaction().Return(nil)
//		db.EXPECT().EndBlock().Return(nil)
//		db.EXPECT().EndSyncPeriod().Return()
//		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
//		err := ss.prime()
//		assert.NoError(t, err)
//	})
//
//	t.Run("failed begin block", func(t *testing.T) {
//		rg := rand.New(rand.NewSource(999))
//		qpdf := make([]float64, 2)
//		contracts := generator.NewSingleUseArgumentSet(
//			generator.NewReusableArgumentSet(
//				1000,
//				generator.NewProxyRandomizer(
//					generator.NewExponentialArgRandomizer(rg, 5.0),
//					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//				)))
//		keys := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//		values := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//
//		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
//		mockErr := errors.New("mock error")
//
//		db := state.NewMockStateDB(ctrl)
//		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
//		db.EXPECT().BeginBlock(gomock.Any()).Return(mockErr)
//		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
//		err := ss.prime()
//		assert.Equal(t, mockErr, err)
//	})
//
//	t.Run("failed begin transaction", func(t *testing.T) {
//		rg := rand.New(rand.NewSource(999))
//		qpdf := make([]float64, 2)
//		contracts := generator.NewSingleUseArgumentSet(
//			generator.NewReusableArgumentSet(
//				1000,
//				generator.NewProxyRandomizer(
//					generator.NewExponentialArgRandomizer(rg, 5.0),
//					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//				)))
//		keys := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//		values := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
//		mockErr := errors.New("mock error")
//
//		db := state.NewMockStateDB(ctrl)
//		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
//		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
//		db.EXPECT().BeginTransaction(gomock.Any()).Return(mockErr)
//		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
//		err := ss.prime()
//		assert.Equal(t, mockErr, err)
//	})
//
//	t.Run("failed end transaction", func(t *testing.T) {
//		rg := rand.New(rand.NewSource(999))
//		qpdf := make([]float64, 2)
//		contracts := generator.NewSingleUseArgumentSet(
//			generator.NewReusableArgumentSet(
//				1000,
//				generator.NewProxyRandomizer(
//					generator.NewExponentialArgRandomizer(rg, 5.0),
//					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//				)))
//		keys := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//		values := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
//		mockErr := errors.New("mock error")
//
//		db := state.NewMockStateDB(ctrl)
//		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
//		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
//		db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
//		db.EXPECT().CreateAccount(gomock.Any()).Return().Times(1002)
//		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).Times(1002)
//		db.EXPECT().EndTransaction().Return(mockErr)
//		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
//		err := ss.prime()
//		assert.Equal(t, mockErr, err)
//	})
//
//	t.Run("failed end block", func(t *testing.T) {
//		rg := rand.New(rand.NewSource(999))
//		qpdf := make([]float64, 2)
//		contracts := generator.NewSingleUseArgumentSet(
//			generator.NewReusableArgumentSet(
//				1000,
//				generator.NewProxyRandomizer(
//					generator.NewExponentialArgRandomizer(rg, 5.0),
//					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//				)))
//		keys := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//		values := generator.NewReusableArgumentSet(
//			1000,
//			generator.NewProxyRandomizer(
//				generator.NewExponentialArgRandomizer(rg, 5.0),
//				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
//			))
//		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
//		mockErr := errors.New("mock error")
//
//		db := state.NewMockStateDB(ctrl)
//		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
//		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
//		db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
//		db.EXPECT().CreateAccount(gomock.Any()).Return().Times(1002)
//		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).Times(1002)
//		db.EXPECT().EndTransaction().Return(nil)
//		db.EXPECT().EndBlock().Return(mockErr)
//		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
//		err := ss.prime()
//		assert.Equal(t, mockErr, err)
//	})
//
//}
//
