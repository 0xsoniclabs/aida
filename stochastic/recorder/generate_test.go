// Copyright 2025 Fantom Foundation
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

package recorder

import (
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGenerateUniformRegistry_Basics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := &utils.Config{
		ContractNumber:    3,
		KeysNumber:        3,
		ValuesNumber:      3,
		SnapshotDepth:     4,
		BlockLength:       5,
		SyncPeriodLength:  6,
		TransactionLength: 7,
	}

	r, err := GenerateUniformStats(cfg, mockLogger)
	assert.NotNil(t, r)
	assert.Nil(t, err)

	for i := 0; i < cfg.SnapshotDepth; i++ {
		assert.Equal(t, uint64(1), r.snapshotFreq[i])
	}

	for i := 0; i < operations.NumArgOps; i++ {
		if operations.IsValidArgOp(i) {
			assert.NotZero(t, r.argOpFreq[i])
		}
	}

	bb, err := operations.EncodeArgOp(operations.BeginBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)
	bt, err := operations.EncodeArgOp(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), r.transitFreq[bb][bt])

	eb, err := operations.EncodeArgOp(operations.EndBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)
	es, err := operations.EncodeArgOp(operations.EndSyncPeriodID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)
	assert.Equal(t, cfg.SyncPeriodLength-1, r.transitFreq[eb][bb])
	assert.Equal(t, uint64(1), r.transitFreq[eb][es])

	gb, err := operations.EncodeArgOp(operations.GetBalanceID, stochastic.NewArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)
	et, err := operations.EncodeArgOp(operations.EndTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
	assert.Nil(t, err)
	if operations.IsValidArgOp(gb) {
		assert.Equal(t, uint64(1), r.transitFreq[gb][et])
	}
}
