package compact

import (
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestCmd_Compact(t *testing.T) {
	_, _, path := utils.CreateTestSubstateDb(t)
	app := cli.NewApp()
	app.Action = compactAction
	app.Flags = []cli.Flag{
		&utils.TargetDbFlag,
	}

	err := app.Run([]string{Command.Name, "--target-db", path})
	require.NoError(t, err)
}
