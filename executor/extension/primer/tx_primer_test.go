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
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/prime"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTxPrimer_MakeTxPrimer(t *testing.T) {
	cfg := &config.Config{}
	ext := MakeTxPrimer(cfg)

	_, ok := ext.(*txPrimer)
	assert.True(t, ok)
}

func TestTxPrimer_PreRun(t *testing.T) {
	cfg := &config.Config{}
	ext := MakeTxPrimer(cfg)

	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1}
	ctx := &executor.Context{}

	err := ext.PreRun(st, ctx)
	assert.NoError(t, err)
}

func TestTxPrimer_PreTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := state.NewMockStateDB(ctrl)
	mockTxContext := txcontext.NewMockTxContext(ctrl)

	cfg := &config.Config{}
	log := logger.NewLogger(cfg.LogLevel, "test")
	ext := &txPrimer{
		primeCtx: prime.NewContext(cfg, mockDb, log),
		cfg:      cfg,
		log:      log,
	}
	alloc, _ := utils.MakeWorldState(t)
	ws := txcontext.NewWorldState(alloc)
	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: mockTxContext}
	ctx := &executor.Context{}
	mockErr := errors.New("mock error")

	mockTxContext.EXPECT().GetInputState().Return(ws)
	mockDb.EXPECT().StartBulkLoad(gomock.Any()).Return(nil, mockErr)

	err := ext.PreTransaction(st, ctx)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "mock error")
}
