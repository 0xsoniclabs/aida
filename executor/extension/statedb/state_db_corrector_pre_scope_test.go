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

package statedb

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	substatedb "github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// dummy exception for testing
var testExceptionTx = &substate.Exception{
	Block: 1001,
	Data: substate.ExceptionBlock{
		Transactions: map[int]substate.ExceptionTx{
			1: {
				PreTransaction:  &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 1, Balance: uint256.NewInt(500)}},
				PostTransaction: &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 2, Balance: uint256.NewInt(1000)}},
			},
		},
	},
}

var testExceptionPreTx = &substate.Exception{
	Block: 1001,
	Data: substate.ExceptionBlock{
		Transactions: map[int]substate.ExceptionTx{
			1: {
				PreTransaction: &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 1, Balance: uint256.NewInt(500)}},
			},
		},
	},
}

// dummy exception for testing
var testExceptionBlock = &substate.Exception{
	Block: 1001,
	Data: substate.ExceptionBlock{
		Transactions: map[int]substate.ExceptionTx{},
		PreBlock:     &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 1, Balance: uint256.NewInt(500)}},
		PostBlock:    &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 1, Balance: uint256.NewInt(500)}},
	},
}

func TestStateDbCorrector_LoadCurrentException(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	corrector := &stateDbCorrector{}

	testcases := []struct {
		name         string
		block        int
		retException *substate.Exception
		retErr       error
	}{
		{"BlockHasException", 1001, testExceptionBlock, nil},
		{"BlockHasNoException", 1002, nil, nil},
		{"BlockHasNoException", 1003, nil, errors.New("err")},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			edb := substatedb.NewMockExceptionDB(ctrl)
			edb.EXPECT().GetException(uint64(test.block)).Return(test.retException, test.retErr).Times(1)
			corrector.db = edb

			err := corrector.loadCurrentException(test.block)
			if test.retErr != nil {
				assert.Error(t, err, "expected error when loading current exception")
			} else {
				assert.NoError(t, err)
			}

			if test.retException != nil {
				assert.NotNil(t, corrector.currentException, "expected currentException to be set")
				assert.Equal(t, test.block, int(corrector.currentException.Block), "expected currentException block to match")
				assert.NotNil(t, corrector.currentException.Data.PreBlock, "expected PreBlock to contain delta state")
				assert.NotNil(t, corrector.currentException.Data.PostBlock, "expected PostBlock to contain delta state")
			} else {
				assert.Nilf(t, corrector.currentException, "expected currentException to be nil, got %v", corrector.currentException)
			}
		})
	}
}

func TestStateDbCorrector_LoadStateFromException(t *testing.T) {
	corrector := &stateDbCorrector{}

	testcases := []struct {
		name      string
		exc       *substate.Exception
		scope     correctorScope
		wantState bool
		wantError bool
	}{
		{"RunPreBlockScopeReturnValid", testExceptionBlock, preBlock, true, false},
		{"RunPostBlockScopeReturnValid", testExceptionBlock, preBlock, true, false},
		{"RunPreBlockScopeReturnEmpty", testExceptionTx, preBlock, false, false},
		{"RunPostBlockScopeReturnEmpty", testExceptionTx, postBlock, false, false},
		{"RunPreTransactionScopeReturnValid", testExceptionTx, preTransaction, true, false},
		{"RunPostTransactionScopeReturnValid", testExceptionTx, postTransaction, true, false},
		{"RunPreTransactionScopeReturnEmpty", testExceptionBlock, preTransaction, false, false},
		{"RunPostTransactionScopeReturnEmpty", testExceptionPreTx, postTransaction, false, false},
		{"RunWrongScopeError", testExceptionBlock, correctorScope(100), false, true},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			corrector.currentException = test.exc
			ws, err := corrector.loadStateFromException(test.scope, 1)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if test.wantState {
				assert.NotNil(t, ws, "expected exception state to contain data")
			} else {
				assert.Nilf(t, ws, "expected exception state to be not nil, got %v", ws)
			}
		})
	}
}

