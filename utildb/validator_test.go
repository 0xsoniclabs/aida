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

package utildb

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
)

const emptyDBHash = "d41d8cd98f00b204e9800998ecf8427e"

func TestGenerateDbHash_EmptyDb(t *testing.T) {
	tmpDir := t.TempDir() + "/aidaDb"
	emptyDB, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	emptyHash, err := GenerateDbHash(emptyDB, "error")
	if err != nil {
		t.Fatalf("failed to hash empty db: %v", err)
	}

	if emptyDBHash != hex.EncodeToString(emptyHash) {
		t.Errorf("expected empty db hash to be %s, got %s", emptyDBHash, hex.EncodeToString(emptyHash))
	}
}

func TestGenerateDbHash_ComponentsAffectingHash(t *testing.T) {
	tests := []struct {
		name   string
		prefix []byte
		key    string
		val    []byte
	}{
		{"AddSubstate", []byte(db.SubstateDBPrefix), "key1", []byte("value1")},
		{"AddDeletion", []byte(db.DestroyedAccountPrefix), "del1", []byte("deleted")},
		{"AddUpdate", []byte(db.UpdateDBPrefix), "upd1", []byte("update")},
		{"AddStateHash", []byte(utils.StateRootHashPrefix), "state1", []byte("statehash")},
		{"AddBlockHash", []byte(utils.BlockHashPrefix), "block1", []byte("blockhash")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir() + "/aidaDb"
			baseDB, err := db.NewDefaultBaseDB(tmpDir)
			if err != nil {
				t.Fatal(err)
			}

			key := append(tc.prefix, []byte(tc.key)...)
			if err := baseDB.Put(key, tc.val); err != nil {
				t.Fatalf("failed to put %s: %v", tc.name, err)
			}
			hash, err := GenerateDbHash(baseDB, "error")
			if err != nil {
				t.Fatalf("failed to hash db: %v", err)
			}
			if hex.EncodeToString(hash) == emptyDBHash {
				t.Errorf("db-hash did not change after adding %s", tc.name)
			}

			// Remove the tmpDir after test
			t.Cleanup(func() {
				_ = baseDB.Close()
				_ = os.RemoveAll(tmpDir)
			})
		})
	}
}
