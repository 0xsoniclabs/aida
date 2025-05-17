package utils

import (
	"errors"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"math/big"
	"testing"
)

func TestPrimeContext_mayApplyBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	// case success
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
	}
	mockBulkLoad.EXPECT().Close().Return(nil)
	mockStateDb.EXPECT().StartBulkLoad(uint64(1)).Return(mockBulkLoad, nil)
	err := prime.mayApplyBulkLoad()
	assert.NoError(t, err)

	// case success
	prime.operations = 0
	err = prime.mayApplyBulkLoad()
	assert.Nil(t, err)

	// case error on close
	prime = &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
	}
	mockBulkLoad.EXPECT().Close().Return(mockErr)
	err = prime.mayApplyBulkLoad()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), err.Error())

	// case error on start bulk load
	prime = &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: OperationThreshold + 1,
		log:        logger.NewLogger("ERROR", "Test"),
	}
	mockBulkLoad.EXPECT().Close().Return(nil)
	mockStateDb.EXPECT().StartBulkLoad(uint64(1)).Return(nil, mockErr)
	err = prime.mayApplyBulkLoad()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), err.Error())
}

func TestPrimeContext_PrimeStateDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil).AnyTimes()
	mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil).AnyTimes()
	mockStateDb.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
	mockStateDb.EXPECT().EndTransaction().Return(nil).AnyTimes()
	mockStateDb.EXPECT().EndBlock().Return(nil).AnyTimes()
	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulkLoad, nil).AnyTimes()
	mockBulkLoad.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().Close().Return(nil).AnyTimes()

	acc := txcontext.NewAccount(
		MakeRandomByteSlice(t, 2048),
		MakeAccountStorage(t),
		big.NewInt(int64(GetRandom(1, 1000*5000))),
		uint64(GetRandom(1, 1000*5000)),
	)
	ws := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc,
	})

	// Case 1
	prime := &PrimeContext{
		cfg: &Config{
			PrimeRandom: true,
		},
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	err := prime.PrimeStateDB(ws, mockStateDb)
	assert.NoError(t, err)

	// Case 2
	prime = &PrimeContext{
		cfg: &Config{
			PrimeRandom: false,
		},
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	err = prime.PrimeStateDB(ws, mockStateDb)
	assert.NoError(t, err)
}

func TestPrimeContext_loadExistingAccountsIntoCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	acc := txcontext.NewAccount(
		MakeRandomByteSlice(t, 2048),
		MakeAccountStorage(t),
		big.NewInt(int64(GetRandom(1, 1000*5000))),
		uint64(GetRandom(1, 1000*5000)),
	)
	mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().Exist(gomock.Any()).Return(true)
	mockStateDb.EXPECT().EndTransaction().Return(nil)
	mockStateDb.EXPECT().EndBlock().Return(nil)
	err := prime.loadExistingAccountsIntoCache(txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc,
	}))
	assert.NoError(t, err)
}

func TestPrimeContext_primeOneAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       mockBulkLoad,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist:      map[common.Address]bool{},
	}
	acc := txcontext.NewAccount(
		MakeRandomByteSlice(t, 2048),
		MakeAccountStorage(t),
		big.NewInt(int64(GetRandom(1, 1000*5000))),
		uint64(GetRandom(1, 1000*5000)),
	)
	mockBulkLoad.EXPECT().CreateAccount(gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return()
	mockBulkLoad.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	err := prime.primeOneAccount(common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"), acc, NewProgressTracker(0, logger.NewLogger("ERROR", "Test")))
	assert.Nil(t, err)
}

func TestPrimeContext_PrimeStateDBRandom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	mockBulkLoad := state.NewMockBulkLoad(ctrl)
	acc1 := txcontext.NewAccount(
		MakeRandomByteSlice(t, 2048),
		MakeAccountStorage(t),
		big.NewInt(int64(GetRandom(1, 1000*5000))),
		uint64(GetRandom(1, 1000*5000)),
	)
	acc2 := txcontext.NewAccount(
		MakeRandomByteSlice(t, 2048),
		MakeAccountStorage(t),
		big.NewInt(int64(GetRandom(1, 1000*5000))),
		uint64(GetRandom(1, 1000*5000)),
	)
	mockWs := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): acc1,
		common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): acc2,
	})
	prime := &PrimeContext{
		cfg: &Config{
			RandomSeed: 0,
		},
		load:       nil,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist: map[common.Address]bool{
			common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): true,
			common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): true,
		},
	}
	mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().EndTransaction().Return(nil)
	mockStateDb.EXPECT().EndBlock().Return(nil)
	mockStateDb.EXPECT().StartBulkLoad(gomock.Any()).Return(mockBulkLoad, nil)
	mockBulkLoad.EXPECT().SetBalance(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetNonce(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetCode(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockBulkLoad.EXPECT().Close().Return(nil)
	err := prime.PrimeStateDBRandom(mockWs, mockStateDb, NewProgressTracker(0, logger.NewLogger("ERROR", "Test")))
	assert.NoError(t, err)
}

func TestPrimeContext_SelfDestructAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       nil,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
		exist: map[common.Address]bool{
			common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"): true,
			common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"): true,
		},
	}
	mockStateDb.EXPECT().BeginSyncPeriod(gomock.Any()).Return()
	mockStateDb.EXPECT().BeginBlock(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	mockStateDb.EXPECT().EndTransaction().Return(nil)
	mockStateDb.EXPECT().EndBlock().Return(nil)
	mockStateDb.EXPECT().EndSyncPeriod().Return()
	mockStateDb.EXPECT().Exist(gomock.Any()).Return(true).AnyTimes()
	mockStateDb.EXPECT().SelfDestruct(gomock.Any()).Return(*uint256.NewInt(99)).AnyTimes()
	prime.SelfDestructAccounts(mockStateDb, []substatetypes.Address{
		substatetypes.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		substatetypes.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
	})
	assert.Equal(t, uint64(1), prime.block)
}

func TestPrimeContext_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	prime := &PrimeContext{
		cfg:        nil,
		load:       nil,
		db:         mockStateDb,
		operations: 0,
		log:        logger.NewLogger("ERROR", "Test"),
		block:      0,
	}
	block := prime.GetBlock()
	assert.Equal(t, uint64(0), block)
}
