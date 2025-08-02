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

package executor

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/tosca/go/tosca"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestPrepareBlockCtx tests a creation of block context from substate environment.
func TestPrepareBlockCtx(t *testing.T) {
	gaslimit := uint64(10000000)
	blocknum := uint64(4600000)
	basefee := big.NewInt(12345)
	env := substatecontext.NewBlockEnvironment(&substate.Env{Difficulty: big.NewInt(1), GasLimit: gaslimit, Number: blocknum, Timestamp: 1675961395, BaseFee: basefee})

	var hashError error
	// BlockHashes are nil, expect an error
	blockCtx := utils.PrepareBlockCtx(env, &hashError)

	if blocknum != blockCtx.BlockNumber.Uint64() {
		t.Fatalf("Wrong block number")
	}
	if gaslimit != blockCtx.GasLimit {
		t.Fatalf("Wrong amount of gas limit")
	}
	if basefee.Cmp(blockCtx.BaseFee) != 0 {
		t.Fatalf("Wrong base fee")
	}
	if hashError != nil {
		t.Fatalf("Hash error; %v", hashError)
	}
}

func TestMakeTxProcessor_CanSelectBetweenProcessorImplementations(t *testing.T) {
	isAida := func(t *testing.T, p processor, name string) {
		_, ok := p.(*aidaProcessor)
		if !ok {
			t.Fatalf("Expected aidaProcessor from '%s', got %T", name, p)
		}
	}
	isTosca := func(t *testing.T, p processor, name string) {
		if _, ok := p.(*toscaProcessor); !ok {
			t.Fatalf("Expected toscaProcessor from '%s', got %T", name, p)
		}
	}

	tests := map[string]func(*testing.T, processor, string){
		"":         isAida,
		"opera":    isAida,
		"ethereum": isAida,
	}

	for name := range tosca.GetAllRegisteredProcessorFactories() {
		if _, present := tests[name]; !present {
			tests[name] = isTosca
		}
	}

	for name, check := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &utils.Config{
				ChainID: utils.MainnetChainID,
				EvmImpl: name,
				VmImpl:  "geth",
			}
			p, err := MakeTxProcessor(cfg)
			if err != nil {
				t.Fatalf("Failed to create tx processor; %v", err)
			}
			check(t, p.processor, name)
		})
	}

}

func TestMakeTxProcessor_InvalidVmImplCausesError(t *testing.T) {
	cfg := &utils.Config{
		ChainID: utils.MainnetChainID,
		EvmImpl: "tosca",
		VmImpl:  "invalid",
	}
	_, err := MakeTxProcessor(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create interpreter invalid, error interpreter not found: invalid")
}

func TestMakeTxProcessor_InvalidEvmImplCausesError(t *testing.T) {
	cfg := &utils.Config{
		ChainID: utils.MainnetChainID,
		EvmImpl: "invalid",
		VmImpl:  "lfvm",
	}
	_, err := MakeTxProcessor(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown EVM implementation: invalid")
}

func TestEthTestProcessor_DoesNotExecuteTransactionWhenBlobGasCouldExceed(t *testing.T) {
	p, err := MakeEthTestProcessor(&utils.Config{})
	if err != nil {
		t.Fatalf("cannot make eth test processor: %v", err)
	}
	ctrl := gomock.NewController(t)
	// Process is returned early - nothing is expected
	stateDb := state.NewMockStateDB(ctrl)

	ctx := &Context{State: stateDb}
	err = p.Process(State[txcontext.TxContext]{Data: ethtest.CreateTransactionThatFailsBlobGasExceedCheck(t)}, ctx)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	_, got := ctx.ExecutionResult.GetRawResult()
	want := "blob gas exceeds maximum"
	if !strings.EqualFold(got.Error(), want) {
		t.Errorf("unexpected error, got: %v, want: %v", got, want)
	}
}

func TestEthTestProcessor_DoesNotExecuteTransactionWithInvalidTxBytes(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
		data          txcontext.TxContext
	}{
		{
			name:          "fails_unmarshal",
			expectedError: "cannot unmarshal tx-bytes",
			data:          ethtest.CreateTransactionWithInvalidTxBytes(t),
		},
		{
			name:          "fails_validation",
			expectedError: "cannot validate sender",
			data:          ethtest.CreateTransactionThatFailsSenderValidation(t),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, err := MakeEthTestProcessor(&utils.Config{ChainID: utils.EthTestsChainID})
			if err != nil {
				t.Fatalf("cannot make eth test processor: %v", err)
			}
			ctrl := gomock.NewController(t)
			// Process is returned early - no calls are expected
			stateDb := state.NewMockStateDB(ctrl)

			ctx := &Context{State: stateDb}
			err = p.Process(State[txcontext.TxContext]{Data: test.data}, ctx)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			_, got := ctx.ExecutionResult.GetRawResult()
			if !strings.Contains(got.Error(), test.expectedError) {
				t.Errorf("unexpected error, got: %v, want: %v", got, test.expectedError)
			}
		})
	}
}

func TestMessageResult(t *testing.T) {
	e := errors.New("error")
	res := executionResult(messageResult{
		failed:     true,
		returnData: []byte{0x12},
		gasUsed:    10,
		err:        e,
	})

	require.True(t, res.Failed())
	require.Equal(t, res.Return(), []byte{0x12})
	require.Equal(t, res.GetGasUsed(), uint64(10))
	require.ErrorIs(t, res.GetError(), e)
}

// TestToscaTxContext_CreateAccount tests the CreateAccount method of toscaTxContext
func TestToscaTxContext_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	// Test case 1: Account doesn't exist
	mockStateDB.EXPECT().Exist(ethAddr).Return(false)
	mockStateDB.EXPECT().CreateAccount(ethAddr)
	mockStateDB.EXPECT().CreateContract(ethAddr)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	ctx.CreateAccount(addr)

	// Test case 2: Account already exists
	mockStateDB.EXPECT().Exist(ethAddr).Return(true)
	mockStateDB.EXPECT().CreateContract(ethAddr)

	ctx.CreateAccount(addr)
}

// TestToscaTxContext_HasEmptyStorage tests the HasEmptyStorage method of toscaTxContext
func TestToscaTxContext_HasEmptyStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	tests := []struct {
		name     string
		rootHash common.Hash
		expected bool
	}{
		{
			name:     "empty_hash",
			rootHash: common.Hash{},
			expected: true,
		},
		{
			name:     "empty_root_hash",
			rootHash: types.EmptyRootHash,
			expected: true,
		},
		{
			name:     "non_empty_hash",
			rootHash: common.HexToHash("0x1234"),
			expected: false,
		},
	}

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStateDB.EXPECT().GetStorageRoot(ethAddr).Return(tt.rootHash)
			result := ctx.HasEmptyStorage(addr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestToscaTxContext_AccountExists tests the AccountExists method of toscaTxContext
func TestToscaTxContext_AccountExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	tests := []struct {
		name   string
		exists bool
	}{
		{
			name:   "account_exists",
			exists: true,
		},
		{
			name:   "account_does_not_exist",
			exists: false,
		},
	}

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStateDB.EXPECT().Exist(ethAddr).Return(tt.exists)
			result := ctx.AccountExists(addr)
			assert.Equal(t, tt.exists, result)
		})
	}
}

