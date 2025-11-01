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

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
)

var (
	testTraceDir = "trace-test"
	testDataDir  = "testdata"
)

// TestMain runs global setup, test cases then global teardown
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

// setup prepares
// substateDB and creates trace directory
func setup() {
	// create trace directory
	err := os.Mkdir(testTraceDir, 0700)
	if err != nil {
		fmt.Printf("unable to create direcotry. %v\n", err)
		os.Exit(1)
	}

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

	fmt.Printf("Setup completed\n")
}

// teardown removes temp directories
func teardown() {
	// Do something here.
	err := os.RemoveAll(testTraceDir)
	if err != nil {
		fmt.Printf("Failed to remove trace dir: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll("substate.test")
	if err != nil {
		fmt.Printf("Failed to remove substate dir: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(testDataDir)
	if err != nil {
		fmt.Printf("Failed to remove temp dir: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Teardown completed\n")
}
