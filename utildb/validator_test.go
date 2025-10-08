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
	t.Run("AddSubstate", func(t *testing.T) {
		tmpDir := t.TempDir() + "/aidaDb"
		baseDB, err := db.NewDefaultBaseDB(tmpDir)
		if err != nil {
			t.Fatal(err)
		}

		key := append([]byte(db.SubstateDBPrefix), []byte("key1")...)
		val := []byte("value1")
		if err := baseDB.Put(key, val); err != nil {
			t.Fatalf("failed to put substate: %v", err)
		}
		hash, err := GenerateDbHash(baseDB, "error")
		if err != nil {
			t.Fatalf("failed to hash db: %v", err)
		}
		if hex.EncodeToString(hash) == emptyDBHash {
			t.Error("db-hash did not change after adding substate")
		}
	})

	t.Run("AddDeletion", func(t *testing.T) {
		tmpDir := t.TempDir() + "/aidaDb"
		baseDB, err := db.NewDefaultBaseDB(tmpDir)
		if err != nil {
			t.Fatal(err)
		}

		key := append([]byte(db.DestroyedAccountPrefix), []byte("del1")...)
		val := []byte("deleted")
		if err := baseDB.Put(key, val); err != nil {
			t.Fatalf("failed to put deletion: %v", err)
		}
		hash, err := GenerateDbHash(baseDB, "error")
		if err != nil {
			t.Fatalf("failed to hash db: %v", err)
		}
		if hex.EncodeToString(hash) == emptyDBHash {
			t.Error("db-hash did not change after adding deletion")
		}
	})

	t.Run("AddUpdate", func(t *testing.T) {
		tmpDir := t.TempDir() + "/aidaDb"
		baseDB, err := db.NewDefaultBaseDB(tmpDir)
		if err != nil {
			t.Fatal(err)
		}

		key := append([]byte(db.UpdateDBPrefix), []byte("upd1")...)
		val := []byte("update")
		if err := baseDB.Put(key, val); err != nil {
			t.Fatalf("failed to put update: %v", err)
		}
		hash, err := GenerateDbHash(baseDB, "error")
		if err != nil {
			t.Fatalf("failed to hash db: %v", err)
		}
		if hex.EncodeToString(hash) == emptyDBHash {
			t.Error("db-hash did not change after adding update")
		}
	})

	t.Run("AddStateHash", func(t *testing.T) {
		tmpDir := t.TempDir() + "/aidaDb"
		baseDB, err := db.NewDefaultBaseDB(tmpDir)
		if err != nil {
			t.Fatal(err)
		}

		key := append([]byte(utils.StateRootHashPrefix), []byte("state1")...)
		val := []byte("statehash")
		if err := baseDB.Put(key, val); err != nil {
			t.Fatalf("failed to put state hash: %v", err)
		}
		hash, err := GenerateDbHash(baseDB, "error")
		if err != nil {
			t.Fatalf("failed to hash db: %v", err)
		}
		if hex.EncodeToString(hash) == emptyDBHash {
			t.Error("db-hash did not change after adding state hash")
		}
	})

	t.Run("AddBlockHash", func(t *testing.T) {
		tmpDir := t.TempDir() + "/aidaDb"
		baseDB, err := db.NewDefaultBaseDB(tmpDir)
		if err != nil {
			t.Fatal(err)
		}

		key := append([]byte(utils.BlockHashPrefix), []byte("block1")...)
		val := []byte("blockhash")
		if err := baseDB.Put(key, val); err != nil {
			t.Fatalf("failed to put block hash: %v", err)
		}
		hash, err := GenerateDbHash(baseDB, "error")
		if err != nil {
			t.Fatalf("failed to hash db: %v", err)
		}
		if hex.EncodeToString(hash) == emptyDBHash {
			t.Error("db-hash did not change after adding block hash")
		}
	})
}
