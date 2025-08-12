package utildb

import (
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestOpenSourceDatabases(t *testing.T) {
	_, ss1, path1 := utils.CreateTestSubstateDb(t)
	_, ss2, path2 := utils.CreateTestSubstateDb(t)
	_, ss3, path3 := utils.CreateTestSubstateDb(t)

	dbs, err := OpenSourceDatabases([]string{path1, path2, path3})
	require.NoError(t, err)
	sdb1 := db.MakeDefaultSubstateDBFromBaseDB(dbs[0])
	sdb2 := db.MakeDefaultSubstateDBFromBaseDB(dbs[1])
	sdb3 := db.MakeDefaultSubstateDBFromBaseDB(dbs[2])

	gotSs1, err := sdb1.GetSubstate(ss1.Block, ss1.Transaction)
	require.NoError(t, err)
	require.NoError(t, gotSs1.Equal(ss1))

	gotSs2, err := sdb2.GetSubstate(ss2.Block, ss2.Transaction)
	require.NoError(t, err)
	require.NoError(t, gotSs2.Equal(ss2))

	gotSs3, err := sdb3.GetSubstate(ss3.Block, ss3.Transaction)
	require.NoError(t, err)
	require.NoError(t, gotSs3.Equal(ss3))
}

func TestCalculateMD5Sum(t *testing.T) {
	name := t.TempDir() + "/testfile"
	f, err := os.Create(name)
	require.NoError(t, err)
	_, err = f.Write([]byte("test"))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	md5Sum, err := calculateMD5Sum(name)
	require.NoError(t, err)
	require.Equal(t, md5Sum, "098f6bcd4621d373cade4e832627b4f6")
}
