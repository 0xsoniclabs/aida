package proxy

import (
	"errors"
	"sync"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProxy_NewLoggerProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := NewLoggerProxy(mockDb, mockLogger, mockChan, mockWg)
	assert.NotNil(t, proxy)
}
func TestLoggingStateDb_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedErr := errors.New("test error")
	mockDb.EXPECT().Error().Return(expectedErr)
	err := proxy.Error()
	assert.Equal(t, expectedErr, err)

}
func TestLoggingStateDb_BeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	block := uint64(1)
	mockDb.EXPECT().BeginBlock(block).Return(nil)

	err := proxy.BeginBlock(block)
	assert.NoError(t, err)
}
func TestLoggingStateDb_EndBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	mockDb.EXPECT().EndBlock().Return(nil)

	err := proxy.EndBlock()
	assert.NoError(t, err)
}
func TestLoggingStateDb_BeginSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedNumber := uint64(42)
	mockDb.EXPECT().BeginSyncPeriod(expectedNumber)

	proxy.BeginSyncPeriod(expectedNumber)
}
func TestLoggingStateDb_EndSyncPeriod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	mockDb.EXPECT().EndSyncPeriod()

	proxy.EndSyncPeriod()
}
func TestLoggingStateDb_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedHash := common.HexToHash("0x1234")
	mockDb.EXPECT().GetHash().Return(expectedHash, nil)

	hash, err := proxy.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}
