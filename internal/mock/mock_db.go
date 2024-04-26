// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: DBManager)
//
// Generated by this command:
//
//	mockgen -destination ../mock/mock_db.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts DBManager
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	gorm "gorm.io/gorm"
)

// MockDBManager is a mock of DBManager interface.
type MockDBManager struct {
	ctrl     *gomock.Controller
	recorder *MockDBManagerMockRecorder
}

// MockDBManagerMockRecorder is the mock recorder for MockDBManager.
type MockDBManagerMockRecorder struct {
	mock *MockDBManager
}

// NewMockDBManager creates a new mock instance.
func NewMockDBManager(ctrl *gomock.Controller) *MockDBManager {
	mock := &MockDBManager{ctrl: ctrl}
	mock.recorder = &MockDBManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBManager) EXPECT() *MockDBManagerMockRecorder {
	return m.recorder
}

// AutoMigrate mocks base method.
func (m *MockDBManager) AutoMigrate(arg0 ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AutoMigrate", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// AutoMigrate indicates an expected call of AutoMigrate.
func (mr *MockDBManagerMockRecorder) AutoMigrate(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AutoMigrate", reflect.TypeOf((*MockDBManager)(nil).AutoMigrate), arg0...)
}

// DB mocks base method.
func (m *MockDBManager) DB() *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DB")
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// DB indicates an expected call of DB.
func (mr *MockDBManagerMockRecorder) DB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DB", reflect.TypeOf((*MockDBManager)(nil).DB))
}

// SetDB mocks base method.
func (m *MockDBManager) SetDB(arg0 *gorm.DB) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetDB", arg0)
}

// SetDB indicates an expected call of SetDB.
func (mr *MockDBManagerMockRecorder) SetDB(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDB", reflect.TypeOf((*MockDBManager)(nil).SetDB), arg0)
}
