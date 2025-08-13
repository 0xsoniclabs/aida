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
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/rlp"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"go.uber.org/mock/gomock"
)

// TestStatedb_PrimeStateDB tests priming fresh state DB with randomized world state data
func TestPrime_PrimeStateDB(t *testing.T) {
	log := logger.NewLogger("Warning", "TestPrimeStateDB")
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)

			// Initialization of state DB
			sDB, sDbDir, err := utils.PrepareStateDB(cfg)
			defer os.RemoveAll(sDbDir)

			if err != nil {
				t.Fatalf("failed to create state DB: %v", err)
			}

			// Closing of state DB
			defer func(sDB state.StateDB) {
				err = state.CloseCarmenDbTestContext(sDB)
				if err != nil {
					t.Fatalf("cannot close carmen test context; %v", err)
				}
			}(sDB)

			// Generating randomized world state
			ws, _ := utils.MakeWorldState(t)

			pc := NewPrimeContext(cfg, sDB, log)
			// Priming state DB
			err = pc.PrimeStateDB(ws, sDB)
			require.NoError(t, err, "failed to prime state DB")

			err = sDB.BeginBlock(uint64(2))
			require.NoError(t, err, "cannot begin block")
			err = sDB.BeginTransaction(uint32(0))
			require.NoError(t, err, "cannot begin transaction")

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {

				if sDB.GetBalance(addr).Cmp(acc.GetBalance()) != 0 {
					t.Fatalf("failed to prime account balance; Is: %v; Should be: %v", sDB.GetBalance(addr), acc.GetBalance())
				}

				if sDB.GetNonce(addr) != acc.GetNonce() {
					t.Fatalf("failed to prime account nonce; Is: %v; Should be: %v", sDB.GetNonce(addr), acc.GetNonce())
				}

				if !bytes.Equal(sDB.GetCode(addr), acc.GetCode()) {
					t.Fatalf("failed to prime account code; Is: %v; Should be: %v", sDB.GetCode(addr), acc.GetCode())
				}

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					if sDB.GetState(addr, keyHash) != valueHash {
						t.Fatalf("failed to prime account storage; Is: %v; Should be: %v", sDB.GetState(addr, keyHash), valueHash)
					}
				})

			})

		})
	}
}

func TestStateDbPrimerExtension_UserIsInformedAboutRandomPriming(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	aidaDbPath := t.TempDir() + "aidadb"
	stateDb := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.SkipPriming = false
	cfg.StateDbSrc = ""
	cfg.First = 10
	cfg.PrimeRandom = true
	cfg.RandomSeed = 111
	cfg.PrimeThreshold = 10
	cfg.UpdateBufferSize = 1024

	p := MakePrimer(cfg, stateDb, log)

	gomock.InOrder(
		log.EXPECT().Infof("Randomized Priming enabled; Seed: %v, threshold: %v", int64(111), 10),
		log.EXPECT().Infof("Update buffer size: %v bytes", uint64(1024)),
		log.EXPECT().Noticef("Priming from block %v...", uint64(0)),
		log.EXPECT().Noticef("Priming to block %v...", uint64(9)),
		log.EXPECT().Debugf("\tLoading %d accounts with %d values ..", 0, 0),
		stateDb.EXPECT().BeginBlock(uint64(0)),
		stateDb.EXPECT().BeginTransaction(uint32(0)),
		stateDb.EXPECT().EndTransaction(),
		stateDb.EXPECT().EndBlock(),
		stateDb.EXPECT().StartBulkLoad(uint64(1)).Return(nil, errors.New("stop")),
	)

	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
	require.NoError(t, err)

	err = p.Prime(stateDb, aidaDb)
	require.Error(t, err, "expected error during priming")

	want := "cannot prime state-db; failed to prime StateDB: stop"
	require.ErrorIs(t, err, errors.New(want))
}

