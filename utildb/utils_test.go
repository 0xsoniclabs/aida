// Copyright 2025 Sonic Labs
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package utildb

import (
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
)

func TestUtils_OpenSourceDatabases(t *testing.T) {
	ss1, path1 := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	ss2, path2 := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	ss3, path3 := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)

	dbs, err := OpenSourceDatabases([]string{path1, path2, path3})
	require.NoError(t, err)
	sdb1, err := db.MakeDefaultSubstateDBFromBaseDB(dbs[0])
	require.NoError(t, err)
	sdb2, err := db.MakeDefaultSubstateDBFromBaseDB(dbs[1])
	require.NoError(t, err)
	sdb3, err := db.MakeDefaultSubstateDBFromBaseDB(dbs[2])
	require.NoError(t, err)

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

func TestUtils_CalculateMD5Sum(t *testing.T) {
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
