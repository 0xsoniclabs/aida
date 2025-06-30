package utildb

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUtils_StartOperaIpc_Timeout(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &utils.Config{
		ClientDb:    tmpDir,
		LogLevel:    "info",
		OperaBinary: "/bin/true", // This is a valid binary that should succeed
	}

	stopChan := make(chan struct{})
	errChan := startOperaIpc(cfg, stopChan)

	select {
	case err, _ := <-errChan:
		if err != nil {
			assert.Equal(t, "timeout waiting for opera ipc to start after 10s", err.Error())
		} else {
			assert.Error(t, err, "expected an error but got none")
		}
	case <-time.After(20 * time.Second):
		t.Error("timeout waiting for startOperaIpc to finish")
	}
}

func TestUtils_StartOperaIpc_InvalidBinary(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &utils.Config{
		ClientDb:    tmpDir,
		LogLevel:    "info",
		OperaBinary: "non-existent-binary", // This binary will fail
	}

	stopChan := make(chan struct{})
	errChan := startOperaIpc(cfg, stopChan)

	select {
	case err, ok := <-errChan:
		if !ok || err == nil {
			t.Error("expected an error but got none")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for startOperaIpc to finish")
	}
}

func TestUtils_StartOperaRecording(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &utils.Config{
		ClientDb:    tmpDir,
		LogLevel:    "info",
		OperaBinary: "/bin/true", // This is a valid binary that should succeed
	}

	errChan := startOperaRecording(cfg, 0)

	select {
	case err, ok := <-errChan:
		if ok && err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for startOperaRecording to finish")
	}
}

func TestUtils_StartOperaRecording_InvalidBinary(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &utils.Config{
		ClientDb:    tmpDir,
		LogLevel:    "info",
		OperaBinary: "non-existent-binary", // This binary will fail
	}

	errChan := startOperaRecording(cfg, 0)

	select {
	case err, ok := <-errChan:
		if !ok || err == nil {
			t.Error("expected an error but got none")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for startOperaRecording to finish")
	}
}

func TestUtils_IpcLoadingProcessWait_ResultSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	resChan := make(chan string, 1)
	errChan := make(chan error, 1)

	timer := time.NewTimer(10 * time.Second)
	waitDuration := 5 * time.Second

	log.EXPECT().Noticef("IPC endpoint opened")

	resChan <- "IPC endpoint opened"
	close(resChan)

	err := ipcLoadingProcessWait(resChan, errChan, timer, waitDuration, log)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

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
