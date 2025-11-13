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
	"fmt"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorldStateUpdate_GenerateUpdateSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &utils.Config{}
	// mockStateDb := state.NewMockStateDB(ctrl)
	mockSubstateDb := db.NewMockSubstateDB(ctrl)
	mockDeletionDb := db.NewMockDestroyedAccountDB(ctrl)
	mockSubstateIter := db.NewMockIIterator[*substate.Substate](ctrl)

	substateBlk2 := &substate.Substate{
		OutputSubstate: substate.NewWorldState().Add(types.Address{3}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:          2,
		Transaction:    0,
	}

	mockDestroyed := []types.Address{{1}, {2}}
	mockResurrected := []types.Address{{3}}

	gomock.InOrder(
		mockSubstateDb.EXPECT().NewSubstateIterator(gomock.Any(), gomock.Any()).Return(mockSubstateIter).AnyTimes(),
		mockSubstateIter.EXPECT().Next().Return(true),
		mockSubstateIter.EXPECT().Value().Return(substateBlk2),
		mockDeletionDb.EXPECT().GetDestroyedAccounts(uint64(2), 0).Return(mockDestroyed, mockResurrected, nil),
		mockSubstateIter.EXPECT().Next().Return(false),
		mockSubstateIter.EXPECT().Release(),
	)
	retUpdateSet, retDestroyed, err := generateUpdateSet(0, 2, cfg, mockSubstateDb, mockDeletionDb)
	assert.NoError(t, err)
	assert.NotNil(t, retUpdateSet)
	assert.Equal(t, len(substateBlk2.OutputSubstate), len(retUpdateSet))
	assert.Equal(t, len(mockDestroyed)+len(mockResurrected), len(retDestroyed))
}

func TestWorldStateUpdate_ClearAccountStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	addr := types.BytesToAddress([]byte("test"))
	ws := substate.WorldState{}
	ws[addr] = &substate.Account{
		Nonce:   1,
		Balance: nil,
		Storage: map[types.Hash]types.Hash{
			types.BytesToHash([]byte("key1")): types.BytesToHash([]byte("value1")),
		},
		Code: nil,
	}
	ClearAccountStorage(ws, []types.Address{addr})
	assert.Equal(t, 0, len(ws[addr].Storage))
}

// TestStatedb_DeleteDestroyedAccountsFromWorldState tests removal of destroyed accounts from given world state
func TestStatedb_DeleteDestroyedAccountsFromWorldState(t *testing.T) {
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)
			// Generating randomized world state
			ws, addrList := utils.MakeWorldState(t)
			// Init directory for destroyed accounts DB
			deletionDb := t.TempDir()
			// Pick two account which will represent destroyed ones
			destroyedAccounts := []types.Address{
				types.Address(addrList[0]),
				types.Address(addrList[50]),
			}

			// Update config to enable removal of destroyed accounts
			cfg.DeletionDb = deletionDb

			// Initializing backend DB for storing destroyed accounts
			daBackend, err := db.NewDefaultSubstateDB(deletionDb)
			if err != nil {
				t.Fatalf("failed to create backend DB: %s; %v", deletionDb, err)
			}

			// Creating new destroyed accounts DB
			daDB, err := db.MakeDefaultDestroyedAccountDBFromBaseDB(daBackend)
			assert.NoError(t, err)

			// Storing two picked accounts from destroyedAccounts slice to destroyed accounts DB
			err = daDB.SetDestroyedAccounts(5, 1, destroyedAccounts, []types.Address{})
			if err != nil {
				t.Fatalf("failed to set destroyed accounts into DB: %v", err)
			}

			// Closing destroyed accounts DB
			err = daDB.Close()
			if err != nil {
				t.Fatalf("failed to close destroyed accounts DB: %v", err)
			}

			// Call for removal of destroyed accounts from given world state
			err = deleteDestroyedAccountsFromWorldState(ws, cfg, 5)
			if err != nil {
				t.Fatalf("failed to delete accounts from the world state: %v", err)
			}

			// check if accounts are not present anymore
			if ws.Get(common.Address(destroyedAccounts[0])) != nil || ws.Get(common.Address(destroyedAccounts[1])) != nil {
				t.Fatalf("failed to delete accounts from the world state")
			}
		})
	}
}

// TestStatedb_DeleteDestroyedAccountsFromWorldState tests removal of deleted accounts from given state DB
func TestStatedb_DeleteDestroyedAccountsFromStateDB(t *testing.T) {
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)
			// Generating randomized world state
			ws, addrList := utils.MakeWorldState(t)
			// Init directory for destroyed accounts DB
			deletedAccountsDir := t.TempDir()
			// Pick two account which will represent destroyed ones
			destroyedAccounts := []types.Address{
				types.Address(addrList[0]),
				types.Address(addrList[50]),
			}

			// Update config to enable removal of destroyed accounts
			cfg.DeletionDb = deletedAccountsDir

			// Initializing backend DB for storing destroyed accounts
			aidaDb, err := db.NewDefaultSubstateDB(deletedAccountsDir)
			if err != nil {
				t.Fatalf("failed to create backend DB: %s; %v", deletedAccountsDir, err)
			}

			// Creating new destroyed accounts DB
			ddb, err := db.MakeDefaultDestroyedAccountDBFromBaseDB(aidaDb)
			assert.NoError(t, err)

			// Storing two picked accounts from destroyedAccounts slice to destroyed accounts DB
			err = ddb.SetDestroyedAccounts(5, 1, destroyedAccounts, []types.Address{})
			if err != nil {
				t.Fatalf("failed to set destroyed accounts into DB: %v", err)
			}

			defer func(daDB db.DestroyedAccountDB) {
				e := daDB.Close()
				if e != nil {
					t.Fatalf("failed to close destroyed accounts DB: %v", e)
				}
			}(ddb)

			// Initialization of state DB
			stateDb, _, err := utils.PrepareStateDB(cfg)
			if err != nil {
				t.Fatalf("failed to create state DB: %v", err)
			}

			log := logger.NewLogger("INFO", "TestStateDb")

			p := &primer{
				cfg:    cfg,
				log:    log,
				ctx:    newContext(cfg, stateDb, log),
				aidadb: aidaDb,
				ddb:    ddb,
			}
			// Priming state DB with given world state
			err = p.ctx.PrimeStateDB(ws)
			if err != nil {
				t.Fatalf("cannot prime statedb; %v", err)
			}

			// Call for removal of destroyed accounts from state DB
			err = p.mayDeleteDestroyedAccountsFromStateDB(5)
			if err != nil {
				t.Fatalf("failed to delete accounts from the state DB: %v", err)
			}

			err = state.BeginCarmenDbTestContext(stateDb)
			if err != nil {
				t.Fatal(err)
			}

			// check if accounts are not present anymore
			for _, da := range destroyedAccounts {
				if stateDb.Exist(common.Address(da)) {
					t.Fatalf("failed to delete destroyed accounts from the state DB")
				}
			}

			err = state.CloseCarmenDbTestContext(stateDb)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
