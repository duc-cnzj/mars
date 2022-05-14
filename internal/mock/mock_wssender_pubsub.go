// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/internal/contracts (interfaces: PubSub)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	contracts "github.com/duc-cnzj/mars/internal/contracts"
	gomock "github.com/golang/mock/gomock"
)

// MockPubSub is a mock of PubSub interface.
type MockPubSub struct {
	ctrl     *gomock.Controller
	recorder *MockPubSubMockRecorder
}

// MockPubSubMockRecorder is the mock recorder for MockPubSub.
type MockPubSubMockRecorder struct {
	mock *MockPubSub
}

// NewMockPubSub creates a new mock instance.
func NewMockPubSub(ctrl *gomock.Controller) *MockPubSub {
	mock := &MockPubSub{ctrl: ctrl}
	mock.recorder = &MockPubSubMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPubSub) EXPECT() *MockPubSubMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockPubSub) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockPubSubMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockPubSub)(nil).Close))
}

// ID mocks base method.
func (m *MockPubSub) ID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ID indicates an expected call of ID.
func (mr *MockPubSubMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockPubSub)(nil).ID))
}

// Info mocks base method.
func (m *MockPubSub) Info() any {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Info")
	ret0, _ := ret[0].(any)
	return ret0
}

// Info indicates an expected call of Info.
func (mr *MockPubSubMockRecorder) Info() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockPubSub)(nil).Info))
}

// Subscribe mocks base method.
func (m *MockPubSub) Subscribe() <-chan []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe")
	ret0, _ := ret[0].(<-chan []byte)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockPubSubMockRecorder) Subscribe() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockPubSub)(nil).Subscribe))
}

// ToAll mocks base method.
func (m *MockPubSub) ToAll(arg0 contracts.WebsocketMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToAll", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ToAll indicates an expected call of ToAll.
func (mr *MockPubSubMockRecorder) ToAll(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToAll", reflect.TypeOf((*MockPubSub)(nil).ToAll), arg0)
}

// ToOthers mocks base method.
func (m *MockPubSub) ToOthers(arg0 contracts.WebsocketMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToOthers", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ToOthers indicates an expected call of ToOthers.
func (mr *MockPubSubMockRecorder) ToOthers(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToOthers", reflect.TypeOf((*MockPubSub)(nil).ToOthers), arg0)
}

// ToSelf mocks base method.
func (m *MockPubSub) ToSelf(arg0 contracts.WebsocketMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ToSelf", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ToSelf indicates an expected call of ToSelf.
func (mr *MockPubSubMockRecorder) ToSelf(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ToSelf", reflect.TypeOf((*MockPubSub)(nil).ToSelf), arg0)
}

// Uid mocks base method.
func (m *MockPubSub) Uid() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Uid")
	ret0, _ := ret[0].(string)
	return ret0
}

// Uid indicates an expected call of Uid.
func (mr *MockPubSubMockRecorder) Uid() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Uid", reflect.TypeOf((*MockPubSub)(nil).Uid))
}