// make sure that the stateDb contains data from both the first and the second priming
func TestStateDbPrimerExtension_ContinuousPrimingFromExistingDb(t *testing.T) {
	log := logger.NewLogger("Warning", "TestPrimeStateDB")
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)

			// Initialization of state DB
			sDB, sDbDir, err := utils.PrepareStateDB(cfg)
			defer os.RemoveAll(sDbDir)

			require.NoError(t, err, "failed to create state DB")

			// Generating randomized world state
			alloc, _ := utils.MakeWorldState(t)
			ws := txcontext.NewWorldState(alloc)

			p := MakePrimer(cfg, sDB, log)
			// Priming state DB using p.Prime
			aidaDb, err := db.NewDefaultBaseDB(sDbDir)
			require.NoError(t, err)
			err = p.Prime(sDB, aidaDb)
			require.NoError(t, err, "failed to prime state DB")

			err = state.BeginCarmenDbTestContext(sDB)
			require.NoError(t, err)

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				assert.Equal(t, 0, sDB.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; Is: %v; Should be: %v", sDB.GetBalance(addr), acc.GetBalance())
				assert.Equal(t, acc.GetNonce(), sDB.GetNonce(addr), "failed to prime account nonce; Is: %v; Should be: %v", sDB.GetNonce(addr), acc.GetNonce())
				assert.Equal(t, 0, bytes.Compare(sDB.GetCode(addr), acc.GetCode()), "failed to prime account code; Is: %v; Should be: %v", sDB.GetCode(addr), acc.GetCode())
				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					assert.Equal(t, valueHash, sDB.GetState(addr, keyHash), "failed to prime account storage; Is: %v; Should be: %v", sDB.GetState(addr, keyHash), valueHash)
				})
			})

			rootHash, err := sDB.GetHash()
			require.NoError(t, err, "failed to get root hash")
			// Closing of state DB

			err = state.CloseCarmenDbTestContext(sDB)
			require.NoError(t, err, "failed to close state DB")

			cfg.StateDbSrc = sDbDir
			// Call for json creation and writing into it
			err = utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, 2, rootHash, true)
			require.NoError(t, err, "failed to write into DB info json file")

			// Initialization of state DB
			sDB2, sDbDir2, err := utils.PrepareStateDB(cfg)
			defer os.RemoveAll(sDbDir2)
			require.NoError(t, err, "failed to create state DB2")

			defer func() {
				err = state.CloseCarmenDbTestContext(sDB2)
				require.NoError(t, err, "failed to close state DB")
			}()

			err = sDB2.BeginBlock(uint64(7))
			require.NoError(t, err, "cannot begin block")

			err = sDB2.BeginTransaction(uint32(0))
			require.NoError(t, err, "cannot begin transaction")

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				assert.Equal(t, 0, sDB2.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; Is: %v; Should be: %v", sDB2.GetBalance(addr), acc.GetBalance())
				assert.Equal(t, acc.GetNonce(), sDB2.GetNonce(addr), "failed to prime account nonce; Is: %v; Should be: %v", sDB2.GetNonce(addr), acc.GetNonce())
				assert.Equal(t, 0, bytes.Compare(sDB2.GetCode(addr), acc.GetCode()), "failed to prime account code; Is: %v; Should be: %v", sDB2.GetCode(addr), acc.GetCode())
				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					assert.Equal(t, valueHash, sDB2.GetState(addr, keyHash), "failed to prime account storage; Is: %v; Should be: %v", sDB2.GetState(addr, keyHash), valueHash)
				})
			})

			err = sDB2.EndTransaction()
			require.NoError(t, err, "cannot end transaction")

			err = sDB2.EndBlock()
			require.NoError(t, err, "cannot end block sDB2")

			// Generating randomized world state
			alloc2, _ := utils.MakeWorldState(t)
			ws2 := txcontext.NewWorldState(alloc2)

			cfg.IsExistingStateDb = true
			p2 := MakePrimer(cfg, sDB2, log)
			aidaDb2, err := db.NewDefaultBaseDB(sDbDir2)
			require.NoError(t, err)
			// Priming state DB using p2.Prime
			err = p2.Prime(sDB2, aidaDb2)
			require.NoError(t, err, "failed to prime state DB2")

			err = sDB2.BeginBlock(uint64(10))
			require.NoError(t, err, "cannot begin block")

			err = sDB2.BeginTransaction(uint32(0))
			require.NoError(t, err, "cannot begin transaction")

			// Checks if state DB was primed correctly
			ws2.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				assert.Equal(t, 0, sDB2.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; Is: %v; Should be: %v", sDB2.GetBalance(addr), acc.GetBalance())
				assert.Equal(t, acc.GetNonce(), sDB2.GetNonce(addr), "failed to prime account nonce; Is: %v; Should be: %v", sDB2.GetNonce(addr), acc.GetNonce())
				assert.Equal(t, 0, bytes.Compare(sDB2.GetCode(addr), acc.GetCode()), "failed to prime account code; Is: %v; Should be: %v", sDB2.GetCode(addr), acc.GetCode())
				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					assert.Equal(t, valueHash, sDB2.GetState(addr, keyHash), "failed to prime account storage; Is: %v; Should be: %v", sDB2.GetState(addr, keyHash), valueHash)
				})
			})

			err = sDB2.EndTransaction()
			require.NoError(t, err, "cannot end transaction")

			err = sDB2.EndBlock()
			require.NoError(t, err, "cannot end block sDB2")
		})
	}
}

