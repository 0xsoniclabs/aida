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

package prime

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

func TestPrime_NewPrimer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := logger.NewLogger("Info", "TestPrime")

	cfg := &utils.Config{}

	mockStateDb := state.NewMockStateDB(ctrl)
	mockAidaDb := db.NewMockBaseDB(ctrl)
	mockAdapter := db.NewMockDbAdapter(ctrl)
	kv := &testutil.KeyValue{}
	iter := iterator.NewArrayIterator(kv)

	mockAidaDb.EXPECT().GetBackend().Return(mockAdapter).AnyTimes()
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iter).AnyTimes()

	p := newPrimer(cfg, mockStateDb, mockAidaDb, log)

	assert.NotNil(t, p)
	assert.Equal(t, cfg, p.cfg)
	assert.Equal(t, log, p.log)
	assert.Equal(t, mockStateDb, p.ctx.db)
	assert.Equal(t, uint64(0), p.block)
	assert.Equal(t, uint64(0), p.target)
}

func TestPrime_Prime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := logger.NewLogger("Info", "TestPrime")
	primeBlock := uint64(5)
	primeFirst := uint64(10)

	// mock data
	cfg := &utils.Config{}
	update := &updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           5,
		DeletedAccounts: []types.Address{},
	}
	substateBlk9 := &substate.Substate{
		InputSubstate: substate.NewWorldState().Add(types.Address{3}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:         9,
		Transaction:   0,
	}
	retError := errors.New("Test Error")

	t.Run("Normal priming flow", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockUpdateIter := db.NewMockIIterator[*updateset.UpdateSet](ctrl)
		mockSubstateIter := db.NewMockIIterator[*substate.Substate](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, mockDeletionDb, log)

		gomock.InOrder(
			mockUpdateDb.EXPECT().NewUpdateSetIterator(gomock.Any(), gomock.Any()).Return(mockUpdateIter),
			mockUpdateIter.EXPECT().Next().Return(true),
			mockUpdateIter.EXPECT().Value().Return(update),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil),
			mockBulk.EXPECT().Close().Return(nil),
			mockUpdateIter.EXPECT().Next().Return(false),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil),
			mockBulk.EXPECT().CreateAccount(gomock.Any()),
			mockBulk.EXPECT().SetBalance(gomock.Any(), gomock.Any()),
			mockBulk.EXPECT().SetNonce(gomock.Any(), gomock.Any()),
			mockBulk.EXPECT().SetCode(gomock.Any(), gomock.Any()),
			mockBulk.EXPECT().Close().Return(nil),
			mockUpdateIter.EXPECT().Release(),

			mockSubstateDb.EXPECT().NewSubstateIterator(gomock.Any(), gomock.Any()).Return(mockSubstateIter).AnyTimes(),
			mockSubstateIter.EXPECT().Next().Return(true),
			mockSubstateIter.EXPECT().Value().Return(substateBlk9),
			mockDeletionDb.EXPECT().GetDestroyedAccounts(uint64(9), 0).Return([]types.Address{}, []types.Address{}, nil),
			mockSubstateIter.EXPECT().Next().Return(false),
			mockSubstateIter.EXPECT().Release(),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil),
			mockBulk.EXPECT().Close().Return(nil),

			mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{}, nil),
		)
		err := p.Prime()
		assert.NoError(t, err)
	})

	t.Run("Skip priming when block is greater than first", func(t *testing.T) {
		p := newTestPrimer(uint64(15), primeFirst, cfg, nil, nil, nil, nil, log)
		err := p.Prime()
		assert.NoError(t, err)
		assert.Equal(t, uint64(15), p.block) // no changes
	})

	t.Run("mayPrimeFromUpdateSet fails", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockUpdateIter := db.NewMockIIterator[*updateset.UpdateSet](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, nil, nil, log)
		gomock.InOrder(
			mockUpdateDb.EXPECT().NewUpdateSetIterator(gomock.Any(), gomock.Any()).Return(mockUpdateIter),
			mockUpdateIter.EXPECT().Next().Return(true),
			mockUpdateIter.EXPECT().Value().Return(update),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, retError),
			mockUpdateIter.EXPECT().Release(),
		)
		err := p.Prime()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cannot prime from update-set")
	})

	t.Run("mayPrimeFromSubstate fails", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		mockUpdateIter := db.NewMockIIterator[*updateset.UpdateSet](ctrl)
		mockSubstateIter := db.NewMockIIterator[*substate.Substate](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, mockDeletionDb, log)
		gomock.InOrder(
			mockUpdateDb.EXPECT().NewUpdateSetIterator(gomock.Any(), gomock.Any()).Return(mockUpdateIter),
			mockUpdateIter.EXPECT().Next().Return(false),
			mockUpdateIter.EXPECT().Release(),
			mockSubstateDb.EXPECT().NewSubstateIterator(gomock.Any(), gomock.Any()).Return(mockSubstateIter).AnyTimes(),
			mockSubstateIter.EXPECT().Next().Return(true),
			mockSubstateIter.EXPECT().Value().Return(substateBlk9),
			mockDeletionDb.EXPECT().GetDestroyedAccounts(uint64(9), 0).Return([]types.Address{}, []types.Address{}, retError),
			mockSubstateIter.EXPECT().Release(),
		)
		err := p.Prime()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cannot prime from substate")
	})

	t.Run("mayDeleteDestroyedAccountsFromStateDB fails", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockUpdateIter := db.NewMockIIterator[*updateset.UpdateSet](ctrl)
		mockSubstateIter := db.NewMockIIterator[*substate.Substate](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, mockDeletionDb, log)

		gomock.InOrder(
			mockUpdateDb.EXPECT().NewUpdateSetIterator(gomock.Any(), gomock.Any()).Return(mockUpdateIter),
			mockUpdateIter.EXPECT().Next().Return(false),
			mockUpdateIter.EXPECT().Release(),

			mockSubstateDb.EXPECT().NewSubstateIterator(gomock.Any(), gomock.Any()).Return(mockSubstateIter).AnyTimes(),
			mockSubstateIter.EXPECT().Next().Return(true),
			mockSubstateIter.EXPECT().Value().Return(substateBlk9),
			mockDeletionDb.EXPECT().GetDestroyedAccounts(uint64(9), 0).Return([]types.Address{}, []types.Address{}, nil),
			mockSubstateIter.EXPECT().Next().Return(false),
			mockSubstateIter.EXPECT().Release(),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil),
			mockBulk.EXPECT().Close().Return(nil),

			mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{}, retError),
		)
		err := p.Prime()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cannot delete destroyed accounts from state-db")
	})
}

