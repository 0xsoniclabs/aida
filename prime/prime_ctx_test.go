package prime

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
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

	// case success
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: utils.OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
	}
	mockBulkLoad.EXPECT().Close().Return(nil)
	mockStateDb.EXPECT().StartBulkLoad(uint64(1)).Return(mockBulkLoad, nil)
	err := prime.mayApplyBulkLoad()
	assert.NoError(t, err)

	// case success
	prime.operations = 0
	err = prime.mayApplyBulkLoad()
	assert.Nil(t, err)

	// case error on close
	prime = &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: utils.OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
	}
	mockBulkLoad.EXPECT().Close().Return(mockErr)
	err = prime.mayApplyBulkLoad()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), err.Error())

	// case error on start bulk load
	prime = &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: utils.OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
	}
	mockBulkLoad.EXPECT().Close().Return(nil)
	mockStateDb.EXPECT().StartBulkLoad(uint64(1)).Return(nil, mockErr)
	err = prime.mayApplyBulkLoad()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), err.Error())
}

func TestPrimeContext_PrimeStateDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulkLoad := state.NewMockBulkLoad(ctrl)

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
			prime := &PrimeContext{
				cfg: &utils.Config{
					PrimeRandom:       tc.primRandom,
					IsExistingStateDb: tc.useSrcDb,
				},
				load:       mockBulkLoad,
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
			mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulkLoad, nil).AnyTimes()
			mockBulkLoad.EXPECT().CreateAccount(gomock.Any()).Return().AnyTimes()
			mockBulkLoad.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulkLoad.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulkLoad.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulkLoad.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
			mockBulkLoad.EXPECT().Close().Return(nil).AnyTimes()

			err := prime.PrimeStateDB(ws)
			assert.NoError(t, err)
		})
	}
}

func TestPrimeContext_PrimeStateDB_EmptyWorldState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	prime := &PrimeContext{
		cfg:        &utils.Config{},
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		exist:      make(map[common.Address]bool),
	}
	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulkLoad, nil).AnyTimes()
	mockBulkLoad.EXPECT().Close().Return(nil).AnyTimes()
	emptyWs := txcontext.NewWorldState(map[common.Address]txcontext.Account{})
	err := prime.PrimeStateDB(emptyWs)
	assert.NoError(t, err)
}

