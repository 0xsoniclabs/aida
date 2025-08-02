package validator

import (
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEthereumPostTransactionUpdater_SkippedExtensionBecauseOfWrongVmImplOrWrongChainId(t *testing.T) {
	tests := []struct {
		name    string
		vmImpl  string
		chainId utils.ChainID
	}{
		{
			name:    "SkipNonLfvm",
			vmImpl:  "geth",
			chainId: utils.EthereumChainID,
		},
		{
			name:    "SkipWrongChainId",
			vmImpl:  "lfvm",
			chainId: utils.MainnetChainID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &utils.Config{}
			cfg.VmImpl = tt.vmImpl
			cfg.ChainID = tt.chainId

			ctrl := gomock.NewController(t)
			log := logger.NewMockLogger(ctrl)
			db := state.NewMockStateDB(ctrl)

			data := createTestTransaction()
			ctx := new(executor.Context)
			ctx.State = db

			st := executor.State[txcontext.TxContext]{Block: getEthereumExceptionBlock(), Transaction: 1, Data: data}

			ext := makeEthereumDbPostTransactionUpdater(cfg, log)
			if _, ok := ext.(extension.NilExtension[txcontext.TxContext]); !ok {
				t.Fatal("unexpected extension initialization")
			}
			err := ext.PostTransaction(st, ctx)
			if err != nil {
				t.Fatal("post-transaction unexpected error: ", err)
			}
		})
	}
}

func TestEthereumPostTransactionUpdater_NonExceptionBlockDoesntGetOverwritten(t *testing.T) {
	cfg := &utils.Config{}
	cfg.VmImpl = "lfvm"
	cfg.ChainID = utils.EthereumChainID

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	data := createTestTransaction()
	ctx := new(executor.Context)
	ctx.State = db

	nonExceptionBlock := 1

	if _, ok := ethereumLfvmBlockExceptions[utils.EthereumChainID][nonExceptionBlock]; ok {
		t.Fatal("nonExceptionBlock is in ethereumLfvmBlockExceptions; invalid test conditions")
	}

	st := executor.State[txcontext.TxContext]{Block: nonExceptionBlock, Transaction: 1, Data: data}

	ext := makeEthereumDbPostTransactionUpdater(cfg, log)
	err := ext.PostTransaction(st, ctx)
	if err != nil {
		t.Fatal("post-transaction unexpected error: ", err)
	}
}

func TestEthereumPostTransactionUpdater_ExceptionBlockGetsOverwritten(t *testing.T) {
	cfg := &utils.Config{}
	cfg.VmImpl = "lfvm"
	cfg.ChainID = utils.EthereumChainID

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)

	data := createTestTransaction()
	ctx := new(executor.Context)
	ctx.State = db
	st := executor.State[txcontext.TxContext]{Block: getEthereumExceptionBlock(), Transaction: 1, Data: data}

	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x1")).Return(true),
		db.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(1)),
		db.EXPECT().SubBalance(common.HexToAddress("0x1"), uint256.NewInt(1), tracing.BalanceChangeUnspecified),
		db.EXPECT().AddBalance(common.HexToAddress("0x1"), uint256.NewInt(1000), tracing.BalanceChangeUnspecified),
		db.EXPECT().GetNonce(common.HexToAddress("0x1")),
		db.EXPECT().SetNonce(common.HexToAddress("0x1"), gomock.Any(), gomock.Any()),
		db.EXPECT().GetCode(common.HexToAddress("0x1")),
		db.EXPECT().GetState(common.HexToAddress("0x1"), common.HexToHash("0x1")),
		db.EXPECT().SetState(common.HexToAddress("0x1"), common.HexToHash("0x1"), common.HexToHash("0x2")),
		db.EXPECT().EndTransaction().Return(nil),
		db.EXPECT().BeginTransaction(uint32(utils.PseudoTx)),
	)

	gomock.InOrder(
		db.EXPECT().Exist(common.HexToAddress("0x2")).Return(true),
		db.EXPECT().GetBalance(common.HexToAddress("0x2")).Return(uint256.NewInt(2)),
		db.EXPECT().SubBalance(common.HexToAddress("0x2"), uint256.NewInt(2), tracing.BalanceChangeUnspecified),
		db.EXPECT().AddBalance(common.HexToAddress("0x2"), uint256.NewInt(2000), tracing.BalanceChangeUnspecified),
		db.EXPECT().GetNonce(common.HexToAddress("0x2")),
		db.EXPECT().SetNonce(common.HexToAddress("0x2"), gomock.Any(), gomock.Any()),
		db.EXPECT().GetCode(common.HexToAddress("0x2")).Return([]byte{0x1}),
		db.EXPECT().SetCode(common.HexToAddress("0x2"), gomock.Any()),
	)

	ext := makeEthereumDbPostTransactionUpdater(cfg, log)
	err := ext.PostTransaction(st, ctx)
	if err != nil {
		t.Fatal("post-transaction unexpected error: ", err)
	}
}

func createTestTransaction() txcontext.TxContext {
	return substatecontext.NewTxContext(&substate.Substate{
		InputSubstate: substate.WorldState{
			substatetypes.HexToAddress("0x1"): substate.NewAccount(1, uint256.NewInt(1000), nil),
			substatetypes.HexToAddress("0x2"): substate.NewAccount(2, uint256.NewInt(2000), nil),
		},
		OutputSubstate: substate.WorldState{
			substatetypes.HexToAddress("0x1"): &substate.Account{
				Nonce:   1,
				Balance: uint256.NewInt(1000),
				Storage: map[substatetypes.Hash]substatetypes.Hash{
					substatetypes.BytesToHash([]byte{0x1}): substatetypes.BytesToHash([]byte{0x2})},
			},
			substatetypes.HexToAddress("0x2"): substate.NewAccount(2, uint256.NewInt(2000), nil),
		},
	})
}

func getEthereumExceptionBlock() int {
	// retrieving exception block
	for key := range ethereumLfvmBlockExceptions[utils.EthereumChainID] {
		return key
	}
	return -1
}

func TestEthereumDbPostTransactionUpdater_PreRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	cfg := &utils.Config{}
	st := executor.State[txcontext.TxContext]{}
	ctx := new(executor.Context)
	log.EXPECT().Warning(gomock.Any())
	ext := &ethereumDbPostTransactionUpdater{
		cfg: cfg,
		log: log,
	}
	err := ext.PreRun(st, ctx)
	assert.NoError(t, err)
}

func TestEthereumDbPostTransactionUpdater_MakeEthereumDbPostTransactionUpdater(t *testing.T) {
	cfg := &utils.Config{}
	cfg.ChainID = utils.PseudoTx
	ext := MakeEthereumDbPostTransactionUpdater(cfg)
	assert.IsType(t, extension.NilExtension[txcontext.TxContext]{}, ext)
}
