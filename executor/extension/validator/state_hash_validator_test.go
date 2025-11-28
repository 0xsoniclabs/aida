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

package validator

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
)

const exampleHashA = "0x0100000000000000000000000000000000000000000000000000000000000000"
const exampleHashB = "0x0102000000000000000000000000000000000000000000000000000000000000"
const exampleHashC = "0x0a0b0c0000000000000000000000000000000000000000000000000000000000"
const exampleHashD = "0x0300000000000000000000000000000000000000000000000000000000000000"

func TestStateHashValidator_NotActiveIfNotEnabledInConfig(t *testing.T) {
	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	cfg.ValidateStateHashes = false

	ext := MakeStateHashValidator[any](cfg)
	if _, ok := ext.(extension.NilExtension[any]); !ok {
		t.Errorf("extension is active although it should not")
	}
}

func TestStateHashValidator_FailsIfHashIsNotFoundInAidaDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	hashProvider := substateDb.NewMockHashProvider(ctrl)

	blockNumber := 1

	gomock.InOrder(
		hashProvider.EXPECT().GetStateRootHash(blockNumber).Return(types.Hash{}, leveldb.ErrNotFound),
	)

	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	ext := makeStateHashValidator[any](cfg, log)
	ext.hashProvider = hashProvider

	// AidaDb doesn't exist
	ctx := &executor.Context{State: db}

	err := ext.PostBlock(executor.State[any]{Block: blockNumber}, ctx)
	if err == nil {
		t.Error("post block must return error")
	}

	wantedErr := fmt.Sprintf("state hash for block %v is not present in the db", blockNumber)

	if strings.Compare(err.Error(), wantedErr) != 0 {
		t.Fatalf("unexpected error\nwant: %v\ngot: %v", wantedErr, err.Error())
	}

	if err := ext.PostRun(executor.State[any]{Block: 1}, ctx, nil); err != nil {
		t.Errorf("failed to finish PostRun: %v", err)
	}
}

func TestStateHashValidator_InvalidHashOfLiveDbIsDetected(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	hashProvider := substateDb.NewMockHashProvider(ctrl)

	blockNumber := 1

	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	ext := makeStateHashValidator[any](cfg, log)
	ext.hashProvider = hashProvider

	gomock.InOrder(
		hashProvider.EXPECT().GetStateRootHash(blockNumber).Return(types.Hash([]byte(exampleHashA)), nil),
		db.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashB)), nil),
	)

	ctx := &executor.Context{State: db}

	if err := ext.PostBlock(executor.State[any]{Block: blockNumber}, ctx); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("unexpected hash for Live block %v", blockNumber)) {
		t.Errorf("failed to detect incorrect hash, err %v", err)
	}
}
func TestStateHashValidator_InvalidHashOfArchiveDbIsDetected(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	hashProvider := substateDb.NewMockHashProvider(ctrl)
	sdb := substateDb.NewMockSubstateDB(ctrl)

	blockNumber := 1
	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	cfg.ArchiveMode = true
	cfg.ArchiveVariant = "s5"

	archive := state.NewMockNonCommittableStateDB(ctrl)

	output := make(map[int]*substate.Substate)
	output[0] = &substate.Substate{
		Block:       uint64(0),
		Transaction: 1,
	}

	gomock.InOrder(
		// live state check goes through
		hashProvider.EXPECT().GetStateRootHash(blockNumber).Return(types.Hash([]byte(exampleHashA)), nil),
		db.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashA)), nil),
		db.EXPECT().GetArchiveBlockHeight().Return(uint64(blockNumber), false, nil),

		// archive state check fails
		hashProvider.EXPECT().GetStateRootHash(blockNumber-1).Return(types.Hash([]byte(exampleHashA)), nil),
		db.EXPECT().GetArchiveState(uint64(blockNumber-1)).Return(archive, nil),
		archive.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashB)), nil),
		archive.EXPECT().Release(),
		sdb.EXPECT().GetBlockSubstates(uint64(blockNumber-1)).Return(output, nil),
	)

	ctx := &executor.Context{State: db}
	ext := makeStateHashValidator[any](cfg, log)
	ext.hashProvider = hashProvider
	ext.sdb = sdb
	if err := ext.PostBlock(executor.State[any]{Block: blockNumber}, ctx); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("unexpected hash for archive block %d", blockNumber-1)) {
		t.Errorf("failed to detect incorrect hash, err %v", err)
	}
}

