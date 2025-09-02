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

	gomock "go.uber.org/mock/gomock"
)

// MockProfileDB is a mock of ProfileDB interface.
type MockProfileDB struct {
	ctrl     *gomock.Controller
	recorder *MockProfileDBMockRecorder
	isgomock struct{}
}

// MockProfileDBMockRecorder is the mock recorder for MockProfileDB.
type MockProfileDBMockRecorder struct {
	mock *MockProfileDB
}

// NewMockProfileDB creates a new mock instance.
func NewMockProfileDB(ctrl *gomock.Controller) *MockProfileDB {
	mock := &MockProfileDB{ctrl: ctrl}
	mock.recorder = &MockProfileDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProfileDB) EXPECT() *MockProfileDBMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockProfileDB) Add(data ProfileData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", data)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockProfileDBMockRecorder) Add(data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockProfileDB)(nil).Add), data)
}

// Close mocks base method.
func (m *MockProfileDB) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockProfileDBMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockProfileDB)(nil).Close))
}

// DeleteByBlockRange mocks base method.
func (m *MockProfileDB) DeleteByBlockRange(firstBlock, lastBlock uint64) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByBlockRange", firstBlock, lastBlock)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteByBlockRange indicates an expected call of DeleteByBlockRange.
func (mr *MockProfileDBMockRecorder) DeleteByBlockRange(firstBlock, lastBlock any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByBlockRange", reflect.TypeOf((*MockProfileDB)(nil).DeleteByBlockRange), firstBlock, lastBlock)
}

// Flush mocks base method.
func (m *MockProfileDB) Flush() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Flush")
	ret0, _ := ret[0].(error)
	return ret0
}

// Flush indicates an expected call of Flush.
func (mr *MockProfileDBMockRecorder) Flush() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockProfileDB)(nil).Flush))
}
