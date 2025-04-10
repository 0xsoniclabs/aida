// Copyright 2024 Fantom Foundation
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

package register

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/register"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"

	//db
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"go.uber.org/mock/gomock"
)

const (
	sqlite3SelectFromStats string = `
		select start, end, memory, live_disk, archive_disk, tx_rate, gas_rate, overall_tx_rate, overall_gas_rate
		from stats
		where start>=:start and end<=:end;
	`
	sqlite3SelectFromMetadata string = `
		select key, value
		from metadata
		where key=:key;
	`
)

type query struct {
	Start int `db:"start"`
	End   int `db:"end"`
}

type statsResponse struct {
	Start          int     `db:"start"`
	End            int     `db:"end"`
	Memory         int     `db:"memory"`
	LiveDisk       int     `db:"live_disk"`
	ArchiveDisk    int     `db:"archive_disk"`
	TxRate         float64 `db:"tx_rate"`
	GasRate        float64 `db:"gas_rate"`
	OverallTxRate  float64 `db:"overall_tx_rate"`
	OverallGasRate float64 `db:"overall_gas_rate"`
}

type metadataQuery struct {
	Key string `db:"key"`
}

type metadataResponse struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

func TestRegisterProgress_DoNothingIfDisabled(t *testing.T) {
	cfg := &utils.Config{}
	cfg.RegisterRun = ""
	ext := MakeRegisterProgress(cfg, 0, OnPreBlock)
	if _, ok := ext.(extension.NilExtension[txcontext.TxContext]); !ok {
		t.Fatalf("extension RegisterProgress is enabled even though not disabled in configuration.")
	}
}

func TestRegisterProgress_TerminatesIfPathToRegisterDirDoesNotExist(t *testing.T) {
	var (
		pathToRegisterDir string = filepath.Join("does", "not", "exist")
	)

	cfg := &utils.Config{}
	cfg.RegisterRun = pathToRegisterDir // enabled here
	cfg.First = 5
	cfg.Last = 25
	interval := 10

	ext := MakeRegisterProgress(cfg, interval, OnPreBlock)
	if _, err := ext.(extension.NilExtension[txcontext.TxContext]); err {
		t.Fatalf("Extension RegisterProgress is disabled even though enabled in configuration.")
	}

	err := ext.PreRun(executor.State[txcontext.TxContext]{}, nil)
	if err == nil {
		t.Fatalf("Error is nil even though registered path is %s.", pathToRegisterDir)
	}
}

func TestRegisterProgress_TerminatesIfPathToStateDBDoesNotExist(t *testing.T) {
	var (
		dummyStateDbPath string = filepath.Join("does", "not", "exist")
	)

	cfg := &utils.Config{}
	cfg.RegisterRun = dummyStateDbPath // enabled here
	cfg.First = 5
	cfg.Last = 25
	interval := 10

	ctrl := gomock.NewController(t)
	stateDb := state.NewMockStateDB(ctrl)

	ext := MakeRegisterProgress(cfg, interval, OnPreBlock)
	if _, err := ext.(extension.NilExtension[txcontext.TxContext]); err {
		t.Fatalf("Extension RegisterProgress is disabled even though enabled in configuration.")
	}

	ctx := &executor.Context{State: stateDb, StateDbPath: dummyStateDbPath}

	err := ext.PreRun(executor.State[txcontext.TxContext]{}, ctx)
	if err == nil {
		t.Fatalf("Error is nil even though dummyStateDbPath is %s.", dummyStateDbPath)
	}
}

