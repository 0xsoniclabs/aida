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

package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
)

// TestMakeMiniClientGethStateDB_RoundTrip checks the wrapper around
// mini-client's pkg/aida.NewGethStateDB plugs into aida's gethStateDB
// shell so every existing proxy (logger, profiler, shadow, cache)
// composes against it without changes. It exercises the same lifecycle
// aida's substate harness hits: open, write one account, read back,
// close.
func TestMakeMiniClientGethStateDB_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	sdb, err := MakeMiniClientGethStateDB(dir, "", common.Hash{}, false, nil)
	if err != nil {
		t.Fatalf("MakeMiniClientGethStateDB: %v", err)
	}
	addr := common.HexToAddress("0xfeed")
	if err := sdb.BeginBlock(0); err != nil {
		t.Fatalf("BeginBlock: %v", err)
	}
	if err := sdb.BeginTransaction(0); err != nil {
		t.Fatalf("BeginTransaction: %v", err)
	}
	sdb.CreateAccount(addr)
	sdb.AddBalance(addr, uint256.NewInt(7), tracing.BalanceChangeUnspecified)
	if got := sdb.GetBalance(addr); got.Uint64() != 7 {
		t.Fatalf("GetBalance = %d, want 7", got.Uint64())
	}
	if err := sdb.EndTransaction(); err != nil {
		t.Fatalf("EndTransaction: %v", err)
	}
	if err := sdb.EndBlock(); err != nil {
		t.Fatalf("EndBlock: %v", err)
	}
	if err := sdb.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestMakeMiniClientGethStateDB_RejectsVariant guards the parameter
// signature: the wrapper accepts the same MakeGethStateDB signature
// but only the empty variant. A non-empty variant is a programming
// error and should surface immediately rather than silently use the
// default.
func TestMakeMiniClientGethStateDB_RejectsVariant(t *testing.T) {
	if _, err := MakeMiniClientGethStateDB(t.TempDir(), "weird", common.Hash{}, false, nil); err == nil {
		t.Fatal("expected error for unknown variant")
	}
}
