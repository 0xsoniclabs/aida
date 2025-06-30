package db

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestPrintCount(t *testing.T) {
	type testCase struct {
		name    string
		args    []string
		wantErr string
	}

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

	tests := []testCase{
		{
			name: "IntegrationTest",
			args: []string{
				"info", "count",
				"--aida-db", aidaDbPath,
				"--db-component=all",
				"1", "2",
			},
			wantErr: "",
		},
		{
			name: "InvalidArgs",
			args: []string{
				"info", "count",
				"--aida-db", aidaDbPath,
				"--db-component=all",
			},
			wantErr: "unable to parse cli arguments; command requires 2 arguments",
		},
		{
			name: "InvalidEncoding",
			args: []string{
				"info", "count",
				"--aida-db", aidaDbPath,
				"--db-component=all",
				"--substate-encoding=invalidEncoding",
				"1", "2",
			},
			wantErr: "cannot set substate encoding; failed to set decoder; encoding not supported: invalidEncoding",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := cli.App{
				Commands: []*cli.Command{
					&cmdCount,
				}}
			err := app.Run(tc.args)
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			}
		})
	}
}

func TestPrintCount_OnlyCalculateGivenRange(t *testing.T) {
	aidaDbPath := generateTestAidaDb(t)

	cfg := &utils.Config{
		AidaDb:      aidaDbPath,
		DbComponent: "all",
		LogLevel:    "info",
		First:       10,
		Last:        11,
	}

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	log.EXPECT().Noticef("Inspecting database between blocks %v-%v", uint64(10), uint64(11))
	log.EXPECT().Noticef("Found %v substates", uint64(1))
	log.EXPECT().Noticef("Found %v updates", uint64(0))
	log.EXPECT().Noticef("Found %v deleted accounts", 1)
	log.EXPECT().Noticef("Found %v state-hashes", uint64(1))
	log.EXPECT().Noticef("Found %v block-hashes", uint64(0))

	err := printCount(cfg, log)
	assert.NoError(t, err)
}

func TestPrintCount_LoggingEmpty(t *testing.T) {
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
				{"Noticef", "Found %v block-hashes", []interface{}{uint64(0)}},
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
		{
			name:        "AllComponents_EmptyDbBlockHash",
			first:       3,
			last:        4,
			dbComponent: "block-hash",
			expectedLogs: []struct {
				method string
				format string
				args   []interface{}
			}{
				{"Noticef", "Inspecting database between blocks %v-%v", []interface{}{uint64(3), uint64(4)}},
				{"Noticef", "Found %v block-hashes", []interface{}{uint64(0)}},
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

func TestPrintRange(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *utils.Config
		wantErr string
	}{
		{
			name: "All",
			cfg: &utils.Config{
				AidaDb:      t.TempDir() + "/emptydb",
				DbComponent: "all",
			},
			wantErr: "",
		},
		{
			name: "NonExistentDb",
			cfg: &utils.Config{
				AidaDb:      t.TempDir() + "non-existent-db",
				DbComponent: "all",
			},
			wantErr: "cannot open aida-db; cannot open leveldb; stat %s: no such file or directory",
		},
		{
			name: "InvalidEncoding",
			cfg: &utils.Config{
				AidaDb:           t.TempDir() + "/emptydb1",
				SubstateEncoding: "errorEncoding",
				DbComponent:      "substate",
			},
			wantErr: "cannot set substate encoding; failed to set decoder; encoding not supported: errorEncoding",
		},
		{
			name: "InvalidDbComponent",
			cfg: &utils.Config{
				AidaDb:      t.TempDir() + "/emptydb2",
				DbComponent: "not-a-component",
			},
			wantErr: "invalid db component: not-a-component",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create empty DB if needed
			if !strings.Contains(tc.cfg.AidaDb, "non-existent-db") {
				testDb, err := db.NewDefaultBaseDB(tc.cfg.AidaDb)
				assert.NoError(t, err)
				err = testDb.Close()
				assert.NoError(t, err)
			}
			err := printRange(tc.cfg, logger.NewLogger("Warning", "TestPrintRange_Errors"))
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err != nil {
					// some expected errors need to have the path replaced
					tc.wantErr = strings.ReplaceAll(tc.wantErr, "%s", tc.cfg.AidaDb)
					assert.Contains(t, err.Error(), tc.wantErr)
				}
			}
		})
	}
}

