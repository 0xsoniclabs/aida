package utils

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestNewProgressTracker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	tracker := NewProgressTracker(100, mockLogger)
	assert.Equal(t, 100, tracker.target)
}

func TestProgressTracker_PrintProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	tracker := NewProgressTracker(OperationThreshold, mockLogger)
	tracker.step = OperationThreshold - 1
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	for i := 0; i < 1; i++ {
		tracker.PrintProgress()
	}
}
