// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v5/internal/event (interfaces: Dispatcher)
//
// Generated by this command:
//
//	mockgen -destination ./mock_event.go -package event github.com/duc-cnzj/mars/v5/internal/event Dispatcher
//

// Package event is a generated GoMock package.
package event

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockDispatcher is a mock of Dispatcher interface.
type MockDispatcher struct {
	ctrl     *gomock.Controller
	recorder *MockDispatcherMockRecorder
}

// MockDispatcherMockRecorder is the mock recorder for MockDispatcher.
type MockDispatcherMockRecorder struct {
	mock *MockDispatcher
}

// NewMockDispatcher creates a new mock instance.
func NewMockDispatcher(ctrl *gomock.Controller) *MockDispatcher {
	mock := &MockDispatcher{ctrl: ctrl}
	mock.recorder = &MockDispatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDispatcher) EXPECT() *MockDispatcherMockRecorder {
	return m.recorder
}

// Dispatch mocks base method.
func (m *MockDispatcher) Dispatch(arg0 Event, arg1 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dispatch", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Dispatch indicates an expected call of Dispatch.
func (mr *MockDispatcherMockRecorder) Dispatch(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dispatch", reflect.TypeOf((*MockDispatcher)(nil).Dispatch), arg0, arg1)
}

// Forget mocks base method.
func (m *MockDispatcher) Forget(arg0 Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Forget", arg0)
}

// Forget indicates an expected call of Forget.
func (mr *MockDispatcherMockRecorder) Forget(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Forget", reflect.TypeOf((*MockDispatcher)(nil).Forget), arg0)
}

// GetListeners mocks base method.
func (m *MockDispatcher) GetListeners(arg0 Event) []Listener {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetListeners", arg0)
	ret0, _ := ret[0].([]Listener)
	return ret0
}

// GetListeners indicates an expected call of GetListeners.
func (mr *MockDispatcherMockRecorder) GetListeners(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListeners", reflect.TypeOf((*MockDispatcher)(nil).GetListeners), arg0)
}

// HasListeners mocks base method.
func (m *MockDispatcher) HasListeners(arg0 Event) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasListeners", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasListeners indicates an expected call of HasListeners.
func (mr *MockDispatcherMockRecorder) HasListeners(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasListeners", reflect.TypeOf((*MockDispatcher)(nil).HasListeners), arg0)
}

// List mocks base method.
func (m *MockDispatcher) List() map[Event][]Listener {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List")
	ret0, _ := ret[0].(map[Event][]Listener)
	return ret0
}

// List indicates an expected call of List.
func (mr *MockDispatcherMockRecorder) List() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockDispatcher)(nil).List))
}

// Listen mocks base method.
func (m *MockDispatcher) Listen(arg0 Event, arg1 Listener) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Listen", arg0, arg1)
}

// Listen indicates an expected call of Listen.
func (mr *MockDispatcherMockRecorder) Listen(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Listen", reflect.TypeOf((*MockDispatcher)(nil).Listen), arg0, arg1)
}

// Run mocks base method.
func (m *MockDispatcher) Run(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockDispatcherMockRecorder) Run(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockDispatcher)(nil).Run), arg0)
}

// Shutdown mocks base method.
func (m *MockDispatcher) Shutdown(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockDispatcherMockRecorder) Shutdown(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockDispatcher)(nil).Shutdown), arg0)
}
