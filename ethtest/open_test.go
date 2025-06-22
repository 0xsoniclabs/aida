package ethtest

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
)

func createConfigFile(t *testing.T, path string) {
	b := make(map[string]*stJSON)
	b["test"] = &stJSON{}
	rawSt, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("cannot marshal st: %v", err)
	}
	filePath := path
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_, err = file.Write(rawSt)
	if err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
}

func TestGetTestWithinPath(t *testing.T) {
	tmp := t.TempDir()
	t.Run("no file", func(t *testing.T) {
		cfg := &utils.Config{
			ArgPath: tmp + "/testdata",
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, utils.StateTests)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, tests)
	})

	t.Run("with json config file", func(t *testing.T) {
		filePath := tmp + "/testdata.json"
		createConfigFile(t, filePath)
		cfg := &utils.Config{
			ArgPath: filePath,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, utils.StateTests)
		assert.NoError(t, err)
		assert.Len(t, tests,
			1)
	})
	t.Run("with json config dir", func(t *testing.T) {
		// create dir name testdata
		err := os.Mkdir(tmp+"/GeneralStateTests", 0755)
		if err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		err = os.Mkdir(tmp+"/EIPTests", 0755)
		if err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		err = os.Mkdir(tmp+"/EIPTests/StateTests", 0755)
		if err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		createConfigFile(t, tmp+"/GeneralStateTests/testdata.json")
		cfg := &utils.Config{
			ArgPath: tmp,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, utils.StateTests)
		assert.NoError(t, err)
		assert.Len(t, tests, 1)
	})

	t.Run("with block test", func(t *testing.T) {
		// create dir name testdata
		createConfigFile(t, tmp+"/GeneralStateTests/testdata.json")
		cfg := &utils.Config{
			ArgPath: tmp,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, utils.BlockTests)
		assert.Error(t, err)
		assert.Nil(t, tests)
	})

	t.Run("with other test", func(t *testing.T) {
		// create dir name testdata
		createConfigFile(t, tmp+"/GeneralStateTests/testdata.json")
		cfg := &utils.Config{
			ArgPath: tmp,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, utils.PseudoTx)
		assert.Error(t, err)
		assert.Nil(t, tests)
	})
}
