package db

import (
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestPrintCount_IntegrationTest(t *testing.T) {
	aidaDbPath := t.TempDir() + "aida-db"

	aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
	if err != nil {
		t.Fatal(err)
	}
	if aidaDb == nil {
		t.Fatal("aidaDb is nil")
	}
	err = aidaDb.Close()
	if err != nil {
		t.Fatal(err)
	}

	args := []string{
		"info", "count",
		"--aida-db", aidaDbPath,
		"--db-component=all",
		"1", "2",
	}

	// app is replacement of UtilDbApp
	app := cli.App{
		Commands: []*cli.Command{
			&cmdCount,
		}}
	err = app.Run(args)
	assert.NoError(t, err)
}

func TestPrintCount_Logging(t *testing.T) {
	type testCase struct {
		name         string
		first        uint64
		last         uint64
		dbComponent  string
		expectedLogs []struct {
			method string
			format string
			args   []interface{}
		}
	}

	tests := []testCase{
		{
			name:        "AllComponents_EmptyDbAll",
			first:       1,
			last:        2,
			dbComponent: "all",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Noticef", "Inspecting database between blocks %v-%v", []interface{}{uint64(1), uint64(2)}},
				{"Noticef", "Found %v substates", []interface{}{uint64(0)}},
				{"Noticef", "Found %v updates", []interface{}{uint64(0)}},
				{"Noticef", "Found %v deleted accounts", []interface{}{0}},
				{"Noticef", "Found %v state-hashes", []interface{}{uint64(0)}},
			},
		},
		{
			name:        "AllComponents_EmptyDbSubstate",
			first:       1,
			last:        2,
			dbComponent: "substate",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Noticef", "Inspecting database between blocks %v-%v", []interface{}{uint64(1), uint64(2)}},
				{"Noticef", "Found %v substates", []interface{}{uint64(0)}},
			},
		},
		{
			name:        "AllComponents_EmptyDbUpdate",
			first:       1,
			last:        2,
			dbComponent: "update",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Noticef", "Inspecting database between blocks %v-%v", []interface{}{uint64(1), uint64(2)}},
				{"Noticef", "Found %v updates", []interface{}{uint64(0)}},
			},
		},
		{
			name:        "AllComponents_EmptyDbDelete",
			first:       1,
			last:        2,
			dbComponent: "delete",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Noticef", "Inspecting database between blocks %v-%v", []interface{}{uint64(1), uint64(2)}},
				{"Noticef", "Found %v deleted accounts", []interface{}{0}},
			},
		},
		{
			name:        "AllComponents_EmptyDbStateHash",
			first:       3,
			last:        4,
			dbComponent: "state-hash",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Noticef", "Inspecting database between blocks %v-%v", []interface{}{uint64(3), uint64(4)}},
				{"Noticef", "Found %v state-hashes", []interface{}{uint64(0)}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aidaDbPath := t.TempDir() + "aida-db"

			aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
			if err != nil {
				t.Fatal(err)
			}
			if aidaDb == nil {
				t.Fatal("aidaDb is nil")
			}
			err = aidaDb.Close()
			if err != nil {
				t.Fatal(err)
			}

			cfg := &utils.Config{
				AidaDb:      aidaDbPath,
				DbComponent: tc.dbComponent,
				LogLevel:    "info",
				First:       tc.first,
				Last:        tc.last,
			}

			ctrl := gomock.NewController(t)
			log := logger.NewMockLogger(ctrl)

			for _, l := range tc.expectedLogs {
				switch l.method {
				case "Noticef":
					log.EXPECT().Noticef(l.format, l.args...)
				case "Warningf":
					log.EXPECT().Warningf(l.format, l.args...)
				}
			}

			err = printCount(cfg, log)
			assert.NoError(t, err)
		})
	}
}

func TestPrintRange_Logging(t *testing.T) {
	type testCase struct {
		name         string
		dbComponent  string
		expectedLogs []struct {
			method string
			format string
			args   []interface{}
		}
	}
	tests := []testCase{
		{
			name:        "AllComponents_EmptyDbAll",
			dbComponent: "all",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Warning", "No substate found", []interface{}{}},
				{"Warningf", "cannot find updateset range; %v", []interface{}{gomock.Any()}},
				{"Warningf", "cannot find deleted range; %v", []interface{}{gomock.Any()}},
				{"Warningf", "cannot find state hash range; %v", []interface{}{gomock.Any()}},
			},
		},
		{
			name:        "SubstateOnly_EmptyDb",
			dbComponent: "substate",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Warning", "No substate found", []interface{}{}},
			},
		},
		{
			name:        "UpdateOnly_EmptyDb",
			dbComponent: "update",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Warningf", "cannot find updateset range; %v", []interface{}{gomock.Any()}},
			},
		},
		{
			name:        "DeleteOnly_EmptyDb",
			dbComponent: "delete",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Warningf", "cannot find deleted range; %v", []interface{}{gomock.Any()}},
			},
		},
		{
			name:        "StateHashOnly_EmptyDb",
			dbComponent: "state-hash",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Warningf", "cannot find state hash range; %v", []interface{}{gomock.Any()}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aidaDbPath := t.TempDir() + "aida-db"

			aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
			if err != nil {
				t.Fatal(err)
			}
			if aidaDb == nil {
				t.Fatal("aidaDb is nil")
			}
			err = aidaDb.Close()
			if err != nil {
				t.Fatal(err)
			}

			cfg := &utils.Config{
				AidaDb:      aidaDbPath,
				DbComponent: tc.dbComponent,
				LogLevel:    "info",
			}

			ctrl := gomock.NewController(t)
			log := logger.NewMockLogger(ctrl)

			for _, l := range tc.expectedLogs {
				switch l.method {
				case "Infof":
					log.EXPECT().Infof(l.format, l.args...)
				case "Warningf":
					log.EXPECT().Warningf(l.format, l.args...)
				case "Warning":
					log.EXPECT().Warning(l.format)
				}

			}

			err = printRangeRun(cfg, log)
			assert.NoError(t, err)
		})
	}
}
