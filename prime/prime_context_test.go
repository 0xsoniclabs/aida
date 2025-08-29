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

package prime

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func makeTestAccount(t *testing.T) txcontext.Account {
	return txcontext.NewAccount(
		utils.MakeRandomByteSlice(t, 2048),
		utils.MakeAccountStorage(t),
		big.NewInt(int64(utils.GetRandom(1, 1000*5000))),
		uint64(utils.GetRandom(1, 1000*5000)),
	)
}

func TestPrimeContext_mayApplyBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	t.Run("success", func(t *testing.T) {
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockStateDb := state.NewMockStateDB(ctrl)
		p := &context{
			load:       mockBulk,
			db:         mockStateDb,
			operations: utils.OperationThreshold + 1,
			log:        logger.NewLogger("ERROR", "Test"),
		}
		mockBulk.EXPECT().Close().Return(nil)
		mockStateDb.EXPECT().StartBulkLoad(uint64(1)).Return(mockBulk, nil)
		err := p.mayApplyBulkLoad()
		assert.NoError(t, err)
	})

	t.Run("success_no_bulk", func(t *testing.T) {
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockStateDb := state.NewMockStateDB(ctrl)
		p := &context{
			load:       mockBulk,
			db:         mockStateDb,
			operations: 0,
			log:        logger.NewLogger("ERROR", "Test"),
		}
		err := p.mayApplyBulkLoad()
		assert.Nil(t, err)
	})

	t.Run("error_on_close", func(t *testing.T) {
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockStateDb := state.NewMockStateDB(ctrl)
		p := &context{
			load:       mockBulk,
			db:         mockStateDb,
			operations: utils.OperationThreshold + 1,
			log:        logger.NewLogger("ERROR", "Test"),
		}
		mockBulk.EXPECT().Close().Return(mockErr)
		err := p.mayApplyBulkLoad()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), mockErr.Error())
	})

	t.Run("error_on_start_bulk_load", func(t *testing.T) {
		mockBulk := state.NewMockBulkLoad(ctrl)
		mockStateDb := state.NewMockStateDB(ctrl)
		p := &context{
			load:       mockBulk,
			db:         mockStateDb,
			operations: utils.OperationThreshold + 1,
			log:        logger.NewLogger("ERROR", "Test"),
		}
		mockBulk.EXPECT().Close().Return(nil)
		mockStateDb.EXPECT().StartBulkLoad(uint64(1)).Return(nil, mockErr)
		err := p.mayApplyBulkLoad()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), mockErr.Error())
	})
}

func TestPrimeContext_PrimeStateDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulk := state.NewMockBulkLoad(ctrl)

	testcases := []struct {
		name       string
		useSrcDb   bool
		primRandom bool
	}{
		{name: "RunFromFreshDB_PrimeRandom", useSrcDb: false, primRandom: true},
		{name: "RunFromExistingDB_PrimeRandom", useSrcDb: true, primRandom: true},
		{name: "RunFromFreshDB_PrimeInOrder", useSrcDb: false, primRandom: false},
		{name: "RunFromExistingDB_PrimeInOrder", useSrcDb: true, primRandom: false},
	}

	acc := makeTestAccount(t)
	ws := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc,
	})

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup the PrimeContext
			p := &context{
				cfg: &utils.Config{
					PrimeRandom:       tc.primRandom,
					IsExistingStateDb: tc.useSrcDb,
				},
				load:       mockBulk,
				db:         mockStateDb,
				operations: 0,
				log:        logger.NewLogger("ERROR", "Test"),
				block:      0,
				exist:      map[common.Address]bool{},
			}

			if tc.useSrcDb {
				mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
				mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
				mockStateDb.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
				mockStateDb.EXPECT().EndTransaction().Return(nil).AnyTimes()
				mockStateDb.EXPECT().EndBlock().Return(nil).AnyTimes()
			}
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil).AnyTimes()
			mockBulk.EXPECT().CreateAccount(gomock.Any()).Return().AnyTimes()
			mockBulk.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulk.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulk.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulk.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulk.EXPECT().Close().Return(nil).AnyTimes()

			err := p.PrimeStateDB(ws)
			assert.NoError(t, err)
		})
	}
}