func TestRegisterProgress_InsertToDbIfEnabled(t *testing.T) {
	var (
		tmpDir           string = t.TempDir()
		dummyStateDbPath string = filepath.Join(tmpDir, "dummy.txt")
		dbName           string = "tmp"
		connection       string = filepath.Join(tmpDir, fmt.Sprintf("%s.db", dbName))
	)
	// Check if path to state db is writable
	if err := os.WriteFile(dummyStateDbPath, []byte("hello world"), 0x600); err != nil {
		t.Fatalf("Failed to prepare disk content for %s.", dummyStateDbPath)
	}

	// Check if path to stats db is writable
	sDb, err := sqlx.Open("sqlite3", connection)
	if err != nil {
		t.Fatalf("Failed to connect to database at %s.", connection)
	}

	_, err = sDb.Exec(registerProgressCreateTableIfNotExist)
	if err != nil {
		t.Fatalf("Unable to create stats table at database %s.\n%s", connection, err)
	}

	_, err = sDb.Exec(register.MetadataCreateTableIfNotExist)
	if err != nil {
		t.Fatalf("Unable to create metadata table at database %s.\n%s", connection, err)
	}

	stmt, err := sDb.PrepareNamed(sqlite3SelectFromStats)
	if err != nil {
		t.Fatalf("Failed to prepare statement using db at %s. \n%s", connection, err)
	}

	meta, err := sDb.PrepareNamed(sqlite3SelectFromMetadata)
	if err != nil {
		t.Fatalf("Failed to prepare statement using db at %s. \n%s", connection, err)
	}

	ctrl := gomock.NewController(t)
	stateDb := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.RegisterRun = tmpDir // enabled here
	cfg.OverwriteRunId = dbName
	cfg.First = 5
	cfg.Last = 25
	interval := 10
	// expects [5-9]P[10-19]P[20-24]P, where P is print

	ext := MakeRegisterProgress(cfg, interval, OnPreBlock)
	if _, err := ext.(extension.NilExtension[txcontext.TxContext]); err {
		t.Fatalf("Extension RegisterProgress is disabled even though enabled in configuration.")
	}

	itv := utils.NewInterval(cfg.First, cfg.Last, uint64(interval))

	ctx := &executor.Context{
		State:           stateDb,
		StateDbPath:     dummyStateDbPath,
		ExecutionResult: substatecontext.NewReceipt(&substate.Result{GasUsed: 100}),
	}

	s := &substate.Substate{
		Result: &substate.Result{
			Status:  0,
			GasUsed: 100,
		},
	}

	expectedRowCount := 0

	/// prints 3 times
	gomock.InOrder(
		stateDb.EXPECT().GetMemoryUsage().Return(&state.MemoryUsage{UsedBytes: 1234}),
		stateDb.EXPECT().GetMemoryUsage().Return(&state.MemoryUsage{UsedBytes: 4321}),
		stateDb.EXPECT().GetMemoryUsage().Return(&state.MemoryUsage{UsedBytes: 5555}),
	)

	ext.PreRun(executor.State[txcontext.TxContext]{}, ctx)

	sub := substatecontext.NewTxContext(s)

	for b := int(cfg.First); b < int(cfg.Last); b++ {
		ext.PreBlock(executor.State[txcontext.TxContext]{Block: b, Data: sub}, ctx)

		// check if a print happens here
		if b > int(itv.End()) {
			itv.Next()
			expectedRowCount++
		}
		stats := []statsResponse{}
		stmt.Select(&stats, query{int(cfg.First), int(cfg.Last)})
		if len(stats) != expectedRowCount {
			t.Errorf("Expected #Row: %d, Actual #Row: %d", expectedRowCount, len(stats))
		}

		ext.PreTransaction(executor.State[txcontext.TxContext]{Data: sub}, ctx)
		ext.PostTransaction(executor.State[txcontext.TxContext]{Data: sub}, ctx)
		ext.PostBlock(executor.State[txcontext.TxContext]{Block: b, Data: sub}, ctx)
	}

	ext.PostRun(executor.State[txcontext.TxContext]{}, ctx, nil)

	// check if a print happens here
	expectedRowCount++
	stats := []statsResponse{}
	stmt.Select(&stats, query{int(cfg.First), int(cfg.Last)})
	if len(stats) != expectedRowCount {
		t.Errorf("Expected #Row: %d, Actual #Row: %d", expectedRowCount, len(stats))
	}

	// Check that metadata is not duplicated
	ms := []metadataResponse{}
	meta.Select(&ms, metadataQuery{"Processor"})
	if len(ms) != 1 {
		t.Errorf("Expected runtime to be recorded once, Actual #Row: %d", len(ms))
	}

	// check if runtime is recorded after postrun
	meta.Select(&ms, metadataQuery{"Runtime"})
	if len(ms) != 1 {
		t.Errorf("Expected runtime to be recorded once, Actual #Row: %d", len(ms))
	}

	// check if RunSucceed is recorded after postrun
	meta.Select(&ms, metadataQuery{"RunSucceed"})
	if len(ms) != 1 {
		t.Errorf("Expected RunSucceed to be recorded once, Actual #Row: %d", len(ms))
	}
	if ms[0].Value != strconv.FormatBool(true) {
		t.Errorf("Expected RunSucceed expected to be true, Actual #Row: %s", ms[0].Value)
	}

	// check if RunError is not recorded
	meta.Select(&ms, metadataQuery{"RunError"})
	if len(ms) != 0 {
		t.Errorf("Expected RunError should not be recorded, Actual: #Row: %d", len(ms))
	}

	meta.Close()
	stmt.Close()
	sDb.Close()
}

