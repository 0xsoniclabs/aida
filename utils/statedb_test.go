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

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/state"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	gomock "go.uber.org/mock/gomock"
)

// TestStatedb_InitCloseStateDB test closing db immediately after initialization
func TestStatedb_InitCloseStateDB(t *testing.T) {
	for _, tc := range GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := MakeTestConfig(tc)

			// Initialization of state DB
			sDB, _, err := PrepareStateDB(cfg)

			if err != nil {
				t.Fatalf("failed to create state DB: %v", err)
			}

			// Closing of state DB
			err = sDB.Close()
			if err != nil {
				t.Fatalf("failed to close state DB: %v", err)
			}
		})
	}
}

// TestStatedb_PrepareStateDB tests preparation and initialization of existing state DB
func TestStatedb_PrepareStateDB(t *testing.T) {
	for _, tc := range GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := MakeTestConfig(tc)
			// Update config for state DB preparation by providing additional information
			cfg.DbTmp = t.TempDir()
			cfg.StateDbSrc = t.TempDir()
			cfg.First = 2
			cfg.Last = 4

			// Create state DB info of existing state DB
			dbInfo := StateDbInfo{
				Impl:           cfg.DbImpl,
				Variant:        cfg.DbVariant,
				ArchiveMode:    cfg.ArchiveMode,
				ArchiveVariant: cfg.ArchiveVariant,
				Schema:         0,
				Block:          cfg.Last,
				RootHash:       common.Hash{},
				GitCommit:      GitCommit,
				CreateTime:     time.Now().UTC().Format(time.UnixDate),
			}

			// Create json file for the existing state DB info
			dbInfoJson, err := json.Marshal(dbInfo)
			if err != nil {
				t.Fatalf("failed to create DB info json: %v", err)
			}

			// Fill the json file with the info
			err = os.WriteFile(filepath.Join(cfg.StateDbSrc, PathToDbInfo), dbInfoJson, 0755)
			if err != nil {
				t.Fatalf("failed to write into DB info json file: %v", err)
			}

			// remove files after test ends
			defer func(path string) {
				err = os.RemoveAll(path)
				if err != nil {
					t.Fatal(err)
				}
			}(cfg.StateDbSrc)

			// Call for state DB preparation and subsequent check if it finished successfully
			sDB, _, err := PrepareStateDB(cfg)
			if err != nil {
				t.Fatalf("failed to create state DB: %v", err)
			}

			// Closing of state DB
			defer func(sDB state.StateDB) {
				err = sDB.Close()
				if err != nil {
					t.Fatalf("failed to close state DB: %v", err)
				}
			}(sDB)
		})
	}
}

// TestStatedb_PrepareStateDB tests preparation and initialization of existing state DB as empty
// because of missing PathToDbInfo file
func TestStatedb_PrepareStateDBEmpty(t *testing.T) {
	tc := GetStateDbTestCases()[0]
	cfg := MakeTestConfig(tc)
	// Update config for state DB preparation by providing additional information
	cfg.ShadowImpl = ""
	cfg.DbTmp = t.TempDir()
	cfg.First = 2

	// Call for state DB preparation and subsequent check if it finished successfully
	sDB, _, err := PrepareStateDB(cfg)
	if err != nil {
		t.Fatalf("failed to create state DB: %v", err)
	}

	// Closing of state DB
	defer func(sDB state.StateDB) {
		err = sDB.Close()
		if err != nil {
			t.Fatalf("failed to close state DB: %v", err)
		}
	}(sDB)
}

func TestStateDB_makeNewStateDB(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		DbImpl:                 "memory",
		DbVariant:              "",
		ShadowImpl:             "geth",
		ShadowDb:               true,
		ArchiveMode:            true,
		ArchiveVariant:         "",
		PathToStateDb:          tempDir,
		StateDbSrc:             tempDir,
		StateDbSrcDirectAccess: true,
		ChainID:                MainnetChainID,
	}

	db, dbPath, err := makeNewStateDB(cfg)
	if err != nil {
		t.Fatalf("failed to create state DB: %v", err)
	}
	defer func(path string) {
		e := os.RemoveAll(path)
		if e != nil {
			t.Fatalf("failed to remove state DB path: %v", e)
		}
	}(dbPath)

	if db == nil {
		t.Fatal("expected non-nil state DB")
	}
}

