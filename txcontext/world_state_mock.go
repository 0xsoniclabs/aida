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
	gomock "go.uber.org/mock/gomock"
)

// MockWorldState is a mock of WorldState interface.
type MockWorldState struct {
	ctrl     *gomock.Controller
	recorder *MockWorldStateMockRecorder
	isgomock struct{}
}

// MockWorldStateMockRecorder is the mock recorder for MockWorldState.
type MockWorldStateMockRecorder struct {
	mock *MockWorldState
}

// NewMockWorldState creates a new mock instance.
func NewMockWorldState(ctrl *gomock.Controller) *MockWorldState {
	mock := &MockWorldState{ctrl: ctrl}
	mock.recorder = &MockWorldStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorldState) EXPECT() *MockWorldStateMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockWorldState) Delete(addr common.Address) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Delete", addr)
}

// Delete indicates an expected call of Delete.
func (mr *MockWorldStateMockRecorder) Delete(addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockWorldState)(nil).Delete), addr)
}

// Equal mocks base method.
func (m *MockWorldState) Equal(arg0 WorldState) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Equal", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Equal indicates an expected call of Equal.
func (mr *MockWorldStateMockRecorder) Equal(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Equal", reflect.TypeOf((*MockWorldState)(nil).Equal), arg0)
}

// ForEachAccount mocks base method.
func (m *MockWorldState) ForEachAccount(arg0 AccountHandler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ForEachAccount", arg0)
}

// ForEachAccount indicates an expected call of ForEachAccount.
func (mr *MockWorldStateMockRecorder) ForEachAccount(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForEachAccount", reflect.TypeOf((*MockWorldState)(nil).ForEachAccount), arg0)
}

// Get mocks base method.
func (m *MockWorldState) Get(addr common.Address) Account {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", addr)
	ret0, _ := ret[0].(Account)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockWorldStateMockRecorder) Get(addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockWorldState)(nil).Get), addr)
}

// Has mocks base method.
func (m *MockWorldState) Has(addr common.Address) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Has", addr)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Has indicates an expected call of Has.
func (mr *MockWorldStateMockRecorder) Has(addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Has", reflect.TypeOf((*MockWorldState)(nil).Has), addr)
}

// Len mocks base method.
func (m *MockWorldState) Len() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Len")
	ret0, _ := ret[0].(int)
	return ret0
}

// Len indicates an expected call of Len.
func (mr *MockWorldStateMockRecorder) Len() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Len", reflect.TypeOf((*MockWorldState)(nil).Len))
}

// String mocks base method.
func (m *MockWorldState) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockWorldStateMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockWorldState)(nil).String))
}
