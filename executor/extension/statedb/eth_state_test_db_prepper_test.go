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
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
)

func Test_ethStateTestDbPrepper_PreBlockPreparesStateDB(t *testing.T) {
	cfg := &utils.Config{
		DbImpl:   "geth",
		ChainID:  1,
		LogLevel: "critical",
	}
	ext := ethStateTestDbPrepper{cfg: cfg, log: logger.NewLogger(cfg.LogLevel, "EthStatePrepper")}

	testData := ethtest.CreateTestTransaction(t)
	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: testData}
	ctx := &executor.Context{}
	err := ext.PreBlock(st, ctx)
	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}

	if ctx.State == nil {
		t.Fatalf("failed to initialize a DB instance")
	}
}

func Test_ethStateTestDbPrepper_PostBlockCleansTmpDir(t *testing.T) {
	cfg := &utils.Config{
		DbImpl:   "geth",
		ChainID:  1,
		LogLevel: "critical",
	}
	ext := ethStateTestDbPrepper{cfg: cfg, log: logger.NewLogger(cfg.LogLevel, "EthStatePrepper")}

	testData := ethtest.CreateTestTransaction(t)
	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: testData}
	ctx := &executor.Context{}
	err := ext.PreBlock(st, ctx)
	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}
	dirPath := ctx.StateDbPath
	// check if exists before removing
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Fatalf("tmp dir not found")
	}

	err = ext.PostBlock(st, ctx)
	if err != nil {
		t.Fatalf("unexpected err; %v", err)
	}

	// check if tmp dir is removed
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		t.Fatalf("tmp dir not removed")
	}
}

func Test_ethStateTestDbPrepper_PreBlock_FailsWithUnknownFork(t *testing.T) {
	cfg := &utils.Config{
		ChainID:  utils.EthTestsChainID,
		LogLevel: "critical",
	}
	ext := ethStateTestDbPrepper{cfg: cfg, log: logger.NewLogger(cfg.LogLevel, "EthStatePrepper")}
	testData := ethtest.CreateTestTransactionWithUnknownFork(t)

	err := ext.PreBlock(executor.State[txcontext.TxContext]{Data: testData}, new(executor.Context))
	require.ErrorContains(t, err, "cannot init chain config")
}
