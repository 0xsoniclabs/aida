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

package utildb

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb/dbcomponent"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTableHash_Empty(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config
	cfg := &utils.Config{
		DbComponent: string(dbcomponent.All), // Set this to the component you want to test
	}

	gomock.InOrder(
		// substate count
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(0)),
		// delete count
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(0)),
		// update count
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(0)),
		// state hash count
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(0)),
		// block hash count
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(0)),
	)

	// Call the function
	err = TableHash(cfg, database, log) // Pass a logger if needed
	assert.NoError(t, err)
}

func TestTableHash_Filled(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config
	cfg := &utils.Config{
		DbComponent: string(dbcomponent.All), // Set this to the component you want to test
		First:       0,
		Last:        100, // None of the following generators must not generate record higher than this number
	}

	substateCount, deleteCount, updateCount, stateHashCount, blockHashCount := fillFakeAidaDb(t, database)

	gomock.InOrder(
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(substateCount)),
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(deleteCount)),
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(updateCount)),
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(stateHashCount)),
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(blockHashCount)),
	)

	// Call the function
	err = TableHash(cfg, database, log) // Pass a logger if needed
	assert.NoError(t, err)
}

func TestTableHash_JustSubstate(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config
	cfg := &utils.Config{
		DbComponent: string(dbcomponent.Substate), // Set this to the component you want to test
		First:       0,
		Last:        100, // None of the following generators must not generate record higher than this number
	}

	substateCount, _, _, _, _ := fillFakeAidaDb(t, database)

	gomock.InOrder(
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(substateCount)),
	)

	// Call the function
	err = TableHash(cfg, database, log) // Pass a logger if needed
	assert.NoError(t, err)
}

func TestTableHash_JustDelete(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config
	cfg := &utils.Config{
		DbComponent: string(dbcomponent.Delete), // Set this to the component you want to test
		First:       0,
		Last:        100, // None of the following generators must not generate record higher than this number
	}

	_, deleteCount, _, _, _ := fillFakeAidaDb(t, database)

	gomock.InOrder(
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(deleteCount)),
	)

	// Call the function
	err = TableHash(cfg, database, log) // Pass a logger if needed
	assert.NoError(t, err)
}

func TestTableHash_JustUpdate(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config
	cfg := &utils.Config{
		DbComponent: string(dbcomponent.Update), // Set this to the component you want to test
		First:       0,
		Last:        100, // None of the following generators must not generate record higher than this number
	}

	_, _, updateCount, _, _ := fillFakeAidaDb(t, database)

	gomock.InOrder(
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(updateCount)),
	)

	// Call the function
	err = TableHash(cfg, database, log) // Pass a logger if needed
	assert.NoError(t, err)
}

func TestTableHash_JustStateHash(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config
	cfg := &utils.Config{
		DbComponent: string(dbcomponent.StateHash), // Set this to the component you want to test
		First:       0,
		Last:        100, // None of the following generators must not generate record higher than this number
	}

	_, _, _, stateHashCount, _ := fillFakeAidaDb(t, database)

	gomock.InOrder(
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(stateHashCount)),
	)

	// Call the function
	err = TableHash(cfg, database, log) // Pass a logger if needed
	assert.NoError(t, err)
}

func TestTableHash_JustBlockHash(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config
	cfg := &utils.Config{
		DbComponent: string(dbcomponent.BlockHash), // Set this to the component you want to test
		First:       0,
		Last:        100, // None of the following generators must not generate record higher than this number
	}

	_, _, _, _, blockHashCount := fillFakeAidaDb(t, database)

	gomock.InOrder(
		log.EXPECT().Info(gomock.Any()),
		log.EXPECT().Infof(gomock.Any(), gomock.Any(), uint64(blockHashCount)),
	)

	// Call the function
	err = TableHash(cfg, database, log) // Pass a logger if needed
	assert.NoError(t, err)
}

func TestTableHash_InvalidSubstateEncoding(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config with an invalid substate encoding
	cfg := &utils.Config{
		DbComponent:      string(dbcomponent.Substate),
		SubstateEncoding: "invalid_encoding",
	}

	gomock.InOrder(
		log.EXPECT().Info(gomock.Any()),
	)

	err = TableHash(cfg, database, log)
	assert.Error(t, err)
}