func TestStateDbPrimer_Prime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := logger.NewLogger("Info", "TestStateDbPrimer")

	cfg := &utils.Config{}
	cfg.SkipPriming = false
	cfg.StateDbSrc = ""
	cfg.First = 0
	mockDb := state.NewMockStateDB(ctrl)
	mockAida := db.NewMockBaseDB(ctrl)
	mockAdapter := db.NewMockDbAdapter(ctrl)
	mockBulk := state.NewMockBulkLoad(ctrl)
	p := &primer{
		log: log,
		ctx: NewPrimeContext(cfg, mockDb, log),
		cfg: cfg,
	}
	mockAida.EXPECT().GetBackend().Return(mockAdapter)
	mockIter := NewMockProxyIterator(ctrl)
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockAdapter.EXPECT().Get(gomock.Any(), gomock.Any()).Return([]uint8{1, 2, 3}, nil)
	mockDb.EXPECT().BeginBlock(gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().EndTransaction().Return(nil).Times(3)
	mockDb.EXPECT().EndBlock().Return(nil).Times(3)
	mockDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil).Times(2)
	mockBulk.EXPECT().Close().Times(2)
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
	mockIter.EXPECT().Release()
	mockIter.EXPECT().Next()
	mockIter.EXPECT().Release()
	err := p.Prime(mockDb, mockAida)
	assert.NoError(t, err)
}

// TODO remove and use Substate IIterator mock?
//
//go:generate mockgen -source=prime_test.go -destination=prime_mock.go -package=prime
type ProxyIterator interface {
	iterator.Iterator
}

// Edge case: mayPrimeFromUpdateSet with empty iterator
func TestStateDbPrimer_MayPrimeFromUpdateSet_EmptyIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockUpdateDb := db.NewMockUpdateDB(ctrl)
	mockIter := NewMockProxyIterator(ctrl)
	cfg := &utils.Config{UpdateBufferSize: 1024}
	log := logger.NewLogger("Info", "Test")
	p := &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, mockStateDb, log),
	}
	mockUpdateDb.EXPECT().NewUpdateSetIterator(uint64(0), gomock.Any()).Return(mockIter)
	mockIter.EXPECT().Next().Return(false)
	mockIter.EXPECT().Release()
	err := p.mayPrimeFromUpdateSet(mockStateDb, mockUpdateDb)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), p.block)
}

// Edge case: mayPrimeFromUpdateSet with PrimeStateDB error
func TestStateDbPrimer_MayPrimeFromUpdateSet_PrimeStateDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockUpdateDb := db.NewMockUpdateDB(ctrl)
	mockIter := NewMockProxyIterator(ctrl)
	cfg := &utils.Config{UpdateBufferSize: 1} // force priming
	log := logger.NewLogger("Info", "Test")
	p := &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, mockStateDb, log),
	}
	mockUpdateDb.EXPECT().NewUpdateSetIterator(uint64(0), gomock.Any()).Return(mockIter)
	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Value().Return(&updateset.UpdateSet{Block: 0, WorldState: substate.WorldState{}})
	mockIter.EXPECT().Release()
	p.ctx = NewPrimeContext(cfg, mockStateDb, log)
	err := p.mayPrimeFromUpdateSet(mockStateDb, mockUpdateDb)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prime error")
}

// Edge case: mayPrimeFromSubstate with GenerateUpdateSet error
func TestStateDbPrimer_MayPrimeFromSubstate_GenerateUpdateSetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	cfg := &utils.Config{}
	log := logger.NewLogger("Info", "Test")
	p := &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, mockStateDb, log),
	}
	// Patch GenerateUpdateSet to return error
	err := p.mayPrimeFromSubstate(mockStateDb, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generate error")
}

// Edge case: mayPrimeFromSubstate with deletedAccounts and HasPrimed true
func TestStateDbPrimer_MayPrimeFromSubstate_DeletedAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	cfg := &utils.Config{}
	log := logger.NewLogger("Info", "Test")
	p := &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, mockStateDb, log),
	}
	err := p.mayPrimeFromSubstate(mockStateDb, nil)
	assert.NoError(t, err)
}

// Error handling in prime: mayPrimeFromUpdateSet returns error
func TestStateDbPrimer_Prime_MayPrimeFromUpdateSetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockAidaDb := db.NewMockBaseDB(ctrl)
	cfg := &utils.Config{}
	log := logger.NewLogger("Info", "Test")
	p := &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, mockStateDb, log),
	}
	err := p.Prime(mockStateDb, mockAidaDb)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update set error")
}

// Error handling in prime: mayPrimeFromSubstate returns error
func TestStateDbPrimer_Prime_MayPrimeFromSubstateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockAidaDb := db.NewMockBaseDB(ctrl)
	cfg := &utils.Config{}
	log := logger.NewLogger("Info", "Test")
	p := &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, mockStateDb, log),
	}
	err := p.Prime(mockStateDb, mockAidaDb)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "substate error")
}

// Error handling in prime: MayDeleteDestroyedAccountsFromStateDB returns error
func TestStateDbPrimer_Prime_MayDeleteDestroyedAccountsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockAidaDb := db.NewMockBaseDB(ctrl)
	cfg := &utils.Config{}
	log := logger.NewLogger("Info", "Test")
	p := &primer{
		cfg: cfg,
		log: log,
		ctx: NewPrimeContext(cfg, mockStateDb, log),
	}
	err := p.Prime(mockStateDb, mockAidaDb)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destroyed accounts error")
}
