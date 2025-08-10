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

func TestCmd_RunGetAddressStatsCommand(t *testing.T) {
	// given
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test_events.json")
	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticEstimateCommand}
	args := utils.NewArgs("test").
		Arg(StochasticEstimateCommand.Name).
		Flag(utils.OutputFlag.Name, outputFile).
		Arg(path.Join(testDataDir, "events.json")).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	stat, err := os.Stat(outputFile)
	require.NoError(t, err)
	assert.NotZero(t, stat.Size())
}
