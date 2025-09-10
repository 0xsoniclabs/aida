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
	"errors"
	"math/rand"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/statistics/classifier"
	"github.com/0xsoniclabs/aida/stochastic/statistics/generator"
	"github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestReplay_ExecuteRevertSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)
	defer ctrl.Finish()
	contracts := generator.NewMockArgumentSet(ctrl)
	keys := generator.NewMockArgumentSet(ctrl)
	values := generator.NewMockArgumentSet(ctrl)
	snapshots := generator.NewMockSnapshotSet(ctrl)
	gomock.InOrder(
		snapshots.EXPECT().SampleSnapshot(gomock.Any()).Return(1).Times(1),
		db.EXPECT().RevertToSnapshot(gomock.Any()).Times(1),
	)

	rg := rand.New(rand.NewSource(999))
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, nil)
	ss.activeSnapshots = []int{1, 2, 3, 4, 5}
	snapshotSize := len(ss.activeSnapshots)
	ss.execute(operations.RevertToSnapshotID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
	assert.GreaterOrEqual(t, len(ss.activeSnapshots), 1)         // must have at lest one snapshot
	assert.LessOrEqual(t, len(ss.activeSnapshots), snapshotSize) // must not have more than 5 snapshots
}

//bfs: revisit this test case after changing the json format
//func TestReplay_RunStochasticReplay(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	tmpDir := t.TempDir()
//	cfg := &utils.Config{
//		ContractNumber:    1000,
//		KeysNumber:        1000,
//		ValuesNumber:      1000,
//		SnapshotDepth:     100,
//		BlockLength:       3,
//		SyncPeriodLength:  10,
//		TransactionLength: 2,
//		BalanceRange:      100,
//		NonceRange:        100,
//		Debug:             true,
//		ShadowImpl:        "geth",
//		DbTmp:             tmpDir,
//		DbImpl:            "carmen",
//		DbVariant:         "go-file",
//	}
//	db := state.NewMockStateDB(ctrl)
//	log := logger.NewLogger("INFO", "test")
//	e, err := recorder.ReadSimulation("data/test_replay.json")
//	if err != nil {
//		t.Fatalf("Failed to read simulation: %v", err)
//	}
//	counter := 0
//	db.EXPECT().CreateAccount(gomock.Any()).Times(1001)
//	db.EXPECT().GetShadowDB().Return(db)
//	db.EXPECT().BeginSyncPeriod(gomock.Any()).Return().Times(2)
//	db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
//	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
//	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(1)).Times(1001)
//	db.EXPECT().EndTransaction().Return(nil)
//	db.EXPECT().EndBlock().Return(nil).Times(2)
//	db.EXPECT().EndSyncPeriod().Return()
//	db.EXPECT().Error().Return(nil)
//	err = RunStochasticReplay(db, e, 0, cfg, log)
//	fmt.Printf("Counter: %d\n", counter)
//	assert.NoError(t, err)
//}

func TestStochasticState_execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// enumerate whole operation space with arguments
	// and check encoding/decoding whether it is symmetric.
	var codes = map[int]int{}
	for op := 0; op < operations.NumOps; op++ {
		for addr := 0; addr < classifier.NumArgKinds; addr++ {
			for key := 0; key < classifier.NumArgKinds; key++ {
				for value := 0; value < classifier.NumArgKinds; value++ {
					// check legality of argument/op combination
					if (operations.OpNumArgs[op] == 0 && addr == classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
						(operations.OpNumArgs[op] == 1 && addr != classifier.NoArgID && key == classifier.NoArgID && value == classifier.NoArgID) ||
						(operations.OpNumArgs[op] == 2 && addr != classifier.NoArgID && key != classifier.NoArgID && value == classifier.NoArgID) ||
						(operations.OpNumArgs[op] == 3 && addr != classifier.NoArgID && key != classifier.NoArgID && value != classifier.NoArgID) {

						// encode to an argument-encoded operation
						var err error
						codes[op], err = operations.EncodeArgOp(op, addr, key, value)
						if err != nil {
							t.Fatalf("Encoding failed for %v", codes[op])
						}
					}
				}
			}
		}
	}

	contracts := generator.NewMockArgumentSet(ctrl)
	keys := generator.NewMockArgumentSet(ctrl)
	values := generator.NewMockArgumentSet(ctrl)
	snapshots := generator.NewMockSnapshotSet(ctrl)
	db := state.NewMockStateDB(ctrl)
	db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0))
	db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
	db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	db.EXPECT().CreateAccount(gomock.Any()).Return()
	db.EXPECT().CreateContract(gomock.Any()).Return()
	db.EXPECT().Empty(gomock.Any()).Return(false)
	db.EXPECT().EndBlock().Return(nil)
	db.EXPECT().EndSyncPeriod().Return()
	db.EXPECT().EndTransaction().Return(nil)
	db.EXPECT().Exist(gomock.Any()).Return(false)
	db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(0))
	db.EXPECT().GetCode(gomock.Any()).Return(nil)
	db.EXPECT().GetCodeHash(gomock.Any()).Return(common.Hash{})
	db.EXPECT().GetCodeSize(gomock.Any()).Return(0)
	db.EXPECT().GetCommittedState(gomock.Any(), gomock.Any()).Return(common.Hash{})
	db.EXPECT().GetNonce(gomock.Any()).Return(uint64(0))
	db.EXPECT().GetState(gomock.Any(), gomock.Any()).Return(common.Hash{})
	db.EXPECT().GetStorageRoot(gomock.Any()).Return(common.Hash{})
	db.EXPECT().GetTransientState(gomock.Any(), gomock.Any()).Return(common.Hash{})
	db.EXPECT().HasSelfDestructed(gomock.Any()).Return(false)
	db.EXPECT().RevertToSnapshot(gomock.Any()).Return()
	db.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(0))
	db.EXPECT().SelfDestruct6780(gomock.Any()).Return(*uint256.NewInt(0), false)
	db.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return(nil)
	db.EXPECT().SetNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return()
	db.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Hash{})
	db.EXPECT().SetTransientState(gomock.Any(), gomock.Any(), gomock.Any()).Return()
	db.EXPECT().Snapshot().Return(0)
	db.EXPECT().GetShadowDB().Return(db)
	db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(10))
	db.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0))

	rg := rand.New(rand.NewSource(999))
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
	ss.enableDebug()
	for i := 0; i < len(codes); i++ {
		ss.activeSnapshots = []int{1, 2, 3, 4, 5}
		op, addr, key, value, err := operations.DecodeArgOp(codes[i])
		if err != nil {
			t.Fatalf("Decoding failed for %v", codes[i])
		}
		ss.execute(op, addr, key, value)
	}
}

