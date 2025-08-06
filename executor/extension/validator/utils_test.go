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

package validator

import (
	"errors"
	"fmt"
	"github.com/0xsoniclabs/aida/config"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUtils_UpdateStateDbOnEthereumChainFailsOnEndTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Define the world state to overwrite
	ws := substate.WorldState{
		substatetypes.Address{0x01}: &substate.Account{
			Nonce:   1,
			Balance: uint256.NewInt(1000),
			Code:    []byte{0x60, 0x60},
			Storage: map[substatetypes.Hash]substatetypes.Hash{{0x01}: {0x02}},
		},
	}

	// Create a patch with the world state
	patch := substatecontext.NewWorldState(ws)

	gomock.InOrder(
		db.EXPECT().Exist(common.Address{0x01}).Times(1),
		db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
		db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(500)).Times(1),
		db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(1000), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(2)).Times(1),
		db.EXPECT().SetNonce(common.Address{0x01}, uint64(1), tracing.NonceChangeUnspecified).Times(1),
		db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{0x60, 0x00}).Times(1),
		db.EXPECT().SetCode(common.Address{0x01}, []byte{0x60, 0x60}).Times(1),
		db.EXPECT().GetState(common.Address{0x01}, common.Hash{0x01}).Return(common.Hash{}).Times(1),
		db.EXPECT().SetState(common.Address{0x01}, common.Hash{0x01}, common.Hash{0x02}).Times(1),
		db.EXPECT().EndTransaction().Return(errors.New("err")).Times(1),
	)
	// Call the method to test
	err := updateStateDbOnEthereumChain(patch, db, true)
	assert.Error(t, err, "Expected an error when ending the transaction")
}

func TestUtils_UpdateStateDbOnEthereumChainFailsOnBeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := state.NewMockStateDB(ctrl)

	// Define the world state to overwrite
	ws := substate.WorldState{
		substatetypes.Address{0x01}: &substate.Account{
			Nonce:   1,
			Balance: uint256.NewInt(1000),
			Code:    []byte{0x60, 0x60},
			Storage: map[substatetypes.Hash]substatetypes.Hash{{0x01}: {0x02}},
		},
	}

	// Create a patch with the world state
	patch := substatecontext.NewWorldState(ws)

	gomock.InOrder(
		db.EXPECT().Exist(common.Address{0x01}).Times(1),
		db.EXPECT().CreateAccount(common.Address{0x01}).Times(1),
		db.EXPECT().GetBalance(common.Address{0x01}).Return(uint256.NewInt(500)).Times(1),
		db.EXPECT().SubBalance(common.Address{0x01}, uint256.NewInt(500), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().AddBalance(common.Address{0x01}, uint256.NewInt(1000), tracing.BalanceChangeUnspecified).Times(1),
		db.EXPECT().GetNonce(common.Address{0x01}).Return(uint64(2)).Times(1),
		db.EXPECT().SetNonce(common.Address{0x01}, uint64(1), tracing.NonceChangeUnspecified).Times(1),
		db.EXPECT().GetCode(common.Address{0x01}).Return([]byte{0x60, 0x00}).Times(1),
		db.EXPECT().SetCode(common.Address{0x01}, []byte{0x60, 0x60}).Times(1),
		db.EXPECT().GetState(common.Address{0x01}, common.Hash{0x01}).Return(common.Hash{}).Times(1),
		db.EXPECT().SetState(common.Address{0x01}, common.Hash{0x01}, common.Hash{0x02}).Times(1),
		db.EXPECT().EndTransaction().Return(nil).Times(1),
		db.EXPECT().BeginTransaction(uint32(config.PseudoTx)).Return(errors.New("err")).Times(1),
	)
	// Call the method to test
	err := updateStateDbOnEthereumChain(patch, db, true)
	assert.Error(t, err, "Expected an error when initializing a new transaction")
}