func TestPrime_MayPrimeFromUpdateSet_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := logger.NewLogger("Info", "TestPrime")
	cfg := &utils.Config{}
	primeBlock := uint64(5)
	primeFirst := uint64(10)

	retError := errors.New("Test Error")

	t.Run("Prime stops when block > first", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockUpdateIter := db.NewMockIIterator[*updateset.UpdateSet](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, nil, nil, log)
		updateBlk15 := &updateset.UpdateSet{
			WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
			Block:           15,
			DeletedAccounts: []types.Address{},
		}
		gomock.InOrder(
			mockUpdateDb.EXPECT().NewUpdateSetIterator(gomock.Any(), gomock.Any()).Return(mockUpdateIter),
			mockUpdateIter.EXPECT().Next().Return(true),
			mockUpdateIter.EXPECT().Value().Return(updateBlk15),
			mockUpdateIter.EXPECT().Release(),
		)
		err := p.mayPrimeFromUpdateSet()
		assert.NoError(t, err)
	})

	t.Run("no iterations", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockUpdateIter := db.NewMockIIterator[*updateset.UpdateSet](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, nil, nil, log)
		gomock.InOrder(
			mockUpdateDb.EXPECT().NewUpdateSetIterator(gomock.Any(), gomock.Any()).Return(mockUpdateIter),
			mockUpdateIter.EXPECT().Next().Return(false),
			mockUpdateIter.EXPECT().Release(),
		)
		err := p.mayPrimeFromUpdateSet()
		assert.NoError(t, err)
	})

	t.Run("PrimeStateDB fails", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockUpdateIter := db.NewMockIIterator[*updateset.UpdateSet](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, nil, nil, log)
		updateBlk5 := &updateset.UpdateSet{
			WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
			Block:           5,
			DeletedAccounts: []types.Address{},
		}
		gomock.InOrder(
			mockUpdateDb.EXPECT().NewUpdateSetIterator(gomock.Any(), gomock.Any()).Return(mockUpdateIter),
			mockUpdateIter.EXPECT().Next().Return(true),
			mockUpdateIter.EXPECT().Value().Return(updateBlk5),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, retError),
			mockUpdateIter.EXPECT().Release(),
		)
		err := p.mayPrimeFromUpdateSet()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cannot prime state-db")
	})
}

