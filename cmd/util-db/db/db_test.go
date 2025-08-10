package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
)

var testDataDir string

func TestMain(m *testing.M) {
	fmt.Println("Performing global setup...")

	// setup
	tempDir, err := os.MkdirTemp("", "profile_test_*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	testDataDir = tempDir
	err = utils.DownloadTestDataset(testDataDir)
	fmt.Printf("Downloaded test data: %s\n", testDataDir)
	if err != nil {
		fmt.Printf("Failed to download test dataset: %v\n", err)
		_ = os.RemoveAll(testDataDir)
		os.Exit(1)
	}

	// run
	exitCode := m.Run()

	// teardown
	err = os.RemoveAll(testDataDir)
	if err != nil {
		fmt.Printf("Failed to remove temp dir: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Performing global teardown...")
	os.Exit(exitCode)
}
