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

	ss := createTestSubstate(t, 1)
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

func TestCloner_CloneCodes_DoesNotCloneDuplicates(t *testing.T) {
	srcPath := t.TempDir()
	srcDb, err := db.NewDefaultSubstateDB(srcPath)
	require.NoError(t, err, "failed to create source db")
	err = srcDb.SetSubstateEncoding("protobuf")
	require.NoError(t, err, "failed to set substate encoding")

	// Create two identical substates with different tx numbers
	ss1 := createTestSubstate(t, 1)
	err = srcDb.PutSubstate(ss1)
	require.NoError(t, err, "failed to put substate")
	ss2 := createTestSubstate(t, 2)
	err = srcDb.PutSubstate(ss2)
	require.NoError(t, err, "failed to put substate")

	// PutCode must be called only once for each code
	ctrl := gomock.NewController(t)
	dstDb := db.NewMockSubstateDB(ctrl)
	dstDb.EXPECT().PutCode([]byte{1}).MaxTimes(1)
	dstDb.EXPECT().PutCode([]byte{2}).MaxTimes(1)

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

func createTestSubstate(t *testing.T, tx int) *substate.Substate {
	t.Helper()
	random := types.Hash{1}
	to := types.Address{1}
	return &substate.Substate{
		InputSubstate: substate.WorldState{
			types.Address{1}: &substate.Account{
				Code:    []byte{1},
				Balance: uint256.NewInt(10),
				Storage: make(map[types.Hash]types.Hash),
			},
		},
		OutputSubstate: substate.WorldState{
			types.Address{2}: &substate.Account{
				Code:    []byte{2},
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
