package updateset

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunUpdateSetStatsCommand(t *testing.T) {
	// given - basic priming test with default settings
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir(path.Join(testDataDir, "sample-rlp-db"), aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&UpdateSetStatsCommand}

	args := utils.NewArgs("test").
		Arg(UpdateSetStatsCommand.Name).
		Flag(utils.UpdateDbFlag.Name, aidaDbPath).
		Arg("first").
		Arg("last").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
