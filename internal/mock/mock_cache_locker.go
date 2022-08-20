// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/internal/contracts (interfaces: Locker)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockLocker is a mock of Locker interface.
type MockLocker struct {
	ctrl     *gomock.Controller
	recorder *MockLockerMockRecorder
}

// MockLockerMockRecorder is the mock recorder for MockLocker.
type MockLockerMockRecorder struct {
	mock *MockLocker
}

// NewMockLocker creates a new mock instance.
func NewMockLocker(ctrl *gomock.Controller) *MockLocker {
	mock := &MockLocker{ctrl: ctrl}
	mock.recorder = &MockLockerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLocker) EXPECT() *MockLockerMockRecorder {
	return m.recorder
}

// Acquire mocks base method.
func (m *MockLocker) Acquire(arg0 string, arg1 int64) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Acquire", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Acquire indicates an expected call of Acquire.
func (mr *MockLockerMockRecorder) Acquire(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Acquire", reflect.TypeOf((*MockLocker)(nil).Acquire), arg0, arg1)
}

// ForceRelease mocks base method.
func (m *MockLocker) ForceRelease(arg0 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceRelease", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// ForceRelease indicates an expected call of ForceRelease.
func (mr *MockLockerMockRecorder) ForceRelease(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceRelease", reflect.TypeOf((*MockLocker)(nil).ForceRelease), arg0)
}

// ID mocks base method.
func (m *MockLocker) ID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ID indicates an expected call of ID.
func (mr *MockLockerMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockLocker)(nil).ID))
}

// Owner mocks base method.
func (m *MockLocker) Owner(arg0 string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Owner", arg0)
	ret0, _ := ret[0].(string)
	return ret0
}

// Owner indicates an expected call of Owner.
func (mr *MockLockerMockRecorder) Owner(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Owner", reflect.TypeOf((*MockLocker)(nil).Owner), arg0)
}

// Release mocks base method.
func (m *MockLocker) Release(arg0 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Release", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Release indicates an expected call of Release.
func (mr *MockLockerMockRecorder) Release(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Release", reflect.TypeOf((*MockLocker)(nil).Release), arg0)
}

// RenewalAcquire mocks base method.
func (m *MockLocker) RenewalAcquire(arg0 string, arg1, arg2 int64) (func(), bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RenewalAcquire", arg0, arg1, arg2)
	ret0, _ := ret[0].(func())
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// RenewalAcquire indicates an expected call of RenewalAcquire.
func (mr *MockLockerMockRecorder) RenewalAcquire(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RenewalAcquire", reflect.TypeOf((*MockLocker)(nil).RenewalAcquire), arg0, arg1, arg2)
}

// Type mocks base method.
func (m *MockLocker) Type() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Type")
	ret0, _ := ret[0].(string)
	return ret0
}

// Type indicates an expected call of Type.
func (mr *MockLockerMockRecorder) Type() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Type", reflect.TypeOf((*MockLocker)(nil).Type))
}
