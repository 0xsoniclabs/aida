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
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"go.uber.org/mock/gomock"
)

func TestSubstateDumpTxTraceDiffFunc_DiffsEqual(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof("Block: %v Transaction: %v\n", uint64(1), 1)
	mockLogger.EXPECT().Debugf(gomock.Any())

	mockGetTransactionTraceFromRpc := func(block uint64, tx int) (map[string]interface{}, error) {
		return map[string]interface{}{
			"post": map[string]interface{}{
				types.Address{0x1}.String(): map[string]interface{}{
					"balance": "0x10",
					"code":    "0x6001600101",
					// unmarshaller guesses that int/string is float64 - so the expected value is in float64
					"nonce": float64(1),
					"storage": map[string]interface{}{
						types.Hash{0x0}.String(): types.Hash{0x1}.String(),
					},
				},
			},
			"pre": map[string]interface{}{
				types.Address{0x1}.String(): map[string]interface{}{
					"balance": "0x0",
				},
			},
		}, nil
	}

	mockSubstate := &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{0x1}: &substate.Account{
				Balance: big.NewInt(0),
				Nonce:   0,
				Code:    []byte{},
				Storage: map[types.Hash]types.Hash{},
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{0x1}: &substate.Account{
				Balance: big.NewInt(16),
				Nonce:   1,
				Code:    []byte{0x60, 0x01, 0x60, 0x01, 0x01},
				Storage: map[types.Hash]types.Hash{
					types.Hash{0x0}: {0x1},
				},
			},
		},
	}

	mockTaskPool := &db.SubstateTaskPool{}

	err := SubstateDumpTxTraceDiffFunc(mockGetTransactionTraceFromRpc, mockLogger)(1, 1, mockSubstate, mockTaskPool)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSubstateDumpTxTraceDiffFunc_DiffsNotEqual(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof("Block: %v Transaction: %v\n", uint64(1), 1)

	mockGetTransactionTraceFromRpc := func(block uint64, tx int) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	}

	mockSubstate := &substate.Substate{
		OutputSubstate: substate.WorldState{
			types.Address{0x1}: &substate.Account{
				Balance: big.NewInt(16),
			},
		},
	}

	mockTaskPool := &db.SubstateTaskPool{}

	err := SubstateDumpTxTraceDiffFunc(mockGetTransactionTraceFromRpc, mockLogger)(1, 1, mockSubstate, mockTaskPool)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
