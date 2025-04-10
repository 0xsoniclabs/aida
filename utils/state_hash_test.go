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
	"testing"

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

	err = StateHashScraper(nil, TestnetChainID, "", database, 0, 1, log)
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
	defer database.Close()

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

	err = StateHashScraper(nil, TestnetChainID, "", database, 0, 100, log)
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
	defer database.Close()

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
