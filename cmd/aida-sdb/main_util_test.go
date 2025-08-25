// Copyright 2024 Fantom Foundation
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
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	testTraceDir = "trace-test"
)

// TestMain runs global setup, test cases then global teardown
func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	err = teardown()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

// setup prepares
// substateDB and creates trace directory
func setup() error {
	// download and extract substate.test
	err := setupTestSubstateDB()
	if err != nil {
		return fmt.Errorf("unable to load substatedb. %v", err)
	}

	// create trace directory
	err = os.Mkdir(testTraceDir, 0700)
	if err != nil {
		return fmt.Errorf("unable to create direcotry. %v", err)
	}

	return nil
}

// teardown removes temp directories
func teardown() error {
	return errors.Join(os.RemoveAll("substate.test"), os.RemoveAll(testTraceDir))
}

// setupTestSubstateDB downloads compressed substates and extract in local directory
func setupTestSubstateDB() (finalErr error) {
	// download substate.test from url
	// set timeout to 1 minutes
	client := http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get("https://github.com/0xsoniclabs/substate-cli/releases/download/substate-test/substate.test.tar.gz")
	if err != nil {
		return err
	}

	// channel downloaded content to gzip stream
	gzipStream, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer func() {
		finalErr = errors.Join(finalErr, gzipStream.Close())
	}()

	tarReader := tar.NewReader(gzipStream)

	// decompress and store each file in archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// if head is a directory, create a new directory
		if header.Typeflag == tar.TypeDir {
			if err = os.MkdirAll(header.Name, 0700); err != nil {
				return err
			}
			// if not a directory, copy to out file
		} else {
			outFile, err := os.OpenFile(header.Name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}
			defer func() {
				finalErr = errors.Join(finalErr, outFile.Close())
			}()
			if _, err = io.Copy(outFile, tarReader); err != nil {
				return err
			}
		}
	}
	return nil
}