func TestPrime_MayPrimeFromSubstate_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := logger.NewLogger("Info", "TestPrime")
	cfg := &utils.Config{}
	primeBlock := uint64(5)
	primeFirst := uint64(10)

	substateBlk9 := &substate.Substate{
		InputSubstate: substate.NewWorldState().Add(types.Address{3}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:         9,
		Transaction:   0,
	}
	substateBlk11 := &substate.Substate{
		InputSubstate: substate.NewWorldState().Add(types.Address{3}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:         11,
		Transaction:   0,
	}
	retError := errors.New("Test Error")

	t.Run("No priming because first substate block is larger than the first block", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockSubstateIter := db.NewMockIIterator[*substate.Substate](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, nil, mockSubstateDb, mockDeletionDb, log)
		gomock.InOrder(
			mockSubstateDb.EXPECT().NewSubstateIterator(gomock.Any(), gomock.Any()).Return(mockSubstateIter).AnyTimes(),
			mockSubstateIter.EXPECT().Next().Return(true),
			mockSubstateIter.EXPECT().Value().Return(substateBlk11),
			mockSubstateIter.EXPECT().Release(),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil),
			mockBulk.EXPECT().Close().Return(nil),
		)
		err := p.mayPrimeFromSubstate()
		assert.NoError(t, err)
	})

	t.Run("generateUpdateSet fails", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		mockSubstateIter := db.NewMockIIterator[*substate.Substate](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, nil, mockSubstateDb, mockDeletionDb, log)
		gomock.InOrder(
			mockSubstateDb.EXPECT().NewSubstateIterator(gomock.Any(), gomock.Any()).Return(mockSubstateIter).AnyTimes(),
			mockSubstateIter.EXPECT().Next().Return(true),
			mockSubstateIter.EXPECT().Value().Return(substateBlk9),
			mockDeletionDb.EXPECT().GetDestroyedAccounts(uint64(9), 0).Return([]types.Address{}, []types.Address{}, retError),
			mockSubstateIter.EXPECT().Release(),
		)
		err := p.mayPrimeFromSubstate()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cannot generate update-set")
	})

	t.Run("PrimeStateDB fails", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockSubstateIter := db.NewMockIIterator[*substate.Substate](ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, nil, mockSubstateDb, mockDeletionDb, log)
		gomock.InOrder(
			mockSubstateDb.EXPECT().NewSubstateIterator(gomock.Any(), gomock.Any()).Return(mockSubstateIter).AnyTimes(),
			mockSubstateIter.EXPECT().Next().Return(true),
			mockSubstateIter.EXPECT().Value().Return(substateBlk9),
			mockDeletionDb.EXPECT().GetDestroyedAccounts(uint64(9), 0).Return([]types.Address{}, []types.Address{}, nil),
			mockSubstateIter.EXPECT().Next().Return(false),
			mockSubstateIter.EXPECT().Release(),
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, retError),
		)
		err := p.mayPrimeFromSubstate()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cannot prime state-db")
	})
}

func TestPrime_MayDeleteDestroyedAccountsFromStateDB_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := logger.NewLogger("Info", "TestPrime")
	cfg := &utils.Config{}
	primeBlock := uint64(5)
	primeFirst := uint64(10)

	t.Run("remove accounts", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, nil, nil, mockDeletionDb, log)
		acc1 := types.Address{1}
		acc2 := types.Address{2}

		gomock.InOrder(
			mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{acc1, acc2}, nil),
			mockStateDb.EXPECT().BeginSyncPeriod(uint64(0)),
			mockStateDb.EXPECT().BeginBlock(uint64(0)).Return(nil),
			mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil),
			mockStateDb.EXPECT().SelfDestruct(common.Address(acc1)),
			mockStateDb.EXPECT().SelfDestruct(common.Address(acc2)),
			mockStateDb.EXPECT().EndTransaction().Return(nil),
			mockStateDb.EXPECT().EndBlock().Return(nil),
			mockStateDb.EXPECT().EndSyncPeriod(),
		)
		err := p.mayDeleteDestroyedAccountsFromStateDB(9)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), p.ctx.block)
	})

	t.Run("shortcut", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, nil, nil, mockDeletionDb, log)
		p.ctx.block = 0

		gomock.InOrder(
			mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{}, nil),
		)
		err := p.mayDeleteDestroyedAccountsFromStateDB(9)
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), p.ctx.block)
	})
}