func TestStochasticState_prime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		rg := rand.New(rand.NewSource(999))
		qpdf := make([]float64, 2)
		contracts := generator.NewSingleUseArgumentSet(
			generator.NewReusableArgumentSet(
				1000,
				generator.NewProxyRandomizer(
					generator.NewExponentialArgRandomizer(rg, 5.0),
					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
				)))

		keys := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))

		values := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))

		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)

		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
		db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
		db.EXPECT().CreateAccount(gomock.Any()).Return().Times(1002)
		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).Times(1002)
		db.EXPECT().EndTransaction().Return(nil)
		db.EXPECT().EndBlock().Return(nil)
		db.EXPECT().EndSyncPeriod().Return()
		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
		err := ss.prime()
		assert.NoError(t, err)
	})

	t.Run("failed begin block", func(t *testing.T) {
		rg := rand.New(rand.NewSource(999))
		qpdf := make([]float64, 2)
		contracts := generator.NewSingleUseArgumentSet(
			generator.NewReusableArgumentSet(
				1000,
				generator.NewProxyRandomizer(
					generator.NewExponentialArgRandomizer(rg, 5.0),
					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
				)))
		keys := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))
		values := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))

		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
		mockErr := errors.New("mock error")

		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
		db.EXPECT().BeginBlock(gomock.Any()).Return(mockErr)
		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
		err := ss.prime()
		assert.Equal(t, mockErr, err)
	})

	t.Run("failed begin transaction", func(t *testing.T) {
		rg := rand.New(rand.NewSource(999))
		qpdf := make([]float64, 2)
		contracts := generator.NewSingleUseArgumentSet(
			generator.NewReusableArgumentSet(
				1000,
				generator.NewProxyRandomizer(
					generator.NewExponentialArgRandomizer(rg, 5.0),
					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
				)))
		keys := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))
		values := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))
		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
		mockErr := errors.New("mock error")

		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
		db.EXPECT().BeginTransaction(gomock.Any()).Return(mockErr)
		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
		err := ss.prime()
		assert.Equal(t, mockErr, err)
	})

	t.Run("failed end transaction", func(t *testing.T) {
		rg := rand.New(rand.NewSource(999))
		qpdf := make([]float64, 2)
		contracts := generator.NewSingleUseArgumentSet(
			generator.NewReusableArgumentSet(
				1000,
				generator.NewProxyRandomizer(
					generator.NewExponentialArgRandomizer(rg, 5.0),
					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
				)))
		keys := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))
		values := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))
		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
		mockErr := errors.New("mock error")

		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
		db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
		db.EXPECT().CreateAccount(gomock.Any()).Return().Times(1002)
		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).Times(1002)
		db.EXPECT().EndTransaction().Return(mockErr)
		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
		err := ss.prime()
		assert.Equal(t, mockErr, err)
	})

	t.Run("failed end block", func(t *testing.T) {
		rg := rand.New(rand.NewSource(999))
		qpdf := make([]float64, 2)
		contracts := generator.NewSingleUseArgumentSet(
			generator.NewReusableArgumentSet(
				1000,
				generator.NewProxyRandomizer(
					generator.NewExponentialArgRandomizer(rg, 5.0),
					generator.NewEmpiricalQueueRandomizer(rg, qpdf),
				)))
		keys := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))
		values := generator.NewReusableArgumentSet(
			1000,
			generator.NewProxyRandomizer(
				generator.NewExponentialArgRandomizer(rg, 5.0),
				generator.NewEmpiricalQueueRandomizer(rg, qpdf),
			))
		snapshots := generator.NewExponentialSnapshotRandomizer(rg, 0.1)
		mockErr := errors.New("mock error")

		db := state.NewMockStateDB(ctrl)
		db.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
		db.EXPECT().BeginBlock(gomock.Any()).Return(nil)
		db.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
		db.EXPECT().CreateAccount(gomock.Any()).Return().Times(1002)
		db.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).Times(1002)
		db.EXPECT().EndTransaction().Return(nil)
		db.EXPECT().EndBlock().Return(mockErr)
		ss := newReplayContext(rg, db, contracts, keys, values, snapshots, logger.NewLogger("INFO", "test"))
		err := ss.prime()
		assert.Equal(t, mockErr, err)
	})

}
