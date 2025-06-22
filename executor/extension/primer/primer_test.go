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

package primer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/0xsoniclabs/substate/types/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/mock/gomock"
)

func TestStateDbPrimerExtension_NoPrimerIsCreatedIfDisabled(t *testing.T) {
	cfg := &utils.Config{}
	cfg.SkipPriming = true

	ext := MakeStateDbPrimer[any](cfg)
	if _, ok := ext.(extension.NilExtension[any]); !ok {
		t.Errorf("Primer is enabled although not set in configuration")
	}

}

func TestStateDbPrimerExtension_PrimingExistingStateDbMissingDbInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	cfg := &utils.Config{}
	cfg.IsExistingStateDb = true
	cfg.First = 2

	ext := makeStateDbPrimer[any](cfg, log)

	expected := errors.New("cannot read state db info; failed to read statedb_info.json; open statedb_info.json: no such file or directory")

	err := ext.PreRun(executor.State[any]{}, nil)
	if err.Error() != expected.Error() {
		t.Errorf("Priming should fail if db info is missing; got: %v; expected: %v", err, expected)
	}
}

func TestStateDbPrimerExtension_PrimingDoesTriggerForNonExistingStateDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	stateDb := state.NewMockStateDB(ctrl)
	aidaDbPath := t.TempDir() + "aidadb"

	cfg := &utils.Config{}
	cfg.SkipPriming = false
	cfg.StateDbSrc = ""
	cfg.First = 2

	gomock.InOrder(
		log.EXPECT().Infof("Update buffer size: %v bytes", cfg.UpdateBufferSize),
		log.EXPECT().Noticef("Priming from block %v...", uint64(0)),
		log.EXPECT().Noticef("Priming to block %v...", cfg.First-1),
		log.EXPECT().Debugf("\tLoading %d accounts with %d values ..", 0, 0),
		stateDb.EXPECT().BeginBlock(uint64(0)),
		stateDb.EXPECT().BeginTransaction(uint32(0)),
		stateDb.EXPECT().EndTransaction(),
		stateDb.EXPECT().EndBlock(),
		stateDb.EXPECT().StartBulkLoad(uint64(1)).Return(nil, errors.New("stop")),
	)

	ext := makeStateDbPrimer[any](cfg, log)

	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
	if err != nil {
		t.Fatal(err)
	}

	err = ext.PreRun(executor.State[any]{}, &executor.Context{AidaDb: aidaDb, State: stateDb})
	if err == nil {
		t.Fatal("run must fail")
	}

	want := "cannot prime state-db; stop"

	if err.Error() != want {
		t.Fatalf("unexpected error\ngot: %v\nwant: %v", err, want)
	}
}