// TestToscaTxContext_GetBalance tests the GetBalance method of toscaTxContext
func TestToscaTxContext_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	balance := uint256.NewInt(1000)

	mockStateDB.EXPECT().GetBalance(ethAddr).Return(balance)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	result := ctx.GetBalance(addr)

	// Convert the returned tosca.Value back to uint256 for comparison
	assert.Equal(t, balance.Uint64(), result.ToUint256().Uint64())
}

// TestToscaTxContext_SetBalance tests the SetBalance method of toscaTxContext
func TestToscaTxContext_SetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	tests := []struct {
		name        string
		curBalance  *uint256.Int
		wantBalance *uint256.Int
		expectCall  bool
	}{
		{
			name:        "equal_balance",
			curBalance:  uint256.NewInt(1000),
			wantBalance: uint256.NewInt(1000),
			expectCall:  false,
		},
		{
			name:        "increase_balance",
			curBalance:  uint256.NewInt(500),
			wantBalance: uint256.NewInt(1000),
			expectCall:  true,
		},
		{
			name:        "decrease_balance",
			curBalance:  uint256.NewInt(1000),
			wantBalance: uint256.NewInt(500),
			expectCall:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockStateDB.EXPECT().GetBalance(ethAddr).Return(tt.curBalance)

			if tt.expectCall {
				if tt.curBalance.Cmp(tt.wantBalance) > 0 {
					// Decrease balance
					diff := new(uint256.Int).Sub(tt.curBalance, tt.wantBalance)
					mockStateDB.EXPECT().SubBalance(ethAddr, diff, gomock.Any())
				} else {
					// Increase balance
					diff := new(uint256.Int).Sub(tt.wantBalance, tt.curBalance)
					mockStateDB.EXPECT().AddBalance(ethAddr, diff, gomock.Any())
				}
			}

			wantValue := tosca.ValueFromUint256(tt.wantBalance)
			ctx.SetBalance(addr, wantValue)
		})
	}
}

// TestToscaTxContext_GetNonce tests the GetNonce method of toscaTxContext
func TestToscaTxContext_GetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	expectedNonce := uint64(10)
	mockStateDB.EXPECT().GetNonce(ethAddr).Return(expectedNonce)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	nonce := ctx.GetNonce(addr)
	assert.Equal(t, expectedNonce, nonce)
}

// TestToscaTxContext_SetNonce tests the SetNonce method of toscaTxContext
func TestToscaTxContext_SetNonce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	nonce := uint64(5)
	mockStateDB.EXPECT().SetNonce(ethAddr, nonce, gomock.Any())

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	ctx.SetNonce(addr, nonce)
}

// TestToscaTxContext_GetCodeSize tests the GetCodeSize method of toscaTxContext
func TestToscaTxContext_GetCodeSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	expectedSize := 100
	mockStateDB.EXPECT().GetCodeSize(ethAddr).Return(expectedSize)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	size := ctx.GetCodeSize(addr)
	assert.Equal(t, expectedSize, size)
}

// TestToscaTxContext_GetCodeHash tests the GetCodeHash method of toscaTxContext
func TestToscaTxContext_GetCodeHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	expectedHash := common.HexToHash("0x9876")
	mockStateDB.EXPECT().GetCodeHash(ethAddr).Return(expectedHash)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	hash := ctx.GetCodeHash(addr)
	assert.Equal(t, tosca.Hash(expectedHash), hash)
}

// TestToscaTxContext_GetCode tests the GetCode method of toscaTxContext
func TestToscaTxContext_GetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	expectedCode := tosca.Code{0x1, 0x2, 0x3, 0x4}
	mockStateDB.EXPECT().GetCode(ethAddr).Return(expectedCode)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	code := ctx.GetCode(addr)
	assert.Equal(t, expectedCode, code)
}

// TestToscaTxContext_SetCode tests the SetCode method of toscaTxContext
func TestToscaTxContext_SetCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	code := []byte{1, 2, 3, 4}
	mockStateDB.EXPECT().SetCode(ethAddr, code)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	ctx.SetCode(addr, code)
}

// TestToscaTxContext_GetStorage tests the GetStorage method of toscaTxContext
func TestToscaTxContext_GetStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	key := tosca.Key(common.HexToHash("0xABCD"))
	ethKey := common.Hash(key)

	expectedValue := common.HexToHash("0xDEAD")
	mockStateDB.EXPECT().GetState(ethAddr, ethKey).Return(expectedValue)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	value := ctx.GetStorage(addr, key)
	assert.Equal(t, tosca.Word(expectedValue), value)
}

// TestToscaTxContext_GetCommittedStorage tests the GetCommittedStorage method of toscaTxContext
func TestToscaTxContext_GetCommittedStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	key := tosca.Key(common.HexToHash("0xABCD"))
	ethKey := common.Hash(key)

	expectedValue := common.HexToHash("0xBEEF")
	mockStateDB.EXPECT().GetCommittedState(ethAddr, ethKey).Return(expectedValue)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	value := ctx.GetCommittedStorage(addr, key)
	assert.Equal(t, tosca.Word(expectedValue), value)
}

// TestToscaTxContext_SetStorage tests the SetStorage method of toscaTxContext
func TestToscaTxContext_SetStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	key := tosca.Key(common.HexToHash("0xABCD"))
	ethKey := common.Hash(key)

	tests := []struct {
		name     string
		original tosca.Word
		current  tosca.Word
		newValue tosca.Word
		expected tosca.StorageStatus
	}{
		{
			name:     "unchanged",
			original: tosca.Word(common.HexToHash("0x1234")),
			current:  tosca.Word(common.HexToHash("0x1234")),
			newValue: tosca.Word(common.HexToHash("0x1234")),
			expected: tosca.StorageAssigned,
		},
		{
			name:     "added",
			original: tosca.Word{},
			current:  tosca.Word{},
			newValue: tosca.Word(common.HexToHash("0x1234")),
			expected: tosca.StorageAdded,
		},
		{
			name:     "deleted",
			original: tosca.Word(common.HexToHash("0x1234")),
			current:  tosca.Word(common.HexToHash("0x1234")),
			newValue: tosca.Word{},
			expected: tosca.StorageDeleted,
		},
		{
			name:     "modified",
			original: tosca.Word(common.HexToHash("0x1234")),
			current:  tosca.Word(common.HexToHash("0x1234")),
			newValue: tosca.Word(common.HexToHash("0x5678")),
			expected: tosca.StorageModified,
		},
		{
			name:     "modified_to_original",
			original: tosca.Word(common.HexToHash("0x1234")),
			current:  tosca.Word(common.HexToHash("0x5678")),
			newValue: tosca.Word(common.HexToHash("0x1234")),
			expected: tosca.StorageModifiedRestored,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockStateDB.EXPECT().GetCommittedState(ethAddr, ethKey).Return(common.Hash(tt.original))
			mockStateDB.EXPECT().GetState(ethAddr, ethKey).Return(common.Hash(tt.current))
			mockStateDB.EXPECT().SetState(ethAddr, ethKey, common.Hash(tt.newValue))

			status := ctx.SetStorage(addr, key, tt.newValue)
			assert.Equal(t, tt.expected, status)
		})
	}
}

