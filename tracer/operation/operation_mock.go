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

// Package operation is a generated GoMock package.
package operation

import (
	io "io"
	reflect "reflect"
	time "time"

	state "github.com/0xsoniclabs/aida/state"
	context "github.com/0xsoniclabs/aida/tracer/context"
	gomock "go.uber.org/mock/gomock"
)

// MockOperation is a mock of Operation interface.
type MockOperation struct {
	ctrl     *gomock.Controller
	recorder *MockOperationMockRecorder
	isgomock struct{}
}

// MockOperationMockRecorder is the mock recorder for MockOperation.
type MockOperationMockRecorder struct {
	mock *MockOperation
}

// NewMockOperation creates a new mock instance.
func NewMockOperation(ctrl *gomock.Controller) *MockOperation {
	mock := &MockOperation{ctrl: ctrl}
	mock.recorder = &MockOperationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOperation) EXPECT() *MockOperationMockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockOperation) Debug(arg0 *context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", arg0)
}

// Debug indicates an expected call of Debug.
func (mr *MockOperationMockRecorder) Debug(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockOperation)(nil).Debug), arg0)
}

// Execute mocks base method.
func (m *MockOperation) Execute(arg0 state.StateDB, arg1 *context.Replay) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", arg0, arg1)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// Execute indicates an expected call of Execute.
func (mr *MockOperationMockRecorder) Execute(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockOperation)(nil).Execute), arg0, arg1)
}

// GetId mocks base method.
func (m *MockOperation) GetId() byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetId")
	ret0, _ := ret[0].(byte)
	return ret0
}

// GetId indicates an expected call of GetId.
func (mr *MockOperationMockRecorder) GetId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetId", reflect.TypeOf((*MockOperation)(nil).GetId))
}

// Write mocks base method.
func (m *MockOperation) Write(arg0 io.Writer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write.
func (mr *MockOperationMockRecorder) Write(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockOperation)(nil).Write), arg0)
}
