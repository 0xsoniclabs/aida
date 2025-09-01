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

package update

import (
	"context"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
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
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(logger.LogLevelFlag.Name, "CRITICAL").
		Flag(utils.ChainIDFlag.Name, int(utils.SonicMainnetChainID)).
		Flag(utils.DbTmpFlag.Name, t.TempDir()).
		Flag(utils.SubstateEncodingFlag.Name, "protobuf").
		Build()
	err = app.Run(args)
	require.NoError(t, err)
}

func TestCmd_UpdateCommand(t *testing.T) {
	// given
	tmpDir := t.TempDir()
	aidaDbPath := filepath.Join(tmpDir, "aida-db")
	tmpDbPath := filepath.Join(tmpDir, "tmp-db")
	require.NoError(t, os.Mkdir(tmpDbPath, os.ModePerm))

	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
		Flag(utils.DbTmpFlag.Name, tmpDbPath).
		Flag(utils.UpdateTypeFlag.Name, "stable").
		Flag(utils.SubstateEncodingFlag.Name, "protobuf").
		Build()

	// Create a context with cancellation to control the app execution
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to communicate the result of app.Run
	resultChan := make(chan error, 1)

	// Run the app in a goroutine
	go func() {
		err := app.Run(args)
		resultChan <- err
	}()

	// Monitor tmp-db folder for gz files
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(1 * time.Minute) // 1 minute timeout
	defer timeout.Stop()

	for {
		select {
		case err := <-resultChan:
			// App completed naturally
			assert.NoError(t, err)
			return
		case <-ticker.C:
			// Check for gz files in tmp-db folder
			if hasGzFile(tmpDbPath) {
				// Found gz file, cancel context and terminate test successfully
				cancel()
				t.Log("Found gz file in tmp-db, terminating test early")
				return
			}
		case <-timeout.C:
			// Timeout reached
			cancel()
			t.Fatal("Test timed out waiting for gz file or completion")
		case <-ctx.Done():
			// Context cancelled
			return
		}
	}
}

// hasGzFile checks if there are any .gz files in the specified directory
func hasGzFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gz") {
			return true
		}
	}
	return false
}

