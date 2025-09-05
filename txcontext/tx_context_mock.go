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
	reflect "reflect"

	common "github.com/ethereum/go-ethereum/common"
	core "github.com/ethereum/go-ethereum/core"
	gomock "go.uber.org/mock/gomock"
)

// MockTxContext is a mock of TxContext interface.
type MockTxContext struct {
	ctrl     *gomock.Controller
	recorder *MockTxContextMockRecorder
	isgomock struct{}
}

// MockTxContextMockRecorder is the mock recorder for MockTxContext.
type MockTxContextMockRecorder struct {
	mock *MockTxContext
}

// NewMockTxContext creates a new mock instance.
func NewMockTxContext(ctrl *gomock.Controller) *MockTxContext {
	mock := &MockTxContext{ctrl: ctrl}
	mock.recorder = &MockTxContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTxContext) EXPECT() *MockTxContextMockRecorder {
	return m.recorder
}

// GetBlockEnvironment mocks base method.
func (m *MockTxContext) GetBlockEnvironment() BlockEnvironment {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockEnvironment")
	ret0, _ := ret[0].(BlockEnvironment)
	return ret0
}

// GetBlockEnvironment indicates an expected call of GetBlockEnvironment.
func (mr *MockTxContextMockRecorder) GetBlockEnvironment() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockEnvironment", reflect.TypeOf((*MockTxContext)(nil).GetBlockEnvironment))
}

// GetInputState mocks base method.
func (m *MockTxContext) GetInputState() WorldState {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInputState")
	ret0, _ := ret[0].(WorldState)
	return ret0
}

// GetInputState indicates an expected call of GetInputState.
func (mr *MockTxContextMockRecorder) GetInputState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInputState", reflect.TypeOf((*MockTxContext)(nil).GetInputState))
}

// GetLogsHash mocks base method.
func (m *MockTxContext) GetLogsHash() common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogsHash")
	ret0, _ := ret[0].(common.Hash)
	return ret0
}

// GetLogsHash indicates an expected call of GetLogsHash.
func (mr *MockTxContextMockRecorder) GetLogsHash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLogsHash", reflect.TypeOf((*MockTxContext)(nil).GetLogsHash))
}

// GetMessage mocks base method.
func (m *MockTxContext) GetMessage() *core.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMessage")
	ret0, _ := ret[0].(*core.Message)
	return ret0
}

// GetMessage indicates an expected call of GetMessage.
func (mr *MockTxContextMockRecorder) GetMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMessage", reflect.TypeOf((*MockTxContext)(nil).GetMessage))
}

// GetOutputState mocks base method.
func (m *MockTxContext) GetOutputState() WorldState {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOutputState")
	ret0, _ := ret[0].(WorldState)
	return ret0
}

// GetOutputState indicates an expected call of GetOutputState.
func (mr *MockTxContextMockRecorder) GetOutputState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOutputState", reflect.TypeOf((*MockTxContext)(nil).GetOutputState))
}

// GetResult mocks base method.
func (m *MockTxContext) GetResult() Result {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResult")
	ret0, _ := ret[0].(Result)
	return ret0
}

// GetResult indicates an expected call of GetResult.
func (mr *MockTxContextMockRecorder) GetResult() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResult", reflect.TypeOf((*MockTxContext)(nil).GetResult))
}

// GetStateHash mocks base method.
func (m *MockTxContext) GetStateHash() common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStateHash")
	ret0, _ := ret[0].(common.Hash)
	return ret0
}

// GetStateHash indicates an expected call of GetStateHash.
func (mr *MockTxContextMockRecorder) GetStateHash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStateHash", reflect.TypeOf((*MockTxContext)(nil).GetStateHash))
}

// MockInputState is a mock of InputState interface.
type MockInputState struct {
	ctrl     *gomock.Controller
	recorder *MockInputStateMockRecorder
	isgomock struct{}
}

// MockInputStateMockRecorder is the mock recorder for MockInputState.
type MockInputStateMockRecorder struct {
	mock *MockInputState
}

// NewMockInputState creates a new mock instance.
func NewMockInputState(ctrl *gomock.Controller) *MockInputState {
	mock := &MockInputState{ctrl: ctrl}
	mock.recorder = &MockInputStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInputState) EXPECT() *MockInputStateMockRecorder {
	return m.recorder
}

