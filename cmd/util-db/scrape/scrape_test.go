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

package scrape

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestCmd_ScrapeCommand(t *testing.T) {
	// given
	tmpDir := t.TempDir()
	targetDbPath := filepath.Join(tmpDir, "target-db")
	clientDbPath := filepath.Join(tmpDir, "client-db")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Flag(utils.TargetDbFlag.Name, targetDbPath).
		Flag(utils.ClientDbFlag.Name, clientDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.OperaMainnetChainID)).
		Arg("1"). // blockNumFirst
		Arg("5"). // blockNumLast - small range for testing
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

// TestStateHash_ZeroHasSameStateHashAsOne tests that the state hash of block 0 is the same as the state hash of block 1
func TestStateHash_ZeroHasSameStateHashAsOne(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashes"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	log := logger.NewLogger("info", "Test state hash")

	err = StateAndBlockHashScraper(context.TODO(), utils.OperaTestnetChainID, "", database, 0, 1, log)
	if err != nil {
		t.Fatalf("error scraping state hashes: %v", err)
	}
	err = database.Close()
	if err != nil {
		t.Fatalf("error closing stateHash leveldb %s: %v", tmpDir, err)
	}

	database, err = db.NewReadOnlyBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	defer func(database db.BaseDB) {
		e := database.Close()
		if e != nil {
			t.Fatalf("error closing stateHash leveldb %s: %v", tmpDir, e)
		}
	}(database)

	shp := utils.MakeHashProvider(database)

	hashZero, err := shp.GetStateRootHash(0)
	if err != nil {
		t.Fatalf("error getting state hash for block 0: %v", err)
	}

	hashOne, err := shp.GetStateRootHash(1)
	if err != nil {
		t.Fatalf("error getting state hash for block 1: %v", err)
	}

	if hashZero != hashOne {
		t.Fatalf("state hash of block 0 (%s) is not the same as the state hash of block 1 (%s)", hashZero.Hex(), hashOne.Hex())
	}
}

func TestStateHash_Log(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashes"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Infof("Connected to RPC at %s", utils.RPCTestnet)
	log.EXPECT().Infof("Scraping block %d done!\n", uint64(10000))

	err = StateAndBlockHashScraper(context.TODO(), utils.OperaTestnetChainID, "", database, 9990, 10100, log)
	if err != nil {
		t.Fatalf("error scraping state hashes: %v", err)
	}
}

// TestStateHash_ZeroHasSameStateHashAsOne tests that the state hash of block 0 is different to the state hash of block 100
// we are expecting that at least some storage has changed between block  and block 100
func TestStateHash_ZeroHasDifferentStateHashAfterHundredBlocks(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashes"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	log := logger.NewLogger("info", "Test state hash")

	err = StateAndBlockHashScraper(context.TODO(), utils.OperaTestnetChainID, "", database, 0, 100, log)
	if err != nil {
		t.Fatalf("error scraping state hashes: %v", err)
	}
	err = database.Close()
	if err != nil {
		t.Fatalf("error closing stateHash leveldb %s: %v", tmpDir, err)
	}

	database, err = db.NewReadOnlyBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	defer func(database db.BaseDB) {
		e := database.Close()
		if e != nil {
			t.Fatalf("error closing stateHash leveldb %s: %v", tmpDir, e)
		}
	}(database)

	shp := utils.MakeHashProvider(database)

	hashZero, err := shp.GetStateRootHash(0)
	if err != nil {
		t.Fatalf("error getting state hash for block 0: %v", err)
	}

	hashHundred, err := shp.GetStateRootHash(100)
	if err != nil {
		t.Fatalf("error getting state hash for block 100: %v", err)
	}

	// block 0 should have a different state hash than block 100
	if hashZero == hashHundred {
		t.Fatalf("state hash of block 0 (%s) is the same as the state hash of block 100 (%s)", hashZero.Hex(), hashHundred.Hex())
	}
}

func Test_getClient(t *testing.T) {
	type args struct {
		ctx     context.Context
		chainId utils.ChainID
		ipcPath string
	}
	log := logger.NewLogger("info", "Test state hash")
	tests := []struct {
		name    string
		args    args
		want    *rpc.Client
		wantErr bool
	}{
		{"testGetClientRpcSonicMainnet", args{context.Background(), utils.SonicMainnetChainID, ""}, &rpc.Client{}, false},
		{"testGetClientRpcOperaMainnet", args{context.Background(), utils.OperaMainnetChainID, ""}, &rpc.Client{}, false},
		{"testGetClientRpcTestnet", args{context.Background(), utils.OperaTestnetChainID, ""}, &rpc.Client{}, false},
		{"testGetClientIpcNonExistant", args{context.Background(), utils.OperaTestnetChainID, "/non-existant-path"}, nil, false},
		{"testGetClientRpcUnknownChainId", args{context.Background(), 88888, ""}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getClient(tt.args.ctx, tt.args.chainId, tt.args.ipcPath, log)
			if (err != nil) != tt.wantErr {
				t.Errorf("getClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got == nil {
				t.Errorf("getClient() got nil, want non-nil")
			}
		})
	}
}

func TestStateHash_GetClientIpcFail(t *testing.T) {
	tmpIpcPath := t.TempDir()
	// create this file
	if err := os.WriteFile(tmpIpcPath+"/geth.ipc", []byte("test"), 0644); err != nil {
		t.Fatalf("error creating ipc file %s: %v", tmpIpcPath+"/geth.ipc", err)
	}

	log := logger.NewLogger("info", "Test state hash")
	_, err := getClient(context.Background(), utils.OperaTestnetChainID, tmpIpcPath, log)
	if err == nil {
		t.Fatalf("expected error when trying to connect to ipc file %s, but got nil", tmpIpcPath)
	}
	assert.Equal(t, fmt.Sprintf("failed to connect to IPC at %v/geth.ipc: dial unix %v/geth.ipc: connect: connection refused", tmpIpcPath, tmpIpcPath), err.Error())
}

func TestStateHash_SaveStateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDb := db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	err := utils.SaveStateRoot(mockDb, "0x1234", "0x5678")
	assert.NoError(t, err)

	// case error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(leveldb.ErrNotFound)
	err = utils.SaveStateRoot(mockDb, "0x1234", "0x5678")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "leveldb: not found")
}

func TestStateHash_retrieveStateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	client := NewMockIRpcClient(ctrl)
	client.EXPECT().Call(gomock.Any(), "eth_getBlockByNumber", "0x1234", false).Return(nil)
	output, err := getBlockByNumber(client, "0x1234")
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}(nil), output)

	// case error
	mockErr := errors.New("error")
	client = NewMockIRpcClient(ctrl)
	client.EXPECT().Call(gomock.Any(), "eth_getBlockByNumber", "0x1234", false).Return(mockErr)
	output, err = getBlockByNumber(client, "0x1234")
	assert.Error(t, err)
	assert.Nil(t, output)
}