func TestPrimeContext_PrimeStateDB_EmptyWorldState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulk := state.NewMockBulkLoad(ctrl)
	p := &context{
		cfg:        &utils.Config{},
		load:       mockBulk,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		exist:      make(map[common.Address]bool),
	}
	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil).AnyTimes()
	mockBulk.EXPECT().Close().Return(nil).AnyTimes()
	emptyWs := txcontext.NewWorldState(map[common.Address]txcontext.Account{})
	err := p.PrimeStateDB(emptyWs)
	assert.NoError(t, err)
}

func TestPrimeContext_loadExistingAccountsIntoCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulk := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	p := &context{
		cfg:        nil,
		load:       mockBulk,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	acc := makeTestAccount(t)

	gomock.InOrder(
		mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil),
		mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil),
		mockStateDb.EXPECT().Exist(gomock.Any()).Return(true),
		mockStateDb.EXPECT().EndTransaction().Return(nil),
		mockStateDb.EXPECT().EndBlock().Return(nil),
	)
	err := p.loadExistingAccountsIntoCache(txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc,
	}))
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), p.block)
}

func TestPrimeContext_primeOneAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulk := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	p := &context{
		cfg:        nil,
		load:       mockBulk,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	acc := makeTestAccount(t)
	mockBulk.EXPECT().CreateAccount(gomock.Any()).Return()
	mockBulk.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return()
	mockBulk.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return()
	mockBulk.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return()
	mockBulk.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	err := p.primeOneAccount(common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"), acc, utils.NewProgressTracker(0, logger.NewLogger("ERROR", "Test")))
	assert.Nil(t, err)
}

func TestPrimeContext_PrimeStateDBRandom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulk := state.NewMockBulkLoad(ctrl)
	acc1 := makeTestAccount(t)
	acc2 := makeTestAccount(t)
	mockWs := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc1,
		common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): acc2,
	})
	p := &context{
		cfg: &utils.Config{
			RandomSeed: 0,
		},
		load:       nil,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist: map[common.Address]bool{
			common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): true,
			common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): true,
		},
	}
	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil).AnyTimes()
	mockBulk.EXPECT().CreateAccount(gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().Close().Return(nil).AnyTimes()
	err := p.primeStateDBRandom(mockWs, utils.NewProgressTracker(0, logger.NewLogger("ERROR", "Test")))
	assert.NoError(t, err)
}

func TestPrimeContext_SelfDestructAccountsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	p := &context{
		db:         mockStateDb,
		operations: 0,
		log:        mockLogger,
		block:      0,
		exist: map[common.Address]bool{
			common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): true,
			common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): true,
		},
	}
	mockStateDb.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
	mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().EndTransaction().Return(nil)
	mockStateDb.EXPECT().EndBlock().Return(nil)
	mockStateDb.EXPECT().EndSyncPeriod().Return()
	mockStateDb.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
	mockStateDb.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(99)).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	err := p.selfDestructAccounts([]substatetypes.Address{
		substatetypes.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		substatetypes.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
	})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), p.block)
}

func TestPrimeContext_SelfDestructAccountsReturnsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockError := errors.New("mock error")

	testcases := []struct {
		name      string
		mockSetup func(*state.MockStateDB)
	}{
		{
			name: "BeginBlockError",
			mockSetup: func(mockStateDb *state.MockStateDB) {
				mockStateDb.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
				mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(mockError)
			},
		},
		{
			name: "BeginTransactionError",
			mockSetup: func(mockStateDb *state.MockStateDB) {
				mockStateDb.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
				mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
				mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(mockError)
			},
		},
		{
			name: "EndTransactionError",
			mockSetup: func(mockStateDb *state.MockStateDB) {
				mockStateDb.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
				mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
				mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
				mockStateDb.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
				mockStateDb.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(99)).AnyTimes()
				mockStateDb.EXPECT().EndTransaction().Return(mockError)
			},
		},
		{
			name: "EndBlockError",
			mockSetup: func(mockStateDb *state.MockStateDB) {
				mockStateDb.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
				mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
				mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
				mockStateDb.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
				mockStateDb.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(99)).AnyTimes()
				mockStateDb.EXPECT().EndTransaction().Return(nil)
				mockStateDb.EXPECT().EndBlock().Return(mockError)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockStateDb := state.NewMockStateDB(ctrl)
			mockLogger := logger.NewMockLogger(ctrl)
			p := &context{
				db:         mockStateDb,
				operations: 0,
				log:        mockLogger,
				block:      0,
				exist: map[common.Address]bool{
					common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): true,
					common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): true,
				},
			}

			mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			tc.mockSetup(mockStateDb)
			err := p.selfDestructAccounts([]substatetypes.Address{
				substatetypes.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
				substatetypes.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
			})
			assert.Error(t, err, "SelfDestructAccounts expected error but got none")
			assert.ErrorContains(t, err, "mock error")
		})

	}

}

