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

package db

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_CompactCommand(t *testing.T) {
	// given - create a test leveldb database to compact
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-db")

	// Create a simple leveldb database with some data
	testDb, err := leveldb.New(dbPath, 1024, 100, "test", false)
	require.NoError(t, err)

	// Add some test data
	batch := testDb.NewBatch()
	for i := 0; i < 100; i++ {
		key := []byte(fmt.Sprintf("key-%d", i))
		value := []byte(fmt.Sprintf("value-%d", i))
		err := batch.Put(key, value)
		require.NoError(t, err)
	}
	err = batch.Write()
	require.NoError(t, err)
	err = testDb.Close()
	require.NoError(t, err)

	// Setup CLI app and command
	app := cli.NewApp()
	app.Commands = []*cli.Command{&CompactCommand}

	args := utils.NewArgs("test").
		Arg(CompactCommand.Name).
		Flag(utils.TargetDbFlag.Name, dbPath).
		Build()

	// when
	err = app.Run(args)

	// then
	assert.NoError(t, err)
}
