package utildb

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"testing"
)

func TestValidator_FindDbHashOnline(t *testing.T) {
	// Create a mock logger
	log := logger.NewLogger("INFO", "test")

	// Create a mock AidaDbMetadata
	md := &utils.AidaDbMetadata{}

	// Call the function with the mock data
	hash, err := FindDbHashOnline(utils.MainnetChainID, log, md)

	// Check for errors
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if the hash is not nil
	if hash == nil {
		t.Fatal("expected a non-nil hash")
	}
}
