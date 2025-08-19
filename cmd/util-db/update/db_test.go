package update

import (
	"math/big"
	"os"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	want, patchPath := utils.CreateTestSubstateDb(t)
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
	ss, targetDbPath := utils.CreateTestSubstateDb(t)
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
	want, patchPath := utils.CreateTestSubstateDb(t)
	_, targetPath := utils.CreateTestSubstateDb(t)
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
