package utils

import (
	"testing"

	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestArgsBuilder_NewArgs(t *testing.T) {
	args := NewArgs("test").
		Arg("a").
		Arg(0).
		Arg(false).
		Arg(true).
		Flag("f1", "v1").
		Flag("f2", 0).
		Flag("f3", false).
		Flag("f4", true).
		Build()
	assert.Equal(t, "test", args[0])
	assert.Equal(t, "a", args[1])
	assert.Equal(t, "0", args[2])
	assert.Equal(t, "false", args[3])
	assert.Equal(t, "true", args[4])
	assert.Equal(t, "--f1", args[5])
	assert.Equal(t, "v1", args[6])
	assert.Equal(t, "--f2", args[7])
	assert.Equal(t, "0", args[8])
	assert.Equal(t, "--f4", args[9])
	assert.Equal(t, 10, len(args))
}

// CreateTestSubstateDb creates a test substate database with a predefined substate.
func CreateTestSubstateDb(t *testing.T) (*substate.Substate, string) {
	path := t.TempDir()
	db, err := substateDb.NewSubstateDB(path, nil, nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.SetSubstateEncoding(substateDb.ProtobufEncodingSchema))

	ss := GetTestSubstate("protobuf")
	err = db.PutSubstate(ss)
	require.NoError(t, err)

	md := NewAidaDbMetadata(db, "CRITICAL")
	require.NoError(t, md.genMetadata(ss.Block-1, ss.Block+1, 0, 0, SonicMainnetChainID, []byte{}))

	require.NoError(t, db.Close())

	return ss, path
}