// TestToscaTxContext_GetTransientStorage tests the GetTransientStorage method of toscaTxContext
func TestToscaTxContext_GetTransientStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	key := tosca.Key(common.HexToHash("0xABCD"))
	ethKey := common.Hash(key)

	expectedValue := common.HexToHash("0xCAFE")
	mockStateDB.EXPECT().GetTransientState(ethAddr, ethKey).Return(expectedValue)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	value := ctx.GetTransientStorage(addr, key)
	assert.Equal(t, tosca.Word(expectedValue), value)
}

// TestToscaTxContext_SetTransientStorage tests the SetTransientStorage method of toscaTxContext
func TestToscaTxContext_SetTransientStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	key := tosca.Key(common.HexToHash("0xABCD"))
	ethKey := common.Hash(key)
	value := tosca.Word(common.HexToHash("0xCAFE"))
	ethValue := common.Hash(value)

	mockStateDB.EXPECT().SetTransientState(ethAddr, ethKey, ethValue)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	ctx.SetTransientStorage(addr, key, value)
}

// TestToscaTxContext_GetBlockHash tests the GetBlockHash method of toscaTxContext
func TestToscaTxContext_GetBlockHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	blockNum := int64(12345)
	expectedHash := common.HexToHash("0xBLOCKHASH")

	mockBlockEnv.EXPECT().GetBlockHash(uint64(blockNum)).Return(expectedHash, nil)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	hash := ctx.GetBlockHash(blockNum)
	assert.Equal(t, tosca.Hash(expectedHash), hash)
}

// TestToscaTxContext_EmitLog tests the EmitLog method of toscaTxContext
func TestToscaTxContext_EmitLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	topics := []tosca.Hash{
		tosca.Hash(common.HexToHash("0xTOPIC1")),
		tosca.Hash(common.HexToHash("0xTOPIC2")),
	}
	data := []byte{1, 2, 3, 4}

	log := tosca.Log{
		Address: addr,
		Topics:  topics,
		Data:    data,
	}

	// Expected Ethereum log
	ethTopics := []common.Hash{
		common.Hash(topics[0]),
		common.Hash(topics[1]),
	}

	mockStateDB.EXPECT().AddLog(gomock.Any()).Do(func(ethLog *types.Log) {
		assert.Equal(t, common.Address(addr), ethLog.Address)
		assert.Equal(t, ethTopics, ethLog.Topics)
		assert.Equal(t, data, ethLog.Data)
	})

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	ctx.EmitLog(log)
}

// TestToscaTxContext_GetLogs tests the GetLogs method of toscaTxContext
func TestToscaTxContext_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	// Create some Ethereum logs
	ethLogs := []*types.Log{
		{
			Address: common.HexToAddress("0x1234"),
			Topics: []common.Hash{
				common.HexToHash("0xTOPIC1"),
				common.HexToHash("0xTOPIC2"),
			},
			Data: []byte{0x1, 0x2, 0x3},
		},
		{
			Address: common.HexToAddress("0x5678"),
			Topics: []common.Hash{
				common.HexToHash("0xTOPIC3"),
			},
			Data: []byte{0x4, 0x5, 0x6},
		},
	}

	mockStateDB.EXPECT().GetLogs(common.Hash{}, uint64(0), common.Hash{}, uint64(0)).Return(ethLogs)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	logs := ctx.GetLogs()

	// Check conversion to tosca logs
	assert.Len(t, logs, 2)

	// Check first log
	assert.Equal(t, tosca.Address(ethLogs[0].Address), logs[0].Address)
	assert.Len(t, logs[0].Topics, 2)
	assert.Equal(t, tosca.Hash(ethLogs[0].Topics[0]), logs[0].Topics[0])
	assert.Equal(t, tosca.Hash(ethLogs[0].Topics[1]), logs[0].Topics[1])
	assert.Equal(t, ethLogs[0].Data, []byte(logs[0].Data))

	// Check second log
	assert.Equal(t, tosca.Address(ethLogs[1].Address), logs[1].Address)
	assert.Len(t, logs[1].Topics, 1)
	assert.Equal(t, tosca.Hash(ethLogs[1].Topics[0]), logs[1].Topics[0])
	assert.Equal(t, ethLogs[1].Data, []byte(logs[1].Data))
}

// TestToscaTxContext_SelfDestruct tests the SelfDestruct method of toscaTxContext
func TestToscaTxContext_SelfDestruct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	beneficiary := tosca.Address(common.HexToAddress("0x5678"))

	tests := []struct {
		name              string
		fork              string
		hasSelfDestructed bool
		expected          bool
	}{
		{
			name:              "cancun_not_self_destructed",
			fork:              tosca.R13_Cancun.String(),
			hasSelfDestructed: false,
			expected:          true,
		},
		{
			name:              "cancun_already_self_destructed",
			fork:              tosca.R13_Cancun.String(),
			hasSelfDestructed: true,
			expected:          false,
		},
		{
			name:              "other_fork_not_self_destructed",
			fork:              tosca.R12_Shanghai.String(),
			hasSelfDestructed: false,
			expected:          true,
		},
		{
			name:              "other_fork_already_self_destructed",
			fork:              tosca.R12_Shanghai.String(),
			hasSelfDestructed: true,
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockBlockEnv.EXPECT().GetFork().Return(tt.fork)
			mockStateDB.EXPECT().HasSelfDestructed(ethAddr).Return(tt.hasSelfDestructed)

			if tt.fork == tosca.R13_Cancun.String() {
				mockStateDB.EXPECT().SelfDestruct6780(ethAddr)
			} else {
				mockStateDB.EXPECT().SelfDestruct(ethAddr)
			}

			result := ctx.SelfDestruct(addr, beneficiary)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestToscaTxContext_AccessAccount tests the AccessAccount method of toscaTxContext
func TestToscaTxContext_AccessAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	tests := []struct {
		name     string
		inList   bool
		expected tosca.AccessStatus
	}{
		{
			name:     "warm_access",
			inList:   true,
			expected: tosca.WarmAccess,
		},
		{
			name:     "cold_access",
			inList:   false,
			expected: tosca.ColdAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockStateDB.EXPECT().AddressInAccessList(ethAddr).Return(tt.inList)
			mockStateDB.EXPECT().AddAddressToAccessList(ethAddr)

			status := ctx.AccessAccount(addr)
			assert.Equal(t, tt.expected, status)
		})
	}
}

// TestToscaTxContext_AccessStorage tests the AccessStorage method of toscaTxContext
func TestToscaTxContext_AccessStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	key := tosca.Key(common.HexToHash("0xABCD"))
	ethKey := common.Hash(key)

	tests := []struct {
		name        string
		addrPresent bool
		slotPresent bool
		expected    tosca.AccessStatus
	}{
		{
			name:        "warm_access",
			addrPresent: true,
			slotPresent: true,
			expected:    tosca.WarmAccess,
		},
		{
			name:        "cold_access_addr_present",
			addrPresent: true,
			slotPresent: false,
			expected:    tosca.ColdAccess,
		},
		{
			name:        "cold_access_nothing_present",
			addrPresent: false,
			slotPresent: false,
			expected:    tosca.ColdAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockStateDB.EXPECT().SlotInAccessList(ethAddr, ethKey).Return(tt.addrPresent, tt.slotPresent)
			mockStateDB.EXPECT().AddSlotToAccessList(ethAddr, ethKey)

			status := ctx.AccessStorage(addr, key)
			assert.Equal(t, tt.expected, status)
		})
	}
}

