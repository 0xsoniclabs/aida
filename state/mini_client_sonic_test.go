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
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestMakeMiniClientSonicStateDB_OpensAndCloses exercises the
// happy path: the factory delegates to MakeCarmenStateDB with
// mini-client's S5 / go-file defaults, returns a usable StateDB, and
// Close() cleans it up. Live-DB mode (no archive) — covers the
// substate-replay flow utils/statedb.go dispatches for
// cfg.ArchiveMode == false.
func TestMakeMiniClientSonicStateDB_OpensAndCloses(t *testing.T) {
	sdb, err := MakeMiniClientSonicStateDB(t.TempDir(), "", common.Hash{}, false, nil, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("MakeMiniClientSonicStateDB: %v", err)
	}
	if sdb == nil {
		t.Fatal("expected non-nil StateDB")
	}
	if err := sdb.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestMakeMiniClientSonicStateDB_RejectsVariant guards the signature
// the same way mini_client_geth_test.go does.
func TestMakeMiniClientSonicStateDB_RejectsVariant(t *testing.T) {
	_, err := MakeMiniClientSonicStateDB(t.TempDir(), "weird", common.Hash{}, false, nil, 0, 0, 0, 0)
	if err == nil {
		t.Fatal("expected error for unknown variant")
	}
	if !strings.Contains(err.Error(), "unknown mini-client-sonic variant") {
		t.Fatalf("error %q didn't mention the variant rejection", err.Error())
	}
}

// TestMakeMiniClientSonicStateDB_AcceptsExplicitDefaultVariant
// ensures users who pass the variant explicitly (matching what they
// would see in statedb_info on disk) get the same path as variant="".
func TestMakeMiniClientSonicStateDB_AcceptsExplicitDefaultVariant(t *testing.T) {
	sdb, err := MakeMiniClientSonicStateDB(t.TempDir(), "go-file", common.Hash{}, false, nil, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("MakeMiniClientSonicStateDB(go-file): %v", err)
	}
	if err := sdb.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