func TestStateHashValidator_ChecksArchiveHashesOfLaggingArchive(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	sdb := substateDb.NewMockSubstateDB(ctrl)
	hashProvider := substateDb.NewMockHashProvider(ctrl)

	output := make(map[int]*substate.Substate)
	output[0] = &substate.Substate{
		Block:       uint64(2),
		Transaction: 1,
	}

	db.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashA)), nil)
	hashProvider.EXPECT().GetStateRootHash(2).Return(types.Hash([]byte(exampleHashA)), nil)

	archive0 := state.NewMockNonCommittableStateDB(ctrl)
	archive1 := state.NewMockNonCommittableStateDB(ctrl)
	archive2 := state.NewMockNonCommittableStateDB(ctrl)

	gomock.InOrder(
		db.EXPECT().GetArchiveBlockHeight().Return(uint64(0), false, nil),
		hashProvider.EXPECT().GetStateRootHash(0).Return(types.Hash([]byte(exampleHashB)), nil),
		db.EXPECT().GetArchiveState(uint64(0)).Return(archive0, nil),
		archive0.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashB)), nil),
		archive0.EXPECT().Release(),

		db.EXPECT().GetArchiveBlockHeight().Return(uint64(0), false, nil),
		db.EXPECT().GetArchiveBlockHeight().Return(uint64(0), false, nil),

		db.EXPECT().GetArchiveBlockHeight().Return(uint64(2), false, nil),
		hashProvider.EXPECT().GetStateRootHash(1).Return(types.Hash([]byte(exampleHashC)), nil),
		db.EXPECT().GetArchiveState(uint64(1)).Return(archive1, nil),
		archive1.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashC)), nil),
		archive1.EXPECT().Release(),
		hashProvider.EXPECT().GetStateRootHash(2).Return(types.Hash([]byte(exampleHashD)), nil),
		db.EXPECT().GetArchiveState(uint64(2)).Return(archive2, nil),
		archive2.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashA)), nil),
		archive2.EXPECT().Release(),
		sdb.EXPECT().GetBlockSubstates(uint64(2)).Return(output, nil),
	)

	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	cfg.Last = 5
	cfg.ArchiveMode = true
	cfg.ArchiveVariant = "s5"

	ext := makeStateHashValidator[any](cfg, log)
	ext.hashProvider = hashProvider
	ext.sdb = sdb
	ctx := &executor.Context{State: db}

	// A PostBlock run should check the LiveDB and the ArchiveDB up to block 0.
	if err := ext.PostBlock(executor.State[any]{Block: 2}, ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// PostRun should finish up checking all remaining archive hashes and detect the error in block 2.
	if err := ext.PostRun(executor.State[any]{Block: 3}, ctx, nil); err == nil || !strings.Contains(err.Error(), "unexpected hash for archive block 2") {
		t.Errorf("failed to detect incorrect hash, err %v", err)
	}
}

