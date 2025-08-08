package dbutils

import (
	"errors"
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
