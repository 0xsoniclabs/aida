package utils

import (
	"testing"

	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
)

func TestUtils_createTestUpdateDB(t *testing.T) {
	dbPath := t.TempDir() + "/testUpdateDB"
	db, err := createTestUpdateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test update DB: %v", err)
	}
	defer func(db substateDb.UpdateDB) {
		e := db.Close()
		if e != nil {
			t.Fatalf("Failed to close test update DB: %v", e)
		}
	}(db)

	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestUtils_getTestSubstate(t *testing.T) {
	ss := GetTestSubstate("default")
	assert.NotNil(t, ss)
}
