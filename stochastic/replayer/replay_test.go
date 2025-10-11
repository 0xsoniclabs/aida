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

package replayer

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
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
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// stubSnapshots implements arguments.SnapshotSet for testing
type stubSnapshots struct{ ret int }

func (s *stubSnapshots) SampleSnapshot(n int) int { return s.ret }

const (
	testBalanceRange int64 = 100
	testNonceRange   int   = 10
)

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
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, log, testBalanceRange, testNonceRange)
	ss.activeSnapshots = []int{1, 2, 3, 4, 5}
	snapshotSize := len(ss.activeSnapshots)

	if err := ss.execute(operations.RevertToSnapshotID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	assert.GreaterOrEqual(t, len(ss.activeSnapshots), 1)         // must have at least one snapshot
	assert.LessOrEqual(t, len(ss.activeSnapshots), snapshotSize) // must not have more than initial snapshots
}

func TestPopulateReplayContextAdjustsRanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := state.NewMockStateDB(ctrl)
	db.EXPECT().BeginSyncPeriod(uint64(0))
	db.EXPECT().BeginBlock(uint64(0)).Return(nil)
	db.EXPECT().BeginTransaction(uint32(0)).Return(nil)
	db.EXPECT().CreateAccount(gomock.Any()).AnyTimes()
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
	db.EXPECT().EndTransaction().Return(nil)
	db.EXPECT().EndBlock().Return(nil)
	db.EXPECT().EndSyncPeriod()

	dist := make([]float64, stochastic.QueueLen)
	for i := range dist {
		dist[i] = 1.0 / float64(stochastic.QueueLen)
	}
	counting := recArgs.ArgStatsJSON{
		N:    16,
		ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
	}
	stats := &recorder.StatsJSON{
		Contracts: recArgs.ClassifierJSON{
			Counting: counting,
			Queuing:  recArgs.QueueStatsJSON{Distribution: dist},
		},
		Keys: recArgs.ClassifierJSON{
			Counting: counting,
			Queuing:  recArgs.QueueStatsJSON{Distribution: dist},
		},
		Values: recArgs.ClassifierJSON{
			Counting: counting,
			Queuing:  recArgs.QueueStatsJSON{Distribution: dist},
		},
		SnapshotECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		Balance: recorder.ScalarStatsJSON{
			Max:  9,
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		},
		Nonce: recorder.ScalarStatsJSON{
			Max:  4,
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		},
		CodeSize: recorder.ScalarStatsJSON{
			Max:  3,
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		},
	}
	rg := rand.New(rand.NewSource(99))
	log := logger.NewLogger("INFO", "test")

	ctx, err := populateReplayContext(stats, db, rg, log, -5, -2)
	require.NoError(t, err)
	require.NotNil(t, ctx)

	assert.Equal(t, stats.Balance.Max+1, ctx.balanceRange)
	assert.Equal(t, int(stats.Nonce.Max+1), ctx.nonceRange)

	for i := 0; i < 5; i++ {
		b := ctx.balanceSampler.Sample(ctx.balanceRange)
		assert.GreaterOrEqual(t, b, int64(0))
		assert.Less(t, b, ctx.balanceRange)

		n := ctx.nonceSampler.Sample(int64(ctx.nonceRange))
		assert.GreaterOrEqual(t, n, int64(0))
		assert.Less(t, n, int64(ctx.nonceRange))

		c := ctx.codeSampler.Sample(MaxCodeSize)
		assert.GreaterOrEqual(t, c, int64(0))
		assert.Less(t, c, int64(MaxCodeSize))
	}
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
	db.EXPECT().GetStateAndCommittedState(gomock.Any(), gomock.Any()).Return(common.Hash{}, common.Hash{}).AnyTimes()
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
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, log, testBalanceRange, testNonceRange)
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
	_ = ss.execute(operations.GetStateAndCommittedStateID, addrCl, keyCl, stochastic.NoArgID)
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
	ss := newReplayContext(rand.New(rand.NewSource(1)), db, nil, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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

	ss := newReplayContext(rand.New(rand.NewSource(7)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
		if err := ss.prime(); err == nil {
			t.Fatalf("expected error from BeginBlock")
		}
	})

	t.Run("begin tx fails", func(t *testing.T) {
		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(uint64(0))
		db.EXPECT().BeginBlock(uint64(0)).Return(nil)
		db.EXPECT().BeginTransaction(uint32(0)).Return(assert.AnError)
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
		ss := newReplayContext(rand.New(rand.NewSource(1)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
	e := &recorder.StatsJSON{
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
	e := &recorder.StatsJSON{
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
	e := &recorder.StatsJSON{
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
	e := &recorder.StatsJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{BalanceRange: 10, NonceRange: 10, RandomSeed: 2, ContinueOnFailure: true}
	if err := RunStochasticReplay(db, e, 1, cfg, logger.NewLogger("INFO", "test")); err == nil {
		t.Fatalf("expected aggregated error even when continuing on failure")
	}
}

// TestEnableDebug covers the trivial enableDebug method.
func TestEnableDebug(t *testing.T) {
	ss := newReplayContext(rand.New(rand.NewSource(1)), nil, nil, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
	e := &recorder.StatsJSON{
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
	e := &recorder.StatsJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
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
	e := &recorder.StatsJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
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
	ss := newReplayContext(rg, db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
	if err := ss.execute(operations.GetBalanceID, stochastic.RandArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected error for contract Choose failure")
	}

	// key choose failure
	contracts = repArgs.NewMockSet(ctrl)
	keys = repArgs.NewMockSet(ctrl)
	values = repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	keys.EXPECT().Choose(gomock.Any()).Return(int64(0), assert.AnError)
	ss = newReplayContext(rg, db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
	ss = newReplayContext(rg, db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
	ss := newReplayContext(rg, db, nil, nil, nil, &stubSnapshots{ret: 100}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
	ss := newReplayContext(rg, db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
	ss := newReplayContext(rand.New(rand.NewSource(3)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
	if err := ss.execute(operations.GetBalanceID, stochastic.RandArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected address conversion error")
	}

	// Key conversion error
	contracts = repArgs.NewMockSet(ctrl)
	keys := repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	keys.EXPECT().Choose(gomock.Any()).Return(int64(-1), nil)
	ss = newReplayContext(rand.New(rand.NewSource(4)), db, contracts, keys, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
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
	ss = newReplayContext(rand.New(rand.NewSource(5)), db, contracts, keys, values, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
	if err := ss.execute(operations.SetStateID, stochastic.RandArgID, stochastic.RandArgID, stochastic.RandArgID); err == nil {
		t.Fatalf("expected value conversion error")
	}
}

// TestExecute_DBOperationErrors ensures database failures are propagated as errors.
func TestExecute_DBOperationErrors(t *testing.T) {
	testCases := []struct {
		name string
		op   int
		set  func(*state.MockStateDB)
	}{
		{
			name: "BeginBlock",
			op:   operations.BeginBlockID,
			set: func(db *state.MockStateDB) {
				db.EXPECT().BeginBlock(gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name: "BeginTransaction",
			op:   operations.BeginTransactionID,
			set: func(db *state.MockStateDB) {
				db.EXPECT().BeginTransaction(gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name: "EndTransaction",
			op:   operations.EndTransactionID,
			set: func(db *state.MockStateDB) {
				db.EXPECT().EndTransaction().Return(assert.AnError)
			},
		},
		{
			name: "EndBlock",
			op:   operations.EndBlockID,
			set: func(db *state.MockStateDB) {
				db.EXPECT().EndBlock().Return(assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			db := state.NewMockStateDB(ctrl)
			tc.set(db)
			ss := newReplayContext(rand.New(rand.NewSource(1)), db, nil, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
			err := ss.execute(tc.op, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if !errors.Is(err, assert.AnError) {
				t.Fatalf("expected wrapped assert.AnError, got %v", err)
			}
		})
	}
}

// TestPopulateReplayContext_PrimeError ensures errors from prime() propagate.
func TestPopulateReplayContext_PrimeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	rg := rand.New(rand.NewSource(7))
	lg := logger.NewLogger("INFO", "test")
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.2
	for i := 1; i < len(qpdf); i++ {
		qpdf[i] = 0.8 / float64(stochastic.QueueLen-1)
	}
	cls := recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: [][2]float64{{0, 0}, {1, 1}}}, Queuing: recArgs.QueueStatsJSON{Distribution: qpdf}}
	e := &recorder.StatsJSON{Operations: []string{operations.OpMnemo(operations.BeginSyncPeriodID)}, StochasticMatrix: [][]float64{{1.0}}, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}

	// Cause prime to fail at BeginBlock
	db.EXPECT().BeginSyncPeriod(uint64(0))
	db.EXPECT().BeginBlock(uint64(0)).Return(assert.AnError)
	if _, err := populateReplayContext(e, db, rg, lg, testBalanceRange, testNonceRange); err == nil {
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
	e := &recorder.StatsJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
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
	e := &recorder.StatsJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
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
	e := &recorder.StatsJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
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
	e := &recorder.StatsJSON{Operations: labels, StochasticMatrix: A, Contracts: cls, Keys: cls, Values: cls, SnapshotECDF: [][2]float64{{0, 0}, {1, 1}}}
	cfg := &utils.Config{RandomSeed: 1, BalanceRange: 10, NonceRange: 10, Debug: true, DebugFrom: 2}
	if err := RunStochasticReplay(db, e, 2, cfg, logger.NewLogger("INFO", "test")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestExecute_SetCodeRandomReadError forces randReadBytes error and propagates it.
func TestExecute_SetCodeRandomReadError(t *testing.T) {
	old := randReadBytes
	randReadBytes = func(_ *rand.Rand, _ []byte) (int, error) { return 0, assert.AnError }
	defer func() { randReadBytes = old }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	// contracts choose to supply addr
	contracts := repArgs.NewMockSet(ctrl)
	contracts.EXPECT().Choose(gomock.Any()).Return(int64(1), nil)
	ss := newReplayContext(rand.New(rand.NewSource(10)), db, contracts, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
	if err := ss.execute(operations.SetCodeID, stochastic.RandArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected error for randReadBytes failure")
	} else if !errors.Is(err, assert.AnError) {
		t.Fatalf("expected wrapped randReadBytes error, got %v", err)
	}
}

// TestPopulateReplayContext_Errors covers randomizer ctor error branches.
func TestPopulateReplayContext_Errors(t *testing.T) {
	rg := rand.New(rand.NewSource(123))
	log := logger.NewLogger("INFO", "test")
	db := state.NewMockStateDB(gomock.NewController(t))

	badDist := make([]float64, stochastic.QueueLen-1) // invalid length causes error
	goodDist := make([]float64, stochastic.QueueLen)
	goodDist[0] = 0.1
	for i := 1; i < len(goodDist); i++ {
		goodDist[i] = 0.9 / float64(stochastic.QueueLen-1)
	}

	ecdf := [][2]float64{{0, 0}, {1, 1}}
	badE := func() *recorder.StatsJSON {
		return &recorder.StatsJSON{
			Contracts:    recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: badDist}},
			Keys:         recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Values:       recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			SnapshotECDF: ecdf,
		}
	}
	if _, err := populateReplayContext(badE(), db, rg, log, testBalanceRange, testNonceRange); err == nil {
		t.Fatalf("expected error for bad contracts distribution")
	}

	badE2 := func() *recorder.StatsJSON {
		return &recorder.StatsJSON{
			Contracts:    recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Keys:         recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: badDist}},
			Values:       recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			SnapshotECDF: ecdf,
		}
	}
	if _, err := populateReplayContext(badE2(), db, rg, log, testBalanceRange, testNonceRange); err == nil {
		t.Fatalf("expected error for bad keys distribution")
	}

	badE3 := func() *recorder.StatsJSON {
		return &recorder.StatsJSON{
			Contracts:    recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Keys:         recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: goodDist}},
			Values:       recArgs.ClassifierJSON{Counting: recArgs.ArgStatsJSON{N: 400, ECDF: ecdf}, Queuing: recArgs.QueueStatsJSON{Distribution: badDist}},
			SnapshotECDF: ecdf,
		}
	}
	if _, err := populateReplayContext(badE3(), db, rg, log, testBalanceRange, testNonceRange); err == nil {
		t.Fatalf("expected error for bad values distribution")
	}
}

// TestExecute_DebugEncodeOpcodeError triggers encode error in debug printing.
func TestExecute_DebugEncodeOpcodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	ss := newReplayContext(rand.New(rand.NewSource(1)), db, nil, nil, nil, &stubSnapshots{ret: 0}, logger.NewLogger("INFO", "test"), testBalanceRange, testNonceRange)
	ss.enableDebug()
	// Use a two-arg op with no args to make EncodeOpcode fail without triggering Choose
	if err := ss.execute(operations.GetStateID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID); err == nil {
		t.Fatalf("expected encode opcode error in debug path")
	}
}
