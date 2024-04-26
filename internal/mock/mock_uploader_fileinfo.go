// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: FileInfo)
//
// Generated by this command:
//
//	mockgen -destination ../mock/mock_uploader_fileinfo.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts FileInfo
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockFileInfo is a mock of FileInfo interface.
type MockFileInfo struct {
	ctrl     *gomock.Controller
	recorder *MockFileInfoMockRecorder
}

// MockFileInfoMockRecorder is the mock recorder for MockFileInfo.
type MockFileInfoMockRecorder struct {
	mock *MockFileInfo
}

// NewMockFileInfo creates a new mock instance.
func NewMockFileInfo(ctrl *gomock.Controller) *MockFileInfo {
	mock := &MockFileInfo{ctrl: ctrl}
	mock.recorder = &MockFileInfoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFileInfo) EXPECT() *MockFileInfoMockRecorder {
	return m.recorder
}

// LastModified mocks base method.
func (m *MockFileInfo) LastModified() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastModified")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// LastModified indicates an expected call of LastModified.
func (mr *MockFileInfoMockRecorder) LastModified() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastModified", reflect.TypeOf((*MockFileInfo)(nil).LastModified))
}

// Path mocks base method.
func (m *MockFileInfo) Path() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Path")
	ret0, _ := ret[0].(string)
	return ret0
}

// Path indicates an expected call of Path.
func (mr *MockFileInfoMockRecorder) Path() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Path", reflect.TypeOf((*MockFileInfo)(nil).Path))
}

// Size mocks base method.
func (m *MockFileInfo) Size() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// Size indicates an expected call of Size.
func (mr *MockFileInfoMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockFileInfo)(nil).Size))
}