func TestValidateWorldState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := state.NewMockVmStateDB(ctrl)
	mockExpectedAlloc := txcontext.NewMockWorldState(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	t.Run("SubsetCheck_Success", func(t *testing.T) {
		cfg := &config.Config{StateValidationMode: config.SubsetCheck}
		// We need to simulate doSubsetValidation. Since it's not easily mockable directly
		// without refactoring or using a global variable (which is bad practice),
		// we'll test its behavior by ensuring no error is returned when underlying checks pass.
		// For this specific test, we assume doSubsetValidation would pass if alloc is empty.
		mockExpectedAlloc.EXPECT().ForEachAccount(gomock.Any()).Times(1) // Simulates doSubsetValidation call

		err := validateWorldState(cfg, mockDB, mockExpectedAlloc, mockLogger)
		assert.NoError(t, err)
	})

	t.Run("SubsetCheck_Failure", func(t *testing.T) {
		cfg := &config.Config{StateValidationMode: config.SubsetCheck}
		// Simulate a scenario where doSubsetValidation would fail.
		// For example, an account exists in expectedAlloc but not in db.
		// This requires more intricate mocking of ForEachAccount and db.Exist inside doSubsetValidation.
		// For simplicity here, we'll assume doSubsetValidation returns an error.
		// To properly test this, doSubsetValidation would ideally be an interface method or take a mockable dependency.
		// Given the current structure, we'll mock the inputs to doSubsetValidation to cause an error.

		addr := common.HexToAddress("0x1")
		mockAccount := txcontext.NewMockAccount(ctrl)

		mockExpectedAlloc.EXPECT().ForEachAccount(gomock.Any()).Do(func(cb func(common.Address, txcontext.Account)) {
			cb(addr, mockAccount)
		})
		mockDB.EXPECT().Exist(addr).Return(false)
		mockDB.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(0))
		mockDB.EXPECT().GetNonce(gomock.Any()).Return(uint64(0))
		mockDB.EXPECT().GetCode(gomock.Any()).Return([]byte{0x61, 0x00})
		mockAccount.EXPECT().GetBalance().Return(uint256.NewInt(1))
		mockAccount.EXPECT().GetNonce().Return(uint64(1)).Times(2)
		mockAccount.EXPECT().GetCode().Return([]byte{0x60, 0x00}).Times(2)
		mockAccount.EXPECT().ForEachStorage(gomock.Any())
		err := validateWorldState(cfg, mockDB, mockExpectedAlloc, mockLogger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("Account %v does not exist", addr.Hex()))
	})

	t.Run("EqualityCheck_Success", func(t *testing.T) {
		cfg := &config.Config{StateValidationMode: config.EqualityCheck}
		mockVmAlloc := txcontext.NewMockWorldState(ctrl)

		mockDB.EXPECT().GetSubstatePostAlloc().Return(mockVmAlloc)
		mockExpectedAlloc.EXPECT().Equal(mockVmAlloc).Return(true)

		err := validateWorldState(cfg, mockDB, mockExpectedAlloc, mockLogger)
		assert.NoError(t, err)
	})

	t.Run("EqualityCheck_Failure", func(t *testing.T) {
		cfg := &config.Config{StateValidationMode: config.EqualityCheck}
		mockVmAlloc := txcontext.NewMockWorldState(ctrl)

		mockDB.EXPECT().GetSubstatePostAlloc().Return(mockVmAlloc)
		mockExpectedAlloc.EXPECT().Equal(mockVmAlloc).Return(false)

		// Mocks for printAllocationDiffSummary
		mockExpectedAlloc.EXPECT().Len().Return(0).AnyTimes()
		mockVmAlloc.EXPECT().Len().Return(0).AnyTimes()
		mockExpectedAlloc.EXPECT().ForEachAccount(gomock.Any()).AnyTimes()
		mockVmAlloc.EXPECT().ForEachAccount(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes() // For diffs

		err := validateWorldState(cfg, mockDB, mockExpectedAlloc, mockLogger)
		require.Error(t, err)
		assert.EqualError(t, err, "inconsistent output: alloc")
	})

	t.Run("UnknownValidationMode", func(t *testing.T) {
		cfg := &config.Config{StateValidationMode: config.ValidationMode(999)} // An invalid mode
		// No mocks should be called as it should just return nil or handle gracefully.
		// Based on current implementation, it will default to no error if mode is not recognized.
		err := validateWorldState(cfg, mockDB, mockExpectedAlloc, mockLogger)
		assert.NoError(t, err, "Unknown validation mode should not produce an error by default or should be handled")
	})
}

func TestPrintIfDifferent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockLogger(ctrl)

	t.Run("DifferentValues_Int", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "label_int", 10, 20)
		changed := printIfDifferent("label_int", 10, 20, mockLogger)
		assert.True(t, changed)
	})

	t.Run("SameValues_Int", func(t *testing.T) {
		changed := printIfDifferent("label_int", 10, 10, mockLogger)
		assert.False(t, changed)
	})

	t.Run("DifferentValues_String", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "label_string", "abc", "def")
		changed := printIfDifferent("label_string", "abc", "def", mockLogger)
		assert.True(t, changed)
	})

	t.Run("SameValues_String", func(t *testing.T) {
		changed := printIfDifferent("label_string", "abc", "abc", mockLogger)
		assert.False(t, changed)
	})
}

