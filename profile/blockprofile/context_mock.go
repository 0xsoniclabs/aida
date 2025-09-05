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

// Package blockprofile is a generated GoMock package.
package blockprofile

import (
	reflect "reflect"
	time "time"

	executor "github.com/0xsoniclabs/aida/executor"
	graphutil "github.com/0xsoniclabs/aida/profile/graphutil"
	txcontext "github.com/0xsoniclabs/aida/txcontext"
	gomock "go.uber.org/mock/gomock"
)

// MockContext is a mock of Context interface.
type MockContext struct {
	ctrl     *gomock.Controller
	recorder *MockContextMockRecorder
	isgomock struct{}
}

// MockContextMockRecorder is the mock recorder for MockContext.
type MockContextMockRecorder struct {
	mock *MockContext
}

// NewMockContext creates a new mock instance.
func NewMockContext(ctrl *gomock.Controller) *MockContext {
	mock := &MockContext{ctrl: ctrl}
	mock.recorder = &MockContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContext) EXPECT() *MockContextMockRecorder {
	return m.recorder
}

// GetProfileData mocks base method.
func (m *MockContext) GetProfileData(curBlock uint64, tBlock time.Duration) (*ProfileData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProfileData", curBlock, tBlock)
	ret0, _ := ret[0].(*ProfileData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProfileData indicates an expected call of GetProfileData.
func (mr *MockContextMockRecorder) GetProfileData(curBlock, tBlock any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfileData", reflect.TypeOf((*MockContext)(nil).GetProfileData), curBlock, tBlock)
}

// RecordTransaction mocks base method.
func (m *MockContext) RecordTransaction(state executor.State[txcontext.TxContext], tTransaction time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordTransaction", state, tTransaction)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordTransaction indicates an expected call of RecordTransaction.
func (mr *MockContextMockRecorder) RecordTransaction(state, tTransaction any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordTransaction", reflect.TypeOf((*MockContext)(nil).RecordTransaction), state, tTransaction)
}

// dependencies mocks base method.
func (m *MockContext) dependencies(addresses AddressSet) graphutil.OrdinalSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "dependencies", addresses)
	ret0, _ := ret[0].(graphutil.OrdinalSet)
	return ret0
}

// dependencies indicates an expected call of dependencies.
func (mr *MockContextMockRecorder) dependencies(addresses any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "dependencies", reflect.TypeOf((*MockContext)(nil).dependencies), addresses)
}

// earliestTimeToRun mocks base method.
func (m *MockContext) earliestTimeToRun(addresses AddressSet) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "earliestTimeToRun", addresses)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// earliestTimeToRun indicates an expected call of earliestTimeToRun.
func (mr *MockContextMockRecorder) earliestTimeToRun(addresses any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "earliestTimeToRun", reflect.TypeOf((*MockContext)(nil).earliestTimeToRun), addresses)
}
