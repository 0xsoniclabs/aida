package stochastic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
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
		Arg("../../dataset/events.json").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	_, err = os.Stat(outputFile)
	assert.NoError(t, err)
}