func TestStateDB_useExistingStateDB(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		DbImpl:                 "memory",
		DbVariant:              "",
		ShadowImpl:             "geth",
		ShadowDb:               true,
		ArchiveMode:            true,
		ArchiveVariant:         "",
		PathToStateDb:          tempDir,
		StateDbSrc:             tempDir,
		StateDbSrcDirectAccess: true,
		ChainID:                MainnetChainID,
	}

	// Create state DB info of existing state DB
	dbInfo := StateDbInfo{
		Impl:           cfg.DbImpl,
		Variant:        cfg.DbVariant,
		ArchiveMode:    cfg.ArchiveMode,
		ArchiveVariant: "xyz",
		Schema:         0,
		Block:          cfg.Last,
		RootHash:       common.Hash{},
		GitCommit:      GitCommit,
		CreateTime:     time.Now().UTC().Format(time.UnixDate),
	}

	// Create json file for the existing state DB info
	dbInfoJson, err := json.Marshal(dbInfo)
	if err != nil {
		t.Fatalf("failed to create DB info json: %v", err)
	}

	// Fill the json file with the info
	err = os.Mkdir(filepath.Join(cfg.PathToStateDb, PathToPrimaryStateDb), 0755)
	if err != nil {
		t.Fatalf("failed to create directory for DB info json file: %v", err)
	}
	err = os.WriteFile(filepath.Join(cfg.PathToStateDb, PathToPrimaryStateDb, PathToDbInfo), dbInfoJson, 0755)
	if err != nil {
		t.Fatalf("failed to write into DB info json file: %v", err)
	}
	err = os.Mkdir(filepath.Join(cfg.PathToStateDb, PathToShadowStateDb), 0755)
	if err != nil {
		t.Fatalf("failed to create directory for DB info json file: %v", err)
	}
	err = os.WriteFile(filepath.Join(cfg.PathToStateDb, PathToShadowStateDb, PathToDbInfo), dbInfoJson, 0755)
	if err != nil {
		t.Fatalf("failed to write into DB info json file: %v", err)
	}

	// remove files after test ends
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(cfg.StateDbSrc)

	db, dbPath, err := useExistingStateDB(cfg)
	if err != nil {
		t.Fatalf("failed to create state DB: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatalf("failed to remove state DB path: %v", err)
		}
	}(dbPath)

	if db == nil {
		t.Fatal("expected non-nil state DB")
	}
}

func TestWorldstateUpdate_OverwriteStateDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockVmStateDB(ctrl)

	// Define the world state to overwrite
	ws := substate.WorldState{
		substatetypes.Address{0x01}: &substate.Account{
			Nonce:   1,
			Balance: uint256.NewInt(1000),
			Code:    []byte{0x60, 0x60},
			Storage: map[substatetypes.Hash]substatetypes.Hash{{0x01}: {0x02}},
		},
	}

	// Create a patch with the world state
	patch := substatecontext.NewWorldState(ws)

	gomock.InOrder(
		db.EXPECT().Exist(common.Address{0x01}).Times(1),
		db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
		db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(500)).Times(1),
		db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(1000), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(2)).Times(1),
		db.EXPECT().SetNonce(common.Address{0x01}, uint64(1), tracing.NonceChangeUnspecified).Times(1),
		db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{0x60, 0x00}).Times(1),
		db.EXPECT().SetCode(common.Address{0x01}, []byte{0x60, 0x60}, tracing.CodeChangeUnspecified).Times(1),
		db.EXPECT().GetState(common.Address{0x01}, common.Hash{0x01}).Return(common.Hash{}).Times(1),
		db.EXPECT().SetState(common.Address{0x01}, common.Hash{0x01}, common.Hash{0x02}).Times(1),
	)
	// Call the method to test
	OverwriteStateDB(patch, db)
}
