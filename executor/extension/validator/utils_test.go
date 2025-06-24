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

package validator

import (
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/state"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUtils_UpdateStateDbOnEthereumChainFailsOnEndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Define the world state to overwrite
	ws := substate.WorldState{
		types.Address{0x01}: &substate.Account{
			Nonce:   1,
			Balance: uint256.NewInt(1000),
			Code:    []byte{0x60, 0x60},
			Storage: map[types.Hash]types.Hash{{0x01}: {0x02}},
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
		db.EXPECT().SetCode(common.Address{0x01}, []byte{0x60, 0x60}).Times(1),
		db.EXPECT().GetState(common.Address{0x01}, common.Hash{0x01}).Return(common.Hash{}).Times(1),
		db.EXPECT().SetState(common.Address{0x01}, common.Hash{0x01}, common.Hash{0x02}).Times(1),
		db.EXPECT().EndTransaction().Return(errors.New("err")).Times(1),
	)
	// Call the method to test
	err := updateStateDbOnEthereumChain(patch, db, true)
	assert.Error(t, err, "Expected an error when ending the transaction")
}

func TestUtils_UpdateStateDbOnEthereumChainFailsOnBeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Define the world state to overwrite
	ws := substate.WorldState{
		types.Address{0x01}: &substate.Account{
			Nonce:   1,
			Balance: uint256.NewInt(1000),
			Code:    []byte{0x60, 0x60},
			Storage: map[types.Hash]types.Hash{{0x01}: {0x02}},
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
		db.EXPECT().SetCode(common.Address{0x01}, []byte{0x60, 0x60}).Times(1),
		db.EXPECT().GetState(common.Address{0x01}, common.Hash{0x01}).Return(common.Hash{}).Times(1),
		db.EXPECT().SetState(common.Address{0x01}, common.Hash{0x01}, common.Hash{0x02}).Times(1),
		db.EXPECT().EndTransaction().Return(nil).Times(1),
		db.EXPECT().BeginTransaction(uint32(utils.PseudoTx)).Return(errors.New("err")).Times(1),
	)
	// Call the method to test
	err := updateStateDbOnEthereumChain(patch, db, true)
	assert.Error(t, err, "Expected an error when initializing a new transaction")
}