// TestToscaTxContext_HasSelfDestructed tests the HasSelfDestructed method of toscaTxContext
func TestToscaTxContext_HasSelfDestructed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	tests := []struct {
		name   string
		result bool
	}{
		{
			name:   "has_self_destructed",
			result: true,
		},
		{
			name:   "has_not_self_destructed",
			result: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockStateDB.EXPECT().HasSelfDestructed(ethAddr).Return(tt.result)

			result := ctx.HasSelfDestructed(addr)
			assert.Equal(t, tt.result, result)
		})
	}
}

// TestToscaTxContext_CreateSnapshot tests the CreateSnapshot method of toscaTxContext
func TestToscaTxContext_CreateSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	snapshotId := 42
	mockStateDB.EXPECT().Snapshot().Return(snapshotId)

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	snapshot := ctx.CreateSnapshot()
	assert.Equal(t, tosca.Snapshot(snapshotId), snapshot)
}

// TestToscaTxContext_RestoreSnapshot tests the RestoreSnapshot method of toscaTxContext
func TestToscaTxContext_RestoreSnapshot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	snapshotId := tosca.Snapshot(42)
	mockStateDB.EXPECT().RevertToSnapshot(int(snapshotId))

	ctx := &toscaTxContext{
		blockEnvironment: mockBlockEnv,
		db:               mockStateDB,
	}

	ctx.RestoreSnapshot(snapshotId)
}

// TestToscaTxContext_IsAddressInAccessList tests the IsAddressInAccessList method of toscaTxContext
func TestToscaTxContext_IsAddressInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)

	tests := []struct {
		name   string
		result bool
	}{
		{
			name:   "address_in_list",
			result: true,
		},
		{
			name:   "address_not_in_list",
			result: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockStateDB.EXPECT().AddressInAccessList(ethAddr).Return(tt.result)

			result := ctx.IsAddressInAccessList(addr)
			assert.Equal(t, tt.result, result)
		})
	}
}

// TestToscaTxContext_IsSlotInAccessList tests the IsSlotInAccessList method of toscaTxContext
func TestToscaTxContext_IsSlotInAccessList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)

	addr := tosca.Address(common.HexToAddress("0x1234"))
	ethAddr := common.Address(addr)
	key := tosca.Key(common.HexToHash("0xABCD"))
	ethKey := common.Hash(key)

	tests := []struct {
		name        string
		addrPresent bool
		slotPresent bool
	}{
		{
			name:        "both_present",
			addrPresent: true,
			slotPresent: true,
		},
		{
			name:        "addr_present_slot_not",
			addrPresent: true,
			slotPresent: false,
		},
		{
			name:        "neither_present",
			addrPresent: false,
			slotPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &toscaTxContext{
				blockEnvironment: mockBlockEnv,
				db:               mockStateDB,
			}

			mockStateDB.EXPECT().SlotInAccessList(ethAddr, ethKey).Return(tt.addrPresent, tt.slotPresent)

			addrPresent, slotPresent := ctx.IsSlotInAccessList(addr, key)
			assert.Equal(t, tt.addrPresent, addrPresent)
			assert.Equal(t, tt.slotPresent, slotPresent)
		})
	}
}

// TestBigToValue tests the bigToValue function
func TestExecutor_BigToValue(t *testing.T) {
	tests := []struct {
		name  string
		input *big.Int
		check func(t *testing.T, result tosca.Value)
	}{
		{
			name:  "nil_value",
			input: nil,
			check: func(t *testing.T, result tosca.Value) {
				assert.Equal(t, tosca.Value{}, result)
			},
		},
		{
			name:  "zero_value",
			input: big.NewInt(0),
			check: func(t *testing.T, result tosca.Value) {
				assert.Equal(t, uint64(0), result.ToUint256().Uint64())
			},
		},
		{
			name:  "small_value",
			input: big.NewInt(42),
			check: func(t *testing.T, result tosca.Value) {
				assert.Equal(t, uint64(42), result.ToUint256().Uint64())
			},
		},
		{
			name:  "large_value",
			input: new(big.Int).Exp(big.NewInt(2), big.NewInt(100), nil),
			check: func(t *testing.T, result tosca.Value) {
				expected := new(big.Int).Exp(big.NewInt(2), big.NewInt(100), nil)
				assert.Equal(t, expected.String(), result.ToUint256().ToBig().String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bigToValue(tt.input)
			tt.check(t, result)
		})
	}
}

// TestUint256ToValue tests the uint256ToValue function
func TestExecutor_Uint256ToValue(t *testing.T) {
	tests := []struct {
		name  string
		input *uint256.Int
		check func(t *testing.T, result tosca.Value)
	}{
		{
			name:  "zero_value",
			input: uint256.NewInt(0),
			check: func(t *testing.T, result tosca.Value) {
				assert.Equal(t, uint64(0), result.ToUint256().Uint64())
			},
		},
		{
			name:  "small_value",
			input: uint256.NewInt(42),
			check: func(t *testing.T, result tosca.Value) {
				assert.Equal(t, uint64(42), result.ToUint256().Uint64())
			},
		},
		{
			name:  "large_value",
			input: new(uint256.Int).Exp(uint256.NewInt(2), uint256.NewInt(100)),
			check: func(t *testing.T, result tosca.Value) {
				expected := new(uint256.Int).Exp(uint256.NewInt(2), uint256.NewInt(100))
				assert.True(t, expected.Eq(result.ToUint256()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uint256ToValue(tt.input)
			tt.check(t, result)
		})
	}
}

func TestToscaProcessor_processRegularTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockTxContext := txcontext.NewMockTxContext(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)
	mockToscaProcessor := tosca.NewMockProcessor(ctrl)

	// Setup test data
	blockNum := 12345
	txNum := 1
	sender := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recipient := common.HexToAddress("0x0987654321098765432109876543210987654321")

	// Create test message
	message := &core.Message{
		From:      sender,
		To:        &recipient,
		Nonce:     10,
		Value:     big.NewInt(1000),
		GasLimit:  21048,
		GasPrice:  big.NewInt(50),
		GasFeeCap: big.NewInt(0),
		GasTipCap: big.NewInt(0),
		Data:      []byte{1, 2, 3},
	}

	// Setup mock behaviors
	mockTxContext.EXPECT().GetBlockEnvironment().Return(mockBlockEnv).AnyTimes()
	mockTxContext.EXPECT().GetMessage().Return(message).AnyTimes()

	mockBlockEnv.EXPECT().GetFork().Return("cancun").AnyTimes()
	mockBlockEnv.EXPECT().GetNumber().Return(uint64(blockNum)).AnyTimes()
	mockBlockEnv.EXPECT().GetTimestamp().Return(uint64(1700000000)).AnyTimes() // After Cancun time
	mockBlockEnv.EXPECT().GetBaseFee().Return(big.NewInt(5)).AnyTimes()
	mockBlockEnv.EXPECT().GetBlobBaseFee().Return(big.NewInt(10)).AnyTimes()
	mockBlockEnv.EXPECT().GetGasLimit().Return(uint64(30000000)).AnyTimes()
	mockBlockEnv.EXPECT().GetCoinbase().Return(common.HexToAddress("0xcoinbase")).AnyTimes()
	mockBlockEnv.EXPECT().GetDifficulty().Return(big.NewInt(2)).AnyTimes()
	mockBlockEnv.EXPECT().GetRandom().Return(&common.Hash{}).AnyTimes()

	// Create the processor instance
	cfg := &utils.Config{
		ChainID: utils.MainnetChainID,
	}

	processor := &toscaProcessor{
		processor: mockToscaProcessor,
		cfg:       cfg,
		log:       logger.NewLogger("info", "dummy logger"),
	}

	// Test successful execution
	t.Run("successful_execution", func(t *testing.T) {
		// Setup mock tosca processor response
		successReceipt := tosca.Receipt{
			Success: true,
			GasUsed: 21000,
			Output:  []byte{4, 5, 6},
			Logs: []tosca.Log{
				{
					Address: tosca.Address(sender),
					Topics:  []tosca.Hash{tosca.Hash(common.HexToHash("0xevent"))},
					Data:    []byte{7, 8, 9},
				},
			},
		}

		mockToscaProcessor.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).Return(successReceipt, nil)

		// Execute test
		result, err := processor.processRegularTx(mockStateDB, blockNum, txNum, mockTxContext)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint64(21000), result.gasUsed)
		assert.Equal(t, []byte{4, 5, 6}, result.result)
		assert.Nil(t, result.err)
		assert.Len(t, result.logs, 1)
	})
}