func TestPrimeContext_GetBlock(t *testing.T) {
	target := uint64(5)
	p := &context{
		log:   logger.NewLogger("ERROR", "Test"),
		block: 0,
	}
	block := p.GetBlock()
	assert.Equal(t, uint64(0), block)
	p.SetBlock(target)
	block = p.GetBlock()
	assert.Equal(t, target, block)
}

func TestPrimeContext_HasPrimedIsUpdatedAfterPrimeStateDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockBulk := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	acc := makeTestAccount(t)
	ws := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc,
	})
	p := &context{
		cfg:        &utils.Config{},
		load:       mockBulk,
		db:         mockStateDb,
		operations: utils.OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
		exist:      make(map[common.Address]bool),
	}

	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil).AnyTimes()
	mockBulk.EXPECT().CreateAccount(gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulk.EXPECT().Close().Return(nil).AnyTimes()

	err := p.PrimeStateDB(ws)
	assert.NoError(t, err)
	assert.True(t, p.HasPrimed())
}

func TestPrimeContext_PrimeStateDB_RealData(t *testing.T) {
	log := logger.NewLogger("Warning", "TestPrimeStateDB")
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)
			cfg.DbTmp = t.TempDir()

			// Initialization of state DB
			sDB, _, err := utils.PrepareStateDB(cfg)

			require.NoError(t, err, "failed to create state DB")

			// Closing of state DB
			defer func(sDB state.StateDB) {
				err = state.CloseCarmenDbTestContext(sDB)
				require.NoError(t, err, "cannot close carmen test context")
			}(sDB)

			// Generating randomized world state
			ws, _ := utils.MakeWorldState(t)

			pc := newContext(cfg, sDB, log)
			// Priming state DB
			err = pc.PrimeStateDB(ws)
			require.NoError(t, err)

			err = sDB.BeginBlock(uint64(2))
			require.NoError(t, err, "cannot begin block")
			err = sDB.BeginTransaction(uint32(0))
			require.NoError(t, err, "cannot begin transaction")

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				assert.Equal(t, 0, sDB.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; have: %v; want: %v", sDB.GetBalance(addr), acc.GetBalance())
				assert.Equal(t, sDB.GetNonce(addr), acc.GetNonce(), "failed to prime account nonce; have: %v; want: %v", sDB.GetNonce(addr), acc.GetNonce())
				assert.Equal(t, 0, bytes.Compare(sDB.GetCode(addr), acc.GetCode()), "failed to prime account code; have: %v; want: %v", sDB.GetCode(addr), acc.GetCode())

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					assert.Equal(t, sDB.GetState(addr, keyHash), valueHash, "failed to prime account storage; have: %v; want: %v", sDB.GetState(addr, keyHash), valueHash)
				})
			})

		})
	}
}

