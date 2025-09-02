// Copyright 2025 Sonic Labs
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

// Package txcontext is a generated GoMock package.
package txcontext

import (
	big "math/big"
	reflect "reflect"

	common "github.com/ethereum/go-ethereum/common"
	gomock "go.uber.org/mock/gomock"
)

// MockBlockEnvironment is a mock of BlockEnvironment interface.
type MockBlockEnvironment struct {
	ctrl     *gomock.Controller
	recorder *MockBlockEnvironmentMockRecorder
	isgomock struct{}
}

// MockBlockEnvironmentMockRecorder is the mock recorder for MockBlockEnvironment.
type MockBlockEnvironmentMockRecorder struct {
	mock *MockBlockEnvironment
}

// NewMockBlockEnvironment creates a new mock instance.
func NewMockBlockEnvironment(ctrl *gomock.Controller) *MockBlockEnvironment {
	mock := &MockBlockEnvironment{ctrl: ctrl}
	mock.recorder = &MockBlockEnvironmentMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBlockEnvironment) EXPECT() *MockBlockEnvironmentMockRecorder {
	return m.recorder
}

// GetBaseFee mocks base method.
func (m *MockBlockEnvironment) GetBaseFee() *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBaseFee")
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// GetBaseFee indicates an expected call of GetBaseFee.
func (mr *MockBlockEnvironmentMockRecorder) GetBaseFee() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBaseFee", reflect.TypeOf((*MockBlockEnvironment)(nil).GetBaseFee))
}

// GetBlobBaseFee mocks base method.
func (m *MockBlockEnvironment) GetBlobBaseFee() *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlobBaseFee")
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// GetBlobBaseFee indicates an expected call of GetBlobBaseFee.
func (mr *MockBlockEnvironmentMockRecorder) GetBlobBaseFee() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlobBaseFee", reflect.TypeOf((*MockBlockEnvironment)(nil).GetBlobBaseFee))
}

// GetBlockHash mocks base method.
func (m *MockBlockEnvironment) GetBlockHash(blockNumber uint64) (common.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockHash", blockNumber)
	ret0, _ := ret[0].(common.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockHash indicates an expected call of GetBlockHash.
func (mr *MockBlockEnvironmentMockRecorder) GetBlockHash(blockNumber any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockHash", reflect.TypeOf((*MockBlockEnvironment)(nil).GetBlockHash), blockNumber)
}

// GetCoinbase mocks base method.
func (m *MockBlockEnvironment) GetCoinbase() common.Address {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCoinbase")
	ret0, _ := ret[0].(common.Address)
	return ret0
}

// GetCoinbase indicates an expected call of GetCoinbase.
func (mr *MockBlockEnvironmentMockRecorder) GetCoinbase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCoinbase", reflect.TypeOf((*MockBlockEnvironment)(nil).GetCoinbase))
}

// GetDifficulty mocks base method.
func (m *MockBlockEnvironment) GetDifficulty() *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDifficulty")
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// GetDifficulty indicates an expected call of GetDifficulty.
func (mr *MockBlockEnvironmentMockRecorder) GetDifficulty() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDifficulty", reflect.TypeOf((*MockBlockEnvironment)(nil).GetDifficulty))
}

// GetFork mocks base method.
func (m *MockBlockEnvironment) GetFork() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFork")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetFork indicates an expected call of GetFork.
func (mr *MockBlockEnvironmentMockRecorder) GetFork() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFork", reflect.TypeOf((*MockBlockEnvironment)(nil).GetFork))
}

// GetGasLimit mocks base method.
func (m *MockBlockEnvironment) GetGasLimit() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGasLimit")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetGasLimit indicates an expected call of GetGasLimit.
func (mr *MockBlockEnvironmentMockRecorder) GetGasLimit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGasLimit", reflect.TypeOf((*MockBlockEnvironment)(nil).GetGasLimit))
}

// GetNumber mocks base method.
func (m *MockBlockEnvironment) GetNumber() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNumber")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetNumber indicates an expected call of GetNumber.
func (mr *MockBlockEnvironmentMockRecorder) GetNumber() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNumber", reflect.TypeOf((*MockBlockEnvironment)(nil).GetNumber))
}

// GetRandom mocks base method.
func (m *MockBlockEnvironment) GetRandom() *common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRandom")
	ret0, _ := ret[0].(*common.Hash)
	return ret0
}

// GetRandom indicates an expected call of GetRandom.
func (mr *MockBlockEnvironmentMockRecorder) GetRandom() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRandom", reflect.TypeOf((*MockBlockEnvironment)(nil).GetRandom))
}

// GetTimestamp mocks base method.
func (m *MockBlockEnvironment) GetTimestamp() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTimestamp")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetTimestamp indicates an expected call of GetTimestamp.
func (mr *MockBlockEnvironmentMockRecorder) GetTimestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTimestamp", reflect.TypeOf((*MockBlockEnvironment)(nil).GetTimestamp))
}
