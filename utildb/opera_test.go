package utildb

import (
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOpera_InitFromGenesis_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	tmpDir := t.TempDir()

	cfg := &utils.Config{
		ClientDb:    tmpDir,
		LogLevel:    "info",
		OperaBinary: "/bin/true", // This is a valid binary that should succeed
	}

	ao := newAidaOpera(nil, cfg, log)
	err := ao.initFromGenesis()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOpera_InitFromGenesis_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	tmpDir := t.TempDir() + "/operaDb"

	cfg := &utils.Config{
		ClientDb:    tmpDir,
		LogLevel:    "info",
		OperaBinary: "non-existent-binary", // This binary will fail
	}

	ao := newAidaOpera(nil, cfg, log)
	err := ao.initFromGenesis()
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	assert.Equal(t, err.Error(), "load opera genesis; unable to start Command non-existent-binary --datadir "+tmpDir+" --genesis  --exitwhensynced.epoch=0 --cache 0 --db.preset=legacy-ldb --maxpeers=0; exec: \"non-existent-binary\": executable file not found in $PATH")
}
