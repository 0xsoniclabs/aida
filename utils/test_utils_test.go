package utils

import (
	"testing"

	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/pkg/errors"
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

func TestUtils_Must(t *testing.T) {
	// Test with a valid value
	mockFn := func() ([]byte, error) {
		return []byte{1, 2, 3}, nil
	}
	validValue := []byte{1, 2, 3}
	result := Must(mockFn())
	assert.Equal(t, validValue, result)

	// Test with an error
	mockFnWithError := func() ([]byte, error) {
		return nil, errors.New("mock error")
	}
	assert.Panics(t, func() {
		_ = Must(mockFnWithError())
	})
}