// TestMessageResult_Failed tests the Failed method of messageResult
func TestMessageResult_Failed(t *testing.T) {
	testCases := []struct {
		name       string
		result     messageResult
		wantFailed bool
	}{
		{
			name: "successful_execution",
			result: messageResult{
				failed:     false,
				returnData: []byte{1, 2, 3, 4},
				gasUsed:    21000,
				err:        nil,
			},
			wantFailed: false,
		},
		{
			name: "failed_execution",
			result: messageResult{
				failed:     true,
				returnData: []byte{},
				gasUsed:    100000,
				err:        errors.New("execution failed"),
			},
			wantFailed: true,
		},
		{
			name: "failed_with_return_data",
			result: messageResult{
				failed:     true,
				returnData: []byte{5, 6, 7, 8}, // Return data can exist even on failure (e.g., revert reason)
				gasUsed:    50000,
				err:        errors.New("reverted"),
			},
			wantFailed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantFailed, tc.result.Failed(), "Failed() returned unexpected value")
		})
	}
}

// TestMessageResult_Return tests the Return method of messageResult
func TestMessageResult_Return(t *testing.T) {
	testCases := []struct {
		name       string
		result     messageResult
		wantReturn []byte
	}{
		{
			name: "normal_return_data",
			result: messageResult{
				failed:     false,
				returnData: []byte{1, 2, 3, 4},
				gasUsed:    21000,
				err:        nil,
			},
			wantReturn: []byte{1, 2, 3, 4},
		},
		{
			name: "empty_return_data",
			result: messageResult{
				failed:     true,
				returnData: []byte{},
				gasUsed:    100000,
				err:        errors.New("execution failed"),
			},
			wantReturn: []byte{},
		},
		{
			name: "revert_reason_data",
			result: messageResult{
				failed:     true,
				returnData: []byte{5, 6, 7, 8}, // Return data for revert reason
				gasUsed:    50000,
				err:        errors.New("reverted"),
			},
			wantReturn: []byte{5, 6, 7, 8},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantReturn, tc.result.Return(), "Return() returned unexpected value")
		})
	}
}

// TestMessageResult_GetGasUsed tests the GetGasUsed method of messageResult
func TestMessageResult_GetGasUsed(t *testing.T) {
	testCases := []struct {
		name    string
		result  messageResult
		wantGas uint64
	}{
		{
			name: "standard_gas_usage",
			result: messageResult{
				failed:     false,
				returnData: []byte{1, 2, 3, 4},
				gasUsed:    21000,
				err:        nil,
			},
			wantGas: 21000,
		},
		{
			name: "high_gas_usage",
			result: messageResult{
				failed:     true,
				returnData: []byte{},
				gasUsed:    100000,
				err:        errors.New("execution failed"),
			},
			wantGas: 100000,
		},
		{
			name: "zero_gas_usage",
			result: messageResult{
				failed:     false,
				returnData: []byte{9, 10},
				gasUsed:    0,
				err:        nil,
			},
			wantGas: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantGas, tc.result.GetGasUsed(), "GetGasUsed() returned unexpected value")
		})
	}
}

// TestMessageResult_GetError tests the GetError method of messageResult
func TestMessageResult_GetError(t *testing.T) {
	testCases := []struct {
		name    string
		result  messageResult
		wantErr error
	}{
		{
			name: "no_error",
			result: messageResult{
				failed:     false,
				returnData: []byte{1, 2, 3, 4},
				gasUsed:    21000,
				err:        nil,
			},
			wantErr: nil,
		},
		{
			name: "execution_error",
			result: messageResult{
				failed:     true,
				returnData: []byte{},
				gasUsed:    100000,
				err:        errors.New("execution failed"),
			},
			wantErr: errors.New("execution failed"),
		},
		{
			name: "revert_error",
			result: messageResult{
				failed:     true,
				returnData: []byte{5, 6, 7, 8},
				gasUsed:    50000,
				err:        errors.New("reverted"),
			},
			wantErr: errors.New("reverted"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantErr != nil {
				assert.NotNil(t, tc.result.GetError(), "GetError() should not return nil")
				assert.Equal(t, tc.wantErr.Error(), tc.result.GetError().Error(), "GetError() returned unexpected value")
			} else {
				assert.Nil(t, tc.result.GetError(), "GetError() should return nil")
			}
		})
	}
}

// TestMessageResult_Interface ensures messageResult correctly implements the executionResult interface
func TestMessageResult_Interface(t *testing.T) {
	// Create a messageResult instance
	result := messageResult{
		failed:     false,
		returnData: []byte{1, 2, 3, 4},
		gasUsed:    21000,
		err:        nil,
	}

	// Test that messageResult correctly implements the executionResult interface
	var _ executionResult = result

	// Verify that all interface methods work properly
	assert.Equal(t, false, result.Failed())
	assert.Equal(t, []byte{1, 2, 3, 4}, result.Return())
	assert.Equal(t, uint64(21000), result.GetGasUsed())
	assert.Nil(t, result.GetError())
}

// TestMakeTxProcessor tests the creation of TxProcessor with different configurations
func TestExecutor_MakeTxProcessor(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *utils.Config
		expectError bool
	}{
		{
			name: "default_opera_processor",
			cfg: &utils.Config{
				ChainID:  utils.MainnetChainID,
				EvmImpl:  "opera",
				LogLevel: "info",
			},
			expectError: false,
		},
		{
			name: "ethereum_processor",
			cfg: &utils.Config{
				ChainID:  1, // Ethereum mainnet
				EvmImpl:  "ethereum",
				LogLevel: "info",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := MakeTxProcessor(tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, processor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, processor)
				assert.Equal(t, tt.cfg, processor.cfg)
				assert.NotNil(t, processor.numErrors)
				assert.NotNil(t, processor.log)
				assert.NotNil(t, processor.processor)
			}
		})
	}
}