func TestLoggingStateDb_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	mockDb.EXPECT().Close().Return(nil)

	err := proxy.Close()
	assert.NoError(t, err)
}
func TestLoggingStateDb_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	mockDb.EXPECT().Finalise(true)

	proxy.Finalise(true)
}
func TestLoggingStateDb_IntermediateRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedRoot := common.HexToHash("0x1234")
	mockDb.EXPECT().IntermediateRoot(true).Return(expectedRoot)

	root := proxy.IntermediateRoot(true)
	assert.Equal(t, expectedRoot, root)
}
func TestLoggingStateDb_Commit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedBlock := uint64(1)
	expectedHash := common.HexToHash("0x1234")
	mockDb.EXPECT().Commit(expectedBlock, true).Return(expectedHash, nil)

	h, err := proxy.Commit(expectedBlock, true)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, h)
}
func TestLoggingStateDb_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedBlock := uint64(1)
	expectedWorldState := txcontext.NewMockWorldState(ctrl)
	mockDb.EXPECT().PrepareSubstate(expectedWorldState, expectedBlock)
	expectedWorldState.EXPECT().String().Return("")
	proxy.PrepareSubstate(expectedWorldState, expectedBlock)
}
func TestLoggingStateDb_StartBulkLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedBlock := uint64(1)
	expectedBulkLoad := state.NewMockBulkLoad(ctrl)
	mockDb.EXPECT().StartBulkLoad(expectedBlock).Return(expectedBulkLoad, nil)

	bl, err := proxy.StartBulkLoad(expectedBlock)
	assert.NoError(t, err)
	assert.NotNil(t, bl)
}
func TestLoggingStateDb_GetArchiveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedBlock := uint64(1)
	expectedArchiveState := state.NewMockNonCommittableStateDB(ctrl)
	mockDb.EXPECT().GetArchiveState(expectedBlock).Return(expectedArchiveState, nil)

	as, err := proxy.GetArchiveState(expectedBlock)
	assert.NoError(t, err)
	assert.NotNil(t, as)
}
func TestLoggingStateDb_GetArchiveBlockHeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedBlockHeight := uint64(42)
	mockDb.EXPECT().GetArchiveBlockHeight().Return(expectedBlockHeight, true, nil)

	height, empty, err := proxy.GetArchiveBlockHeight()
	assert.Equal(t, expectedBlockHeight, height)
	assert.True(t, empty)
	assert.NoError(t, err)
}
func TestLoggingStateDb_GetMemoryUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedMemoryUsage := &state.MemoryUsage{}
	mockDb.EXPECT().GetMemoryUsage().Return(expectedMemoryUsage)

	mu := proxy.GetMemoryUsage()
	assert.Equal(t, expectedMemoryUsage, mu)
}
func TestLoggingStateDb_GetShadowDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		state: mockDb,
	}
	expectedShadowDB := state.NewMockStateDB(ctrl)
	mockDb.EXPECT().GetShadowDB().Return(expectedShadowDB)

	sdb := proxy.GetShadowDB()
	assert.Equal(t, expectedShadowDB, sdb)
}
func TestLoggingNonCommittableStateDb_GetHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockNonDb := state.NewMockNonCommittableStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &loggingNonCommittableStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		nonCommittableStateDB: mockNonDb,
	}
	expectedAddr := common.HexToHash("0x1234")
	mockNonDb.EXPECT().GetHash().Return(expectedAddr, nil)
	addr, err := proxy.GetHash()
	assert.NoError(t, err)
	assert.Equal(t, expectedAddr, addr)
}
func TestLoggingNonCommittableStateDb_Release(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockNonDb := state.NewMockNonCommittableStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}
	proxy := &loggingNonCommittableStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     mockDb,
			log:    mockLogger,
			output: mockChan,
			wg:     mockWg,
		},
		nonCommittableStateDB: mockNonDb,
	}
	mockNonDb.EXPECT().Release()

	err := proxy.Release()
	assert.NoError(t, err)
}
func TestLoggingBulkLoad_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockBulkLoad(ctrl)
	proxy := &loggingBulkLoad{
		nested: mockDb,
		writeLog: func(format string, a ...any) {

		},
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().CreateAccount(addr)

	proxy.CreateAccount(addr)
}
func TestLoggingBulkLoad_SetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockBulkLoad(ctrl)
	proxy := &loggingBulkLoad{
		nested: mockDb,
		writeLog: func(format string, a ...any) {

		},
	}
	addr := common.HexToAddress("0x1234")
	value := uint256.NewInt(100)
	mockDb.EXPECT().SetBalance(addr, value)

	proxy.SetBalance(addr, value)
}
func TestLoggingBulkLoad_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockBulkLoad(ctrl)
	proxy := &loggingBulkLoad{
		nested: mockDb,
		writeLog: func(format string, a ...any) {

		},
	}
	addr := common.HexToAddress("0x1234")
	nonce := uint64(42)
	mockDb.EXPECT().SetNonce(addr, nonce)

	proxy.SetNonce(addr, nonce)
}
func TestLoggingBulkLoad_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockBulkLoad(ctrl)
	proxy := &loggingBulkLoad{
		nested: mockDb,
		writeLog: func(format string, a ...any) {

		},
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	mockDb.EXPECT().SetState(addr, key, value)

	proxy.SetState(addr, key, value)
}
func TestLoggingBulkLoad_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockBulkLoad(ctrl)
	proxy := &loggingBulkLoad{
		nested: mockDb,
		writeLog: func(format string, a ...any) {

		},
	}
	addr := common.HexToAddress("0x1234")
	code := []byte{0x01, 0x02}
	mockDb.EXPECT().SetCode(addr, code)

	proxy.SetCode(addr, code)
}
func TestLoggingBulkLoad_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockBulkLoad(ctrl)
	proxy := &loggingBulkLoad{
		nested: mockDb,
		writeLog: func(format string, a ...any) {

		},
	}
	mockDb.EXPECT().Close().Return(nil)

	err := proxy.Close()
	assert.NoError(t, err)
}
func TestLoggingVmStateDb_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().CreateAccount(addr)

	proxy.CreateAccount(addr)
}
func TestLoggingVmStateDb_Exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().Exist(addr).Return(true)

	exists := proxy.Exist(addr)
	assert.True(t, exists)
}
func TestLoggingVmStateDb_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().Empty(addr).Return(true)

	isEmpty := proxy.Empty(addr)
	assert.True(t, isEmpty)
}
func TestLoggingVmStateDb_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().SelfDestruct(addr)

	proxy.SelfDestruct(addr)
}
func TestLoggingVmStateDb_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().HasSelfDestructed(addr).Return(true)

	hasDestructed := proxy.HasSelfDestructed(addr)
	assert.True(t, hasDestructed)
}
func TestLoggingVmStateDb_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().GetBalance(addr).Return(nil)

	balance := proxy.GetBalance(addr)
	assert.Nil(t, balance)
}
func TestLoggingVmStateDb_AddBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	value := uint256.NewInt(0)
	reason := tracing.BalanceChangeUnspecified
	expectedBalance := uint256.NewInt(999)
	mockDb.EXPECT().AddBalance(addr, value, reason).Return(*expectedBalance)
	mockDb.EXPECT().GetBalance(addr).Return(expectedBalance)
	balance := proxy.AddBalance(addr, value, reason)
	assert.Equal(t, *expectedBalance, balance)
}
func TestLoggingVmStateDb_SubBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	value := uint256.NewInt(0)
	reason := tracing.BalanceChangeUnspecified
	expectedBalance := uint256.NewInt(999)
	mockDb.EXPECT().SubBalance(addr, value, reason).Return(*expectedBalance)
	mockDb.EXPECT().GetBalance(addr).Return(expectedBalance)
	balance := proxy.SubBalance(addr, value, reason)
	assert.Equal(t, *expectedBalance, balance)
}
func TestLoggingVmStateDb_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	expectedNonce := uint64(42)
	mockDb.EXPECT().GetNonce(addr).Return(expectedNonce)

	nonce := proxy.GetNonce(addr)
	assert.Equal(t, expectedNonce, nonce)
}
func TestLoggingVmStateDb_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	nonce := uint64(42)
	reason := tracing.NonceChangeUnspecified
	mockDb.EXPECT().SetNonce(addr, nonce, reason)

	proxy.SetNonce(addr, nonce, reason)
}
func TestLoggingVmStateDb_GetCommittedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetCommittedState(addr, key).Return(expected)

	res := proxy.GetCommittedState(addr, key)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_GetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetState(addr, key).Return(expected)

	res := proxy.GetState(addr, key)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_SetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	expected := common.HexToHash("0xdef0")
	mockDb.EXPECT().SetState(addr, key, value).Return(expected)

	res := proxy.SetState(addr, key, value)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_SetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	value := common.HexToHash("0x9abc")
	mockDb.EXPECT().SetTransientState(addr, key, value)

	proxy.SetTransientState(addr, key, value)
}

