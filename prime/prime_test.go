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
	"fmt"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/rlp"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"go.uber.org/mock/gomock"
)

func TestPrime_Prime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := logger.NewLogger("Info", "TestPrime")

	cfg := &utils.Config{
		SkipPriming:       false,
		StateDbSrc:        "",
		First:             0,
		IsExistingStateDb: false,
	}

	mockDb := state.NewMockStateDB(ctrl)
	mockAida := db.NewMockBaseDB(ctrl)
	mockAdapter := db.NewMockDbAdapter(ctrl)
	mockSubstate := db.NewMockSubstateDB(ctrl)
	mockUpdateDb := db.NewMockUpdateDB(ctrl)
	mockDeletion := db.NewMockDestroyedAccountDB(ctrl)
	mockBulk := state.NewMockBulkLoad(ctrl)
	p := &primer{
		log:    log,
		ctx:    NewPrimeContext(cfg, mockDb, log),
		cfg:    cfg,
		aidadb: mockAida,
		sdb:    mockSubstate,
		udb:    mockUpdateDb,
		ddb:    mockDeletion,
		block:  5,
		first:  10,
	}

	// Normal priming flow
	mockAida.EXPECT().GetBackend().Return(mockAdapter)
	mockIter := NewMockProxyIterator(ctrl)
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockAdapter.EXPECT().Get(gomock.Any(), gomock.Any()).Return([]uint8{1, 2, 3}, nil)
	mockDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockDb.EXPECT().EndTransaction().Return(nil)
	mockDb.EXPECT().EndBlock().Return(nil)
	mockDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil)
	mockBulk.EXPECT().Close().Return(nil)
	mockDb.EXPECT().BeginSyncPeriod(gomock.Any())
	mockDb.EXPECT().EndSyncPeriod()
	mockDb.EXPECT().Exist(gomock.Any()).Return(false)
	mockAida.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Next().Return(false)
	mockBulk.EXPECT().CreateAccount(gomock.Any())
	mockBulk.EXPECT().SetBalance(gomock.Any(), gomock.Any())
	mockBulk.EXPECT().SetNonce(gomock.Any(), gomock.Any())
	mockBulk.EXPECT().SetCode(gomock.Any(), gomock.Any())
	mockIter.EXPECT().Release()
	blockNum := uint64(12345)
	key := db.UpdateDBKey(blockNum)
	value, _ := rlp.EncodeToBytes(updateset.UpdateSetRLP{
		WorldState: updateset.UpdateSet{
			WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
			Block:           0,
			DeletedAccounts: []types.Address{},
		}.ToWorldStateRLP(),
		DeletedAccounts: []types.Address{},
	})
	mockIter.EXPECT().Key().Return(key).AnyTimes()
	mockIter.EXPECT().Value().Return(value).AnyTimes()
	mockIter.EXPECT().Release().Times(2)
	err := p.Prime()
	assert.NoError(t, err)

	// Edge case: nil adapter
	p.aidadb = mockAida
	mockAida.EXPECT().GetBackend().Return(nil)
	err = p.Prime()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend is nil")

	// Edge case: error from BeginBlock
	p.aidadb = mockAida
	mockAida.EXPECT().GetBackend().Return(mockAdapter)
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockDb.EXPECT().BeginBlock(gomock.Any()).Return(fmt.Errorf("begin block error"))
	err = p.Prime()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin block error")

	// Edge case: error from StartBulkLoad
	mockDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockDb.EXPECT().StartBulkLoad(gomock.Any()).Return(nil, fmt.Errorf("bulk load error"))
	err = p.Prime()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bulk load error")

	// Edge case: error from EndTransaction
	mockDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil)
	mockBulk.EXPECT().Close()
	mockDb.EXPECT().EndTransaction().Return(fmt.Errorf("end tx error"))
	err = p.Prime()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end tx error")
}

//go:generate mockgen -source=prime_test.go -destination=prime_mock.go -package=prime
type ProxyIterator interface {
	iterator.Iterator
}

func TestPrimer_SetFirstPrimableBlock(t *testing.T) {
	tests := []struct {
		name              string
		isExistingStateDb bool
		stateDbSrc        string
		substateBlock     uint64
		updateSetBlock    uint64
		getFirstKeyErr    error
		expectBlock       uint64
	}{
		{
			name:              "ExistingStateDb",
			isExistingStateDb: true,
			stateDbSrc:        t.TempDir(),
			expectBlock:       5, // adjust as needed for your logic
		},
		{
			name:              "NonExistingStateDb_SubstateNil",
			isExistingStateDb: false,
			substateBlock:     0,
			expectBlock:       0,
		},
		{
			name:              "NonExistingStateDb_UpdateSetError",
			isExistingStateDb: false,
			substateBlock:     100,
			getFirstKeyErr:    fmt.Errorf("not found"),
			expectBlock:       100,
		},
		{
			name:              "NonExistingStateDb_UpdateSetLower",
			isExistingStateDb: false,
			substateBlock:     100,
			updateSetBlock:    50,
			expectBlock:       50,
		},
		{
			name:              "NonExistingStateDb_SubstateLower",
			isExistingStateDb: false,
			substateBlock:     10,
			updateSetBlock:    50,
			expectBlock:       10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := &utils.Config{
				IsExistingStateDb: tc.isExistingStateDb,
				StateDbSrc:        tc.stateDbSrc,
			}
			mockStateDb := state.NewMockStateDB(ctrl)
			mockUpdateDb := db.NewMockUpdateDB(ctrl)
			mockSubstateDb := db.NewMockSubstateDB(ctrl)
			log := logger.NewLogger("Info", "Test")

			p := &primer{
				cfg: cfg,
				log: log,
				ctx: NewPrimeContext(cfg, mockStateDb, log),
				udb: mockUpdateDb,
				sdb: mockSubstateDb,
			}

			if tc.isExistingStateDb {
				_ = utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, tc.expectBlock-1, common.Hash{}, true)
			}

			if !tc.isExistingStateDb {
				if tc.substateBlock == 0 {
					mockSubstateDb.EXPECT().GetFirstSubstate().Return(nil)
				} else {
					mockSubstate := &substate.Substate{Block: tc.substateBlock}
					mockSubstateDb.EXPECT().GetFirstSubstate().Return(mockSubstate)
				}
				if tc.getFirstKeyErr != nil {
					mockUpdateDb.EXPECT().GetFirstKey().Return(uint64(0), tc.getFirstKeyErr)
				} else if tc.updateSetBlock != 0 {
					mockUpdateDb.EXPECT().GetFirstKey().Return(tc.updateSetBlock, nil)
				}
			}
			p.trySetFirstPrimableBlock()
			assert.Equal(t, tc.expectBlock, p.block)
		})
	}
}