// GetInputState mocks base method.
func (m *MockInputState) GetInputState() WorldState {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInputState")
	ret0, _ := ret[0].(WorldState)
	return ret0
}

// GetInputState indicates an expected call of GetInputState.
func (mr *MockInputStateMockRecorder) GetInputState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInputState", reflect.TypeOf((*MockInputState)(nil).GetInputState))
}

// MockTransaction is a mock of Transaction interface.
type MockTransaction struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionMockRecorder
	isgomock struct{}
}

// MockTransactionMockRecorder is the mock recorder for MockTransaction.
type MockTransactionMockRecorder struct {
	mock *MockTransaction
}

// NewMockTransaction creates a new mock instance.
func NewMockTransaction(ctrl *gomock.Controller) *MockTransaction {
	mock := &MockTransaction{ctrl: ctrl}
	mock.recorder = &MockTransactionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransaction) EXPECT() *MockTransactionMockRecorder {
	return m.recorder
}

// GetBlockEnvironment mocks base method.
func (m *MockTransaction) GetBlockEnvironment() BlockEnvironment {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockEnvironment")
	ret0, _ := ret[0].(BlockEnvironment)
	return ret0
}

// GetBlockEnvironment indicates an expected call of GetBlockEnvironment.
func (mr *MockTransactionMockRecorder) GetBlockEnvironment() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockEnvironment", reflect.TypeOf((*MockTransaction)(nil).GetBlockEnvironment))
}

// GetMessage mocks base method.
func (m *MockTransaction) GetMessage() *core.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMessage")
	ret0, _ := ret[0].(*core.Message)
	return ret0
}

// GetMessage indicates an expected call of GetMessage.
func (mr *MockTransactionMockRecorder) GetMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMessage", reflect.TypeOf((*MockTransaction)(nil).GetMessage))
}

// GetOutputState mocks base method.
func (m *MockTransaction) GetOutputState() WorldState {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOutputState")
	ret0, _ := ret[0].(WorldState)
	return ret0
}

// GetOutputState indicates an expected call of GetOutputState.
func (mr *MockTransactionMockRecorder) GetOutputState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOutputState", reflect.TypeOf((*MockTransaction)(nil).GetOutputState))
}

// MockOutputState is a mock of OutputState interface.
type MockOutputState struct {
	ctrl     *gomock.Controller
	recorder *MockOutputStateMockRecorder
	isgomock struct{}
}

// MockOutputStateMockRecorder is the mock recorder for MockOutputState.
type MockOutputStateMockRecorder struct {
	mock *MockOutputState
}

// NewMockOutputState creates a new mock instance.
func NewMockOutputState(ctrl *gomock.Controller) *MockOutputState {
	mock := &MockOutputState{ctrl: ctrl}
	mock.recorder = &MockOutputStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOutputState) EXPECT() *MockOutputStateMockRecorder {
	return m.recorder
}

// GetLogsHash mocks base method.
func (m *MockOutputState) GetLogsHash() common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogsHash")
	ret0, _ := ret[0].(common.Hash)
	return ret0
}

// GetLogsHash indicates an expected call of GetLogsHash.
func (mr *MockOutputStateMockRecorder) GetLogsHash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLogsHash", reflect.TypeOf((*MockOutputState)(nil).GetLogsHash))
}

// GetResult mocks base method.
func (m *MockOutputState) GetResult() Result {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResult")
	ret0, _ := ret[0].(Result)
	return ret0
}

// GetResult indicates an expected call of GetResult.
func (mr *MockOutputStateMockRecorder) GetResult() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResult", reflect.TypeOf((*MockOutputState)(nil).GetResult))
}

// GetStateHash mocks base method.
func (m *MockOutputState) GetStateHash() common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStateHash")
	ret0, _ := ret[0].(common.Hash)
	return ret0
}

// GetStateHash indicates an expected call of GetStateHash.
func (mr *MockOutputStateMockRecorder) GetStateHash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStateHash", reflect.TypeOf((*MockOutputState)(nil).GetStateHash))
}
