package db

import (
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCmd_ScrapeCommand(t *testing.T) {
	// given
	tmpDir := t.TempDir()
	targetDbPath := filepath.Join(tmpDir, "target-db")
	clientDbPath := filepath.Join(tmpDir, "client-db")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&ScrapeCommand}

	args := utils.NewArgs("test").
		Arg(ScrapeCommand.Name).
		Flag(utils.TargetDbFlag.Name, targetDbPath).
		Flag(utils.ClientDbFlag.Name, clientDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Arg("1"). // blockNumFirst
		Arg("5"). // blockNumLast - small range for testing
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