func TestLoggingVmStateDb_GetTransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	key := common.HexToHash("0x5678")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetTransientState(addr, key).Return(expected)

	res := proxy.GetTransientState(addr, key)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	expected := []byte{0x01, 0x02}
	mockDb.EXPECT().GetCode(addr).Return(expected)

	res := proxy.GetCode(addr)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	expected := 42
	mockDb.EXPECT().GetCodeSize(addr).Return(expected)

	res := proxy.GetCodeSize(addr)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetCodeHash(addr).Return(expected)

	res := proxy.GetCodeHash(addr)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	code := []byte{0x01, 0x02}
	expected := []byte{0x03, 0x04}
	mockDb.EXPECT().SetCode(addr, code).Return(expected)

	res := proxy.SetCode(addr, code)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_Snapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	expected := 42
	mockDb.EXPECT().Snapshot().Return(expected)

	res := proxy.Snapshot()
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_RevertToSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	id := 42
	mockDb.EXPECT().RevertToSnapshot(id)

	proxy.RevertToSnapshot(id)
}

func TestLoggingVmStateDb_BeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	tx := uint32(1)
	mockDb.EXPECT().BeginTransaction(tx).Return(nil)

	err := proxy.BeginTransaction(tx)
	assert.NoError(t, err)
}

func TestLoggingVmStateDb_EndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	mockDb.EXPECT().EndTransaction().Return(nil)

	err := proxy.EndTransaction()
	assert.NoError(t, err)
}

func TestLoggingVmStateDb_Finalise(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	mockDb.EXPECT().Finalise(true)

	proxy.Finalise(true)
}

func TestLoggingVmStateDb_AddRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	amount := uint64(100)
	mockDb.EXPECT().AddRefund(amount)
	mockDb.EXPECT().GetRefund().Return(amount)

	proxy.AddRefund(amount)
}

func TestLoggingVmStateDb_SubRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	amount := uint64(100)
	mockDb.EXPECT().SubRefund(amount)
	mockDb.EXPECT().GetRefund().Return(amount)

	proxy.SubRefund(amount)
}

