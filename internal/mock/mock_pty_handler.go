// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: PtyHandler)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	websocket "github.com/duc-cnzj/mars-client/v4/websocket"
	contracts "github.com/duc-cnzj/mars/v4/internal/contracts"
	gomock "github.com/golang/mock/gomock"
	remotecommand "k8s.io/client-go/tools/remotecommand"
)

// MockPtyHandler is a mock of PtyHandler interface.
type MockPtyHandler struct {
	ctrl     *gomock.Controller
	recorder *MockPtyHandlerMockRecorder
}

// MockPtyHandlerMockRecorder is the mock recorder for MockPtyHandler.
type MockPtyHandlerMockRecorder struct {
	mock *MockPtyHandler
}

// NewMockPtyHandler creates a new mock instance.
func NewMockPtyHandler(ctrl *gomock.Controller) *MockPtyHandler {
	mock := &MockPtyHandler{ctrl: ctrl}
	mock.recorder = &MockPtyHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPtyHandler) EXPECT() *MockPtyHandlerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockPtyHandler) Close(arg0 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockPtyHandlerMockRecorder) Close(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockPtyHandler)(nil).Close), arg0)
}

// Cols mocks base method.
func (m *MockPtyHandler) Cols() uint16 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Cols")
	ret0, _ := ret[0].(uint16)
	return ret0
}

// Cols indicates an expected call of Cols.
func (mr *MockPtyHandlerMockRecorder) Cols() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Cols", reflect.TypeOf((*MockPtyHandler)(nil).Cols))
}

// Container mocks base method.
func (m *MockPtyHandler) Container() contracts.Container {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Container")
	ret0, _ := ret[0].(contracts.Container)
	return ret0
}

// Container indicates an expected call of Container.
func (mr *MockPtyHandlerMockRecorder) Container() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Container", reflect.TypeOf((*MockPtyHandler)(nil).Container))
}

// IsClosed mocks base method.
func (m *MockPtyHandler) IsClosed() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsClosed")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsClosed indicates an expected call of IsClosed.
func (mr *MockPtyHandlerMockRecorder) IsClosed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsClosed", reflect.TypeOf((*MockPtyHandler)(nil).IsClosed))
}

// Next mocks base method.
func (m *MockPtyHandler) Next() *remotecommand.TerminalSize {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].(*remotecommand.TerminalSize)
	return ret0
}

// Next indicates an expected call of Next.
func (mr *MockPtyHandlerMockRecorder) Next() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*MockPtyHandler)(nil).Next))
}

// Read mocks base method.
func (m *MockPtyHandler) Read(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockPtyHandlerMockRecorder) Read(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockPtyHandler)(nil).Read), arg0)
}

// Recorder mocks base method.
func (m *MockPtyHandler) Recorder() contracts.RecorderInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recorder")
	ret0, _ := ret[0].(contracts.RecorderInterface)
	return ret0
}

// Recorder indicates an expected call of Recorder.
func (mr *MockPtyHandlerMockRecorder) Recorder() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recorder", reflect.TypeOf((*MockPtyHandler)(nil).Recorder))
}

// ResetTerminalRowCol mocks base method.
func (m *MockPtyHandler) ResetTerminalRowCol(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ResetTerminalRowCol", arg0)
}

// ResetTerminalRowCol indicates an expected call of ResetTerminalRowCol.
func (mr *MockPtyHandlerMockRecorder) ResetTerminalRowCol(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetTerminalRowCol", reflect.TypeOf((*MockPtyHandler)(nil).ResetTerminalRowCol), arg0)
}

// Resize mocks base method.
func (m *MockPtyHandler) Resize(arg0 remotecommand.TerminalSize) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Resize", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Resize indicates an expected call of Resize.
func (mr *MockPtyHandlerMockRecorder) Resize(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resize", reflect.TypeOf((*MockPtyHandler)(nil).Resize), arg0)
}

// Rows mocks base method.
func (m *MockPtyHandler) Rows() uint16 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rows")
	ret0, _ := ret[0].(uint16)
	return ret0
}

// Rows indicates an expected call of Rows.
func (mr *MockPtyHandlerMockRecorder) Rows() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rows", reflect.TypeOf((*MockPtyHandler)(nil).Rows))
}

// Send mocks base method.
func (m *MockPtyHandler) Send(arg0 *websocket.TerminalMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockPtyHandlerMockRecorder) Send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockPtyHandler)(nil).Send), arg0)
}

// SetShell mocks base method.
func (m *MockPtyHandler) SetShell(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetShell", arg0)
}

// SetShell indicates an expected call of SetShell.
func (mr *MockPtyHandlerMockRecorder) SetShell(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetShell", reflect.TypeOf((*MockPtyHandler)(nil).SetShell), arg0)
}

// Toast mocks base method.
func (m *MockPtyHandler) Toast(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Toast", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Toast indicates an expected call of Toast.
func (mr *MockPtyHandlerMockRecorder) Toast(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Toast", reflect.TypeOf((*MockPtyHandler)(nil).Toast), arg0)
}

// Write mocks base method.
func (m *MockPtyHandler) Write(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockPtyHandlerMockRecorder) Write(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockPtyHandler)(nil).Write), arg0)
}
