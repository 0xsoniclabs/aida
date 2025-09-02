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

// Package state is a generated GoMock package.
package state

import (
	reflect "reflect"

	common "github.com/ethereum/go-ethereum/common"
	gomock "go.uber.org/mock/gomock"
)

// MockCodeCache is a mock of CodeCache interface.
type MockCodeCache struct {
	ctrl     *gomock.Controller
	recorder *MockCodeCacheMockRecorder
	isgomock struct{}
}

// MockCodeCacheMockRecorder is the mock recorder for MockCodeCache.
type MockCodeCacheMockRecorder struct {
	mock *MockCodeCache
}

// NewMockCodeCache creates a new mock instance.
func NewMockCodeCache(ctrl *gomock.Controller) *MockCodeCache {
	mock := &MockCodeCache{ctrl: ctrl}
	mock.recorder = &MockCodeCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCodeCache) EXPECT() *MockCodeCacheMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockCodeCache) Get(addr common.Address, code []byte) common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", addr, code)
	ret0, _ := ret[0].(common.Hash)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockCodeCacheMockRecorder) Get(addr, code any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCodeCache)(nil).Get), addr, code)
}
