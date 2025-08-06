package ethtest

import (
	"encoding/json"
	"github.com/0xsoniclabs/aida/config"
	"os"
	"testing"

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

func TestEthTest_getTestWithinPath(t *testing.T) {
	tmp := t.TempDir()
	t.Run("no file", func(t *testing.T) {
		cfg := &config.Config{
			ArgPath: tmp + "/testdata",
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, config.StateTests)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, tests)
	})

	t.Run("with json config file", func(t *testing.T) {
		filePath := tmp + "/testdata.json"
		createConfigFile(t, filePath)
		cfg := &config.Config{
			ArgPath: filePath,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, config.StateTests)
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
		cfg := &config.Config{
			ArgPath: tmp,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, config.StateTests)
		assert.NoError(t, err)
		assert.Len(t, tests, 1)
	})

	t.Run("with block test", func(t *testing.T) {
		// create dir name testdata
		createConfigFile(t, tmp+"/GeneralStateTests/testdata.json")
		cfg := &config.Config{
			ArgPath: tmp,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, config.BlockTests)
		assert.Error(t, err)
		assert.Nil(t, tests)
	})

	t.Run("with other test", func(t *testing.T) {
		// create dir name testdata
		createConfigFile(t, tmp+"/GeneralStateTests/testdata.json")
		cfg := &config.Config{
			ArgPath: tmp,
		}
		tests, err := getTestsWithinPath[*stJSON](cfg, config.PseudoTx)
		assert.Error(t, err)
		assert.Nil(t, tests)
	})
}
