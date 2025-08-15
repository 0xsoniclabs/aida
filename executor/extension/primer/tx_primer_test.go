package primer

import (
	"errors"
	"testing"

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
	cfg := &utils.Config{}
	ext := MakeTxPrimer(cfg)

	_, ok := ext.(*txPrimer)
	assert.True(t, ok)
}

func TestTxPrimer_PreRun(t *testing.T) {
	cfg := &utils.Config{}
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

	cfg := &utils.Config{}
	log := logger.NewLogger(cfg.LogLevel, "test")
	ext := &txPrimer{
		primeCtx: prime.NewPrimeContext(cfg, mockDb, log),
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