func TestStateDbCorrector_FixExceptionAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	testcases := []struct {
		name         string
		scope        correctorScope
		block        int
		exception    *substate.Exception
		wantError    bool
		wantDbUpdate bool
	}{
		{"NoException", preBlock, 1001, nil, false, false},
		{"RunFixWrongBlock", preBlock, 1002, testExceptionBlock, true, false}, // Expect block 1001
		{"RunFixWrongScope", correctorScope(100), 1001, testExceptionBlock, true, false},
		{"RunFixScopeButHasNoFix", preTransaction, 1001, testExceptionBlock, false, false},
		{"RunFixTxScopeSuccessful", preTransaction, 1001, testExceptionTx, false, true},
		{"RunFixBlockScopeSuccessful", preBlock, 1001, testExceptionBlock, false, true},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			tx := 1 // Assume transaction index 1 for block scope
			if test.wantDbUpdate {
				if test.scope == preBlock || test.scope == postBlock {
					db.EXPECT().BeginTransaction(uint32(tx)).Return(nil).Times(1)
					db.EXPECT().EndTransaction().Return(nil).Times(1)
				}
				gomock.InOrder(
					db.EXPECT().Exist(common.Address{0x01}).Times(1),
					db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
					db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(100)).Times(1),
					db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(100), tracing.BalanceChangeUnspecified).Times(1),
					db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
					db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(2)).Times(1),
					db.EXPECT().SetNonce(common.Address{0x01}, uint64(1), tracing.NonceChangeUnspecified).Times(1),
					db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{}).Times(1),
				)
			}

			corrector := &stateDbCorrector{
				cfg:              &utils.Config{},
				currentException: test.exception,
			}
			err := corrector.fixExceptionAt(db, test.scope, test.block, tx)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateDbCorrector_FixExceptionAtWithBeginTransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	targetException := testExceptionBlock
	tx := 1 // Assume transaction index 1 for block scope

	db.EXPECT().BeginTransaction(uint32(tx)).Return(errors.New("err")).Times(1)

	corrector := &stateDbCorrector{
		cfg:              &utils.Config{},
		currentException: targetException,
	}
	err := corrector.fixExceptionAt(db, preBlock, int(targetException.Block), tx)
	assert.Error(t, err)
}

func TestStateDbCorrector_FixExceptionAtWithEndTransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	targetException := testExceptionBlock
	tx := 1 // Assume transaction index 1 for block scope

	gomock.InOrder(
		db.EXPECT().BeginTransaction(uint32(tx)).Return(nil).Times(1),
		db.EXPECT().Exist(common.Address{0x01}).Times(1),
		db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
		db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(100)).Times(1),
		db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(100), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(2)).Times(1),
		db.EXPECT().SetNonce(common.Address{0x01}, uint64(1), tracing.NonceChangeUnspecified).Times(1),
		db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{}).Times(1),
		db.EXPECT().EndTransaction().Return(errors.New("err")).Times(1),
	)

	corrector := &stateDbCorrector{
		cfg:              &utils.Config{},
		currentException: targetException,
	}
	err := corrector.fixExceptionAt(db, preBlock, int(targetException.Block), tx)
	assert.Error(t, err)
}

func TestStateDbCorrector_PreRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var err error
	ctx := &executor.Context{}

	testcases := []struct {
		name       string
		withAidaDb bool
	}{
		{"RunWithAidaDB", true},
		{"RunWihtoutAidaDB", false}, // test do not fail if aida-db is not set
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			ext := MakeStateDbCorrector(&utils.Config{})
			if test.withAidaDb {
				// Init directory for aida-db
				aidaDbDir := t.TempDir() + "aida-db"
				ctx.AidaDb, err = substatedb.NewDefaultBaseDB(aidaDbDir)
				assert.NoError(t, err, "failed to create aida-db")
			}
			err = ext.PreRun(executor.State[txcontext.TxContext]{}, ctx)
			assert.NoError(t, err)
		})
	}
}

func TestStateDbCorrector_PreBlockProcessesLastFixBlockInitialize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	edb := substatedb.NewMockExceptionDB(ctrl)

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		db:        edb,
		nextBlock: 0, // uninitialized lastFixedBlock
	}

	edb.EXPECT().GetException(uint64(3)).Return(nil, nil).Times(1)

	targetBlock := 3
	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: targetBlock}
	err = corrector.PreBlock(state, ctx)
	assert.NoError(t, err)
	assert.Equal(
		t,
		targetBlock+1,
		corrector.nextBlock,
		"expected nextBlock to be the block after, expected %d, got %d", targetBlock+1, corrector.nextBlock,
	)
}