func TestRegisterProgress_IfErrorRecordIntoMetadata(t *testing.T) {
	var (
		tmpDir           string = t.TempDir()
		dummyStateDbPath string = filepath.Join(tmpDir, "dummy.txt")
		dbName           string = "tmp"
		connection       string = filepath.Join(tmpDir, fmt.Sprintf("%s.db", dbName))
	)
	// Check if path to state db is writable
	if err := os.WriteFile(dummyStateDbPath, []byte("hello world"), 0x600); err != nil {
		t.Fatalf("failed to prepare disk content for %s.", dummyStateDbPath)
	}

	// Check if path to stats db is writable
	sDb, err := sqlx.Open("sqlite3", connection)
	if err != nil {
		t.Fatalf("Failed to connect to database at %s.", connection)
	}

	_, err = sDb.Exec(registerProgressCreateTableIfNotExist)
	if err != nil {
		t.Fatalf("Unable to create stats table at database %s.\n%s", connection, err)
	}

	_, err = sDb.Exec(register.MetadataCreateTableIfNotExist)
	if err != nil {
		t.Fatalf("Unable to create metadata table at database %s.\n%s", connection, err)
	}

	meta, err := sDb.PrepareNamed(sqlite3SelectFromMetadata)
	if err != nil {
		t.Fatalf("Failed to prepare statement using db at %s. \n%s", connection, err)
	}

	ctrl := gomock.NewController(t)
	stateDb := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{}
	cfg.RegisterRun = tmpDir // enabled here
	cfg.OverwriteRunId = dbName

	ctx := &executor.Context{State: stateDb, StateDbPath: dummyStateDbPath}
	gomock.InOrder(
		stateDb.EXPECT().GetMemoryUsage().Return(&state.MemoryUsage{UsedBytes: 1234}),
	)

	ext := MakeRegisterProgress(cfg, 123, OnPreBlock)
	if _, err := ext.(extension.NilExtension[txcontext.TxContext]); err {
		t.Fatalf("RegisterProgress is disabled even though enabled in configuration.")
	}

	// this is the run
	errorText := "This is one random error!"
	ext.PreRun(executor.State[txcontext.TxContext]{}, ctx)
	ext.PostRun(executor.State[txcontext.TxContext]{}, ctx, fmt.Errorf("%s", errorText))

	// check if RunSucceed is recorded after postrun
	ms := []metadataResponse{}
	meta.Select(&ms, metadataQuery{"RunSucceed"})
	if len(ms) != 1 {
		t.Errorf("Expected RunSucceed to be recorded once, Actual #Row: %d", len(ms))
	}
	if ms[0].Value != strconv.FormatBool(false) {
		t.Errorf("Expected RunSucceed expected to be true, Actual #Row: %s", ms[0].Value)
	}

	// check if RunError is recorded after postrun
	meta.Select(&ms, metadataQuery{"RunError"})
	if len(ms) != 1 {
		t.Errorf("Expected RunError to be recorded once, Actual #Row: %d", len(ms))
	}
	if ms[0].Value != errorText {
		t.Errorf("Expected RunError expected to be %s, Actual #Row: %s", errorText, ms[0].Value)
	}

	meta.Close()
	sDb.Close()
}

func TestRegisterProgress_ExtensionContinuesDespiteFetchEnvFailure(t *testing.T) {
	var (
		tmpDir     string = t.TempDir()
		dbName     string = "tmp"
		connection string = filepath.Join(tmpDir, fmt.Sprintf("%s.db", dbName))
		noBash     error  = errors.New("I'm using Windows! I need help!")
	)

	mockEnvInfoFetcher := func() (map[string]string, error) {
		var errs error
		return map[string]string{}, errors.Join(errs, noBash)
	}

	rm, err := register.MakeRunMetadata(
		connection,
		register.MakeRunIdentity(time.Now().Unix(), &utils.Config{}),
		mockEnvInfoFetcher,
	)

	if rm == nil {
		t.Fatalf("Object RunMetadata is nil where it should not.")
	}
	if err == nil {
		t.Fatalf("Error is nil even though user cannot execute bash script.")
	}
	if !errors.Is(err, noBash) {
		t.Fatalf("Error not from intended source: %v.", noBash)
	}
}

func TestRegisterProgress_ChecksDefaultReportInterval(t *testing.T) {
	tests := map[*utils.Config]uint64{
		{
			RegisterRun: "enabled",
			CommandName: "substate",
			First:       0,
			Last:        1_000_000,
			BlockLength: 0,
		}: defaultReportFrequency,
		{
			RegisterRun: "enabled",
			CommandName: "tx-generator",
			First:       0,
			Last:        1_000_000,
			BlockLength: 50_000,
		}: 1,
		{
			RegisterRun: "enabled",
			CommandName: "tx-generator",
			First:       0,
			Last:        1_000_000,
			BlockLength: 1,
		}: 50_000,
		{
			RegisterRun: "enabled",
			CommandName: "tx-generator",
			First:       0,
			Last:        1_000_000,
			BlockLength: 5_000_000,
		}: 1,
	}

	for cfg, expectedFreq := range tests {
		var freq int = int(cfg.BlockLength)
		if cfg.CommandName == "tx-generator" {
			freq = int(math.Ceil(float64(50_000) / float64(cfg.BlockLength)))
		}

		ext := MakeRegisterProgress(cfg, freq, OnPreBlock) // 0 to see defaults
		if _, ok := ext.(extension.NilExtension[txcontext.TxContext]); ok {
			t.Fatalf("Extension RegisterProgress is disabled even though enabled in configuration.")
		}

		rp, ok := ext.(*registerProgress)
		if !ok {
			t.Errorf("Could not cast extension to registerProgress even though it should be possible.")
		}

		if rp.interval.End()-rp.interval.Start()+1 != expectedFreq {
			t.Errorf("Printing Interval incorrect. Expected = %d, Actual: %d", expectedFreq, rp.interval.End()-rp.interval.Start()+1)
		}
	}
}
