package profile

import (
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunGetKeyStatsCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&GetKeyStatsCommand}
	args := utils.NewArgs("test").
		Arg(GetKeyStatsCommand.Name).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.WorkersFlag.Name, 1).
		Arg("1").
		Arg("1000").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
