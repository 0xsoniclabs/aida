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
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
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
		PreBlock:  &substate.WorldState{},
		PostBlock: &substate.WorldState{},
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
		PreBlock:  &substate.WorldState{},
		PostBlock: &substate.WorldState{},
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

// Tests the loadCurrentException method when the exception is not empty.
func TestStateDbCorrector_LoadCurrentExceptionNotEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cfg := &utils.Config{
		First: 1001,
	}

	log := logger.NewLogger(cfg.LogLevel, "TestStateDbCorrector")

	mockDB := db.NewMockExceptionDB(ctrl)
	mockDB.EXPECT().GetException(cfg.First).Return(testExceptionBlock, nil).Times(1)

	corrector := &stateDbCorrector{
		cfg:              cfg,
		log:              log,
		db:               mockDB,
		currentException: nil,
	}

	err := corrector.loadCurrentException(int(cfg.First))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if corrector.currentException == nil {
		t.Fatal("expected currentException to be set")
	}
}

// Tests the loadCurrentException method when the exception is empty.
func TestStateDbCorrector_LoadCurrentExceptionEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cfg := &utils.Config{
		First: 1001,
	}

	log := logger.NewLogger(cfg.LogLevel, "TestStateDbCorrector")

	mockDB := db.NewMockExceptionDB(ctrl)
	mockDB.EXPECT().GetException(cfg.First).Return(nil, nil).Times(1)

	corrector := &stateDbCorrector{
		cfg:              cfg,
		log:              log,
		db:               mockDB,
		currentException: nil,
	}
	err := corrector.loadCurrentException(int(cfg.First))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if corrector.currentException != nil {
		t.Fatal("expected currentException to be nil")
	}
}

// Write a test for loadStateFromException
func TestStateDbCorrector_LoadStateFromExceptionBlockScopes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &utils.Config{
		First: 1001,
	}

	log := logger.NewLogger(cfg.LogLevel, "TestStateDbCorrector")
	mockDB := db.NewMockExceptionDB(ctrl)

	corrector := &stateDbCorrector{
		cfg: cfg,
		log: log,
		db:  mockDB,
	}

	testcases := []struct {
		name      string
		exc       *substate.Exception
		scope     correctorScope
		wantState bool
		wantError bool
	}{
		{"RunPreBlockScopeReturnsValid", testExceptionBlock, preBlock, true, false},
		{"RunPostBlockScopeReturnsValid", testExceptionBlock, preBlock, true, false},
		{"RunPreBlockScopeReturnsEmpty", testExceptionTx, preBlock, false, false},
		{"RunPostBlockScopeReturnsEmpty", testExceptionTx, postBlock, false, false},
		{"RunPreTransactionScopeReturnsValid", testExceptionTx, preTransaction, true, false},
		{"RunPostTransactionScopeReturnsValid", testExceptionTx, postTransaction, true, false},
		{"RunPreTransactionScopeReturnsEmpty", testExceptionBlock, preTransaction, true, false},
		{"RunPostTransactionScopeReturnsEmpty", testExceptionPreTx, postTransaction, true, false},
		{"RunWrongScopeError", testExceptionBlock, correctorScope(100), false, true},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			corrector.currentException = test.exc
			ws, err := corrector.loadStateFromException(test.scope, 1)
			if test.wantError {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
			if test.wantState {
				if ws == nil {
					t.Fatal("expected world state to be not nil")
				}
			} else {
				if (ws != &substate.WorldState{}) {
					t.Fatalf("expected world state to be nil, got %v", ws)
				}
			}
		})
	}
	// Add more assertions based on the expected behavior of loadStateFromException
}
