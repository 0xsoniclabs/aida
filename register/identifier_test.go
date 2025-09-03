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

package register

import (
	"strconv"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/config"
	"github.com/stretchr/testify/assert"
)

func TestIdentity_SameIdIfSameRun(t *testing.T) {
	cfg := &config.Config{}
	cfg.DbImpl = "carmen"
	cfg.DbVariant = "go-file"
	cfg.CarmenSchema = 5
	cfg.VmImpl = "lfvm"

	timestamp := time.Now().Unix()

	info := &RunIdentity{timestamp, cfg}

	//Same info = same id
	i, err := info.GetId()
	if err != nil {
		t.Fatalf("Failed to get ID: %v", err)
	}
	j, err := info.GetId()
	if err != nil {
		t.Fatalf("Failed to get ID: %v", err)
	}
	if i != j {
		t.Errorf("Same Info but ID doesn't matched: %s vs %s", i, j)
	}

	//Same timestamp, cfg = same id
	info2 := &RunIdentity{timestamp, cfg}
	k, err := info2.GetId()
	if err != nil {
		t.Fatalf("Failed to get ID: %v", err)
	}
	if i != k {
		t.Errorf("Same timestamp, cfg but ID doesn't matched: %s vs %s", i, k)
	}
}

func TestIdentity_DiffIdIfDiffRun(t *testing.T) {
	cfg := &config.Config{}
	cfg.DbImpl = "carmen"
	cfg.DbVariant = "go-file"
	cfg.CarmenSchema = 5
	cfg.VmImpl = "lfvm"

	cfg2 := &config.Config{}
	cfg2.DbImpl = "carmen"
	cfg2.DbVariant = "go-file"
	cfg2.CarmenSchema = 3
	cfg2.VmImpl = "geth"

	timestamp := time.Now().Unix()
	timestamp2 := timestamp + 10_000

	info := &RunIdentity{timestamp, cfg}
	info2 := &RunIdentity{timestamp2, cfg}
	info3 := &RunIdentity{timestamp, cfg2}

	infoId1, err := info.GetId()
	if err != nil {
		t.Fatalf("Failed to get ID: %v", err)
	}
	infoId2, err := info2.GetId()
	if err != nil {
		t.Fatalf("Failed to get ID: %v", err)
	}
	infoId3, err := info3.GetId()
	if err != nil {
		t.Fatalf("Failed to get ID: %v", err)
	}

	//Different timestamp = Different Id
	if infoId1 == infoId2 {
		t.Errorf("Different timestamp but ID still matched: %s vs %s", infoId1, infoId2)
	}

	//Different cfg = Different Id
	if infoId1 == infoId3 {
		t.Errorf("Different cfg but ID still matched: %s vs %s", infoId1, infoId3)
	}

	//Different everything = different Id
	if infoId2 == infoId3 {
		t.Errorf("Different info but ID still matched: %s vs %s", infoId2, infoId3)
	}
}

func TestIdentity_OverwriteRunIdWorks(t *testing.T) {
	cfg := &config.Config{}
	cfg.DbImpl = "carmen"
	cfg.DbVariant = "go-file"
	cfg.CarmenSchema = 5
	cfg.VmImpl = "lfvm"
	cfg.OverwriteRunId = "DummyTest"

	timestamp := time.Now().Unix()

	info := &RunIdentity{timestamp, cfg}

	s, err := info.GetId()
	if err != nil {
		t.Fatalf("Failed to get ID: %v", err)
	}
	if s != cfg.OverwriteRunId {
		t.Errorf("RunId should be overwritten as %s but is %s", s, cfg.OverwriteRunId)
	}
}

func TestRegister_MakeRunIdentity(t *testing.T) {
	cfg := &config.Config{
		DbImpl:         "carmen",
		DbVariant:      "go-file",
		CarmenSchema:   5,
		VmImpl:         "lfvm",
		OverwriteRunId: "TestRunId",
	}

	timestamp := time.Now().Unix()
	runIdentity := MakeRunIdentity(timestamp, cfg)

	assert.Equal(t, timestamp, runIdentity.Timestamp)
	assert.Equal(t, cfg, runIdentity.Cfg)
}

func TestRunIdentity_fetchConfigInfo(t *testing.T) {
	cfg := &config.Config{
		AppName:          "TestApp",
		CommandName:      "TestCommand",
		RegisterRun:      "TestRegisterRun",
		OverwriteRunId:   "TestRunId",
		DbImpl:           "carmen",
		DbVariant:        "go-file",
		CarmenSchema:     5,
		VmImpl:           "lfvm",
		ArchiveMode:      true,
		ArchiveQueryRate: 10,
		ArchiveVariant:   "test-variant",
		ChainID:          1,
		StateDbSrc:       "test-db-src",
		RpcRecordingPath: "test-rpc-path",
		First:            1000,
		Last:             2000,
	}

	timestamp := time.Now().Unix()
	runIdentity := MakeRunIdentity(timestamp, cfg)

	info, err := runIdentity.fetchConfigInfo()
	assert.NoError(t, err)

	assert.Equal(t, cfg.AppName, info["AppName"])
	assert.Equal(t, cfg.CommandName, info["CommandName"])
	assert.Equal(t, cfg.RegisterRun, info["RegisterRun"])
	assert.Equal(t, cfg.OverwriteRunId, info["OverwriteRunId"])
	assert.Equal(t, cfg.DbImpl, info["DbImpl"])
	assert.Equal(t, cfg.DbVariant, info["DbVariant"])
	assert.Equal(t, strconv.Itoa(cfg.CarmenSchema), info["CarmenSchema"])
	assert.Equal(t, cfg.VmImpl, info["VmImpl"])
	assert.Equal(t, strconv.FormatBool(cfg.ArchiveMode), info["ArchiveMode"])
	assert.Equal(t, strconv.Itoa(cfg.ArchiveQueryRate), info["ArchiveQueryRate"])
	assert.Equal(t, cfg.ArchiveVariant, info["ArchiveVariant"])
	assert.Equal(t, strconv.Itoa(int(cfg.ChainID)), info["ChainId"])
	assert.Equal(t, cfg.StateDbSrc, info["DbSrc"])
	assert.Equal(t, cfg.RpcRecordingPath, info["RpcRecordings"])
	assert.Equal(t, strconv.Itoa(int(cfg.First)), info["First"])
	assert.Equal(t, strconv.Itoa(int(cfg.Last)), info["Last"])
	assert.Equal(t, strconv.Itoa(int(runIdentity.Timestamp)), info["Timestamp"])
}