func TestStateDbPrimerExtension_AttemptToPrimeBlockZeroDoesNotFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	tmpStateDb := t.TempDir()

	cfg := &utils.Config{}
	cfg.SkipPriming = false
	cfg.StateDbSrc = tmpStateDb
	cfg.IsExistingStateDb = true
	cfg.First = 2

	err := utils.WriteStateDbInfo(tmpStateDb, cfg, 1, common.Hash{}, false)
	if err != nil {
		t.Fatalf("cannot write state db info: %v", err)
	}

	ext := makeStateDbPrimer[any](cfg, log)

	log.EXPECT().Debugf("skipping priming; first priming block %v; first block %v", uint64(1), uint64(2))

	err = ext.PreRun(executor.State[any]{}, &executor.Context{})
	if err != nil {
		t.Errorf("priming should not happen hence should not fail")
	}
}

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

			pc := utils.NewPrimeContext(cfg, sDB, 0, log)
			// Priming state DB
			err = pc.PrimeStateDB(ws, sDB)
			if err != nil {
				t.Fatal(err)
			}

			err = sDB.BeginBlock(uint64(2))
			if err != nil {
				t.Fatalf("cannot begin block; %v", err)
			}
			err = sDB.BeginTransaction(uint32(0))
			if err != nil {
				t.Fatalf("cannot begin transaction; %v", err)
			}

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {

				if sDB.GetBalance(addr).Cmp(acc.GetBalance()) != 0 {
					t.Fatalf("failed to prime account balance; Is: %v; Should be: %v", sDB.GetBalance(addr), acc.GetBalance())
				}

				if sDB.GetNonce(addr) != acc.GetNonce() {
					t.Fatalf("failed to prime account nonce; Is: %v; Should be: %v", sDB.GetNonce(addr), acc.GetNonce())
				}

				if bytes.Compare(sDB.GetCode(addr), acc.GetCode()) != 0 {
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

	ext := makeStateDbPrimer[any](cfg, log)

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
	if err != nil {
		t.Fatalf("cannot open test aida-db; %v", err)
	}

	err = ext.PreRun(executor.State[any]{}, &executor.Context{AidaDb: aidaDb, State: stateDb})
	if err == nil {
		t.Fatal("run must fail")
	}

	want := "cannot prime state-db; failed to prime StateDB: stop"

	if err.Error() != want {
		t.Fatalf("unexpected error\ngot: %v\nwant: %v", err, want)
	}
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

			if err != nil {
				t.Fatalf("failed to create state DB: %v", err)
			}

			// Generating randomized world state
			alloc, _ := utils.MakeWorldState(t)
			ws := txcontext.NewWorldState(alloc)

			pc := utils.NewPrimeContext(cfg, sDB, 0, log)
			// Priming state DB
			err = pc.PrimeStateDB(ws, sDB)
			if err != nil {
				t.Fatalf("failed to prime state DB: %v", err)
			}

			err = state.BeginCarmenDbTestContext(sDB)
			if err != nil {
				return
			}

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				if sDB.GetBalance(addr).Cmp(acc.GetBalance()) != 0 {
					t.Fatalf("failed to prime account balance; Is: %v; Should be: %v", sDB.GetBalance(addr), acc.GetBalance())
				}

				if sDB.GetNonce(addr) != acc.GetNonce() {
					t.Fatalf("failed to prime account nonce; Is: %v; Should be: %v", sDB.GetNonce(addr), acc.GetNonce())
				}

				if bytes.Compare(sDB.GetCode(addr), acc.GetCode()) != 0 {
					t.Fatalf("failed to prime account code; Is: %v; Should be: %v", sDB.GetCode(addr), acc.GetCode())
				}

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					if sDB.GetState(addr, keyHash) != valueHash {
						t.Fatalf("failed to prime account storage; Is: %v; Should be: %v", sDB.GetState(addr, keyHash), valueHash)
					}
				})
			})

			rootHash, err := sDB.GetHash()
			if err != nil {
				t.Fatalf("failed to get root hash: %v", err)
			}
			// Closing of state DB

			err = state.CloseCarmenDbTestContext(sDB)
			if err != nil {
				t.Fatalf("failed to close state DB: %v", err)
			}

			cfg.StateDbSrc = sDbDir
			// Call for json creation and writing into it
			err = utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, 2, rootHash, true)
			if err != nil {
				t.Fatalf("failed to write into DB info json file: %v", err)
			}

			// Initialization of state DB
			sDB2, sDbDir2, err := utils.PrepareStateDB(cfg)
			defer os.RemoveAll(sDbDir2)
			if err != nil {
				t.Fatalf("failed to create state DB2: %v", err)
			}

			defer func() {
				err = state.CloseCarmenDbTestContext(sDB2)
				if err != nil {
					t.Fatalf("failed to close state DB: %v", err)
				}
			}()

			err = sDB2.BeginBlock(uint64(7))
			if err != nil {
				t.Fatalf("cannot begin block; %v", err)
			}

			err = sDB2.BeginTransaction(uint32(0))
			if err != nil {
				t.Fatalf("cannot begin transaction; %v", err)
			}

			// Checks if state DB was primed correctly
			ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				if sDB2.GetBalance(addr).Cmp(acc.GetBalance()) != 0 {
					t.Fatalf("failed to prime account balance; Is: %v; Should be: %v", sDB2.GetBalance(addr), acc.GetBalance())
				}

				if sDB2.GetNonce(addr) != acc.GetNonce() {
					t.Fatalf("failed to prime account nonce; Is: %v; Should be: %v", sDB2.GetNonce(addr), acc.GetNonce())
				}

				if bytes.Compare(sDB2.GetCode(addr), acc.GetCode()) != 0 {
					t.Fatalf("failed to prime account code; Is: %v; Should be: %v", sDB2.GetCode(addr), acc.GetCode())
				}

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					if sDB2.GetState(addr, keyHash) != valueHash {
						t.Fatalf("failed to prime account storage; Is: %v; Should be: %v", sDB2.GetState(addr, keyHash), valueHash)
					}
				})
			})

			err = sDB2.EndTransaction()
			if err != nil {
				t.Fatalf("cannot end transaction; %v", err)
			}

			err = sDB2.EndBlock()
			if err != nil {
				t.Fatalf("cannot end block sDB2; %v", err)
			}

			// Generating randomized world state
			alloc2, _ := utils.MakeWorldState(t)
			ws2 := txcontext.NewWorldState(alloc2)

			pc2 := utils.NewPrimeContext(cfg, sDB2, 8, log)
			// Priming state DB
			err = pc2.PrimeStateDB(ws2, sDB2)
			if err != nil {
				t.Fatalf("failed to prime state DB2: %v", err)
			}

			err = sDB2.BeginBlock(uint64(10))
			if err != nil {
				t.Fatalf("cannot begin block; %v", err)
			}

			err = sDB2.BeginTransaction(uint32(0))
			if err != nil {
				t.Fatalf("cannot begin transaction; %v", err)
			}

			// Checks if state DB was primed correctly
			ws2.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
				if sDB2.GetBalance(addr).Cmp(acc.GetBalance()) != 0 {
					t.Fatalf("failed to prime account balance; Is: %v; Should be: %v", sDB2.GetBalance(addr), acc.GetBalance())
				}

				if sDB2.GetNonce(addr) != acc.GetNonce() {
					t.Fatalf("failed to prime account nonce; Is: %v; Should be: %v", sDB2.GetNonce(addr), acc.GetNonce())
				}

				if bytes.Compare(sDB2.GetCode(addr), acc.GetCode()) != 0 {
					t.Fatalf("failed to prime account code; Is: %v; Should be: %v", sDB2.GetCode(addr), acc.GetCode())
				}

				acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
					if sDB2.GetState(addr, keyHash) != valueHash {
						t.Fatalf("failed to prime account storage; Is: %v; Should be: %v", sDB2.GetState(addr, keyHash), valueHash)
					}
				})
			})
		})
	}
}

