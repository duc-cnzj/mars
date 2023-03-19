// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: RecorderInterface)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRecorderInterface is a mock of RecorderInterface interface.
type MockRecorderInterface struct {
	ctrl     *gomock.Controller
	recorder *MockRecorderInterfaceMockRecorder
}

// MockRecorderInterfaceMockRecorder is the mock recorder for MockRecorderInterface.
type MockRecorderInterfaceMockRecorder struct {
	mock *MockRecorderInterface
}

// NewMockRecorderInterface creates a new mock instance.
func NewMockRecorderInterface(ctrl *gomock.Controller) *MockRecorderInterface {
	mock := &MockRecorderInterface{ctrl: ctrl}
	mock.recorder = &MockRecorderInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRecorderInterface) EXPECT() *MockRecorderInterfaceMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockRecorderInterface) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockRecorderInterfaceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockRecorderInterface)(nil).Close))
}

// GetShell mocks base method.
func (m *MockRecorderInterface) GetShell() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetShell")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetShell indicates an expected call of GetShell.
func (mr *MockRecorderInterfaceMockRecorder) GetShell() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetShell", reflect.TypeOf((*MockRecorderInterface)(nil).GetShell))
}

// Resize mocks base method.
func (m *MockRecorderInterface) Resize(arg0, arg1 uint16) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Resize", arg0, arg1)
}

// Resize indicates an expected call of Resize.
func (mr *MockRecorderInterfaceMockRecorder) Resize(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resize", reflect.TypeOf((*MockRecorderInterface)(nil).Resize), arg0, arg1)
}

// SetShell mocks base method.
func (m *MockRecorderInterface) SetShell(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetShell", arg0)
}

// SetShell indicates an expected call of SetShell.
func (mr *MockRecorderInterfaceMockRecorder) SetShell(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetShell", reflect.TypeOf((*MockRecorderInterface)(nil).SetShell), arg0)
}

// Write mocks base method.
func (m *MockRecorderInterface) Write(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write.
func (mr *MockRecorderInterfaceMockRecorder) Write(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockRecorderInterface)(nil).Write), arg0)
}
