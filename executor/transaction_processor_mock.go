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

// Package executor is a generated GoMock package.
package executor

import (
	reflect "reflect"

	state "github.com/0xsoniclabs/aida/state"
	txcontext "github.com/0xsoniclabs/aida/txcontext"
	gomock "go.uber.org/mock/gomock"
)

// Mockprocessor is a mock of processor interface.
type Mockprocessor struct {
	ctrl     *gomock.Controller
	recorder *MockprocessorMockRecorder
	isgomock struct{}
}

// MockprocessorMockRecorder is the mock recorder for Mockprocessor.
type MockprocessorMockRecorder struct {
	mock *Mockprocessor
}

// NewMockprocessor creates a new mock instance.
func NewMockprocessor(ctrl *gomock.Controller) *Mockprocessor {
	mock := &Mockprocessor{ctrl: ctrl}
	mock.recorder = &MockprocessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockprocessor) EXPECT() *MockprocessorMockRecorder {
	return m.recorder
}

// processRegularTx mocks base method.
func (m *Mockprocessor) processRegularTx(db state.VmStateDB, block, tx int, st txcontext.TxContext) (transactionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "processRegularTx", db, block, tx, st)
	ret0, _ := ret[0].(transactionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// processRegularTx indicates an expected call of processRegularTx.
func (mr *MockprocessorMockRecorder) processRegularTx(db, block, tx, st any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "processRegularTx", reflect.TypeOf((*Mockprocessor)(nil).processRegularTx), db, block, tx, st)
}

// MockexecutionResult is a mock of executionResult interface.
type MockexecutionResult struct {
	ctrl     *gomock.Controller
	recorder *MockexecutionResultMockRecorder
	isgomock struct{}
}

// MockexecutionResultMockRecorder is the mock recorder for MockexecutionResult.
type MockexecutionResultMockRecorder struct {
	mock *MockexecutionResult
}

// NewMockexecutionResult creates a new mock instance.
func NewMockexecutionResult(ctrl *gomock.Controller) *MockexecutionResult {
	mock := &MockexecutionResult{ctrl: ctrl}
	mock.recorder = &MockexecutionResultMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockexecutionResult) EXPECT() *MockexecutionResultMockRecorder {
	return m.recorder
}

// Failed mocks base method.
func (m *MockexecutionResult) Failed() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Failed")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Failed indicates an expected call of Failed.
func (mr *MockexecutionResultMockRecorder) Failed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Failed", reflect.TypeOf((*MockexecutionResult)(nil).Failed))
}

// GetError mocks base method.
func (m *MockexecutionResult) GetError() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetError")
	ret0, _ := ret[0].(error)
	return ret0
}

// GetError indicates an expected call of GetError.
func (mr *MockexecutionResultMockRecorder) GetError() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetError", reflect.TypeOf((*MockexecutionResult)(nil).GetError))
}

// GetGasUsed mocks base method.
func (m *MockexecutionResult) GetGasUsed() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGasUsed")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetGasUsed indicates an expected call of GetGasUsed.
func (mr *MockexecutionResultMockRecorder) GetGasUsed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGasUsed", reflect.TypeOf((*MockexecutionResult)(nil).GetGasUsed))
}

// Return mocks base method.
func (m *MockexecutionResult) Return() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Return")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Return indicates an expected call of Return.
func (mr *MockexecutionResultMockRecorder) Return() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Return", reflect.TypeOf((*MockexecutionResult)(nil).Return))
}
