// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/cron (interfaces: Runner)
//
// Generated by this command:
//
//	mockgen -destination ./mock_cron.go -package cron github.com/duc-cnzj/mars/v4/internal/cron Runner
//

// Package cron is a generated GoMock package.
package cron

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockRunner is a mock of Runner interface.
type MockRunner struct {
	ctrl     *gomock.Controller
	recorder *MockRunnerMockRecorder
}

// MockRunnerMockRecorder is the mock recorder for MockRunner.
type MockRunnerMockRecorder struct {
	mock *MockRunner
}

// NewMockRunner creates a new mock instance.
func NewMockRunner(ctrl *gomock.Controller) *MockRunner {
	mock := &MockRunner{ctrl: ctrl}
	mock.recorder = &MockRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRunner) EXPECT() *MockRunnerMockRecorder {
	return m.recorder
}

// AddCommand mocks base method.
func (m *MockRunner) AddCommand(arg0, arg1 string, arg2 func()) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddCommand", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddCommand indicates an expected call of AddCommand.
func (mr *MockRunnerMockRecorder) AddCommand(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCommand", reflect.TypeOf((*MockRunner)(nil).AddCommand), arg0, arg1, arg2)
}

// Run mocks base method.
func (m *MockRunner) Run(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockRunnerMockRecorder) Run(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockRunner)(nil).Run), arg0)
}

// Shutdown mocks base method.
func (m *MockRunner) Shutdown(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockRunnerMockRecorder) Shutdown(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockRunner)(nil).Shutdown), arg0)
}
