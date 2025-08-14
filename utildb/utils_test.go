package utildb

import (
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
)

func TestUtils_OpenSourceDatabases(t *testing.T) {
	ss1, path1 := utils.CreateTestSubstateDb(t)
	ss2, path2 := utils.CreateTestSubstateDb(t)
	ss3, path3 := utils.CreateTestSubstateDb(t)

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

func TestUtils_OpenSourceDatabases_Error(t *testing.T) {
	tests := []struct {
		name        string
		sourcePaths []string
		wantErr     string
	}{
		{
			name:        "No_Source_Paths",
			sourcePaths: []string{},
			wantErr:     "no source database were specified",
		},
		{
			name:        "No_Source_Paths",
			sourcePaths: []string{"/non/existent/path"},
			wantErr:     "source database /non/existent/path; doesn't exist",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := OpenSourceDatabases(test.sourcePaths)
			require.ErrorContains(t, err, test.wantErr)
		})
	}

}