func TestTableHash_InvalidKeys(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, dbInst db.BaseDB)
		dbComponent dbcomponent.DbComponent
		logMsg      string
		errWant     string
	}{
		{
			name: "InvalidStateHashKey",
			setup: func(t *testing.T, dbInst db.BaseDB) {
				// must be bigger than 32 bytes
				junkValue := "asffsafasfassafsafkjlasffasklsfaklasfjagqeiojgqeiogewiogewjogieweowvniboiewgioewjgfewiofewijofewjeiqoqwfio"
				err := dbInst.Put([]byte(utils.StateRootHashPrefix+"0x1"), []byte(junkValue))
				assert.NoError(t, err)
			},
			dbComponent: dbcomponent.StateHash,
			logMsg:      "Generating State-Hashes hash...",
			errWant:     "invalid state root length for block 1: expected 32 bytes, got 106 bytes",
		},
		{
			name: "InvalidDeleteKey",
			setup: func(t *testing.T, dbInst db.BaseDB) {
				junkValue := "asffsafasfassafsafkjlasffasklsfaklasfjagqeiojgqeiogewiogewjogieweowvniboiewgioewjgfewiofewijofewjeiqoqwfio"
				err := dbInst.Put(db.EncodeDestroyedAccountKey(1, 0), []byte(junkValue))
				assert.NoError(t, err)
			},
			dbComponent: dbcomponent.Delete,
			logMsg:      "Generating Deletion hash...",
			errWant:     "rlp: expected input list for db.SuicidedAccountLists",
		},
		{
			name: "InvalidBlockHashKey",
			setup: func(t *testing.T, dbInst db.BaseDB) {
				junkValue := "asffsafasfassafsafkjlasffasklsfaklasfjagqeiojgqeiogewiogewjogieweowvniboiewgioewjgfewiofewijofewjeiqoqwfio"
				err := dbInst.Put(utils.BlockHashDBKey(1), []byte(junkValue))
				assert.NoError(t, err)
			},
			dbComponent: dbcomponent.BlockHash,
			logMsg:      "Generating Block-Hashes hash...",
			errWant:     "invalid block hash length for block 1: expected 32 bytes, got 106 bytes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir() + "/aidaDb"
			database, err := db.NewDefaultBaseDB(tmpDir)
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()

			tc.setup(t, database)

			ctrl := gomock.NewController(t)
			log := logger.NewMockLogger(ctrl)

			cfg := &utils.Config{
				First:       1,
				Last:        1,
				DbComponent: string(tc.dbComponent),
			}

			gomock.InOrder(
				log.EXPECT().Info(tc.logMsg),
			)

			err = TableHash(cfg, database, log)
			if err == nil {
				t.Fatalf("expected an error: %v, but got nil", tc.errWant)
			}
			assert.Equal(t, tc.errWant, err.Error(), "error message mismatch")
		})
	}
}

func TestTableHash_InvalidDbComponent(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	defer database.Close()

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	// Create a config with an invalid db component
	cfg := &utils.Config{
		DbComponent: "invalid_component",
	}

	errWant := "invalid db component: invalid_component. Usage: (\"all\", \"substate\", \"delete\", \"update\", \"state-hash\", \"block-hash\")"
	err = TableHash(cfg, database, log)
	if err == nil {
		t.Fatalf("expected an error: %v, but got nil", errWant)
	}
	assert.Equal(t, errWant, err.Error())
}

