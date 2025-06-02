package utildb

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestAutogen_SetLock(t *testing.T) {
	cfg := &utils.Config{
		AidaDb: t.TempDir() + "/testdb",
	}
	message := "Test message"
	err := SetLock(cfg, message)
	assert.NoError(t, err)
	assert.FileExists(t, cfg.AidaDb+".autogen.lock")
}

func TestAutogen_GetLock(t *testing.T) {
	// no file
	cfg := &utils.Config{
		AidaDb: t.TempDir() + "/testdb",
	}
	str, err := GetLock(cfg)
	assert.NoError(t, err)
	assert.Empty(t, str)

	// file exists
	err = os.WriteFile(cfg.AidaDb+".autogen.lock", []byte("test"), 0655)
	if err != nil {
		t.Fatalf("failed to create lock file: %v", err)
	}
	str, err = GetLock(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "test", str)
}

func TestAutogen_AutogenRun(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	t.Run("Success", func(t *testing.T) {

		cfg := &utils.Config{
			AidaDb: tempDir + "/testdb",
		}
		g := &Generator{
			Cfg: cfg,
			Log: logger.NewLogger("INFO", "test"),
			Opera: &aidaOpera{
				FirstEpoch: 1,
			},
			TargetEpoch: 2,
		}

		tempDb, err := db.NewDefaultBaseDB(cfg.AidaDb)
		if err != nil {
			t.Fatalf("failed to create new db: %v", err)
		}
		defer func(tempDb db.BaseDB) {
			e := tempDb.Close()
			if e != nil {
				t.Fatalf("failed to close db: %v", e)
			}
		}(tempDb)

		err = AutogenRun(cfg, g)
		assert.Error(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		cfg := &utils.Config{
			AidaDb: tempDir + "/testdb",
		}
		g := &Generator{
			Cfg: cfg,
			Log: logger.NewLogger("INFO", "test"),
			Opera: &aidaOpera{
				FirstEpoch: 1,
			},
			TargetEpoch: 2,
		}
		err := AutogenRun(cfg, g)
		assert.Error(t, err)
	})

}
