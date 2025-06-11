package utildb

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/types/hash"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"math/big"
	"testing"
)

func TestCloner_CloneCodes_ClonesCodesFromInputAndOutputSubstate(t *testing.T) {
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	ss := createTestSubstate(t, 1, []byte{1}, []byte{2})
	err = srcDb.PutSubstate(ss)
	require.NoError(t, err, "failed to put substate")

	dstPath := t.TempDir()
	dstDb, err := db.NewDefaultSubstateDB(dstPath)
	require.NoError(t, err, "failed to create destination db")
	err = dstDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	clnr := cloner{
		cfg: &utils.Config{
			First:            1,
			Last:             10,
			Workers:          1,
			SubstateEncoding: "protobuf",
		},
		aidaDb:  srcDb,
		cloneDb: dstDb,
		log:     logger.NewLogger("warn", "CloneCodesTest"),
	}

	err = clnr.cloneCodes()
	require.NoError(t, err, "failed to clone codes")

	codeDb := db.MakeDefaultCodeDBFromBaseDB(dstDb)
	ok, err := codeDb.HasCode(hash.Keccak256Hash([]byte{1}))
	require.NoError(t, err, "failed to check if code exists")
	require.True(t, ok, "code does not exist")
	ok, err = codeDb.HasCode(hash.Keccak256Hash([]byte{2}))
	require.NoError(t, err, "failed to check if code exists")
	require.True(t, ok, "code does not exist")
}

func TestCloner_PutCode_DoesNotPutNilCode(t *testing.T) {
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	// Create one substate with nil code and one with empty code
	ss1 := createTestSubstate(t, 1, nil, []byte{123})
	err = srcDb.PutSubstate(ss1)
	require.NoError(t, err, "failed to put substate")

	// PutCode must be called only once for each code
	ctrl := gomock.NewController(t)
	dstDb := db.NewMockSubstateDB(ctrl)
	// only one code should be put
	dstDb.EXPECT().PutCode([]byte{123}).Times(1)

	clnr := cloner{
		cfg: &utils.Config{
			First:            1,
			Last:             10,
			Workers:          1,
			SubstateEncoding: "protobuf",
		},
		aidaDb:  srcDb,
		cloneDb: dstDb,
		log:     logger.NewLogger("warn", "CloneCodesTest"),
	}

	err = clnr.cloneCodes()
	require.NoError(t, err, "failed to clone codes")
}

func TestCloner_CloneCodes_DoesNotCloneDuplicates(t *testing.T) {
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	// Create two identical substates with different tx numbers
	ss1 := createTestSubstate(t, 1, []byte{1}, []byte{1})
	err = srcDb.PutSubstate(ss1)
	require.NoError(t, err, "failed to put substate")

	// PutCode must be called only once for each code
	ctrl := gomock.NewController(t)
	dstDb := db.NewMockSubstateDB(ctrl)
	dstDb.EXPECT().PutCode([]byte{1}).Times(1)

	clnr := cloner{
		cfg: &utils.Config{
			First:            1,
			Last:             10,
			Workers:          1,
			SubstateEncoding: "protobuf",
		},
		aidaDb:  srcDb,
		cloneDb: dstDb,
		log:     logger.NewLogger("warn", "CloneCodesTest"),
	}

	err = clnr.cloneCodes()
	require.NoError(t, err, "failed to clone codes")
}

func TestOpenCloningDbs_OpensDbsCorrectly(t *testing.T) {
	tmp := t.TempDir()
	srcPath := tmp + "/src"
	dstPath := tmp + "/dst"
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	ss1 := createTestSubstate(t, 1, []byte{1}, []byte{1})
	err = srcDb.PutSubstate(ss1)
	require.NoError(t, err, "failed to put substate")

	// Close the db to test opening
	require.NoError(t, srcDb.Close())

	srcDb, dstDb, err := OpenCloningDbs(srcPath, dstPath, "protobuf")
	require.NoError(t, err, "failed to open cloning dbs")

	// check correct opening of source db
	srcDbSs, err := srcDb.GetSubstate(1, 1)
	require.NoError(t, err, "failed to get substate")
	require.NoError(t, srcDbSs.Equal(ss1))
	// Make sure destination db is empty
	iter := dstDb.NewSubstateIterator(0, 1)
	require.False(t, iter.Next())
}

func createTestSubstate(t *testing.T, tx int, codeA, codeB []byte) *substate.Substate {
	t.Helper()
	random := types.Hash{1}
	to := types.Address{1}
	return &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{1}: &substate.Account{
				Code:    codeA,
				Balance: uint256.NewInt(10),
				Storage: make(map[types.Hash]types.Hash),
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{2}: &substate.Account{
				Code:    codeB,
				Balance: uint256.NewInt(10),
				Storage: make(map[types.Hash]types.Hash),
			},
		},
		Env: &substate.Env{
			Difficulty:  big.NewInt(10),
			BaseFee:     big.NewInt(10),
			BlobBaseFee: big.NewInt(10),
			BlockHashes: make(map[uint64]types.Hash),
			Random:      &random,
		},
		Message: &substate.Message{
			CheckNonce:            true,
			GasPrice:              big.NewInt(10),
			To:                    &to,
			Value:                 big.NewInt(10),
			AccessList:            make(types.AccessList, 0),
			GasFeeCap:             big.NewInt(10),
			GasTipCap:             big.NewInt(10),
			BlobHashes:            make([]types.Hash, 0),
			SetCodeAuthorizations: make([]types.SetCodeAuthorization, 0),
		},
		Result:      &substate.Result{},
		Block:       uint64(1),
		Transaction: tx,
	}
}

func TestClone_CorrectlyClonesData(t *testing.T) {
	// prepare the source db
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	md := utils.NewAidaDbMetadata(srcDb, "INFO")
	err = md.SetChainID(utils.MainnetChainID)
	require.NoError(t, err, "failed to set chain id")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")
	ss := createTestSubstate(t, 1, []byte{1}, []byte{1})
	err = srcDb.PutSubstate(ss)

	targetPath := t.TempDir()
	targetDb, err := db.NewDefaultSubstateDB(targetPath)
	require.NoError(t, err, "failed to create target db")

	cfg := &utils.Config{First: 0, Last: 1, ChainID: utils.MainnetChainID, Workers: 1}
	err = CreatePatchClone(cfg, srcDb, targetDb, 5577, 5578, true)
	require.NoError(t, err, "failed to clone codes")

	gotSs, err := targetDb.GetSubstate(1, 1)
	require.NoError(t, err, "failed to get substate")

	err = ss.Equal(gotSs)
	require.NoError(t, err)
}