func TestPrime_MayDeleteDestroyedAccountsFromStateDB_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := logger.NewLogger("Info", "TestPrime")
	cfg := &utils.Config{}
	retError := errors.New("Test Error")
	acc1 := types.Address{1}
	primeBlock := uint64(5)
	primeFirst := uint64(10)

	testcases := []struct {
		name       string
		setupMocks func(
			mockStateDb *state.MockStateDB,
			mockDeletionDb *db.MockDestroyedAccountDB,
		)
	}{
		{
			name: "GetAccountsDestroyedInRange",
			setupMocks: func(mockStateDb *state.MockStateDB, mockDeletionDb *db.MockDestroyedAccountDB) {
				gomock.InOrder(
					mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{acc1}, retError),
				)
			},
		},
		{
			name: "BeginBlock",
			setupMocks: func(mockStateDb *state.MockStateDB, mockDeletionDb *db.MockDestroyedAccountDB) {
				gomock.InOrder(
					mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{acc1}, nil),
					mockStateDb.EXPECT().BeginSyncPeriod(uint64(0)),
					mockStateDb.EXPECT().BeginBlock(uint64(0)).Return(retError),
				)
			},
		},
		{
			name: "BeginTransaction",
			setupMocks: func(mockStateDb *state.MockStateDB, mockDeletionDb *db.MockDestroyedAccountDB) {
				gomock.InOrder(
					mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{acc1}, nil),
					mockStateDb.EXPECT().BeginSyncPeriod(uint64(0)),
					mockStateDb.EXPECT().BeginBlock(uint64(0)).Return(nil),
					mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(retError),
				)
			},
		},
		{
			name: "EndTransaction",
			setupMocks: func(mockStateDb *state.MockStateDB, mockDeletionDb *db.MockDestroyedAccountDB) {
				gomock.InOrder(
					mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{acc1}, nil),
					mockStateDb.EXPECT().BeginSyncPeriod(uint64(0)),
					mockStateDb.EXPECT().BeginBlock(uint64(0)).Return(nil),
					mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil),
					mockStateDb.EXPECT().SelfDestruct(common.Address(acc1)),
					mockStateDb.EXPECT().EndTransaction().Return(retError),
				)
			},
		},
		{
			name: "EndBlock",
			setupMocks: func(mockStateDb *state.MockStateDB, mockDeletionDb *db.MockDestroyedAccountDB) {
				gomock.InOrder(
					mockDeletionDb.EXPECT().GetAccountsDestroyedInRange(uint64(0), uint64(9)).Return([]types.Address{acc1}, nil),
					mockStateDb.EXPECT().BeginSyncPeriod(uint64(0)),
					mockStateDb.EXPECT().BeginBlock(uint64(0)).Return(nil),
					mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil),
					mockStateDb.EXPECT().SelfDestruct(common.Address(acc1)),
					mockStateDb.EXPECT().EndTransaction().Return(nil),
					mockStateDb.EXPECT().EndBlock().Return(retError),
				)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockStateDb := state.NewMockStateDB(ctrl)
			mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
			p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, nil, nil, mockDeletionDb, log)
			tc.setupMocks(mockStateDb, mockDeletionDb)
			err := p.mayDeleteDestroyedAccountsFromStateDB(9)
			assert.Error(t, err, tc.name+" does not fail.")
			assert.ErrorContains(t, err, "Test Error")
			assert.Equal(t, uint64(0), p.ctx.block)
		})
	}
}