// TestTxProcessor_isErrFatal tests the isErrFatal method of TxProcessor
func TestTxProcessor_isErrFatal(t *testing.T) {
	tests := []struct {
		name              string
		continueOnFailure bool
		maxNumErrors      int
		currentErrors     int32
		expectFatal       bool
	}{
		{
			name:              "no_continue_on_failure",
			continueOnFailure: false,
			maxNumErrors:      10,
			currentErrors:     5,
			expectFatal:       true,
		},
		{
			name:              "continue_unlimited_errors",
			continueOnFailure: true,
			maxNumErrors:      0,
			currentErrors:     100,
			expectFatal:       false,
		},
		{
			name:              "continue_under_max_errors",
			continueOnFailure: true,
			maxNumErrors:      10,
			currentErrors:     5,
			expectFatal:       false,
		},
		{
			name:              "continue_at_max_errors",
			continueOnFailure: true,
			maxNumErrors:      10,
			currentErrors:     10,
			expectFatal:       true,
		},
		{
			name:              "continue_over_max_errors",
			continueOnFailure: true,
			maxNumErrors:      10,
			currentErrors:     11,
			expectFatal:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &TxProcessor{
				cfg: &utils.Config{
					ContinueOnFailure: tt.continueOnFailure,
					MaxNumErrors:      tt.maxNumErrors,
				},
				numErrors: new(atomic.Int32),
				log:       logger.NewLogger("info", "test"),
			}

			processor.numErrors.Store(tt.currentErrors)

			fatal := processor.isErrFatal()
			assert.Equal(t, tt.expectFatal, fatal)

			// If not fatal and errors are being counted, check that the counter was incremented
			if !fatal && tt.continueOnFailure && tt.maxNumErrors > 0 {
				assert.Equal(t, tt.currentErrors+1, processor.numErrors.Load())
			} else {
				assert.Equal(t, tt.currentErrors, processor.numErrors.Load())
			}
		})
	}
}

// TestTxProcessor_ProcessTransaction tests the ProcessTransaction method of TxProcessor
func TestTxProcessor_ProcessTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockTxContext := txcontext.NewMockTxContext(ctrl)
	mockWorldState := txcontext.NewMockWorldState(ctrl)
	mockProcessor := NewMockprocessor(ctrl)

	// Setup for regular transaction test
	block := 12345
	tx := 0 // Regular transaction

	processor := &TxProcessor{
		cfg:       &utils.Config{},
		numErrors: new(atomic.Int32),
		processor: mockProcessor,
		log:       logger.NewLogger("info", "test"),
	}

	// Test regular transaction processing
	t.Run("regular_transaction", func(t *testing.T) {
		expectedResult := transactionResult{gasUsed: 21000}
		mockProcessor.EXPECT().processRegularTx(mockStateDB, block, tx, mockTxContext).Return(expectedResult, nil)

		result, err := processor.ProcessTransaction(mockStateDB, block, tx, mockTxContext)

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	// Setup for pseudo transaction test
	pseudoTx := utils.PseudoTx

	// Test pseudo transaction processing
	t.Run("pseudo_transaction", func(t *testing.T) {
		mockTxContext.EXPECT().GetOutputState().Return(mockWorldState)

		// Mock the ForEachAccount to call the callback with test data
		mockWorldState.EXPECT().ForEachAccount(gomock.Any()).DoAndReturn(
			func(callback func(common.Address, txcontext.Account)) {
				addr := common.HexToAddress("0x1234")
				mockAccount := txcontext.NewMockAccount(ctrl)

				mockAccount.EXPECT().GetBalance().Return(uint256.NewInt(1000))
				mockAccount.EXPECT().GetNonce().Return(uint64(5))
				mockAccount.EXPECT().GetCode().Return([]byte{1, 2, 3})

				// Mock storage iteration
				mockAccount.EXPECT().ForEachStorage(gomock.Any()).DoAndReturn(
					func(storageCallback func(common.Hash, common.Hash)) {
						key := common.HexToHash("0xABCD")
						value := common.HexToHash("0x1234")
						storageCallback(key, value)
					},
				)

				callback(addr, mockAccount)
			},
		)

		// Setup state DB expectations
		mockStateDB.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(0))
		mockStateDB.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any())
		mockStateDB.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any())
		mockStateDB.EXPECT().SetNonce(gomock.Any(), uint64(5), gomock.Any())
		mockStateDB.EXPECT().SetCode(gomock.Any(), []byte{1, 2, 3})
		mockStateDB.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any())

		result, err := processor.ProcessTransaction(mockStateDB, block, pseudoTx, mockTxContext)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestAidaProcessor_processRegularTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockStateDB := state.NewMockVmStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)
	mockTxContext := txcontext.NewMockTxContext(ctrl)

	// Test data
	block := 12345
	tx := 67890
	txHash := common.HexToHash(fmt.Sprintf("0x%016d%016d", block, tx))
	blockHash := common.HexToHash(fmt.Sprintf("0x%016d", block))
	sender := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recipient := common.HexToAddress("0x0987654321098765432109876543210987654321")

	// Create test message
	message := &core.Message{
		From:             sender,
		To:               &recipient,
		Nonce:            10,
		Value:            big.NewInt(1000),
		GasLimit:         21048,
		GasPrice:         big.NewInt(50),
		Data:             []byte{1, 2, 3},
		SkipFromEOACheck: true,
		SkipNonceChecks:  true,
	}

	// Common setup for all tests
	cfg := &utils.Config{
		EvmImpl:  "opera",
		LogLevel: "info",
		ChainID:  utils.MainnetChainID,
	}

	processor := &aidaProcessor{
		cfg: cfg,
		log: logger.NewLogger("info", "test"),
	}

	t.Run("successful_transaction", func(t *testing.T) {
		// Set up mock expectations
		mockTxContext.EXPECT().GetBlockEnvironment().Return(mockBlockEnv).AnyTimes()
		mockTxContext.EXPECT().GetMessage().Return(message).AnyTimes()

		mockBlockEnv.EXPECT().GetFork().Return("cancun").Times(1)
		mockBlockEnv.EXPECT().GetGasLimit().Return(uint64(30000000)).Times(1)
		mockBlockEnv.EXPECT().GetNumber().Return(uint64(block)).AnyTimes()
		mockBlockEnv.EXPECT().GetTimestamp().Return(uint64(1700000000)).AnyTimes()
		mockBlockEnv.EXPECT().GetBaseFee().Return(big.NewInt(5)).AnyTimes()
		mockBlockEnv.EXPECT().GetBlobBaseFee().Return(big.NewInt(10)).AnyTimes()
		mockBlockEnv.EXPECT().GetCoinbase().Return(common.HexToAddress("0xcoinbase")).AnyTimes()
		mockBlockEnv.EXPECT().GetDifficulty().Return(big.NewInt(2)).AnyTimes()
		mockBlockEnv.EXPECT().GetRandom().Return(&common.Hash{}).AnyTimes()
		mockBlockEnv.EXPECT().GetBlockHash(gomock.Any()).Return(common.Hash{}, nil).AnyTimes()
		mockBlockEnv.EXPECT().GetGasLimit().Return(uint64(30000000)).Times(1)

		snapshot := 42
		mockStateDB.EXPECT().SetTxContext(txHash, tx).Times(1)
		mockStateDB.EXPECT().Snapshot().Return(snapshot).AnyTimes()
		mockStateDB.EXPECT().GetBalance(sender).Return(uint256.NewInt(10510000)).AnyTimes()
		mockStateDB.EXPECT().GetLogs(txHash, uint64(block), blockHash, uint64(1700000000)).Return([]*types.Log{}).Times(1)
		mockStateDB.EXPECT().SubBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()
		mockStateDB.EXPECT().Prepare(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockStateDB.EXPECT().GetNonce(sender).Return(uint64(10)).Times(1)
		mockStateDB.EXPECT().SetNonce(sender, uint64(11), gomock.Any()).Times(1)
		mockStateDB.EXPECT().GetCode(recipient).Return([]byte{}).AnyTimes()
		mockStateDB.EXPECT().Exist(recipient).Return(true).Times(1)
		mockStateDB.EXPECT().GetRefund().Return(uint64(0)).AnyTimes()
		mockStateDB.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(*uint256.NewInt(0)).AnyTimes()

		// Call the method being tested
		result, err := processor.processRegularTx(mockStateDB, block, tx, mockTxContext)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint64(0x5238), result.gasUsed)
		assert.Equal(t, []byte(nil), result.result)
		assert.Nil(t, result.err)
		assert.Empty(t, result.logs)
	})
}