// make sure that the stateDb contains data from both the first and the second priming
func TestPrime_ContinuousPrimingFromExistingDb(t *testing.T) {
	log := logger.NewLogger("Warning", "TestPrimeStateDB")
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(
			fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
				cfg := utils.MakeTestConfig(tc)

				// Initialization of state DB
				stateDb, stateDir, err := utils.PrepareStateDB(cfg)
				aidaDir := stateDir + "/aida"
				require.NoError(t, err, "failed to create state DB")

				// Generating randomized world state
				alloc, _ := utils.MakeWorldState(t)
				ws := txcontext.NewWorldState(alloc)

				aidaDb, err := db.NewDefaultBaseDB(aidaDir)
				require.NoError(t, err)
				p := NewPrimer(cfg, stateDb, aidaDb, log)
				err = p.ctx.PrimeStateDB(ws)
				require.NoError(t, err, "failed to prime state DB")

				err = state.BeginCarmenDbTestContext(stateDb)
				require.NoError(t, err)

				// Checks if state DB was primed correctly
				ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
					assert.Equal(t, 0, stateDb.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; Is: %v; Should be: %v", stateDb.GetBalance(addr), acc.GetBalance())
					assert.Equal(t, acc.GetNonce(), stateDb.GetNonce(addr), "failed to prime account nonce; Is: %v; Should be: %v", stateDb.GetNonce(addr), acc.GetNonce())
					assert.Equal(t, 0, bytes.Compare(stateDb.GetCode(addr), acc.GetCode()), "failed to prime account code; Is: %v; Should be: %v", stateDb.GetCode(addr), acc.GetCode())
					acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
						assert.Equal(t, valueHash, stateDb.GetState(addr, keyHash), "failed to prime account storage; Is: %v; Should be: %v", stateDb.GetState(addr, keyHash), valueHash)
					})
				})

				rootHash, err := stateDb.GetHash()
				require.NoError(t, err, "failed to get root hash")
				// Closing of state DB

				err = state.CloseCarmenDbTestContext(stateDb)
				require.NoError(t, err, "failed to close state DB")

				cfg.StateDbSrc = stateDir
				// Call for json creation and writing into it
				// reserve one block for validation
				err = utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, uint64(10), rootHash, true)
				require.NoError(t, err, "failed to write into DB info json file")

				// Initialization of state DB
				stateDb2, stateDir2, err := utils.PrepareStateDB(cfg)
				aidaDir2 := stateDir2 + "/aida"
				require.NoError(t, err, "failed to create state DB2")

				// Use next block to validate state DB content
				err = stateDb2.BeginBlock(uint64(10))
				require.NoError(t, err, "cannot begin block")

				err = stateDb2.BeginTransaction(uint32(0))
				require.NoError(t, err, "cannot begin transaction")

				// Checks if state DB was primed correctly
				ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
					assert.Equal(t, 0, stateDb2.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; Is: %v; Should be: %v", stateDb2.GetBalance(addr), acc.GetBalance())
					assert.Equal(t, acc.GetNonce(), stateDb2.GetNonce(addr), "failed to prime account nonce; Is: %v; Should be: %v", stateDb2.GetNonce(addr), acc.GetNonce())
					assert.Equal(t, 0, bytes.Compare(stateDb2.GetCode(addr), acc.GetCode()), "failed to prime account code; Is: %v; Should be: %v", stateDb2.GetCode(addr), acc.GetCode())
					acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
						assert.Equal(t, valueHash, stateDb2.GetState(addr, keyHash), "failed to prime account storage; Is: %v; Should be: %v", stateDb2.GetState(addr, keyHash), valueHash)
					})
				})

				err = stateDb2.EndTransaction()
				require.NoError(t, err, "cannot end transaction")

				err = stateDb2.EndBlock()
				require.NoError(t, err, "cannot end block sDB2")

				// Generating randomized world state
				alloc2, _ := utils.MakeWorldState(t)
				ws2 := txcontext.NewWorldState(alloc2)

				cfg.IsExistingStateDb = true
				aidaDb2, err := db.NewDefaultBaseDB(aidaDir2)
				require.NoError(t, err)
				p2 := NewPrimer(cfg, stateDb2, aidaDb2, log)
				// Priming state DB using p2.Prime
				err = p2.ctx.PrimeStateDB(ws2)
				require.NoError(t, err, "failed to prime state DB2")

				err = stateDb2.BeginBlock(uint64(20))
				require.NoError(t, err, "cannot begin block")

				err = stateDb2.BeginTransaction(uint32(0))
				require.NoError(t, err, "cannot begin transaction")

				// Checks if state DB was primed correctly
				ws2.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
					assert.Equal(t, 0, stateDb2.GetBalance(addr).Cmp(acc.GetBalance()), "failed to prime account balance; Is: %v; Should be: %v", stateDb2.GetBalance(addr), acc.GetBalance())
					assert.Equal(t, acc.GetNonce(), stateDb2.GetNonce(addr), "failed to prime account nonce; Is: %v; Should be: %v", stateDb2.GetNonce(addr), acc.GetNonce())
					assert.Equal(t, 0, bytes.Compare(stateDb2.GetCode(addr), acc.GetCode()), "failed to prime account code; Is: %v; Should be: %v", stateDb2.GetCode(addr), acc.GetCode())
					acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
						assert.Equal(t, valueHash, stateDb2.GetState(addr, keyHash), "failed to prime account storage; Is: %v; Should be: %v", stateDb2.GetState(addr, keyHash), valueHash)
					})
				})

				err = stateDb2.EndTransaction()
				require.NoError(t, err, "cannot end transaction")

				err = stateDb2.EndBlock()
				require.NoError(t, err, "cannot end block sDB2")

				err = state.CloseCarmenDbTestContext(stateDb2)
				require.NoError(t, err, "failed to close state DB 2")

				os.RemoveAll(stateDir)
				os.RemoveAll(stateDir2)

			})

	}
}