func TestStateHashValidator_ChecksArchiveHashesOfLaggingArchiveDoesNotWaitForNonExistingBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	hashProvider := substateDb.NewMockHashProvider(ctrl)

	db.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashA)), nil)
	hashProvider.EXPECT().GetStateRootHash(2).Return(types.Hash([]byte(exampleHashA)), nil)

	archive0 := state.NewMockNonCommittableStateDB(ctrl)
	archive1 := state.NewMockNonCommittableStateDB(ctrl)
	archive2 := state.NewMockNonCommittableStateDB(ctrl)

	gomock.InOrder(
		db.EXPECT().GetArchiveBlockHeight().Return(uint64(0), false, nil),
		hashProvider.EXPECT().GetStateRootHash(0).Return(types.Hash([]byte(exampleHashB)), nil),
		db.EXPECT().GetArchiveState(uint64(0)).Return(archive0, nil),
		archive0.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashB)), nil),
		archive0.EXPECT().Release(),

		db.EXPECT().GetArchiveBlockHeight().Return(uint64(2), false, nil),
		hashProvider.EXPECT().GetStateRootHash(1).Return(types.Hash([]byte(exampleHashC)), nil),
		db.EXPECT().GetArchiveState(uint64(1)).Return(archive1, nil),
		archive1.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashC)), nil),
		archive1.EXPECT().Release(),
		hashProvider.EXPECT().GetStateRootHash(2).Return(types.Hash([]byte(exampleHashD)), nil),
		db.EXPECT().GetArchiveState(uint64(2)).Return(archive2, nil),
		archive2.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashD)), nil),
		archive2.EXPECT().Release(),
	)

	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	cfg.Last = 5
	cfg.ArchiveMode = true
	cfg.ArchiveVariant = "s5"

	ext := makeStateHashValidator[any](cfg, log)
	ext.hashProvider = hashProvider
	ctx := &executor.Context{State: db}

	// A PostBlock run should check the LiveDB and the ArchiveDB up to block 0.
	if err := ext.PostBlock(executor.State[any]{Block: 2}, ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// PostRun should finish up checking all remaining archive blocks, even if the
	// there are some blocks missing at the end of the range.
	if err := ext.PostRun(executor.State[any]{Block: 10}, ctx, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStateHashValidator_ValidatingLaggingArchivesIsSkippedIfRunIsAborted(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	hashProvider := substateDb.NewMockHashProvider(ctrl)

	db.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashA)), nil)
	hashProvider.EXPECT().GetStateRootHash(2).Return(types.Hash([]byte(exampleHashA)), nil)

	archive0 := state.NewMockNonCommittableStateDB(ctrl)

	gomock.InOrder(
		db.EXPECT().GetArchiveBlockHeight().Return(uint64(0), false, nil),
		hashProvider.EXPECT().GetStateRootHash(0).Return(types.Hash([]byte(exampleHashB)), nil),
		db.EXPECT().GetArchiveState(uint64(0)).Return(archive0, nil),
		archive0.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashB)), nil),
		archive0.EXPECT().Release(),
	)

	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	cfg.Last = 5
	cfg.ArchiveMode = true
	cfg.ArchiveVariant = "s5"

	ext := makeStateHashValidator[any](cfg, log)
	ext.hashProvider = hashProvider
	ctx := &executor.Context{State: db}

	// A PostBlock run should check the LiveDB and the ArchiveDB up to block 0.
	if err := ext.PostBlock(executor.State[any]{Block: 2}, ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// PostRun should finish up checking all remaining archive hashes and detect the error in block 2.
	if err := ext.PostRun(executor.State[any]{Block: 2}, ctx, fmt.Errorf("dummy")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
func TestStateHashValidator_PreRunAdjustFirstArchiveBlockToCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSubstateDb := substateDb.NewMockSubstateDB(ctrl)
	aidaDb, err := substateDb.NewDefaultSubstateDB(t.TempDir())
	require.NoError(t, err)
	cfg := &utils.Config{
		First:          3,
		Last:           5,
		DbImpl:         "carmen",
		CarmenSchema:   5,
		ArchiveMode:    true,
		ArchiveVariant: "s5",
	}
	ctx := &executor.Context{
		AidaDb: aidaDb,
	}
	input := utils.GetTestSubstate("default")
	input.Block = 4
	input.Transaction = 1

	mockSubstateDb.EXPECT().GetFirstSubstate().Return(input)

	// run
	ext := makeStateHashValidator[any](cfg, logger.NewLogger("Info", "TestPrime"))
	ext.sdb = mockSubstateDb
	err = ext.PreRun(executor.State[any]{}, ctx)
	assert.NoError(t, err)
	assert.Equal(t, 4, ext.nextArchiveBlockToCheck)

}
func TestStateHashValidator_PreRunReturnsErrorIfWrongDbImplIsChosen(t *testing.T) {
	cfg := &utils.Config{}
	cfg.DbImpl = "wrong"
	cfg.Last = 5

	ext := makeStateHashValidator[any](cfg, nil)

	// A PostBlock run should check the LiveDB and the ArchiveDB up to block 0.
	err := ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatal("pre run must return an error")
	}

	if strings.Compare(err.Error(), "state-hash-validation only works with db-impl carmen or geth") != 0 {
		t.Fatalf("unexpected err")
	}
}

func TestStateHashValidator_PreRunReturnsErrorIfWrongDbVariantIsChosen(t *testing.T) {
	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 3
	cfg.Last = 5

	ext := makeStateHashValidator[any](cfg, nil)

	// A PostBlock run should check the LiveDB and the ArchiveDB up to block 0.
	err := ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatal("pre run must return an error")
	}

	if strings.Compare(err.Error(), "state-hash-validation only works with carmen schema 5") != 0 {
		t.Fatalf("unexpected err")
	}
}