func TestInfo_PrintRange_LoggingEmpty(t *testing.T) {
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
				{"Warningf", "cannot find state hash range; %s", []interface{}{gomock.Any()}},
				{"Warningf", "cannot find block hash range; %v", []interface{}{gomock.Any()}},
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
				{"Warningf", "cannot find state hash range; %s", []interface{}{gomock.Any()}},
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

			err = printRange(cfg, log)
			assert.NoError(t, err)
		})
	}
}

func TestPrintRange_Success(t *testing.T) {
	aidaDbPath := generateTestAidaDb(t)

	// mock logger
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	log.EXPECT().Infof("Substate block range: %v - %v", uint64(11), uint64(11))
	log.EXPECT().Infof("Updateset block range: %v - %v", uint64(12), uint64(12))
	log.EXPECT().Infof("Deleted block range: %v - %v", uint64(1), uint64(10))
	log.EXPECT().Warningf("cannot find state hash range; %s", "cannot get first state hash; not implemented")
	log.EXPECT().Infof("Block Hash range: %v - %v", uint64(21), uint64(30))

	cfg := &utils.Config{
		AidaDb:      aidaDbPath,
		DbComponent: "all",
	}

	err := printRange(cfg, log)
	if err != nil {
		t.Fatalf("printRange failed: %v", err)
	}
}

func TestPrintRange_IntegrationTest(t *testing.T) {
	aidaDbPath := generateTestAidaDb(t)
	args := []string{
		"info", "range",
		"--aida-db", aidaDbPath,
		"--db-component=all",
	}

	app := cli.App{
		Commands: []*cli.Command{
			&cmdRange,
		}}
	err := app.Run(args)
	assert.NoError(t, err)
}

