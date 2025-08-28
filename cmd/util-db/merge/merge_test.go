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

package merge

import (
	"os"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestMerge_Command(t *testing.T) {
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

func TestMerge_Command_Errors(t *testing.T) {
	dstDb := t.TempDir() + "/dstDb"
	wrongFile := t.TempDir() + "testfile.txt"
	f, err := os.Create(wrongFile)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	tests := []struct {
		name    string
		srcDb   []string
		wantErr string
	}{
		{
			name:    "No_Source_Dbs",
			srcDb:   nil,
			wantErr: "this command requires at least 1 argument",
		},
		{
			name:    "Wrong_Source_Db",
			srcDb:   []string{wrongFile},
			wantErr: "cannot open source databases",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			app.Action = mergeAction
			app.Flags = Command.Flags

			err = app.Run(append([]string{
				Command.Name,
				"--aida-db",
				dstDb,
				"-l",
				"CRITICAL",
				"--substate-encoding",
				"pb",
			}, test.srcDb...))
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}
