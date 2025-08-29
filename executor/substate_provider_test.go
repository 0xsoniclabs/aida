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
package executor

import (
	"errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestSubstateProvider_IterateOverExistingDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockTxConsumer(ctrl)

	// Prepare a directory containing some substate data.
	path := t.TempDir()
	if err := createSubstateDb(t, path); err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}

	// Open the substate data for reading.
	provider := openSubstateDb(path, t)
	defer provider.Close()

	gomock.InOrder(
		consumer.EXPECT().Consume(10, 7, gomock.Any()),
		consumer.EXPECT().Consume(10, 9, gomock.Any()),
		consumer.EXPECT().Consume(12, 5, gomock.Any()),
	)

	if err := provider.Run(0, 20, toSubstateConsumer(consumer)); err != nil {
		t.Fatalf("failed to iterate through states: %v", err)
	}
}

func TestSubstateProvider_LowerBoundIsInclusive(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockTxConsumer(ctrl)

	// Prepare a directory containing some substate data.
	path := t.TempDir()
	if err := createSubstateDb(t, path); err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}

	// Open the substate data for reading.
	provider := openSubstateDb(path, nil)
	defer provider.Close()

	gomock.InOrder(
		consumer.EXPECT().Consume(10, 7, gomock.Any()),
		consumer.EXPECT().Consume(10, 9, gomock.Any()),
		consumer.EXPECT().Consume(12, 5, gomock.Any()),
	)

	if err := provider.Run(10, 20, toSubstateConsumer(consumer)); err != nil {
		t.Fatalf("failed to iterate through states: %v", err)
	}
}

func TestSubstateProvider_UpperBoundIsExclusive(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockTxConsumer(ctrl)

	// Prepare a directory containing some substate data.
	path := t.TempDir()
	if err := createSubstateDb(t, path); err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}

	// Open the substate data for reading.
	provider := openSubstateDb(path, nil)
	defer provider.Close()

	gomock.InOrder(
		consumer.EXPECT().Consume(10, 7, gomock.Any()),
		consumer.EXPECT().Consume(10, 9, gomock.Any()),
	)

	if err := provider.Run(10, 12, toSubstateConsumer(consumer)); err != nil {
		t.Fatalf("failed to iterate through states: %v", err)
	}
}

func TestSubstateProvider_RangeCanBeEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockTxConsumer(ctrl)

	// Prepare a directory containing some substate data.
	path := t.TempDir()
	if err := createSubstateDb(t, path); err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}

	// Open the substate data for reading.
	provider := openSubstateDb(path, nil)
	defer provider.Close()

	if err := provider.Run(5, 10, toSubstateConsumer(consumer)); err != nil {
		t.Fatalf("failed to iterate through states: %v", err)
	}
}

func TestSubstateProvider_IterationCanBeAbortedByConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockTxConsumer(ctrl)

	// Prepare a directory containing some substate data.
	path := t.TempDir()
	if err := createSubstateDb(t, path); err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}

	// Open the substate data for reading.
	provider := openSubstateDb(path, nil)
	defer provider.Close()

	stop := errors.New("stop!")
	gomock.InOrder(
		consumer.EXPECT().Consume(10, 7, gomock.Any()),
		consumer.EXPECT().Consume(10, 9, gomock.Any()).Return(stop),
	)

	if got, want := provider.Run(10, 20, toSubstateConsumer(consumer)), stop; !errors.Is(got, want) {
		t.Errorf("provider run did not finish with expected exception, wanted %d, got %d", want, got)
	}
}

func openSubstateDb(path string, t *testing.T) Provider[txcontext.TxContext] {
	cfg := utils.Config{}
	cfg.AidaDb = path
	cfg.Workers = 1
	aidaDb, err := db.NewReadOnlyBaseDB(path)
	if err != nil {
		t.Fatal(err)
	}
	return OpenSubstateProvider(&cfg, nil, aidaDb)
}

func createSubstateDb(t *testing.T, path string) error {
	sdb, err := db.NewDefaultSubstateDB(path)
	if err != nil {
		t.Fatal(err)
	}
	state := substate.Substate{
		Block:       10,
		Transaction: 7,
		Env: &substate.Env{
			Number:     11,
			Difficulty: big.NewInt(1),
			GasLimit:   uint64(15),
		},
		Message: &substate.Message{
			Value:    big.NewInt(12),
			GasPrice: big.NewInt(14),
		},
		InputSubstate:  substate.WorldState{},
		OutputSubstate: substate.WorldState{},
		Result:         &substate.Result{},
	}

	err = sdb.PutSubstate(&state)
	if err != nil {
		t.Fatal(err)
	}

	state.Block = 10
	state.Transaction = 9
	err = sdb.PutSubstate(&state)
	if err != nil {
		t.Fatal(err)
	}

	state.Block = 12
	state.Transaction = 5
	err = sdb.PutSubstate(&state)
	if err != nil {
		t.Fatal(err)
	}

	err = sdb.Close()
	if err != nil {
		t.Fatal(err)
	}
	return nil
}

func TestExecutor_OpenSubstateProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		cfg := &utils.Config{
			AidaDb: "testdb",
		}
		ctxt := cli.NewContext(nil, nil, nil)
		kv := &testutil.KeyValue{}

		mockBaseDb := db.NewMockBaseDB(ctrl)
		mockDb := db.NewMockDbAdapter(ctrl)
		// Try catch mechanism for finding encoding
		mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).MinTimes(1)
		mockBaseDb.EXPECT().GetBackend().Return(mockDb)

		provider := OpenSubstateProvider(cfg, ctxt, mockBaseDb)
		assert.NotNil(t, provider)
	})
}

func TestSubstateProvider_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	mockIter := db.NewMockIIterator[*substate.Substate](ctrl)
	mockDb.EXPECT().NewSubstateIterator(0, 0).Return(mockIter)
	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Next().Return(false)
	mockIter.EXPECT().Value().Return(&substate.Substate{
		Block: 0,
	})
	mockIter.EXPECT().Release().Return()
	mockIter.EXPECT().Error().Return(nil)

	provider := &substateProvider{
		db: mockDb,
	}
	err := provider.Run(0, 1, func(info TransactionInfo[txcontext.TxContext]) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestSubstateProvider_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	provider := &substateProvider{
		db: mockDb,
	}
	assert.NotPanics(t, func() {
		provider.Close()
	})
}
