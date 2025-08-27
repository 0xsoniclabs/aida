package utils

import (
	"errors"
	"testing"

	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestUtils_CreateTestSubstateDb(t *testing.T) {
	ss, path := CreateTestSubstateDb(t)
	sdb, err := substateDb.NewDefaultSubstateDB(path)
	require.NoError(t, err)
	gotSs, err := sdb.GetSubstate(ss.Block, ss.Transaction)
	require.NoError(t, err)
	require.NoError(t, ss.Equal(gotSs))
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