func TestPrime_TrySetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &utils.Config{
		First: 0,
	}
	primeBlock := uint64(0)
	primeFirst := uint64(0)

	t.Run("Existing state db returns error", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockLog := logger.NewMockLogger(ctrl)

		cfg.IsExistingStateDb = true
		cfg.StateDbSrc = t.TempDir()
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, nil, mockLog)
		mockLog.EXPECT().Warningf("cannot read state db info; %v", gomock.Any())
		p.trySetBlocks()
		assert.Equal(t, uint64(1), p.block)
		assert.Equal(t, uint64(1), p.target)
	})

	t.Run("Existing state db success", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockLog := logger.NewMockLogger(ctrl)

		cfg.IsExistingStateDb = true
		cfg.StateDbSrc = t.TempDir()
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, nil, mockLog)
		_ = utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, uint64(19), common.Hash{}, true)
		p.trySetBlocks()
		assert.Equal(t, uint64(20), p.block)
		assert.Equal(t, uint64(20), p.target)
	})

	t.Run("Non-existing state db, empty substate db", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockLog := logger.NewMockLogger(ctrl)

		cfg.IsExistingStateDb = false
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, nil, mockLog)
		mockLog.EXPECT().Warning("cannot get first substate; substate db is empty")
		mockSubstateDb.EXPECT().GetFirstSubstate().Return(nil)
		p.trySetBlocks()
		assert.Equal(t, uint64(0), p.block)
		assert.Equal(t, uint64(0), p.target)
	})

	t.Run("Non-existing state db, substate first < update-set first", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockLog := logger.NewMockLogger(ctrl)

		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, nil, mockLog)
		mockSubstate := &substate.Substate{Block: uint64(10)}
		gomock.InOrder(
			mockSubstateDb.EXPECT().GetFirstSubstate().Return(mockSubstate),
			mockUpdateDb.EXPECT().GetFirstKey().Return(uint64(20), nil),
		)
		p.trySetBlocks()
		assert.Equal(t, uint64(10), p.block)
		assert.Equal(t, uint64(10), p.target)
	})

	t.Run("Non-existing state db, substate first > update-set first", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockLog := logger.NewMockLogger(ctrl)

		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, nil, mockLog)
		mockSubstate := &substate.Substate{Block: uint64(30)}
		gomock.InOrder(
			mockSubstateDb.EXPECT().GetFirstSubstate().Return(mockSubstate),
			mockUpdateDb.EXPECT().GetFirstKey().Return(uint64(20), nil),
		)
		p.trySetBlocks()
		assert.Equal(t, uint64(20), p.block)
		assert.Equal(t, uint64(30), p.target)
	})

	t.Run("State db exist, first processable block is cfg.First", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockLog := logger.NewMockLogger(ctrl)

		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, nil, mockLog)
		cfg.First = 100
		cfg.IsExistingStateDb = true
		cfg.StateDbSrc = t.TempDir()
		_ = utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, uint64(9), common.Hash{}, true)
		p.trySetBlocks()
		assert.Equal(t, uint64(10), p.block)
		assert.Equal(t, uint64(100), p.target)
	})

	t.Run("Non-existing state db, first processable block is cfg.First", func(t *testing.T) {
		mockStateDb := state.NewMockStateDB(ctrl)
		mockUpdateDb := db.NewMockUpdateDB(ctrl)
		mockSubstateDb := db.NewMockSubstateDB(ctrl)
		mockLog := logger.NewMockLogger(ctrl)

		cfg.First = 100
		cfg.IsExistingStateDb = false
		mockSubstate := &substate.Substate{Block: uint64(20)}
		p := newTestPrimer(primeBlock, primeFirst, cfg, mockStateDb, mockUpdateDb, mockSubstateDb, nil, mockLog)
		gomock.InOrder(
			mockSubstateDb.EXPECT().GetFirstSubstate().Return(mockSubstate),
			mockUpdateDb.EXPECT().GetFirstKey().Return(uint64(20), nil),
		)
		p.trySetBlocks()
		assert.Equal(t, uint64(20), p.block)
		assert.Equal(t, uint64(100), p.target)
	})
}

func newTestPrimer(
	block uint64,
	first uint64,
	cfg *utils.Config,
	stateDb state.StateDB,
	updateDb db.UpdateDB,
	substateDb db.SubstateDB,
	deletionDb db.DestroyedAccountDB,
	log logger.Logger,
) *primer {
	return &primer{
		cfg:    cfg,
		log:    log,
		ctx:    newPrimeContext(cfg, stateDb, log),
		udb:    updateDb,
		sdb:    substateDb,
		ddb:    deletionDb,
		block:  block,
		target: first,
	}
}
