// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: CronRunner)
//
// Generated by this command:
//
//	mockgen -destination ../mock/mock_cron_runner.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts CronRunner
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockCronRunner is a mock of CronRunner interface.
type MockCronRunner struct {
	ctrl     *gomock.Controller
	recorder *MockCronRunnerMockRecorder
}

// MockCronRunnerMockRecorder is the mock recorder for MockCronRunner.
type MockCronRunnerMockRecorder struct {
	mock *MockCronRunner
}

// NewMockCronRunner creates a new mock instance.
func NewMockCronRunner(ctrl *gomock.Controller) *MockCronRunner {
	mock := &MockCronRunner{ctrl: ctrl}
	mock.recorder = &MockCronRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCronRunner) EXPECT() *MockCronRunnerMockRecorder {
	return m.recorder
}

// AddCommand mocks base method.
func (m *MockCronRunner) AddCommand(arg0, arg1 string, arg2 func()) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddCommand", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddCommand indicates an expected call of AddCommand.
func (mr *MockCronRunnerMockRecorder) AddCommand(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCommand", reflect.TypeOf((*MockCronRunner)(nil).AddCommand), arg0, arg1, arg2)
}

// Run mocks base method.
func (m *MockCronRunner) Run(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockCronRunnerMockRecorder) Run(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockCronRunner)(nil).Run), arg0)
}

// Shutdown mocks base method.
func (m *MockCronRunner) Shutdown(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockCronRunnerMockRecorder) Shutdown(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockCronRunner)(nil).Shutdown), arg0)
}
