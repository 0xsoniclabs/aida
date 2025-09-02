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

// Package utils is a generated GoMock package.
package utils

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockPrinter is a mock of Printer interface.
type MockPrinter struct {
	ctrl     *gomock.Controller
	recorder *MockPrinterMockRecorder
	isgomock struct{}
}

// MockPrinterMockRecorder is the mock recorder for MockPrinter.
type MockPrinterMockRecorder struct {
	mock *MockPrinter
}

// NewMockPrinter creates a new mock instance.
func NewMockPrinter(ctrl *gomock.Controller) *MockPrinter {
	mock := &MockPrinter{ctrl: ctrl}
	mock.recorder = &MockPrinterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPrinter) EXPECT() *MockPrinterMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockPrinter) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockPrinterMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockPrinter)(nil).Close))
}

// Print mocks base method.
func (m *MockPrinter) Print() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Print")
	ret0, _ := ret[0].(error)
	return ret0
}

// Print indicates an expected call of Print.
func (mr *MockPrinterMockRecorder) Print() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Print", reflect.TypeOf((*MockPrinter)(nil).Print))
}

// MockIFlusher is a mock of IFlusher interface.
type MockIFlusher struct {
	ctrl     *gomock.Controller
	recorder *MockIFlusherMockRecorder
	isgomock struct{}
}

// MockIFlusherMockRecorder is the mock recorder for MockIFlusher.
type MockIFlusherMockRecorder struct {
	mock *MockIFlusher
}

// NewMockIFlusher creates a new mock instance.
func NewMockIFlusher(ctrl *gomock.Controller) *MockIFlusher {
	mock := &MockIFlusher{ctrl: ctrl}
	mock.recorder = &MockIFlusherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIFlusher) EXPECT() *MockIFlusherMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockIFlusher) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockIFlusherMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIFlusher)(nil).Close))
}

// Print mocks base method.
func (m *MockIFlusher) Print() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Print")
	ret0, _ := ret[0].(error)
	return ret0
}

// Print indicates an expected call of Print.
func (mr *MockIFlusherMockRecorder) Print() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Print", reflect.TypeOf((*MockIFlusher)(nil).Print))
}