func TestEthTestProcessor_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test message
	sender := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recipient := common.HexToAddress("0x0987654321098765432109876543210987654321")
	message := &core.Message{
		From:     sender,
		To:       &recipient,
		Nonce:    10,
		Value:    big.NewInt(1000),
		GasLimit: 21000,
		GasPrice: big.NewInt(50),
		Data:     []byte{1, 2, 3},
	}

	mockStateDB := state.NewMockStateDB(ctrl)
	mockNonCommitStateDB := state.NewMockNonCommittableStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)
	mockTxProcessor := NewMockprocessor(ctrl)

	// Base configuration
	cfg := &utils.Config{
		ChainID:  utils.SepoliaChainID,
		EvmImpl:  "opera",
		LogLevel: "info",
	}

	// Create processor with mocked internal TxProcessor
	processor := &ethTestProcessor{
		TxProcessor: &TxProcessor{
			cfg:       cfg,
			numErrors: new(atomic.Int32),
			log:       logger.NewLogger("info", "test"),
			processor: mockTxProcessor,
		},
	}

	// Various test scenarios
	t.Run("unknown_fork", func(t *testing.T) {

		// Create mocks
		mockTx := types.NewTx(&types.DynamicFeeTx{
			ChainID: big.NewInt(int64(utils.SepoliaChainID)),
			V:       common.Big0,
			R:       common.Big1,
			S:       common.Big1,
		})
		mockTxContext := ethtest.NewStateTestContext(message, mockBlockEnv, utils.Must(mockTx.MarshalBinary()))

		// Create a context that would be passed to the processor
		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		// Setup base state
		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}

		// Setup fork to an unknown value
		mockBlockEnv.EXPECT().GetFork().Return("unknown_fork").Times(1)

		// Execute test
		err := processor.Process(testState, execContext)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown fork")
	})

	t.Run("successful_processing", func(t *testing.T) {

		// Create mocks
		mockTx := types.NewTx(&types.DynamicFeeTx{
			ChainID: big.NewInt(int64(utils.SepoliaChainID)),
			V:       common.Big0,
			R:       common.Big1,
			S:       common.Big1,
		})
		mockTxContext := ethtest.NewStateTestContext(message, mockBlockEnv, utils.Must(mockTx.MarshalBinary()))

		// Create a context that would be passed to the processor
		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		// Setup base state
		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}

		// Setup to a known fork with valid blob limits
		mockBlockEnv.EXPECT().GetFork().Return("cancun").AnyTimes()

		// Mock the ProcessTransaction call with a successful result
		expectedResult := transactionResult{
			gasUsed: 0,
			result:  []byte(nil),
		}
		mockTxProcessor.EXPECT().processRegularTx(mockStateDB, testState.Block, testState.Transaction, mockTxContext).Return(expectedResult, nil).Times(1)

		// Execute test
		err := processor.Process(testState, execContext)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, execContext.ExecutionResult)
	})

	t.Run("failed_unmarshal", func(t *testing.T) {

		mockTxContext := ethtest.NewStateTestContext(message, mockBlockEnv, hexutil.Bytes{1, 2, 3})

		// Create a context that would be passed to the processor
		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		// Setup base state
		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}
		err := processor.Process(testState, execContext)
		assert.Nil(t, err)
	})

	t.Run("failed_signer", func(t *testing.T) {
		mockTx := types.NewTx(&types.DynamicFeeTx{
			V: common.Big0,
			R: common.Big1,
			S: common.Big1,
		})
		mockTxContext := ethtest.NewStateTestContext(message, mockBlockEnv, utils.Must(mockTx.MarshalBinary()))

		// Create a context that would be passed to the processor
		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		// Setup base state
		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}
		err := processor.Process(testState, execContext)
		assert.Nil(t, err)
	})

}

func TestExecutor_MessageToTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test data
	sender := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recipient := common.HexToAddress("0x0987654321098765432109876543210987654321")
	message := &core.Message{
		From:      sender,
		To:        &recipient,
		Nonce:     10,
		Value:     big.NewInt(1000),
		GasLimit:  21000,
		GasPrice:  big.NewInt(50),
		GasFeeCap: big.NewInt(100),
		GasTipCap: big.NewInt(10),
		Data:      []byte{1, 2, 3},
		BlobHashes: []common.Hash{
			common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		},
		AccessList: types.AccessList{
			{
				Address: common.HexToAddress("0x1234"),
			},
		},
	}

	tx := messageToTransaction(message)
	assert.NotNil(t, tx, "Transaction should not be nil")
	assert.Equal(t, sender, common.Address(tx.Sender), "Sender address should match")
	assert.Equal(t, recipient, common.Address(*tx.Recipient), "Recipient address should match")
	assert.Equal(t, uint64(10), tx.Nonce, "Nonce should match")
	assert.Equal(t, tosca.ValueFromUint256(uint256.NewInt(1000)), tx.Value, "Value should match")
	assert.Equal(t, tosca.Gas(21000), tx.GasLimit, "Gas limit should match")
}

