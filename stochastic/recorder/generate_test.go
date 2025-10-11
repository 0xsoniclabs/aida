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
		BalanceRange:      10,
		NonceRange:        8,
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

	midBalance := cfg.BalanceRange / 2
	if midBalance <= 0 {
		midBalance = 1
	}
	assert.Equal(t, uint64(1), r.balance.freq[0])
	assert.Equal(t, uint64(1), r.balance.freq[midBalance])
	assert.Equal(t, uint64(1), r.balance.freq[cfg.BalanceRange-1])

	assert.Equal(t, uint64(1), r.nonce.freq[0])
	assert.Equal(t, uint64(1), r.nonce.freq[int64(cfg.NonceRange/2)])
	assert.Equal(t, uint64(1), r.nonce.freq[int64(cfg.NonceRange-1)])

	assert.Equal(t, uint64(1), r.code.freq[1])
	assert.Equal(t, uint64(1), r.code.freq[1024])
	assert.Equal(t, uint64(1), r.code.freq[24576])
}

func TestGenerateUniformStats_ValidationErrors(t *testing.T) {
	baseCfg := utils.Config{
		ContractNumber:    1,
		KeysNumber:        1,
		ValuesNumber:      1,
		SnapshotDepth:     1,
		BlockLength:       1,
		SyncPeriodLength:  1,
		TransactionLength: 1,
	}

	testCases := []struct {
		name   string
		mutate func(cfg *utils.Config)
		errMsg string
	}{
		{
			name: "ZeroBlockLength",
			mutate: func(cfg *utils.Config) {
				cfg.BlockLength = 0
			},
			errMsg: "block-length",
		},
		{
			name: "ZeroSyncPeriod",
			mutate: func(cfg *utils.Config) {
				cfg.SyncPeriodLength = 0
			},
			errMsg: "sync-period",
		},
		{
			name: "ZeroTransactionLength",
			mutate: func(cfg *utils.Config) {
				cfg.TransactionLength = 0
			},
			errMsg: "transaction-length",
		},
		{
			name: "ZeroContracts",
			mutate: func(cfg *utils.Config) {
				cfg.ContractNumber = 0
			},
			errMsg: "num-contracts",
		},
		{
			name: "ZeroKeys",
			mutate: func(cfg *utils.Config) {
				cfg.KeysNumber = 0
			},
			errMsg: "num-keys",
		},
		{
			name: "ZeroValues",
			mutate: func(cfg *utils.Config) {
				cfg.ValuesNumber = 0
			},
			errMsg: "num-values",
		},
		{
			name: "ZeroSnapshotDepth",
			mutate: func(cfg *utils.Config) {
				cfg.SnapshotDepth = 0
			},
			errMsg: "snapshot-depth",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := baseCfg
			tc.mutate(&cfg)

			mockLogger := logger.NewMockLogger(ctrl)
			mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(0)

			_, err := GenerateUniformStats(&cfg, mockLogger)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestGenerateUniformStats_EncodeErrorsBubbleUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	original := operations.OpNumArgs[operations.BeginBlockID]
	operations.OpNumArgs[operations.BeginBlockID] = 4
	t.Cleanup(func() {
		operations.OpNumArgs[operations.BeginBlockID] = original
	})

	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := &utils.Config{
		ContractNumber:    1,
		KeysNumber:        1,
		ValuesNumber:      1,
		SnapshotDepth:     1,
		BlockLength:       1,
		SyncPeriodLength:  1,
		TransactionLength: 1,
	}
	stats, err := GenerateUniformStats(cfg, mockLogger)
	assert.Nil(t, stats)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operation/arguments")
}