func TestPrintStateHash_IntegrationTest(t *testing.T) {
	tests := []struct {
		name        string
		insertKey   string
		insertValue string
		queryArg    string
		expectErr   bool
	}{
		{
			name:        "Success",
			insertKey:   "0x1",
			insertValue: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			queryArg:    "1",
			expectErr:   false,
		},
		{
			name:        "NotFound",
			insertKey:   "0x2",
			insertValue: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			queryArg:    "1",
			expectErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aidaDbPath := t.TempDir() + "aida-db"

			aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
			if err != nil {
				t.Fatal(err)
			}
			if aidaDb == nil {
				t.Fatal("aidaDb is nil")
			}

			// insert state hash
			err = utils.SaveStateRoot(aidaDb, tc.insertKey, tc.insertValue)
			assert.NoError(t, err)

			err = aidaDb.Close()
			assert.NoError(t, err)

			args := []string{
				"info", "state-hash",
				"--aida-db", aidaDbPath,
				tc.queryArg,
			}

			app := cli.App{
				Commands: []*cli.Command{
					&cmdPrintStateHash,
				}}
			err = app.Run(args)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrintBlockHash_IntegrationTest(t *testing.T) {
	tests := []struct {
		name        string
		insertKey   string
		insertValue string
		queryArg    string
		expectErr   string
	}{
		{
			name:        "Success",
			insertKey:   "0x1",
			insertValue: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			queryArg:    "1",
			expectErr:   "",
		},
		{
			name:        "NotFound",
			insertKey:   "0x2",
			insertValue: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			queryArg:    "1",
			expectErr:   "cannot get block hash for block 1; leveldb: not found",
		},
		{
			name:        "Success",
			insertKey:   "0x1",
			insertValue: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			queryArg:    "",
			expectErr:   "cannot parse block number ; strconv.ParseInt: parsing \"\": invalid syntax",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aidaDbPath := t.TempDir() + "aida-db"

			aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
			if err != nil {
				t.Fatal(err)
			}
			if aidaDb == nil {
				t.Fatal("aidaDb is nil")
			}

			// insert block hash
			err = utils.SaveBlockHash(aidaDb, tc.insertKey, tc.insertValue)
			assert.NoError(t, err)

			err = aidaDb.Close()
			assert.NoError(t, err)

			args := []string{
				"info", "block-hash",
				"--aida-db", aidaDbPath,
				tc.queryArg,
			}

			app := cli.App{
				Commands: []*cli.Command{
					&cmdPrintBlockHash,
				}}
			err = app.Run(args)
			if len(tc.expectErr) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrintHash_EmptyDb(t *testing.T) {
	aidaDbPath := t.TempDir() + "non-existent-db"
	args := []string{
		"info", "block-hash",
		"--aida-db", aidaDbPath,
		"1",
	}

	app := cli.App{
		Commands: []*cli.Command{
			&cmdPrintBlockHash,
		}}
	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot open leveldb; stat "+aidaDbPath+": no such file or directory")
}

func TestPrintHash_InvalidArg(t *testing.T) {
	aidaDbPath := t.TempDir() + "aida-db"

	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
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
		"info", "block-hash",
		"--aida-db", aidaDbPath,
		"invalid-arg", "invalid-arg2",
	}

	app := cli.App{
		Commands: []*cli.Command{
			&cmdPrintBlockHash,
		}}
	err = app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "block-hash command requires exactly 1 argument")
}

func TestPrintHash_MissingArg(t *testing.T) {
	aidaDbPath := t.TempDir() + "aida-db"

	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
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
		"info", "block-hash",
		"--aida-db", aidaDbPath,
	}

	app := cli.App{
		Commands: []*cli.Command{
			&cmdPrintBlockHash,
		}}
	err = app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to parse cli arguments; this command requires at least 1 argument")
}

func generateTestAidaDb(t *testing.T) string {
	aidaDbPath := t.TempDir() + "aida-db"

	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
	if err != nil {
		t.Fatal(err)
	}
	if aidaDb == nil {
		t.Fatal("aidaDb is nil")
	}

	// insert substate
	state := substate.Substate{
		Block:       10,
		Transaction: 7,
		Env:         &substate.Env{Difficulty: big.NewInt(1), GasLimit: uint64(15), Number: 11},
		Message: &substate.Message{
			Value:    big.NewInt(12),
			GasPrice: big.NewInt(14),
		},
		InputSubstate:  substate.WorldState{},
		OutputSubstate: substate.WorldState{},
		Result:         &substate.Result{},
	}
	sdb := db.MakeDefaultSubstateDBFromBaseDB(aidaDb)
	err = sdb.SetSubstateEncoding("pb")
	assert.NoError(t, err)
	err = sdb.PutSubstate(&state)
	assert.NoError(t, err)

	state.Block = 15
	err = sdb.PutSubstate(&state)
	assert.NoError(t, err)

	us := updateset.UpdateSet{
		WorldState:      substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		Block:           12,
		DeletedAccounts: []types.Address{},
	}

	// insert update
	udb := db.MakeDefaultUpdateDBFromBaseDB(aidaDb)
	err = udb.PutUpdateSet(&us, us.DeletedAccounts)
	assert.NoError(t, err)

	// write delete accounts to the database
	for i := 1; i <= 10; i++ {
		err = aidaDb.Put(db.EncodeDestroyedAccountKey(uint64(i), i), []byte("0x1234567812345678123456781234567812345678123456781234567812345678"))
		if err != nil {
			t.Fatal(err)
		}
	}

	// write state hashes to the database
	for i := 11; i <= 20; i++ {
		key := "0x" + strconv.FormatInt(int64(i), 16)
		err = utils.SaveStateRoot(aidaDb, key, "0x1234567812345678123456781234567812345678123456781234567812345678")
		if err != nil {
			t.Fatal(err)
		}
	}

	// write block hashes to the database
	for i := 21; i <= 30; i++ {
		key := "0x" + strconv.FormatInt(int64(i), 16)
		err = utils.SaveBlockHash(aidaDb, key, "0x1234567812345678123456781234567812345678123456781234567812345678")
		if err != nil {
			t.Fatal(err)
		}
	}

	err = aidaDb.Close()
	if err != nil {
		t.Fatal(err)
	}
	return aidaDbPath
}

func TestPrintHashForBlock_InvalidHashtype(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)

	aidaDbPath := t.TempDir() + "aida-db"

	cfg := &utils.Config{
		AidaDb: aidaDbPath,
	}

	// Create an empty database
	aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
	if err != nil {
		t.Fatalf("error opening aida-db %s: %v", aidaDbPath, err)
	}
	err = aidaDb.Close()
	if err != nil {
		t.Fatalf("error closing aida-db %s: %v", aidaDbPath, err)
	}

	err = printHashForBlock(cfg, log, 0, "invalidHashType")
	if err == nil {
		t.Fatal("expected an error for invalid hash type, but got nil")
	}
}