func fillFakeAidaDb(t *testing.T, aidaDb db.BaseDB) (int, int, int, int, int) {
	// Seed the random number generator
	rand.NewSource(time.Now().UnixNano())

	sdb := db.MakeDefaultSubstateDBFromBaseDB(aidaDb)
	// Generate a random number between 1 and 5
	numSubstates := rand.Intn(5) + 1
	acc := substate.NewAccount(1, uint256.NewInt(1), []byte{1})

	for i := 0; i < numSubstates; i++ {
		state := substate.Substate{
			Block:       uint64(i),
			Transaction: 0,
			Env: &substate.Env{
				Number:     uint64(i),
				Difficulty: big.NewInt(int64(i)),
				GasLimit:   uint64(i),
			},
			Message: &substate.Message{
				Value:    big.NewInt(int64(rand.Intn(100))),
				GasPrice: big.NewInt(int64(rand.Intn(100))),
			},
			InputSubstate:  substate.WorldState{substatetypes.Address{0x0}: acc},
			OutputSubstate: substate.WorldState{substatetypes.Address{0x0}: acc},
			Result:         &substate.Result{},
		}

		err := sdb.PutSubstate(&state)
		if err != nil {
			t.Fatal(err)
		}
	}

	ddb := db.MakeDefaultDestroyedAccountDBFromBaseDB(aidaDb)

	// Generate random number between 6-10
	numDestroyedAccounts := rand.Intn(5) + 6

	for i := 0; i < numDestroyedAccounts; i++ {
		err := ddb.SetDestroyedAccounts(uint64(i), 0, []substatetypes.Address{substatetypes.BytesToAddress(utils.MakeRandomByteSlice(t, 40))}, []substatetypes.Address{})
		if err != nil {
			t.Fatalf("error setting destroyed accounts: %v", err)
		}
	}

	udb := db.MakeDefaultUpdateDBFromBaseDB(aidaDb)

	// Generate random number between 11-15
	numUpdates := rand.Intn(5) + 11

	for i := 0; i < numUpdates; i++ {
		sa := new(substate.Account)
		sa.Balance = uint256.NewInt(uint64(utils.GetRandom(1, 1000*5000)))
		randomAddress := substatetypes.BytesToAddress(utils.MakeRandomByteSlice(t, 40))
		worldState := substate.WorldState{

			randomAddress: sa,
		}
		err := udb.PutUpdateSet(&updateset.UpdateSet{
			WorldState: worldState,
			Block:      uint64(i),
		}, []substatetypes.Address{})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Generate random number between 16-20
	numStateHashes := rand.Intn(5) + 16

	for i := 0; i < numStateHashes; i++ {
		err := utils.SaveStateRoot(aidaDb, fmt.Sprintf("0x%x", i), strings.Repeat("0", 64))
		if err != nil {
			t.Fatalf("error saving state root: %v", err)
		}
	}

	// Generate random number between 21-25
	numBlockHashes := rand.Intn(5) + 21
	for i := 0; i < numBlockHashes; i++ {
		err := utils.SaveBlockHash(aidaDb, fmt.Sprintf("0x%x", i), strings.Repeat("0", 64))
		if err != nil {
			t.Fatalf("error saving block hash: %v", err)
		}
	}

	return numSubstates, numDestroyedAccounts, numUpdates, numStateHashes, numBlockHashes
}

func TestTableHash_GetHashesHash_Ticker(t *testing.T) {
	tests := []struct {
		name        string
		getHashFunc func(cfg *utils.Config, db db.BaseDB, log logger.Logger) ([]byte, uint64, error)
		logMsg      string
	}{
		{
			name:        "StateRootHashes",
			getHashFunc: GetStateRootHashesHash,
			logMsg:      "State-Hashes hash progress: %v/%v",
		},
		{
			name:        "BlockHashes",
			getHashFunc: GetBlockHashesHash,
			logMsg:      "Block-Hashes hash progress: %v/%v",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			log := logger.NewMockLogger(ctrl)
			aidaDb := db.NewMockBaseDB(ctrl)

			cfg := &utils.Config{
				First: 0,
				Last:  1,
			}

			aidaDb.EXPECT().Get(gomock.Any()).DoAndReturn(func(key []byte) ([]byte, error) {
				time.Sleep(60 * time.Second) // Simulate a delay
				return []byte("12345678123456781234567812345678"), nil
			})
			aidaDb.EXPECT().Get(gomock.Any()).Return([]byte("12345678123456781234567812345678"), nil)
			log.EXPECT().Infof(tc.logMsg, uint64(1), uint64(1))

			_, count, err := tc.getHashFunc(cfg, aidaDb, log)
			assert.NoError(t, err)
			assert.Equal(t, uint64(2), count, "Expected count to be 2")
		})
	}
}
