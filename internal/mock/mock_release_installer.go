// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: ReleaseInstaller)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	contracts "github.com/duc-cnzj/mars/v4/internal/contracts"
	gomock "github.com/golang/mock/gomock"
	chart "helm.sh/helm/v3/pkg/chart"
	release "helm.sh/helm/v3/pkg/release"
)

// MockReleaseInstaller is a mock of ReleaseInstaller interface.
type MockReleaseInstaller struct {
	ctrl     *gomock.Controller
	recorder *MockReleaseInstallerMockRecorder
}

// MockReleaseInstallerMockRecorder is the mock recorder for MockReleaseInstaller.
type MockReleaseInstallerMockRecorder struct {
	mock *MockReleaseInstaller
}

// NewMockReleaseInstaller creates a new mock instance.
func NewMockReleaseInstaller(ctrl *gomock.Controller) *MockReleaseInstaller {
	mock := &MockReleaseInstaller{ctrl: ctrl}
	mock.recorder = &MockReleaseInstallerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReleaseInstaller) EXPECT() *MockReleaseInstallerMockRecorder {
	return m.recorder
}

// Chart mocks base method.
func (m *MockReleaseInstaller) Chart() *chart.Chart {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Chart")
	ret0, _ := ret[0].(*chart.Chart)
	return ret0
}

// Chart indicates an expected call of Chart.
func (mr *MockReleaseInstallerMockRecorder) Chart() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Chart", reflect.TypeOf((*MockReleaseInstaller)(nil).Chart))
}

// Logs mocks base method.
func (m *MockReleaseInstaller) Logs() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logs")
	ret0, _ := ret[0].([]string)
	return ret0
}

// Logs indicates an expected call of Logs.
func (mr *MockReleaseInstallerMockRecorder) Logs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logs", reflect.TypeOf((*MockReleaseInstaller)(nil).Logs))
}

// Run mocks base method.
func (m *MockReleaseInstaller) Run(arg0 context.Context, arg1 contracts.SafeWriteMessageChInterface, arg2 contracts.Percentable, arg3 bool, arg4 string) (*release.Release, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*release.Release)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Run indicates an expected call of Run.
func (mr *MockReleaseInstallerMockRecorder) Run(arg0, arg1, arg2, arg3, arg4 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockReleaseInstaller)(nil).Run), arg0, arg1, arg2, arg3, arg4)
}
