package recorder

import (
    "testing"

    "github.com/0xsoniclabs/aida/logger"
    "github.com/0xsoniclabs/aida/stochastic/operations"
    "github.com/0xsoniclabs/aida/stochastic/statistics/classifier"
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
        ContractNumber:     3,
        KeysNumber:         3,
        ValuesNumber:       3,
        SnapshotDepth:      4,
        BlockLength:        5,
        SyncPeriodLength:   6,
        TransactionLength:  7,
    }

    r := GenerateUniformRegistry(cfg, mockLogger)
    assert.NotNil(t, r)

    for i := 0; i < cfg.SnapshotDepth; i++ {
        assert.Equal(t, uint64(1), r.snapshotFreq[i])
    }

    for i := 0; i < operations.NumArgOps; i++ {
        if operations.IsValidArgOp(i) {
            assert.NotZero(t, r.argOpFreq[i])
        }
    }

    bb := operations.EncodeArgOp(operations.BeginBlockID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
    bt := operations.EncodeArgOp(operations.BeginTransactionID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
    assert.Equal(t, uint64(1), r.transitFreq[bb][bt])

    eb := operations.EncodeArgOp(operations.EndBlockID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
    es := operations.EncodeArgOp(operations.EndSyncPeriodID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
    assert.Equal(t, cfg.SyncPeriodLength-1, r.transitFreq[eb][bb])
    assert.Equal(t, uint64(1), r.transitFreq[eb][es])

    gb := operations.EncodeArgOp(operations.GetBalanceID, classifier.NewArgID, classifier.NoArgID, classifier.NoArgID)
    et := operations.EncodeArgOp(operations.EndTransactionID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
    if operations.IsValidArgOp(gb) {
        assert.Equal(t, uint64(1), r.transitFreq[gb][et])
    }
}

