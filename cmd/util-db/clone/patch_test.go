package clone

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"os"
	"strconv"
	"testing"
)

func TestCommand_clonePatchAction(t *testing.T) {
	_, ss, srcDbPath := utils.CreateTestSubstateDb(t)
	app := cli.NewApp()
	app.Action = clonePatchAction
	app.Flags = []cli.Flag{
		&utils.AidaDbFlag,
		&utils.TargetDbFlag,
		&logger.LogLevelFlag,
	}

	targetDbPath := t.TempDir() + "/target.db"

	err := app.Run([]string{
		clonePatchCommand.Name,
		"--aida-db",
		srcDbPath,
		"--target-db",
		targetDbPath,
		"-l",
		"CRITICAL",
		strconv.FormatUint(ss.Block-1, 10),
		strconv.FormatUint(ss.Block+1, 10),
		"0",
		"0",
	})
	require.NoError(t, err)

	require.Condition(t, func() bool {
		stat, err := os.Stat(targetDbPath)
		require.NoError(t, err)
		return stat != nil && stat.IsDir() && stat.Size() > 0
	}, "Target database seems to be empty")
}
