// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: SessionMapper)
//
// Generated by this command:
//
//	mockgen -destination ../mock/mock_socket_session_mapper.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts SessionMapper
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	websocket "github.com/duc-cnzj/mars/api/v4/websocket"
	contracts "github.com/duc-cnzj/mars/v4/internal/contracts"
	gomock "go.uber.org/mock/gomock"
)

// MockSessionMapper is a mock of SessionMapper interface.
type MockSessionMapper struct {
	ctrl     *gomock.Controller
	recorder *MockSessionMapperMockRecorder
}

// MockSessionMapperMockRecorder is the mock recorder for MockSessionMapper.
type MockSessionMapperMockRecorder struct {
	mock *MockSessionMapper
}

// NewMockSessionMapper creates a new mock instance.
func NewMockSessionMapper(ctrl *gomock.Controller) *MockSessionMapper {
	mock := &MockSessionMapper{ctrl: ctrl}
	mock.recorder = &MockSessionMapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionMapper) EXPECT() *MockSessionMapperMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockSessionMapper) Close(arg0 string, arg1 uint32, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close", arg0, arg1, arg2)
}

// Close indicates an expected call of Close.
func (mr *MockSessionMapperMockRecorder) Close(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockSessionMapper)(nil).Close), arg0, arg1, arg2)
}

// CloseAll mocks base method.
func (m *MockSessionMapper) CloseAll() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CloseAll")
}

// CloseAll indicates an expected call of CloseAll.
func (mr *MockSessionMapperMockRecorder) CloseAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseAll", reflect.TypeOf((*MockSessionMapper)(nil).CloseAll))
}

// Get mocks base method.
func (m *MockSessionMapper) Get(arg0 string) (contracts.PtyHandler, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(contracts.PtyHandler)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSessionMapperMockRecorder) Get(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSessionMapper)(nil).Get), arg0)
}

// Send mocks base method.
func (m *MockSessionMapper) Send(arg0 *websocket.TerminalMessage) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Send", arg0)
}

// Send indicates an expected call of Send.
func (mr *MockSessionMapperMockRecorder) Send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSessionMapper)(nil).Send), arg0)
}

// Set mocks base method.
func (m *MockSessionMapper) Set(arg0 string, arg1 contracts.PtyHandler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Set", arg0, arg1)
}

// Set indicates an expected call of Set.
func (mr *MockSessionMapperMockRecorder) Set(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockSessionMapper)(nil).Set), arg0, arg1)
}
