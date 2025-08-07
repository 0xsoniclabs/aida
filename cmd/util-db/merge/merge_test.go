package merge

import (
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestMergeCommand(t *testing.T) {
	path1 := t.TempDir() + "/sdb1"
	sdb1, err := db.NewDefaultSubstateDB(path1)
	require.NoError(t, err)
	s1 := utils.GetTestSubstate("pb")
	s1.Block = 10
	s1.Transaction = 2
	err = sdb1.PutSubstate(s1)
	require.NoError(t, err)
	err = sdb1.Close()
	require.NoError(t, err)

	path2 := t.TempDir() + "/sdb2"
	sdb2, err := db.NewDefaultSubstateDB(path2)
	require.NoError(t, err)
	s2 := utils.GetTestSubstate("pb")
	s2.Block = 20
	s2.Transaction = 3
	err = sdb2.PutSubstate(s2)
	require.NoError(t, err)
	err = sdb2.Close()
	require.NoError(t, err)

	_, aidaDbPath := utils.CreateTestSubstateDb(t)
	app := cli.NewApp()
	app.Action = mergeAction
	app.Flags = Command.Flags

	err = app.Run([]string{
		Command.Name,
		"--aida-db",
		aidaDbPath,
		"-l",
		"CRITICAL",
		"--substate-encoding",
		"pb",
		path1,
		path2,
	})
	require.NoError(t, err)
	aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
	require.NoError(t, err)

	gotS1, err := aidaDb.GetSubstate(10, 2)
	require.NoError(t, err)
	require.NoError(t, gotS1.Equal(s1))

	gotS2, err := aidaDb.GetSubstate(20, 3)
	require.NoError(t, err)
	require.NoError(t, gotS2.Equal(s2))
}
