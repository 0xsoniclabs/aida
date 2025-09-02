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

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	state "github.com/0xsoniclabs/aida/state"
	common "github.com/ethereum/go-ethereum/common"
	vm "github.com/ethereum/go-ethereum/core/vm"
	gomock "go.uber.org/mock/gomock"
)

// MockiEvmProcessor is a mock of iEvmProcessor interface.
type MockiEvmProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockiEvmProcessorMockRecorder
	isgomock struct{}
}

// MockiEvmProcessorMockRecorder is the mock recorder for MockiEvmProcessor.
type MockiEvmProcessorMockRecorder struct {
	mock *MockiEvmProcessor
}

// NewMockiEvmProcessor creates a new mock instance.
func NewMockiEvmProcessor(ctrl *gomock.Controller) *MockiEvmProcessor {
	mock := &MockiEvmProcessor{ctrl: ctrl}
	mock.recorder = &MockiEvmProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockiEvmProcessor) EXPECT() *MockiEvmProcessorMockRecorder {
	return m.recorder
}

// ProcessParentBlockHash mocks base method.
func (m *MockiEvmProcessor) ProcessParentBlockHash(arg0 common.Hash, arg1 *vm.EVM, arg2 state.StateDB) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ProcessParentBlockHash", arg0, arg1, arg2)
}

// ProcessParentBlockHash indicates an expected call of ProcessParentBlockHash.
func (mr *MockiEvmProcessorMockRecorder) ProcessParentBlockHash(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessParentBlockHash", reflect.TypeOf((*MockiEvmProcessor)(nil).ProcessParentBlockHash), arg0, arg1, arg2)
}