// make sure that the stateDb contains data from both the first and the second priming
func TestPrimeContext_PrimeStateDB_ContinuousPrimingFromSrcDB(t *testing.T) {
	log := logger.NewLogger("Warning", "TestPrimeStateDB")
	srcDbBlock := uint64(8)
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)
			cfg.DbTmp = t.TempDir()

			// Initialization of state DB
			sDB, sDbDir, err := utils.PrepareStateDB(cfg)

			require.NoError(t, err, "failed to create state DB")

			// Generating randomized world state
			alloc, _ := utils.MakeWorldState(t)
			ws := txcontext.NewWorldState(alloc)

			pc := newContext(cfg, sDB, log)
			// Priming state DB
			err = pc.PrimeStateDB(ws)
			require.NoError(t, err, "failed to prime state DB")

			err = state.BeginCarmenDbTestContext(sDB)
			require.NoError(t, err, "failed to begin carmen db test context")

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				assert.Equal(t, 0, sDB.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; have: %v; want: %v", sDB.GetBalance(addr), acc.GetBalance())
				assert.Equal(t, sDB.GetNonce(addr), acc.GetNonce(), "failed to prime account nonce; have: %v; want: %v", sDB.GetNonce(addr), acc.GetNonce())
				assert.Equal(t, 0, bytes.Compare(sDB.GetCode(addr), acc.GetCode()), "failed to prime account code; have: %v; want: %v", sDB.GetCode(addr), acc.GetCode())

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					assert.Equal(t, sDB.GetState(addr, keyHash), valueHash, "failed to prime account storage; have: %v; want: %v", sDB.GetState(addr, keyHash), valueHash)
				})
			})

			rootHash, err := sDB.GetHash()
			require.NoError(t, err, "failed to get root hash")
			// Closing of state DB

			err = state.CloseCarmenDbTestContext(sDB)
			require.NoError(t, err, "failed to close state DB")

			cfg.StateDbSrc = sDbDir
			// Call for json creation and writing into it
			err = utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, srcDbBlock, rootHash, true)
			require.NoError(t, err, "failed to write into DB info json file")
			cfg.IsExistingStateDb = true

			// Initialization of state DB
			sDB2, _, err := utils.PrepareStateDB(cfg)
			require.NoError(t, err, "failed to create state DB2")

			defer func() {
				err = state.CloseCarmenDbTestContext(sDB2)
				require.NoError(t, err, "failed to close state DB2")
			}()

			err = sDB2.BeginBlock(srcDbBlock - 1)
			require.NoError(t, err, "cannot begin block")

			err = sDB2.BeginTransaction(uint32(0))
			require.NoError(t, err, "cannot begin transaction")

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				assert.Equal(t, 0, sDB2.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; have: %v; want: %v", sDB2.GetBalance(addr), acc.GetBalance())
				assert.Equal(t, sDB2.GetNonce(addr), acc.GetNonce(), "failed to prime account nonce; have: %v; want: %v", sDB2.GetNonce(addr), acc.GetNonce())
				assert.Equal(t, 0, bytes.Compare(sDB2.GetCode(addr), acc.GetCode()), "failed to prime account code; have: %v; want: %v", sDB2.GetCode(addr), acc.GetCode())

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					assert.Equal(t, sDB2.GetState(addr, keyHash), valueHash, "failed to prime account storage; have: %v; want: %v", sDB2.GetState(addr, keyHash), valueHash)
				})
			})

			err = sDB2.EndTransaction()
			require.NoError(t, err, "cannot end transaction")

			err = sDB2.EndBlock()
			require.NoError(t, err, "cannot end block sDB 2")

			// Generating randomized world state
			alloc2, _ := utils.MakeWorldState(t)
			ws2 := txcontext.NewWorldState(alloc2)

			pc2 := newContext(cfg, sDB2, log)
			pc2.block = srcDbBlock
			// Priming state DB
			err = pc2.PrimeStateDB(ws2)
			require.NoError(t, err, "failed to prime state DB2")

			err = sDB2.BeginBlock(srcDbBlock + 2)
			require.NoError(t, err, "cannot begin block")

			err = sDB2.BeginTransaction(uint32(0))
			require.NoError(t, err, "cannot begin transaction")

			// Checks if state DB was primed correctly
			ws2.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				assert.Equal(t, 0, sDB2.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; have: %v; want: %v", sDB2.GetBalance(addr), acc.GetBalance())
				assert.Equal(t, sDB2.GetNonce(addr), acc.GetNonce(), "failed to prime account nonce; have: %v; want: %v", sDB2.GetNonce(addr), acc.GetNonce())
				assert.Equal(t, 0, bytes.Compare(sDB2.GetCode(addr), acc.GetCode()), "failed to prime account code; have: %v; want: %v", sDB2.GetCode(addr), acc.GetCode())

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					assert.Equal(t, sDB2.GetState(addr, keyHash), valueHash, "failed to prime account storage; have: %v; want: %v", sDB2.GetState(addr, keyHash), valueHash)
				})
			})
		})
	}
}
