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

func TestCmd_RunStochasticReplayCommand(t *testing.T) {
	// given
	tempDir := t.TempDir()
	traceFile := filepath.Join(tempDir, "trace.bin")

	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticReplayCommand}
	args := utils.NewArgs("test").
		Arg(StochasticReplayCommand.Name).
		Flag(utils.BalanceRangeFlag.Name, 100).
		Flag(utils.NonceRangeFlag.Name, 100).
		Flag(utils.TraceFlag.Name, true).
		Flag(utils.TraceFileFlag.Name, traceFile).
		Flag(utils.MemoryBreakdownFlag.Name, true).
		Flag(utils.ShadowDbImplementationFlag.Name, "geth").
		Flag(utils.StateDbImplementationFlag.Name, "carmen").
		Flag(utils.StateDbVariantFlag.Name, "go-file").
		Arg(10).
		Arg(path.Join(testDataDir, "replay.json")).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	stat, err := os.Stat(traceFile)
	require.NoError(t, err)
	assert.NotZero(t, stat.Size())
}