func TestStateDbCorrector_PreBlockSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	edb := substatedb.NewMockExceptionDB(ctrl)
	targetBlock := 1001

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		db:        edb,
		nextBlock: 1000, // Assume nextBlock is 1000, meaning we need to process block 1000 and 1001 (target)
	}

	edb.EXPECT().GetException(uint64(1000)).Return(nil, nil).Times(1)
	edb.EXPECT().GetException(uint64(1001)).Return(testExceptionBlock, nil).Times(1)

	gomock.InOrder(
		db.EXPECT().BeginTransaction(uint32(utils.PseudoTx)).Return(nil).Times(1),
		db.EXPECT().Exist(common.Address{0x01}).Times(1),
		db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
		db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(100)).Times(1),
		db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(100), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(2)).Times(1),
		db.EXPECT().SetNonce(common.Address{0x01}, uint64(1), tracing.NonceChangeUnspecified).Times(1),
		db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{}).Times(1),
		db.EXPECT().EndTransaction().Return(nil).Times(1),
	)

	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: targetBlock}
	err = corrector.PreBlock(state, ctx)
	assert.NoError(t, err, "expected no error when processing block with exception")
	assert.NotNil(t, corrector.currentException, "expected currentException to be set")
	assert.Equal(t, corrector.currentException.Block, uint64(targetBlock), "expected currentException block to match")
	assert.Equal(
		t,
		targetBlock+1,
		corrector.nextBlock,
		"expected nextBlock to be the block after, expected %d, got %d", targetBlock+1, corrector.nextBlock,
	)
}

func TestStateDbCorrector_PreBlockFailsLoadCurrentException(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	edb := substatedb.NewMockExceptionDB(ctrl)

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		db:        edb,
		nextBlock: 0, // uninitialized lastFixedBlock
	}

	edb.EXPECT().GetException(uint64(3)).Return(nil, errors.New("err")).Times(1)

	targetBlock := 3
	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: targetBlock}
	err = corrector.PreBlock(state, ctx)
	assert.Error(t, err)
}

func TestStateDbCorrector_PreBlockFailsFixException(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)
	edb := substatedb.NewMockExceptionDB(ctrl)
	targetBlock := 1001

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		db:        edb,
		nextBlock: 1001,
	}

	edb.EXPECT().GetException(uint64(1001)).Return(testExceptionBlock, nil).Times(1)

	gomock.InOrder(
		db.EXPECT().BeginTransaction(uint32(utils.PseudoTx)).Return(errors.New("err")).Times(1),
	)

	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: targetBlock}
	err = corrector.PreBlock(state, ctx)
	assert.Error(t, err)
}

func TestStateDbCorrector_PreTransactionProcessesException(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		currentException: testExceptionTx,
	}

	gomock.InOrder(
		db.EXPECT().Exist(common.Address{0x01}).Times(1),
		db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
		db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(100)).Times(1),
		db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(100), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(1)).Times(1),
		db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{}).Times(1),
	)

	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: 1001, Transaction: 1}
	err = corrector.PreTransaction(state, ctx)
	assert.NoError(t, err)
}

func TestStateDbCorrector_PreTransactionFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		currentException: testExceptionTx,
	}

	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: 1000, Transaction: 1}
	// expect error because the block is not the same as the exception block
	err = corrector.PreTransaction(state, ctx)
	assert.Error(t, err)
}

func TestStateDbCorrector_PostBlockProcessesException(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		currentException: testExceptionBlock,
	}

	gomock.InOrder(
		db.EXPECT().BeginTransaction(uint32(utils.PseudoTx)).Return(nil).Times(1),
		db.EXPECT().Exist(common.Address{0x01}).Times(1),
		db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
		db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(100)).Times(1),
		db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(100), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(1)).Times(1),
		db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{}).Times(1),
		db.EXPECT().EndTransaction().Return(nil).Times(1),
	)

	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: 1001}
	err = corrector.PostBlock(state, ctx)
	assert.NoError(t, err)
	assert.Nil(t, corrector.currentException, "expected currentException to be reset after processing the block")
}

func TestStateDbCorrector_PostBlockFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Initialize aida-db directory
	aidaDbDir := t.TempDir() + "aida-db"
	aidaDb, err := substatedb.NewDefaultBaseDB(aidaDbDir)
	assert.NoError(t, err)

	corrector := &stateDbCorrector{
		currentException: testExceptionTx,
	}

	ctx := &executor.Context{AidaDb: aidaDb, State: db}
	state := executor.State[txcontext.TxContext]{Block: 1000, Transaction: 1}
	// expect error because the block is not the same as the exception block
	err = corrector.PostBlock(state, ctx)
	assert.Error(t, err)
}