func TestLoggingVmStateDb_GetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	expected := uint64(100)
	mockDb.EXPECT().GetRefund().Return(expected)

	res := proxy.GetRefund()
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	rules := params.TestRules
	sender := common.HexToAddress("0x1234")
	coinbase := common.HexToAddress("0x5678")
	dest := &common.Address{}
	precompiles := []common.Address{}
	accessList := types.AccessList{}
	mockDb.EXPECT().Prepare(rules, sender, coinbase, dest, precompiles, accessList)

	proxy.Prepare(rules, sender, coinbase, dest, precompiles, accessList)
}

func TestLoggingVmStateDb_AddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().AddressInAccessList(addr).Return(true)

	res := proxy.AddressInAccessList(addr)
	assert.True(t, res)
}

func TestLoggingVmStateDb_SlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	mockDb.EXPECT().SlotInAccessList(addr, slot).Return(true, false)

	a, b := proxy.SlotInAccessList(addr, slot)
	assert.True(t, a)
	assert.False(t, b)
}

func TestLoggingVmStateDb_AddAddressToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().AddAddressToAccessList(addr)

	proxy.AddAddressToAccessList(addr)
}

func TestLoggingVmStateDb_AddSlotToAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	slot := common.HexToHash("0x5678")
	mockDb.EXPECT().AddSlotToAccessList(addr, slot)

	proxy.AddSlotToAccessList(addr, slot)
}

func TestLoggingVmStateDb_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	entry := &types.Log{}
	mockDb.EXPECT().AddLog(entry)

	proxy.AddLog(entry)
}

func TestLoggingVmStateDb_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	hash := common.HexToHash("0x1234")
	block := uint64(1)
	blockHash := common.HexToHash("0x5678")
	blkTimestamp := uint64(42)
	expected := []*types.Log{}
	mockDb.EXPECT().GetLogs(hash, block, blockHash, blkTimestamp).Return(expected)

	res := proxy.GetLogs(hash, block, blockHash, blkTimestamp)
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_PointCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	expected := &utils.PointCache{}
	mockDb.EXPECT().PointCache().Return(expected)

	res := proxy.PointCache()
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_Witness(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	expected := &stateless.Witness{}
	mockDb.EXPECT().Witness().Return(expected)

	res := proxy.Witness()
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_SetTxContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	thash := common.HexToHash("0x1234")
	ti := 42
	mockDb.EXPECT().SetTxContext(thash, ti)

	proxy.SetTxContext(thash, ti)
}

func TestLoggingVmStateDb_GetSubstatePostAlloc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	mockWorldState := txcontext.NewMockWorldState(ctrl)
	mockDb.EXPECT().GetSubstatePostAlloc().Return(mockWorldState)
	mockWorldState.EXPECT().String()

	res := proxy.GetSubstatePostAlloc()
	assert.Equal(t, mockWorldState, res)
}

func TestLoggingVmStateDb_AddPreimage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	hash := common.HexToHash("0x1234")
	data := []byte{0x01, 0x02}
	mockDb.EXPECT().AddPreimage(hash, data)

	proxy.AddPreimage(hash, data)
}

func TestLoggingVmStateDb_AccessEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	expected := &geth.AccessEvents{}
	mockDb.EXPECT().AccessEvents().Return(expected)

	res := proxy.AccessEvents()
	assert.Equal(t, expected, res)
}

func TestLoggingVmStateDb_CreateContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	mockDb.EXPECT().CreateContract(addr)

	proxy.CreateContract(addr)
}

func TestLoggingVmStateDb_SelfDestruct6780(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	expectedBalance := uint256.NewInt(999)
	mockDb.EXPECT().SelfDestruct6780(addr).Return(*expectedBalance, true)

	balance, success := proxy.SelfDestruct6780(addr)
	assert.Equal(t, *expectedBalance, balance)
	assert.True(t, success)
}

func TestLoggingVmStateDb_GetStorageRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewLogger("info", "test")
	mockChan := make(chan string, 1)
	mockWg := &sync.WaitGroup{}

	proxy := &loggingVmStateDb{
		db:     mockDb,
		log:    mockLogger,
		output: mockChan,
		wg:     mockWg,
	}
	addr := common.HexToAddress("0x1234")
	expected := common.HexToHash("0x9abc")
	mockDb.EXPECT().GetStorageRoot(addr).Return(expected)

	res := proxy.GetStorageRoot(addr)
	assert.Equal(t, expected, res)
}
