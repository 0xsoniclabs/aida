package statedb

import (
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/executor/extension/statedb/mocks"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substateCtx "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"math"
	"testing"
)

func TestParentBlockHashProcessor_PreBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProvider := utils.NewMockStateHashProvider(ctrl)
	mockState := state.NewMockStateDB(ctrl)
	mockProcessor := mocks.NewMockiEvmProcessor(ctrl)
	hash := common.Hash{123}
	// Processor is called only once
	gomock.InOrder(
		mockProvider.EXPECT().GetStateHash(2).Return(hash, nil),
		// Parent hash must be processed in a separate transaction!
		mockState.EXPECT().BeginTransaction(uint32(utils.PseudoTx)).Return(nil),
		mockProcessor.EXPECT().ProcessParentBlockHash(hash, gomock.Any()),
		mockState.EXPECT().EndTransaction().Return(nil),
	)

	hashProcessor := parentBlockHashProcessor{
		hashProvider: mockProvider,
		processor:    mockProcessor,
		// At the time of implementation, Sonic does not have Prague time yet
		cfg:          utils.NewTestConfig(t, utils.HoleskyChainID, 1, 10, false, "Prague"),
		NilExtension: extension.NilExtension[txcontext.TxContext]{},
	}

	// First call is skipped because block number is the first block number of given chain id
	err := hashProcessor.PreBlock(executor.State[txcontext.TxContext]{Block: int(utils.KeywordBlocks[utils.HoleskyChainID]["first"]), Data: substateCtx.NewTxContext(&substate.Substate{
		Env:   &substate.Env{Timestamp: math.MaxUint64},
		Block: utils.KeywordBlocks[utils.HoleskyChainID]["first"],
	})}, &executor.Context{State: mockState})
	require.NoError(t, err, "PreBlock failed")

	chainCfg, err := hashProcessor.cfg.GetChainConfig("Prague")
	require.NoError(t, err, "GetChainConfig failed")

	// Second call is skipped because it does not have prague time yet
	err = hashProcessor.PreBlock(executor.State[txcontext.TxContext]{Block: 2, Data: substateCtx.NewTxContext(&substate.Substate{
		Env:   &substate.Env{Timestamp: *chainCfg.PragueTime - 100},
		Block: 2,
	})}, &executor.Context{State: mockState})
	require.NoError(t, err, "PreBlock failed")

	// Third call calls the StateHashProvider with block-1
	err = hashProcessor.PreBlock(executor.State[txcontext.TxContext]{Block: 3, Data: substateCtx.NewTxContext(&substate.Substate{
		Env:         &substate.Env{Timestamp: math.MaxUint64},
		Result:      nil,
		Block:       3,
		Transaction: 0,
	})}, &executor.Context{State: mockState})
	require.NoError(t, err, "PreBlock failed")
}

func TestParentBlockHashProcessor_PreRunInitializesHashProvider(t *testing.T) {
	cfg := utils.NewTestConfig(t, utils.HoleskyChainID, 1, 10, false, "Prague")
	hp := NewParentBlockHashProcessor(cfg)
	ctrl := gomock.NewController(t)
	aidaDb := db.NewMockSubstateDB(ctrl)

	stateRoot := types.Hash{1}
	aidaDb.EXPECT().Get([]byte(utils.StateHashPrefix+hexutil.EncodeUint64(10))).Return(stateRoot.Bytes(), nil)

	err := hp.PreRun(executor.State[txcontext.TxContext]{}, &executor.Context{AidaDb: aidaDb})
	require.NoError(t, err, "PreBlock failed")

	hash, err := hp.(*parentBlockHashProcessor).hashProvider.GetStateHash(10)
	require.NoError(t, err, "hashProvider.GetStateHash failed")
	require.Equal(t, stateRoot.Bytes(), hash.Bytes())
}
