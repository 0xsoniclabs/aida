package update

import (
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestUpdate_Command(t *testing.T) {
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
	app.Action = updateAction
	app.Flags = Command.Flags

	err = app.Run([]string{
		Command.Name,
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
	}, 0, 28_000_000)
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
