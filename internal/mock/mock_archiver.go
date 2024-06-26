// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: Archiver)
//
// Generated by this command:
//
//	mockgen -destination ../mock/mock_archiver.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts Archiver
//

// Package mock is a generated GoMock package.
package mock

import (
	io "io"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockArchiver is a mock of Archiver interface.
type MockArchiver struct {
	ctrl     *gomock.Controller
	recorder *MockArchiverMockRecorder
}

// MockArchiverMockRecorder is the mock recorder for MockArchiver.
type MockArchiverMockRecorder struct {
	mock *MockArchiver
}

// NewMockArchiver creates a new mock instance.
func NewMockArchiver(ctrl *gomock.Controller) *MockArchiver {
	mock := &MockArchiver{ctrl: ctrl}
	mock.recorder = &MockArchiverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArchiver) EXPECT() *MockArchiverMockRecorder {
	return m.recorder
}

// Archive mocks base method.
func (m *MockArchiver) Archive(arg0 []string, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Archive", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Archive indicates an expected call of Archive.
func (mr *MockArchiverMockRecorder) Archive(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Archive", reflect.TypeOf((*MockArchiver)(nil).Archive), arg0, arg1)
}

// Open mocks base method.
func (m *MockArchiver) Open(arg0 string) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Open", arg0)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Open indicates an expected call of Open.
func (mr *MockArchiverMockRecorder) Open(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Open", reflect.TypeOf((*MockArchiver)(nil).Open), arg0)
}

// Remove mocks base method.
func (m *MockArchiver) Remove(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockArchiverMockRecorder) Remove(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockArchiver)(nil).Remove), arg0)
}
