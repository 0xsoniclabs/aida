package dbutils

import (
	"errors"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUtils_ErrorRelayer_Success(t *testing.T) {
	errChan := make(chan error, 1)
	errChanParser := make(chan error, 1)
	resChan := make(chan string)

	go errorRelayer(resChan, errChan, errChanParser)

	// testing that result is non-blocking
	resChan <- "test result"
	close(errChan)

	select {
	case err := <-errChanParser:
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for errorRelayer to finish")
	}
}

func TestUtils_ErrorRelayer_Fail(t *testing.T) {
	errChan := make(chan error, 1)
	errChanParser := make(chan error, 1)
	resChan := make(chan string, 1)

	errMsg := "test error"

	errChan <- errors.New(errMsg)
	close(errChan)

	errorRelayer(resChan, errChan, errChanParser)

	select {
	case err := <-errChanParser:
		if err != nil {
			assert.Equal(t, "opera error after ipc initialization; "+errMsg, err.Error(), "expected error message to match")
		} else {
			t.Fatal("expected an error but got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for errorRelayer to finish")
	}
}

func TestMustCloseDB(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "CloseDBSuccess",
			wantErr: nil,
		},
		{
			name:    "CloseDBError",
			wantErr: leveldb.ErrClosed,
		},
		{
			name:    "CloseDBPanic",
			wantErr: errors.New("mock err"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockDb := db.NewMockBaseDB(ctrl)
			mockDb.EXPECT().Close().Return(test.wantErr)
			MustCloseDB(mockDb)
		})
	}

}
func TestPrintMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testDb, tmpDir := GenerateTestAidaDb(t)
	require.NoError(t, testDb.Close())
	err := PrintMetadata(tmpDir)
	assert.NoError(t, err)
}
