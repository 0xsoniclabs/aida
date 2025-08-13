package update

import (
	"math"
	"strconv"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestUpdateCommand(t *testing.T) {
	aidaDbPath := t.TempDir() + "/aida-db"
	aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
	require.NoError(t, err)

	// Put substate with max latest block to avoid any updating
	ss := utils.GetTestSubstate("pb")
	ss.Block = math.MaxUint64
	ss.Env.Number = math.MaxUint64
	err = aidaDb.PutSubstate(ss)
	require.NoError(t, err)

	err = aidaDb.Close()
	require.NoError(t, err)

	app := cli.NewApp()
	app.Action = updateAction
	app.Flags = Command.Flags

	err = app.Run([]string{
		Command.Name,
		"--aida-db",
		aidaDbPath,
		"-l",
		"CRITICAL",
		"--chainid",
		strconv.FormatInt(int64(utils.SonicMainnetChainID), 10),
		"--db-tmp",
		t.TempDir(),
		"--substate-encoding",
		"pb",
	})
	require.NoError(t, err)
}