func TestPrintIfDifferentBytes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockLogger(ctrl)

	wantBytes := []byte{1, 2, 3}
	haveBytesDifferent := []byte{1, 2, 4}
	haveBytesSame := []byte{1, 2, 3}

	t.Run("DifferentBytes", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "bytes_label", wantBytes, haveBytesDifferent)
		changed := printIfDifferentBytes("bytes_label", wantBytes, haveBytesDifferent, mockLogger)
		assert.True(t, changed)
	})

	t.Run("SameBytes", func(t *testing.T) {
		changed := printIfDifferentBytes("bytes_label", wantBytes, haveBytesSame, mockLogger)
		assert.False(t, changed)
	})

	t.Run("WantNilHaveNotNil", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "bytes_label", ([]byte)(nil), haveBytesSame)
		changed := printIfDifferentBytes("bytes_label", nil, haveBytesSame, mockLogger)
		assert.True(t, changed)
	})

	t.Run("WantNotNilHaveNil", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "bytes_label", wantBytes, ([]byte)(nil))
		changed := printIfDifferentBytes("bytes_label", wantBytes, nil, mockLogger)
		assert.True(t, changed)
	})

	t.Run("BothNil", func(t *testing.T) {
		changed := printIfDifferentBytes("bytes_label", nil, nil, mockLogger)
		assert.False(t, changed)
	})

	t.Run("BothEmpty", func(t *testing.T) {
		emptyBytes1 := []byte{}
		emptyBytes2 := []byte{}
		changed := printIfDifferentBytes("bytes_label", emptyBytes1, emptyBytes2, mockLogger)
		assert.False(t, changed)
	})
}

func TestPrintIfDifferentUint256(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockLogger(ctrl)

	wantVal := uint256.NewInt(100)
	haveValDifferent := uint256.NewInt(200)
	haveValSame := uint256.NewInt(100)

	t.Run("DifferentUint256", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "uint256_label", wantVal, haveValDifferent)
		changed := printIfDifferentUint256("uint256_label", wantVal, haveValDifferent, mockLogger)
		assert.True(t, changed)
	})

	t.Run("SameUint256", func(t *testing.T) {
		changed := printIfDifferentUint256("uint256_label", wantVal, haveValSame, mockLogger)
		assert.False(t, changed)
	})

	t.Run("WantNilHaveNotNil", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "uint256_label", (*uint256.Int)(nil), haveValSame)
		changed := printIfDifferentUint256("uint256_label", nil, haveValSame, mockLogger)
		assert.True(t, changed)
	})

	t.Run("WantNotNilHaveNil", func(t *testing.T) {
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "uint256_label", wantVal, (*uint256.Int)(nil))
		changed := printIfDifferentUint256("uint256_label", wantVal, nil, mockLogger)
		assert.True(t, changed)
	})

	t.Run("BothNil", func(t *testing.T) {
		changed := printIfDifferentUint256("uint256_label", nil, nil, mockLogger)
		assert.False(t, changed)
	})
}

func TestPrintLogDiffSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := logger.NewMockLogger(ctrl)

	addr1 := common.HexToAddress("0x1111")
	addr2 := common.HexToAddress("0x2222")
	topic1 := common.HexToHash("0xaaaa")
	topic2 := common.HexToHash("0xbbbb")
	topic3 := common.HexToHash("0xcccc")
	data1 := []byte{1, 2, 3}
	data2 := []byte{4, 5, 6}

	baseLog := &types.Log{
		Address: addr1,
		Topics:  []common.Hash{topic1, topic2},
		Data:    data1,
	}

	t.Run("SameLogs", func(t *testing.T) {
		// No EXPECT calls on mockLogger as nothing should be logged
		logWant := &types.Log{Address: addr1, Topics: []common.Hash{topic1, topic2}, Data: data1}
		logHave := &types.Log{Address: addr1, Topics: []common.Hash{topic1, topic2}, Data: data1}
		printLogDiffSummary("log_same", logWant, logHave, mockLogger)
	})

	t.Run("DifferentAddress", func(t *testing.T) {
		logWant := baseLog
		logHave := &types.Log{Address: addr2, Topics: baseLog.Topics, Data: baseLog.Data}
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_diff_addr.address", logWant.Address, logHave.Address)
		printLogDiffSummary("log_diff_addr", logWant, logHave, mockLogger)
	})

	t.Run("DifferentTopicLength_WantMore", func(t *testing.T) {
		logWant := baseLog
		logHave := &types.Log{Address: baseLog.Address, Topics: []common.Hash{topic1}, Data: baseLog.Data} // Have has fewer topics
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_diff_topiclen.Topics size", len(logWant.Topics), len(logHave.Topics))
		// The loop for individual topics won't run due to size mismatch
		printLogDiffSummary("log_diff_topiclen", logWant, logHave, mockLogger)
	})

	t.Run("DifferentTopicLength_HaveMore", func(t *testing.T) {
		logWant := &types.Log{Address: baseLog.Address, Topics: []common.Hash{topic1}, Data: baseLog.Data} // Want has fewer topics
		logHave := baseLog
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_diff_topiclen2.Topics size", len(logWant.Topics), len(logHave.Topics))
		printLogDiffSummary("log_diff_topiclen2", logWant, logHave, mockLogger)
	})

	t.Run("DifferentTopicValue", func(t *testing.T) {
		logWant := baseLog
		logHave := &types.Log{Address: baseLog.Address, Topics: []common.Hash{topic1, topic3}, Data: baseLog.Data} // topic2 vs topic3
		// First topic is same, so no log for Topics[0]
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_diff_topicval.Topics[1]", topic2, topic3)
		printLogDiffSummary("log_diff_topicval", logWant, logHave, mockLogger)
	})

	t.Run("DifferentData", func(t *testing.T) {
		logWant := baseLog
		logHave := &types.Log{Address: baseLog.Address, Topics: baseLog.Topics, Data: data2}
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_diff_data.data", data1, data2)
		printLogDiffSummary("log_diff_data", logWant, logHave, mockLogger)
	})

	t.Run("AllDifferent", func(t *testing.T) {
		logWant := &types.Log{Address: addr1, Topics: []common.Hash{topic1}, Data: data1}
		logHave := &types.Log{Address: addr2, Topics: []common.Hash{topic2}, Data: data2}

		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_all_diff.address", addr1, addr2)
		// Topic lengths are same (1)
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_all_diff.Topics[0]", topic1, topic2)
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_all_diff.data", data1, data2)
		printLogDiffSummary("log_all_diff", logWant, logHave, mockLogger)
	})

	t.Run("EmptyTopicsVsNonEmpty", func(t *testing.T) {
		logWant := &types.Log{Address: addr1, Topics: []common.Hash{}, Data: data1}
		logHave := &types.Log{Address: addr1, Topics: []common.Hash{topic1}, Data: data1}
		mockLogger.EXPECT().Errorf("Different %s:\nwant: %v\nhave: %v\n", "log_empty_topics.Topics size", 0, 1)
		printLogDiffSummary("log_empty_topics", logWant, logHave, mockLogger)
	})

}

func TestPrintAllocationDiffSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockAllocWant := txcontext.NewMockWorldState(ctrl)
	mockAllocHave := txcontext.NewMockWorldState(ctrl)
	mockAllocWant.EXPECT().Len().Return(1).AnyTimes()
	mockAllocHave.EXPECT().Len().Return(1).AnyTimes()
	t.Run("Success", func(t *testing.T) {
		mockAllocWant.EXPECT().ForEachAccount(gomock.Any()).Do(func(cb func(common.Address, txcontext.Account)) {
			addr := common.HexToAddress("0x1234")
			mockAccount := txcontext.NewMockAccount(ctrl)
			mockAllocHave.EXPECT().Get(addr).Return(mockAccount)
			cb(addr, mockAccount)
		})
		mockAllocHave.EXPECT().ForEachAccount(gomock.Any()).Do(func(cb func(common.Address, txcontext.Account)) {
			addr := common.HexToAddress("0x1234")
			mockAccount := txcontext.NewMockAccount(ctrl)
			mockAllocWant.EXPECT().Get(addr).Return(mockAccount)
			cb(addr, mockAccount)
		})
		mockAllocHave.EXPECT().ForEachAccount(gomock.Any()).Do(func(cb func(common.Address, txcontext.Account)) {
			addr := common.HexToAddress("0x1234")
			mockAccount := txcontext.NewMockAccount(ctrl)
			mockAllocWant.EXPECT().Get(addr).Return(nil)
			cb(addr, mockAccount)
		})
		printAllocationDiffSummary(mockAllocWant, mockAllocHave, mockLogger)
	})

	t.Run("Fail", func(t *testing.T) {
		mockAllocWant.EXPECT().ForEachAccount(gomock.Any()).Do(func(cb func(common.Address, txcontext.Account)) {
			addr := common.HexToAddress("0x1234")
			mockAccount := txcontext.NewMockAccount(ctrl)
			mockAllocHave.EXPECT().Get(addr).Return(nil)
			mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			cb(addr, mockAccount)
		})
		mockAllocHave.EXPECT().ForEachAccount(gomock.Any()).Do(func(cb func(common.Address, txcontext.Account)) {
			addr := common.HexToAddress("0x1234")
			mockAccount := txcontext.NewMockAccount(ctrl)
			mockAllocWant.EXPECT().Get(addr).Return(nil)
			mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			cb(addr, mockAccount)
		})
		mockAllocHave.EXPECT().ForEachAccount(gomock.Any()).Do(func(cb func(common.Address, txcontext.Account)) {
			addr := common.HexToAddress("0x1234")
			mockAccount := txcontext.NewMockAccount(ctrl)
			mockAllocWant.EXPECT().Get(addr).Return(nil)
			cb(addr, mockAccount)
		})
		printAllocationDiffSummary(mockAllocWant, mockAllocHave, mockLogger)
	})
}

func TestPrintAccountDiffSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockAllocWant := txcontext.NewMockAccount(ctrl)
	mockAllocHave := txcontext.NewMockAccount(ctrl)
	mockAllocWant.EXPECT().GetNonce().Return(uint64(0)).AnyTimes()
	mockAllocHave.EXPECT().GetNonce().Return(uint64(0)).AnyTimes()
	mockAllocWant.EXPECT().GetBalance().Return(uint256.NewInt(100)).AnyTimes()
	mockAllocHave.EXPECT().GetBalance().Return(uint256.NewInt(100)).AnyTimes()
	mockAllocWant.EXPECT().GetCode().Return([]byte{0x61, 0x00}).AnyTimes()
	mockAllocHave.EXPECT().GetCode().Return([]byte{0x61, 0x00}).AnyTimes()
	mockAllocWant.EXPECT().GetStorageSize().Return(100).AnyTimes()
	mockAllocHave.EXPECT().GetStorageSize().Return(100).AnyTimes()

	t.Run("Success", func(t *testing.T) {
		mockAllocWant.EXPECT().ForEachStorage(gomock.Any()).Do(func(cb func(common.Hash, common.Hash)) {
			hash1 := common.HexToHash("0x1234")
			hash2 := common.HexToHash("0x5678")
			mockAllocHave.EXPECT().GetStorageAt(hash1).Return(hash2)
			cb(hash1, hash2)
		})
		mockAllocHave.EXPECT().ForEachStorage(gomock.Any()).Do(func(cb func(common.Hash, common.Hash)) {
			hash1 := common.HexToHash("0x1234")
			hash2 := common.HexToHash("0x5678")
			mockAllocWant.EXPECT().GetStorageAt(hash1).Return(hash2)
			cb(hash1, hash2)
		})
		mockAllocHave.EXPECT().ForEachStorage(gomock.Any()).Do(func(cb func(common.Hash, common.Hash)) {
			hash1 := common.HexToHash("0x1234")
			hash2 := common.HexToHash("0x5678")
			mockAllocWant.EXPECT().GetStorageAt(hash1).Return(hash2)
			cb(hash1, hash2)
		})
		printAccountDiffSummary("test", mockAllocWant, mockAllocHave, mockLogger)
	})

	t.Run("Fail", func(t *testing.T) {
		mockAllocWant.EXPECT().ForEachStorage(gomock.Any()).Do(func(cb func(common.Hash, common.Hash)) {
			hash1 := common.HexToHash("0x1234")
			hash2 := common.HexToHash("0x5678")
			hash3 := common.HexToHash("0x9abc")
			mockAllocHave.EXPECT().GetStorageAt(hash1).Return(hash3)
			mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			cb(hash1, hash2)
		})
		mockAllocHave.EXPECT().ForEachStorage(gomock.Any()).Do(func(cb func(common.Hash, common.Hash)) {
			hash1 := common.HexToHash("0x1234")
			hash2 := common.HexToHash("0x5678")
			hash3 := common.HexToHash("0x9abc")
			mockAllocWant.EXPECT().GetStorageAt(hash1).Return(hash3)
			mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			cb(hash1, hash2)
		})
		mockAllocHave.EXPECT().ForEachStorage(gomock.Any()).Do(func(cb func(common.Hash, common.Hash)) {
			hash1 := common.HexToHash("0x1234")
			hash2 := common.HexToHash("0x5678")
			mockAllocWant.EXPECT().GetStorageAt(hash1).Return(hash2)
			cb(hash1, hash2)
		})
		printAccountDiffSummary("test", mockAllocWant, mockAllocHave, mockLogger)
	})
}