func TestPrimeContext_loadExistingAccountsIntoCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	acc := makeTestAccount(t)
	mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().Exist(gomock.Any()).Return(true)
	mockStateDb.EXPECT().EndTransaction().Return(nil)
	mockStateDb.EXPECT().EndBlock().Return(nil)
	err := prime.loadExistingAccountsIntoCache(txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc,
	}))
	assert.NoError(t, err)
}

func TestPrimeContext_primeOneAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	acc := makeTestAccount(t)
	mockBulkLoad.EXPECT().CreateAccount(gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	err := prime.primeOneAccount(common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"), acc, utils.NewProgressTracker(0, logger.NewLogger("ERROR", "Test")))
	assert.Nil(t, err)
}

func TestPrimeContext_PrimeStateDBRandom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	acc1 := makeTestAccount(t)
	acc2 := makeTestAccount(t)
	mockWs := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc1,
		common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): acc2,
	})
	prime := &PrimeContext{
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
	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulkLoad, nil).AnyTimes()
	mockBulkLoad.EXPECT().CreateAccount(gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().Close().Return(nil).AnyTimes()
	err := prime.PrimeStateDBRandom(mockWs, utils.NewProgressTracker(0, logger.NewLogger("ERROR", "Test")))
	assert.NoError(t, err)
}

func TestPrimeContext_SelfDestructAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {

		mockStateDb := state.NewMockStateDB(ctrl)
		mockLogger := logger.NewMockLogger(ctrl)
		prime := &PrimeContext{
			cfg:        nil,
			load:       nil,
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
		prime.SelfDestructAccounts([]substatetypes.Address{
			substatetypes.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
			substatetypes.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
		})
		assert.Equal(t, uint64(1), prime.block)
	})

	t.Run("error", func(t *testing.T) {

		mockStateDb := state.NewMockStateDB(ctrl)
		mockLogger := logger.NewMockLogger(ctrl)
		prime := &PrimeContext{
			cfg:        nil,
			load:       nil,
			db:         mockStateDb,
			operations: 0,
			log:        mockLogger,
			block:      0,
			exist: map[common.Address]bool{
				common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): true,
				common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): true,
			},
		}
		mockError := errors.New("mock error")
		mockStateDb.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
		mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(mockError)
		mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(mockError)
		mockStateDb.EXPECT().EndTransaction().Return(mockError)
		mockStateDb.EXPECT().EndBlock().Return(mockError)
		mockStateDb.EXPECT().EndSyncPeriod().Return()
		mockStateDb.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
		mockStateDb.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(99)).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
		prime.SelfDestructAccounts([]substatetypes.Address{
			substatetypes.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
			substatetypes.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
		})
		assert.Equal(t, uint64(1), prime.block)
	})

}

func TestPrimeContext_GetBlock(t *testing.T) {
	target := uint64(5)
	prime := &PrimeContext{
		log:   logger.NewLogger("ERROR", "Test"),
		block: 0,
	}
	block := prime.GetBlock()
	assert.Equal(t, uint64(0), block)
	prime.SetBlock(target)
	block = prime.GetBlock()
	assert.Equal(t, target, block)
}

func TestPrimeContext_HasPrimedIsUpdatedAfterPrimeStateDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	acc := makeTestAccount(t)
	ws := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc,
	})
	prime := &PrimeContext{
		cfg:        &utils.Config{},
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: utils.OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
		exist:      make(map[common.Address]bool),
	}

	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulkLoad, nil).AnyTimes()
	mockBulkLoad.EXPECT().CreateAccount(gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().Close().Return(nil).AnyTimes()

	err := prime.PrimeStateDB(ws)
	assert.NoError(t, err)
	assert.True(t, prime.HasPrimed())
}
