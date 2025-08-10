package stochastic

import (
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunStochasticReplayCommand(t *testing.T) {
	// given
	app := cli.NewApp()
	app.Commands = []*cli.Command{&StochasticReplayCommand}
	args := utils.NewArgs("test").
		Arg(StochasticReplayCommand.Name).
		Flag(utils.BalanceRangeFlag.Name, 100).
		Flag(utils.NonceRangeFlag.Name, 100).
		Flag(utils.ShadowDbImplementationFlag.Name, "geth").
		Flag(utils.StateDbImplementationFlag.Name, "carmen").
		Flag(utils.StateDbVariantFlag.Name, "go-file").
		Arg(100).
		Arg("../../dataset/replay.json").
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