func TestArchiveDbTxProcessor_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockStateDB(ctrl)
	mockNonCommitStateDB := state.NewMockNonCommittableStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)
	mockTxProcessor := NewMockprocessor(ctrl)

	sender := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recipient := common.HexToAddress("0x0987654321098765432109876543210987654321")
	message := &core.Message{
		From:     sender,
		To:       &recipient,
		Nonce:    10,
		Value:    big.NewInt(1000),
		GasLimit: 21000,
		GasPrice: big.NewInt(50),
		Data:     []byte{1, 2, 3},
	}

	// Base configuration
	cfg := &utils.Config{
		ChainID:  utils.SepoliaChainID,
		EvmImpl:  "opera",
		LogLevel: "info",
	}

	// Create processor with mocked internal TxProcessor
	processor := &ArchiveDbTxProcessor{
		TxProcessor: &TxProcessor{
			cfg:       cfg,
			numErrors: new(atomic.Int32),
			log:       logger.NewLogger("info", "test"),
			processor: mockTxProcessor,
		},
	}

	t.Run("success", func(t *testing.T) {

		mockTxContext := txcontext.NewMockTxContext(ctrl)
		mockTxContext.EXPECT().GetMessage().Return(message).AnyTimes()
		mockTxContext.EXPECT().GetBlockEnvironment().Return(mockBlockEnv).AnyTimes()

		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}

		expectedResult := transactionResult{
			gasUsed: 0,
			result:  []byte(nil),
		}

		mockTxProcessor.EXPECT().processRegularTx(mockNonCommitStateDB, testState.Block, testState.Transaction, mockTxContext).Return(expectedResult, nil).Times(1)

		err := processor.Process(testState, execContext)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, execContext.ExecutionResult)
	})

	t.Run("failed", func(t *testing.T) {

		mockTxContext := txcontext.NewMockTxContext(ctrl)
		mockTxContext.EXPECT().GetMessage().Return(message).AnyTimes()
		mockTxContext.EXPECT().GetBlockEnvironment().Return(mockBlockEnv).AnyTimes()

		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}

		expectedResult := transactionResult{
			gasUsed: 0,
			result:  []byte(nil),
		}

		mockTxProcessor.EXPECT().processRegularTx(mockNonCommitStateDB, testState.Block, testState.Transaction, mockTxContext).Return(expectedResult, errors.New("mock error")).Times(1)

		err := processor.Process(testState, execContext)
		assert.Error(t, err)
	})
}

func TestLiveDbTxProcessor_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDB := state.NewMockStateDB(ctrl)
	mockNonCommitStateDB := state.NewMockNonCommittableStateDB(ctrl)
	mockBlockEnv := txcontext.NewMockBlockEnvironment(ctrl)
	mockTxProcessor := NewMockprocessor(ctrl)

	sender := common.HexToAddress("0x1234567890123456789012345678901234567890")
	recipient := common.HexToAddress("0x0987654321098765432109876543210987654321")
	message := &core.Message{
		From:     sender,
		To:       &recipient,
		Nonce:    10,
		Value:    big.NewInt(1000),
		GasLimit: 21000,
		GasPrice: big.NewInt(50),
		Data:     []byte{1, 2, 3},
	}

	// Base configuration
	cfg := &utils.Config{
		ChainID:  utils.SepoliaChainID,
		EvmImpl:  "opera",
		LogLevel: "info",
	}

	// Create processor with mocked internal TxProcessor
	processor := &LiveDbTxProcessor{
		TxProcessor: &TxProcessor{
			cfg:       cfg,
			numErrors: new(atomic.Int32),
			log:       logger.NewLogger("info", "test"),
			processor: mockTxProcessor,
		},
	}

	t.Run("success", func(t *testing.T) {

		mockTxContext := txcontext.NewMockTxContext(ctrl)
		mockTxContext.EXPECT().GetMessage().Return(message).AnyTimes()
		mockTxContext.EXPECT().GetBlockEnvironment().Return(mockBlockEnv).AnyTimes()

		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}

		expectedResult := transactionResult{
			gasUsed: 0,
			result:  []byte(nil),
		}

		mockTxProcessor.EXPECT().processRegularTx(mockStateDB, testState.Block, testState.Transaction, mockTxContext).Return(expectedResult, nil).Times(1)

		err := processor.Process(testState, execContext)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, execContext.ExecutionResult)
	})

	t.Run("failed", func(t *testing.T) {

		mockTxContext := txcontext.NewMockTxContext(ctrl)
		mockTxContext.EXPECT().GetMessage().Return(message).AnyTimes()
		mockTxContext.EXPECT().GetBlockEnvironment().Return(mockBlockEnv).AnyTimes()

		execContext := &Context{
			State:   mockStateDB,
			Archive: mockNonCommitStateDB,
		}

		testState := State[txcontext.TxContext]{
			Block:       123,
			Transaction: 456,
			Data:        mockTxContext,
		}

		expectedResult := transactionResult{
			gasUsed: 0,
			result:  []byte(nil),
		}

		mockTxProcessor.EXPECT().processRegularTx(mockStateDB, testState.Block, testState.Transaction, mockTxContext).Return(expectedResult, errors.New("mock error")).Times(1)

		err := processor.Process(testState, execContext)
		assert.Error(t, err)
	})
}

func TestExecutor_MakeLiveDbTxProcessor(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *utils.Config
		expectError bool
	}{
		{
			name: "valid_live_db_processor",
			cfg: &utils.Config{
				ChainID:  utils.SepoliaChainID,
				EvmImpl:  "opera",
				LogLevel: "info",
			},
			expectError: false,
		},
		{
			name: "invalid_live_db_processor",
			cfg: &utils.Config{
				ChainID:  utils.MainnetChainID,
				EvmImpl:  "invalid_impl",
				LogLevel: "info",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := MakeLiveDbTxProcessor(tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, processor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, processor)
				assert.Equal(t, tt.cfg, processor.cfg)
				assert.NotNil(t, processor.numErrors)
				assert.NotNil(t, processor.log)
				assert.NotNil(t, processor.processor)
			}
		})
	}
}

func TestExecutor_MakeArchiveDbTxProcessor(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *utils.Config
		expectError bool
	}{
		{
			name: "valid_archive_db_processor",
			cfg: &utils.Config{
				ChainID:  utils.SepoliaChainID,
				EvmImpl:  "opera",
				LogLevel: "info",
			},
			expectError: false,
		},
		{
			name: "invalid_archive_db_processor",
			cfg: &utils.Config{
				ChainID:  utils.MainnetChainID,
				EvmImpl:  "invalid_impl",
				LogLevel: "info",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := MakeArchiveDbTxProcessor(tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, processor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, processor)
				assert.Equal(t, tt.cfg, processor.cfg)
				assert.NotNil(t, processor.numErrors)
				assert.NotNil(t, processor.log)
				assert.NotNil(t, processor.processor)
			}
		})
	}
}

func TestExecutor_MakeMakeEthTestProcessor(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *utils.Config
		expectError bool
	}{
		{
			name: "valid_eth_test_processor",
			cfg: &utils.Config{
				ChainID:  utils.SepoliaChainID,
				EvmImpl:  "opera",
				LogLevel: "info",
			},
			expectError: false,
		},
		{
			name: "invalid_eth_test_processor",
			cfg: &utils.Config{
				ChainID:  utils.MainnetChainID,
				EvmImpl:  "invalid_impl",
				LogLevel: "info",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := MakeEthTestProcessor(tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, processor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, processor)
				assert.Equal(t, tt.cfg, processor.cfg)
				assert.NotNil(t, processor.numErrors)
				assert.NotNil(t, processor.log)
				assert.NotNil(t, processor.processor)
			}
		})
	}
}