func TestUpdate_getTargetDbBlockRange(t *testing.T) {
	tests := []struct {
		name      string
		wantFirst uint64
		wantLast  uint64
		wantErr   string
		setup     func(t *testing.T) *utils.Config
	}{
		{
			name: "db does not exist",
			setup: func(t *testing.T) *utils.Config {
				return &utils.Config{AidaDb: t.TempDir() + "/nonexistent-db"}
			},
			wantFirst: 0, wantLast: 0, wantErr: "",
		},
		{
			name: "db exists but no substates",
			setup: func(t *testing.T) *utils.Config {
				aidaDbPath := t.TempDir() + "/aida-db"
				aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
				require.NoError(t, err)
				require.NoError(t, aidaDb.Close())
				return &utils.Config{AidaDb: aidaDbPath, SubstateEncoding: "pb"}
			},
			wantFirst: 0, wantLast: 0, wantErr: "cannot find blocks in substate",
		},
		{
			name: "db exists with substates",
			setup: func(t *testing.T) *utils.Config {
				aidaDbPath := t.TempDir() + "/aida-db"
				aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
				require.NoError(t, err)
				ss := utils.GetTestSubstate("pb")
				ss.Block = 100
				ss.Env.Number = 100
				require.NoError(t, aidaDb.PutSubstate(ss))
				require.NoError(t, aidaDb.Close())
				return &utils.Config{AidaDb: aidaDbPath, SubstateEncoding: "pb"}
			},
			wantFirst: 100, wantLast: 100, wantErr: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := test.setup(t)
			first, last, err := getTargetDbBlockRange(cfg)
			assert.Equal(t, test.wantFirst, first)
			assert.Equal(t, test.wantLast, last)
			if test.wantErr != "" {
				assert.ErrorContains(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestUpdate_CalculateMD5Sum(t *testing.T) {
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

func TestUpdate_pushToChanel(t *testing.T) {
	patches := []utils.PatchJson{
		{FileName: "patch1.tar.gz", ToBlock: 10},
		{FileName: "patch2.tar.gz", ToBlock: 20},
	}

	ch := pushPatchToChanel(patches)

	var received []utils.PatchJson
	for patch := range ch {
		received = append(received, patch)
	}

	assert.Equal(t, patches, received)
}

func TestUpdate_retrievePatchesToDownload(t *testing.T) {
	utils.AidaDbRepositoryUrl = utils.AidaDbRepositorySonicUrl
	defer func() {
		utils.AidaDbRepositoryUrl = ""
	}()
	patches, err := retrievePatchesToDownload(&utils.Config{
		ChainID:    utils.SonicMainnetChainID,
		UpdateType: "nightly",
	}, 28_000_000)
	require.NoError(t, err)
	require.NotEmpty(t, patches)
}

func TestUpdate_update(t *testing.T) {
	aidaDbPath := t.TempDir() + "/aida-db"
	utils.AidaDbRepositoryUrl = utils.AidaDbRepositoryTestUrl
	defer func() {
		utils.AidaDbRepositoryUrl = ""
	}()
	err := update(&utils.Config{
		AidaDb:     aidaDbPath,
		UpdateType: "nightly",
		DbTmp:      t.TempDir(),
	})
	require.NoError(t, err)
	aidaDb, err := db.NewDefaultSubstateDB(aidaDbPath)
	require.NoError(t, err)
	ss := aidaDb.GetFirstSubstate()
	assert.Equal(t, uint64(1), ss.Block)
	ss, err = aidaDb.GetLastSubstate()
	require.NoError(t, err)
	assert.Equal(t, uint64(210080), ss.Block)
}
func TestUpdate_update_downloadFails(t *testing.T) {
	aidaDbPath := t.TempDir() + "/aida-db"
	utils.AidaDbRepositoryUrl = "https://unknownrepository.com"
	defer func() {
		utils.AidaDbRepositoryUrl = ""
	}()
	err := update(&utils.Config{
		AidaDb:     aidaDbPath,
		UpdateType: "nightly",
		DbTmp:      t.TempDir(),
	})
	require.ErrorContains(t, err, "unable to download patches.json")
}

func TestUpdate_mergeToExistingAidaDb_ClassicPatch(t *testing.T) {
	// Create patch with a substate
	want, patchPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	// Open target db and fill it
	targetDb, err := db.NewDefaultSubstateDB(t.TempDir() + "/target-db")
	require.NoError(t, err)
	txType := int32(substate.SetCodeTxType)
	err = targetDb.PutSubstate(&substate.Substate{
		InputSubstate:  make(substate.WorldState),
		OutputSubstate: make(substate.WorldState),
		Env: &substate.Env{
			Difficulty: new(big.Int).SetUint64(1),
			BaseFee:    new(big.Int).SetUint64(1),
		},
		Message: substate.NewMessage(
			1,
			true,
			new(big.Int).SetUint64(1),
			1,
			types.Address{1},
			new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, &txType,
			types.AccessList{{Address: types.Address{1}, StorageKeys: []types.Hash{{1}, {2}}}}, new(big.Int).SetUint64(1),
			new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0),
			[]types.SetCodeAuthorization{
				{ChainID: *uint256.NewInt(1), Address: types.Address{1}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)},
			}),
		Result:      new(substate.Result),
		Block:       0,
		Transaction: 0,
	})
	require.NoError(t, err)
	// Set correct metadata block range
	targetMD := utils.NewAidaDbMetadata(targetDb, "CRITICAL")
	err = targetMD.SetBlockRange(0, want.Block-1)
	require.NoError(t, err)
	err = targetMD.SetChainID(utils.SonicMainnetChainID)
	require.NoError(t, err)

	cfg := &utils.Config{
		ChainID:  utils.SonicMainnetChainID,
		LogLevel: "CRITICAL",
	}
	err = mergeToExistingAidaDb(cfg, targetMD, patchPath)
	require.NoError(t, err)
	// Check that merge has happened
	got, err := targetDb.GetSubstate(want.Block, want.Transaction)
	require.NoError(t, err)
	assert.NoError(t, want.Equal(got))
}

func TestUpdate_mergeToExistingAidaDb_StateHashPatch(t *testing.T) {
	ss, targetDbPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	// Create patch with a state root
	patchPath := t.TempDir() + stateHashPatchFileName
	patchDb, err := db.NewDefaultBaseDB(patchPath)
	require.NoError(t, err)
	wantHash := common.Hash{0x12}
	err = utils.SaveStateRoot(patchDb, hexutil.EncodeUint64(ss.Block), wantHash.String())
	require.NoError(t, err)
	err = patchDb.Close()
	require.NoError(t, err)

	// Create target db
	targetDb, err := db.NewDefaultBaseDB(targetDbPath)
	require.NoError(t, err)

	// Set correct metadata block range
	targetMD := utils.NewAidaDbMetadata(targetDb, "CRITICAL")
	require.NoError(t, err)
	err = targetMD.SetChainID(utils.SonicMainnetChainID)
	require.NoError(t, err)

	cfg := &utils.Config{
		ChainID:  utils.SonicMainnetChainID,
		LogLevel: "CRITICAL",
	}
	err = mergeToExistingAidaDb(cfg, targetMD, patchPath)
	require.NoError(t, err)
	hp := utils.MakeHashProvider(targetDb)

	gotHash, err := hp.GetStateRootHash(int(ss.Block))
	require.NoError(t, err)
	require.Zero(t, wantHash.Cmp(gotHash))
}

func TestUpdate_mergeToExistingAidaDb_BlocksDoesNotAlign(t *testing.T) {
	want, patchPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	_, targetPath := utils.CreateTestSubstateDb(t, db.ProtobufEncodingSchema)
	targetDb, err := db.NewDefaultBaseDB(targetPath)
	require.NoError(t, err)
	targetMD := utils.NewAidaDbMetadata(targetDb, "CRITICAL")
	// set wrong block range to target db
	err = targetMD.SetBlockRange(0, want.Block-1000)
	require.NoError(t, err)
	err = targetMD.SetChainID(utils.SonicMainnetChainID)
	require.NoError(t, err)

	cfg := &utils.Config{
		ChainID:  utils.SonicMainnetChainID,
		LogLevel: "CRITICAL",
	}
	err = mergeToExistingAidaDb(cfg, targetMD, patchPath)
	require.ErrorContains(t, err, "metadata blocks does not align")
}

func TestUpdate_retrievePatchesToDownload_MustChooseUpdateType(t *testing.T) {
	_, err := retrievePatchesToDownload(&utils.Config{
		UpdateType: "", // empty
	}, 0)
	require.ErrorContains(t, err, "please choose correct data-type")
}

func TestByToBlock_CanBeUsedToSortByToBlock(t *testing.T) {
	patches := []utils.PatchJson{
		{FileName: "patch1.tar.gz", ToBlock: 10},
		{FileName: "patch2.tar.gz", ToBlock: 20},
		{FileName: "patch3.tar.gz", ToBlock: 15},
	}

	expected := []utils.PatchJson{
		{FileName: "patch1.tar.gz", ToBlock: 10},
		{FileName: "patch3.tar.gz", ToBlock: 15},
		{FileName: "patch2.tar.gz", ToBlock: 20},
	}

	sort.Sort(ByToBlock(patches))
	assert.Equal(t, expected, patches)
}
