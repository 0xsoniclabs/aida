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

package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/ethereum/go-ethereum/rpc"
)

// TestStateHash_ZeroHasSameStateHashAsOne tests that the state hash of block 0 is the same as the state hash of block 1
func TestStateHash_ZeroHasSameStateHashAsOne(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashes"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	log := logger.NewLogger("info", "Test state hash")

	err = StateHashScraper(context.TODO(), TestnetChainID, "", database, 0, 1, log)
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

	shp := MakeStateHashProvider(database)

	hashZero, err := shp.GetStateHash(0)
	if err != nil {
		t.Fatalf("error getting state hash for block 0: %v", err)
	}

	hashOne, err := shp.GetStateHash(1)
	if err != nil {
		t.Fatalf("error getting state hash for block 1: %v", err)
	}

	if hashZero != hashOne {
		t.Fatalf("state hash of block 0 (%s) is not the same as the state hash of block 1 (%s)", hashZero.Hex(), hashOne.Hex())
	}
}

// TestStateHash_ZeroHasSameStateHashAsOne tests that the state hash of block 0 is different to the state hash of block 100
// we are expecting that at least some storage has changed between block 0 and block 100
func TestStateHash_ZeroHasDifferentStateHashAfterHundredBlocks(t *testing.T) {
	tmpDir := t.TempDir() + "/blockHashes"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	log := logger.NewLogger("info", "Test state hash")

	err = StateHashScraper(context.TODO(), TestnetChainID, "", database, 0, 100, log)
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

	shp := MakeStateHashProvider(database)

	hashZero, err := shp.GetStateHash(0)
	if err != nil {
		t.Fatalf("error getting state hash for block 0: %v", err)
	}

	hashHundred, err := shp.GetStateHash(100)
	if err != nil {
		t.Fatalf("error getting state hash for block 100: %v", err)
	}

	// block 0 should have a different state hash than block 100
	if hashZero == hashHundred {
		t.Fatalf("state hash of block 0 (%s) is the same as the state hash of block 100 (%s)", hashZero.Hex(), hashHundred.Hex())
	}
}

func TestStateHash_KeyToUint64(t *testing.T) {
	type args struct {
		hexBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"testZeroConvert", args{[]byte(StateHashPrefix + "0x0")}, 0, false},
		{"testOneConvert", args{[]byte(StateHashPrefix + "0x1")}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StateHashKeyToUint64(tt.args.hexBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("StateHashKeyToUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StateHashKeyToUint64() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getClient(t *testing.T) {
	type args struct {
		ctx     context.Context
		chainId ChainID
		ipcPath string
	}
	log := logger.NewLogger("info", "Test state hash")
	tests := []struct {
		name    string
		args    args
		want    *rpc.Client
		wantErr bool
	}{
		{"testGetClientRpcSonicMainnet", args{context.Background(), SonicMainnetChainID, ""}, &rpc.Client{}, false},
		{"testGetClientRpcOperaMainnet", args{context.Background(), MainnetChainID, ""}, &rpc.Client{}, false},
		{"testGetClientRpcTestnet", args{context.Background(), TestnetChainID, ""}, &rpc.Client{}, false},
		{"testGetClientIpcNonExistant", args{context.Background(), TestnetChainID, "/non-existant-path"}, nil, false},
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

func TestStateHash_GetStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDb := db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return([]byte("abcdefghijabcdefghijabcdefghij32"), nil)
	stateHash := MakeStateHashProvider(mockDb)
	hash, err := stateHash.GetStateHash(1234)
	assert.NoError(t, err)
	assert.Equal(t, "0x6162636465666768696a6162636465666768696a6162636465666768696a3332", hash.String())

	// case error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(nil, leveldb.ErrNotFound)
	stateHash = MakeStateHashProvider(mockDb)
	hash, err = stateHash.GetStateHash(1234)
	assert.Equal(t, leveldb.ErrNotFound, err)
	assert.Equal(t, common.Hash{}, hash)

	// case empty
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(nil, nil)
	stateHash = MakeStateHashProvider(mockDb)
	hash, err = stateHash.GetStateHash(1234)
	assert.NoError(t, err)
	assert.Equal(t, common.Hash{}, hash)
}

func TestStateHash_SaveStateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	mockDb := db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	err := SaveStateRoot(mockDb, "0x1234", "0x5678")
	assert.NoError(t, err)

	// case error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(leveldb.ErrNotFound)
	err = SaveStateRoot(mockDb, "0x1234", "0x5678")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "leveldb: not found")
}

func TestStateHash_StateHashKeyToUint64(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	output, err := StateHashKeyToUint64([]byte("dbh0x1234"))
	assert.NoError(t, err)
	assert.Equal(t, uint64(0x1234), output)

	// case error
	output, err = StateHashKeyToUint64([]byte("ggggggg"))
	assert.Equal(t, uint64(0), output)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")
}

func TestStateHash_retrieveStateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case success
	client := NewMockIRpcClient(ctrl)
	client.EXPECT().Call(gomock.Any(), "eth_getBlockByNumber", "0x1234", false).Return(nil)
	output, err := retrieveStateRoot(client, "0x1234")
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}(nil), output)

	// case error
	mockErr := errors.New("error")
	client = NewMockIRpcClient(ctrl)
	client.EXPECT().Call(gomock.Any(), "eth_getBlockByNumber", "0x1234", false).Return(mockErr)
	output, err = retrieveStateRoot(client, "0x1234")
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestStateHash_GetFirstStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	output, err := GetFirstStateHash(mockDb)
	assert.Equal(t, uint64(0x0), output)
	assert.Error(t, err)

}

func TestStateHash_GetLastStateHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	output, err := GetLastStateHash(mockDb)
	assert.Equal(t, uint64(0x0), output)
	assert.Error(t, err)
}
