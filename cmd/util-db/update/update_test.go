package update

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestUpdate_UpdateDbCommand(t *testing.T) {
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
	app.Action = updateDbAction
	app.Flags = updateDbCommand.Flags

	err = app.Run([]string{
		updateDbCommand.Name,
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

func TestUpdate_UpdateHashesCommand(t *testing.T) {
	aidaDbPath := t.TempDir() + "/aida-db"
	clientDbPath := t.TempDir() + "/client-db"
	// given
	app := cli.NewApp()
	app.Action = updateHashesAction
	app.Flags = updateHashesCommand.Flags
	err := app.Run([]string{
		updateHashesCommand.Name,
		"--target-db",
		aidaDbPath,
		"--datadir",
		clientDbPath,
		"-l",
		"CRITICAL",
		"--chainid",
		strconv.FormatInt(int64(utils.SonicMainnetChainID), 10),
		strconv.FormatInt(int64(1), 10),
		strconv.FormatInt(int64(5), 10),
	})

	// then
	assert.NoError(t, err)
}