func TestPrimer_EthereumGenesisPriming(t *testing.T) {
	ctrl := gomock.NewController(t)

	addr := common.Address{0x1}
	for chainID, name := range utils.AllowedChainIDs {

		t.Run(name, func(t *testing.T) {
			// setup
			cfg := utils.NewTestConfig(t, chainID, utils.KeywordBlocks[chainID]["first"], 100, false, "Prague")
			cfg.SkipPriming = false
			log := logger.NewMockLogger(ctrl)
			stateDb := state.NewMockStateDB(ctrl)
			bulk := state.NewMockBulkLoad(ctrl)
			aidaDb, err := db.NewDefaultUpdateDB(t.TempDir())
			require.NoError(t, err, "cannot create updatedb")
			aidaDb.PutUpdateSet(&updateset.UpdateSet{
				WorldState: map[types.Address]*substate.Account{
					types.Address(addr): {
						Nonce:   12,
						Balance: uint256.NewInt(11),
						Storage: map[types.Hash]types.Hash{{0x2}: {0x3}},
						Code:    []byte{0x3},
					},
				},
				Block:           0,
				DeletedAccounts: nil,
			}, nil)

			// Genesis priming should only be triggered by ethereum data sets
			if _, isEthChainID := utils.EthereumChainIDs[chainID]; isEthChainID {
				gomock.InOrder(
					log.EXPECT().Noticef("Priming ethereum genesis..."),
					log.EXPECT().Debugf("\tLoading %d accounts with %d values ..", 1, 1),
					stateDb.EXPECT().BeginBlock(uint64(0)).Return(nil),
					stateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil),
					stateDb.EXPECT().Exist(addr).Return(false),
					stateDb.EXPECT().EndTransaction(),
					stateDb.EXPECT().EndBlock(),
					stateDb.EXPECT().StartBulkLoad(uint64(1)).Return(bulk, nil),
					bulk.EXPECT().CreateAccount(addr),
					bulk.EXPECT().SetBalance(addr, uint256.NewInt(11)),
					bulk.EXPECT().SetNonce(addr, uint64(12)),
					bulk.EXPECT().SetCode(addr, []byte{0x3}),
					bulk.EXPECT().SetState(addr, common.Hash{0x2}, common.Hash{0x3}),
					bulk.EXPECT().Close().Return(nil),
					log.EXPECT().Debugf("\t\tPriming completed ..."),
				)
			}

			primer := makeStateDbPrimer[any](cfg, log)
			err = primer.PreRun(executor.State[any]{}, &executor.Context{
				State:  stateDb,
				AidaDb: aidaDb,
			})
			require.NoError(t, err, "pre-run failed")
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
	p := &stateDbPrimer[any]{
		log: log,
		ctx: utils.NewPrimeContext(cfg, mockDb, 0, log),
		cfg: cfg,
	}
	mockAida.EXPECT().GetBackend().Return(mockAdapter)
	mockIter := NewMockProxyIterator(ctrl)
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(mockIter)
	mockAdapter.EXPECT().Get(gomock.Any(), gomock.Any()).Return([]uint8{1, 2, 3}, nil).AnyTimes()
	mockDb.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	mockDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	mockDb.EXPECT().EndTransaction().Return(nil).AnyTimes()
	mockDb.EXPECT().EndBlock().Return(nil).AnyTimes()
	mockDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulk, nil).AnyTimes()
	mockBulk.EXPECT().Close().AnyTimes()
	mockDb.EXPECT().BeginSyncPeriod(gomock.Any()).AnyTimes()
	mockDb.EXPECT().EndSyncPeriod().AnyTimes()
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
	err := p.prime(mockDb, mockAida)
	assert.NoError(t, err)
}

//go:generate mockgen -source=primer_test.go -destination=primer_mock.go -package=primer
type ProxyIterator interface {
	iterator.Iterator
}