func TestStateHashValidator_PreRunReturnsErrorIfArchiveIsEnabledAndWrongVariantIsChosen(t *testing.T) {
	cfg := &utils.Config{}
	cfg.DbImpl = "carmen"
	cfg.CarmenSchema = 5
	cfg.Last = 5
	cfg.ArchiveMode = true
	cfg.ArchiveVariant = "wrong"

	ext := makeStateHashValidator[any](cfg, nil)

	// A PostBlock run should check the LiveDB and the ArchiveDB up to block 0.
	err := ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatal("pre run must return an error")
	}

	if strings.Compare(err.Error(), "archive state-hash-validation only works with archive variant s5") != 0 {
		t.Fatalf("unexpected err")
	}
}

func TestStateHashValidator_CheckArchiveHashesBlocksReturnsError(t *testing.T) {
	blockNumber := 1
	testcases := []struct {
		name      string
		output    map[int]*substate.Substate
		mockSetup func(sdb *substateDb.MockSubstateDB, log *logger.MockLogger, output map[int]*substate.Substate)
		hasError  bool
		errString string
	}{
		{
			name:   "empty block",
			output: make(map[int]*substate.Substate), // no transactions
			mockSetup: func(sdb *substateDb.MockSubstateDB, log *logger.MockLogger, output map[int]*substate.Substate) {
				sdb.EXPECT().GetBlockSubstates(uint64(blockNumber)).Return(output, nil)
				log.EXPECT().Warningf("Empty block %d has mismatch hash; %v", uint64(blockNumber), gomock.Any())
			},
			hasError: false,
		},
		{
			name: "unexpected hash error",
			output: map[int]*substate.Substate{
				0: {
					Block:       uint64(blockNumber),
					Transaction: 0,
				},
			},
			mockSetup: func(sdb *substateDb.MockSubstateDB, log *logger.MockLogger, output map[int]*substate.Substate) {
				sdb.EXPECT().GetBlockSubstates(uint64(blockNumber)).Return(output, nil)
			},
			hasError:  true,
			errString: "unexpected hash for archive block",
		},
		{
			name: "unexpected substate error",
			output: map[int]*substate.Substate{
				0: {
					Block:       uint64(blockNumber),
					Transaction: 0,
				},
			},
			mockSetup: func(sdb *substateDb.MockSubstateDB, log *logger.MockLogger, output map[int]*substate.Substate) {
				sdb.EXPECT().GetBlockSubstates(uint64(blockNumber)).Return(output, errors.New("substate error"))
			},
			hasError:  true,
			errString: "cannot get substates for block",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			log := logger.NewMockLogger(ctrl)
			db := state.NewMockStateDB(ctrl)
			hashProvider := substateDb.NewMockHashProvider(ctrl)
			baseDb := substateDb.NewMockBaseDB(ctrl)
			sdb := substateDb.NewMockSubstateDB(ctrl)

			cfg := &utils.Config{}
			cfg.DbImpl = "carmen"
			cfg.CarmenSchema = 5
			cfg.ArchiveMode = true
			cfg.ArchiveVariant = "s5"

			ext := makeStateHashValidator[any](cfg, log)
			ext.hashProvider = hashProvider
			ext.sdb = sdb

			archive := state.NewMockNonCommittableStateDB(ctrl)

			gomock.InOrder(
				// live state check goes through
				hashProvider.EXPECT().GetStateRootHash(blockNumber).Return(types.Hash([]byte(exampleHashA)), nil),
				db.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashA)), nil),
				db.EXPECT().GetArchiveBlockHeight().Return(uint64(blockNumber), false, nil),

				// archive state check goes through
				hashProvider.EXPECT().GetStateRootHash(0).Return(types.Hash([]byte(exampleHashA)), nil),
				db.EXPECT().GetArchiveState(uint64(0)).Return(archive, nil),
				archive.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashA)), nil),
				archive.EXPECT().Release(),

				// archive state check fails
				hashProvider.EXPECT().GetStateRootHash(blockNumber).Return(types.Hash([]byte(exampleHashA)), nil),
				db.EXPECT().GetArchiveState(uint64(blockNumber)).Return(archive, nil),
				archive.EXPECT().GetHash().Return(common.Hash([]byte(exampleHashB)), nil),
				archive.EXPECT().Release(),
			)

			// test variants
			tc.mockSetup(sdb, log, tc.output)

			ctx := &executor.Context{State: db, AidaDb: baseDb}

			err := ext.PostBlock(executor.State[any]{Block: blockNumber}, ctx)
			if tc.hasError {
				assert.Error(t, err, "post block must return error")
				assert.ErrorContains(t, err, tc.errString)
			} else {
				assert.NoError(t, err, "post block must not return error")
			}
		})
	}
}
