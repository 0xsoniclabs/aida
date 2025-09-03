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

package primer

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/rlp"
	trlp "github.com/0xsoniclabs/substate/types/rlp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

func TestStateDbPrimerExtension_NoPrimerIsCreatedIfDisabled(t *testing.T) {
	cfg := &config.Config{}
	cfg.SkipPriming = true

	ext := MakeStateDbPrimer[any](cfg)
	if _, ok := ext.(extension.NilExtension[any]); !ok {
		t.Errorf("Primer is enabled although not set in configuration")
	}

}

func TestStateDbPrimerExtension_PrimingDoesTriggerForExistingStateDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAidaDb := db.NewMockBaseDB(ctrl)
	mockAdapter := db.NewMockDbAdapter(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)

	cfg := &config.Config{
		IsExistingStateDb: true,
		First:             10,
	}

	log := logger.NewLogger("Info", "Test")
	input := utils.GetTestSubstate("default")
	input.Block = 9
	input.Transaction = 1
	encoded, err := trlp.EncodeToBytes(rlp.NewRLP(input))
	if err != nil {
		t.Fatalf("Failed to encode substate: %v", err)
	}

	kv := &testutil.KeyValue{}
	kv.PutU(db.SubstateDBKey(input.Block, input.Transaction), encoded)
	iter := iterator.NewArrayIterator(kv)

	// start priming
	mockAidaDb.EXPECT().GetBackend().Return(mockAdapter).AnyTimes()
	mockAdapter.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iter).AnyTimes()
	// loadExistingAccountsIntoCache is executed only if an existing db is used
	mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().EndTransaction().Return(nil)
	mockStateDb.EXPECT().EndBlock().Return(nil)
	// primeContext block starts from 1 because StateDbInfo file doesn't exist and db block is assumed to be 0.
	// The first primable block is 1 (the next block).
	// primeContext block gets incremented in loadExistingAccountsIntoCache to 2.
	mockStateDb.EXPECT().StartBulkLoad(uint64(2)).Return(nil, errors.New("stop"))

	ext := makeStateDbPrimer[any](cfg, log)
	err = ext.PreRun(executor.State[any]{}, &executor.Context{AidaDb: mockAidaDb, State: mockStateDb})
	assert.Error(t, err)
	assert.ErrorContains(t, err, "stop")
}

func TestStateDbPrimerExtension_PrimingDoesTriggerForNonExistingStateDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	stateDb := state.NewMockStateDB(ctrl)
	aidaDbPath := t.TempDir() + "aidadb"

	cfg := &config.Config{
		SkipPriming: false,
		StateDbSrc:  "",
		First:       10,
	}

	gomock.InOrder(
		log.EXPECT().Infof("Update buffer size: %v bytes", cfg.UpdateBufferSize),
		log.EXPECT().Warning("cannot get first substate; substate db is empty"),
		log.EXPECT().Noticef("Priming from block %v...", uint64(0)),
		log.EXPECT().Noticef("Priming to block %v...", cfg.First-1),
		log.EXPECT().Debugf("\tLoading %d accounts with %d values ..", 0, 0),
		stateDb.EXPECT().StartBulkLoad(uint64(0)).Return(nil, errors.New("stop")),
	)

	ext := makeStateDbPrimer[any](cfg, log)

	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
	assert.NoError(t, err, "cannot open test aida-db")
	err = ext.PreRun(executor.State[any]{}, &executor.Context{AidaDb: aidaDb, State: stateDb})
	assert.Error(t, err)
	assert.ErrorContains(t, err, "stop")
}

func TestStateDbPrimerExtension_NoBlockToPrime_Skip(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	tmpStateDb := t.TempDir()

	cfg := &config.Config{
		SkipPriming:       false,
		StateDbSrc:        tmpStateDb,
		IsExistingStateDb: true,
		First:             2,
	}

	err := utils.WriteStateDbInfo(tmpStateDb, cfg, 1, common.Hash{}, false)
	if err != nil {
		t.Fatalf("cannot write state db info: %v", err)
	}

	ext := makeStateDbPrimer[any](cfg, log)

	log.EXPECT().Infof("Update buffer size: %v bytes", cfg.UpdateBufferSize)
	log.EXPECT().Debugf("skipping priming; first priming block %v; first block %v", uint64(2), uint64(2))

	err = ext.PreRun(executor.State[any]{}, &executor.Context{})
	assert.NoError(t, err, "PreRun should not return an error when there is no block to prime")
}

func TestStateDbPrimerExtension_UserIsInformedAboutRandomPriming(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	aidaDbPath := t.TempDir() + "aidadb"
	stateDb := state.NewMockStateDB(ctrl)

	cfg := &config.Config{}
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
		log.EXPECT().Warning("cannot get first substate; substate db is empty"),
		log.EXPECT().Noticef("Priming from block %v...", uint64(0)),
		log.EXPECT().Noticef("Priming to block %v...", uint64(9)),
		log.EXPECT().Debugf("\tLoading %d accounts with %d values ..", 0, 0),
		stateDb.EXPECT().StartBulkLoad(uint64(0)).Return(nil, errors.New("stop")),
	)

	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
	assert.NoError(t, err)

	err = ext.PreRun(executor.State[any]{}, &executor.Context{AidaDb: aidaDb, State: stateDb})
	assert.Error(t, err)
	assert.ErrorContains(t, err, "stop")
}
