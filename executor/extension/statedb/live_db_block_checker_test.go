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

package statedb

import (
	"os"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
)

func TestLiveDbBlockChecker_PreRunReturnsErrorIfStateDbLastBlockIsTooSmallEthereum(t *testing.T) {
	cfg := &utils.Config{}
	cfg.First = 15
	cfg.IsExistingStateDb = true
	cfg.ChainID = 1

	cfg.StateDbSrc = t.TempDir()
	err := utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, 10, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info; %v", err)
	}

	ext := MakeLiveDbBlockChecker[any](cfg)
	err = ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatalf("pre-run must return error")
	}

	wantedErr := "if using existing live-db with vm-sdb first block needs to be last block of live-db + 1, in your case 11"

	if strings.Compare(err.Error(), wantedErr) != 0 {
		t.Fatalf("unexpected err\ngot: %v\n want: %v", err, wantedErr)
	}
}

func TestLiveDbBlockChecker_PreRunIgnoresBlockAlignmentIfDbIsMarkedAsNotFinished(t *testing.T) {
	cfg := &utils.Config{}
	cfg.First = 10
	cfg.IsExistingStateDb = true
	cfg.ChainID = 250

	cfg.StateDbSrc = t.TempDir()
	err := utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, 0, common.Hash{}, false)
	if err != nil {
		t.Fatalf("cannot create testing state db info; %v", err)
	}

	ext := MakeLiveDbBlockChecker[any](cfg)
	err = ext.PreRun(executor.State[any]{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLiveDbBlockChecker_PreRunReturnsErrorIfShadowDbLastBlockIsTooSmallEthereum(t *testing.T) {
	cfg := &utils.Config{}
	cfg.First = 10
	cfg.IsExistingStateDb = true
	cfg.ShadowDb = true
	cfg.ChainID = 1

	cfg.StateDbSrc = t.TempDir()
	if err := os.Mkdir(cfg.StateDbSrc+utils.PathToPrimaryStateDb, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	err := utils.WriteStateDbInfo(cfg.StateDbSrc+utils.PathToPrimaryStateDb, cfg, 15, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info %v", err)
	}

	if err = os.Mkdir(cfg.StateDbSrc+utils.PathToShadowStateDb, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	err = utils.WriteStateDbInfo(cfg.StateDbSrc+utils.PathToShadowStateDb, cfg, 15, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info %v", err)
	}

	ext := MakeLiveDbBlockChecker[any](cfg)
	err = ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatalf("pre-run must return error")
	}

	wantedErr := "if using existing live-db with vm-sdb first block needs to be last block of live-db + 1, in your case 16"

	if strings.Compare(err.Error(), wantedErr) != 0 {
		t.Fatalf("unexpected err\ngot: %v\n want: %v", err, wantedErr)
	}
}

func TestLiveDbBlockChecker_PreRunReturnsErrorIfStateDbLastBlockIsTooHigh(t *testing.T) {
	cfg := &utils.Config{}
	cfg.First = 10
	cfg.IsExistingStateDb = true

	cfg.StateDbSrc = t.TempDir()
	err := utils.WriteStateDbInfo(cfg.StateDbSrc, cfg, 15, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info; %v", err)
	}

	ext := MakeLiveDbBlockChecker[any](cfg)
	err = ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatalf("pre-run must return error")
	}

	wantedErr := "if using existing live-db with vm-sdb first block needs to be higher than last block of live-db, in your case 16"

	if strings.Compare(err.Error(), wantedErr) != 0 {
		t.Fatalf("unexpected err\ngot: %v\n want: %v", err, wantedErr)
	}
}

func TestLiveDbBlockChecker_PreRunReturnsErrorIfShadowDbLastBlockIsTooHigh(t *testing.T) {
	cfg := &utils.Config{}
	cfg.First = 10
	cfg.IsExistingStateDb = true
	cfg.ShadowDb = true

	cfg.StateDbSrc = t.TempDir()
	if err := os.Mkdir(cfg.StateDbSrc+utils.PathToPrimaryStateDb, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	err := utils.WriteStateDbInfo(cfg.StateDbSrc+utils.PathToPrimaryStateDb, cfg, 15, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info %v", err)
	}

	if err = os.Mkdir(cfg.StateDbSrc+utils.PathToShadowStateDb, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	err = utils.WriteStateDbInfo(cfg.StateDbSrc+utils.PathToShadowStateDb, cfg, 15, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info %v", err)
	}

	ext := MakeLiveDbBlockChecker[any](cfg)
	err = ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatalf("pre-run must return error")
	}

	wantedErr := "if using existing live-db with vm-sdb first block needs to be higher than last block of live-db, in your case 16"

	if strings.Compare(err.Error(), wantedErr) != 0 {
		t.Fatalf("unexpected err\ngot: %v\n want: %v", err, wantedErr)
	}
}

func TestShadowDbBlockChecker_PreRunReturnsErrorIfPrimeAndShadowDbHaveDifferentLastBlock(t *testing.T) {
	cfg := &utils.Config{}
	cfg.First = 15
	cfg.IsExistingStateDb = true
	cfg.ShadowDb = true

	cfg.StateDbSrc = t.TempDir()
	if err := os.Mkdir(cfg.StateDbSrc+utils.PathToPrimaryStateDb, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	err := utils.WriteStateDbInfo(cfg.StateDbSrc+utils.PathToPrimaryStateDb, cfg, 11, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info %v", err)
	}

	if err = os.Mkdir(cfg.StateDbSrc+utils.PathToShadowStateDb, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	err = utils.WriteStateDbInfo(cfg.StateDbSrc+utils.PathToShadowStateDb, cfg, 10, common.Hash{}, true)
	if err != nil {
		t.Fatalf("cannot create testing state db info %v", err)
	}

	ext := MakeLiveDbBlockChecker[any](cfg)
	err = ext.PreRun(executor.State[any]{}, nil)
	if err == nil {
		t.Fatalf("pre-run must return error")
	}

	wantedErr := "shadow (11) and prime (10) state dbs have different last block"

	if strings.Compare(err.Error(), wantedErr) != 0 {
		t.Fatalf("unexpected err\ngot: %v\n want: %v", err, wantedErr)
	}
}
