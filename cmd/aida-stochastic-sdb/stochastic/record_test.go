package stochastic

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunStochasticRecordCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir(path.Join(testDataDir, "sample-pb-db"), aidaDbPath))
	outputFile := filepath.Join(tempDir, "test_events.json")
	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticRecordCommand}
	args := utils.NewArgs("test").
		Arg(StochasticRecordCommand.Name).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Flag(utils.WorkersFlag.Name, 1).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.OutputFlag.Name, outputFile).
		Arg(0).
		Arg(1000).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	stat, err := os.Stat(outputFile)
	require.NoError(t, err)
	assert.NotZero(t, stat.Size())
}
