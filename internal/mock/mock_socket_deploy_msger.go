// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: DeployMsger)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	types "github.com/duc-cnzj/mars-client/v4/types"
	websocket "github.com/duc-cnzj/mars-client/v4/websocket"
	contracts "github.com/duc-cnzj/mars/v4/internal/contracts"
	gomock "github.com/golang/mock/gomock"
)

// MockDeployMsger is a mock of DeployMsger interface.
type MockDeployMsger struct {
	ctrl     *gomock.Controller
	recorder *MockDeployMsgerMockRecorder
}

// MockDeployMsgerMockRecorder is the mock recorder for MockDeployMsger.
type MockDeployMsgerMockRecorder struct {
	mock *MockDeployMsger
}

// NewMockDeployMsger creates a new mock instance.
func NewMockDeployMsger(ctrl *gomock.Controller) *MockDeployMsger {
	mock := &MockDeployMsger{ctrl: ctrl}
	mock.recorder = &MockDeployMsgerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeployMsger) EXPECT() *MockDeployMsgerMockRecorder {
	return m.recorder
}

// SendDeployedResult mocks base method.
func (m *MockDeployMsger) SendDeployedResult(arg0 websocket.ResultType, arg1 string, arg2 *types.ProjectModel) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendDeployedResult", arg0, arg1, arg2)
}

// SendDeployedResult indicates an expected call of SendDeployedResult.
func (mr *MockDeployMsgerMockRecorder) SendDeployedResult(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendDeployedResult", reflect.TypeOf((*MockDeployMsger)(nil).SendDeployedResult), arg0, arg1, arg2)
}

// SendEndError mocks base method.
func (m *MockDeployMsger) SendEndError(arg0 error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendEndError", arg0)
}

// SendEndError indicates an expected call of SendEndError.
func (mr *MockDeployMsgerMockRecorder) SendEndError(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendEndError", reflect.TypeOf((*MockDeployMsger)(nil).SendEndError), arg0)
}

// SendError mocks base method.
func (m *MockDeployMsger) SendError(arg0 error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendError", arg0)
}

// SendError indicates an expected call of SendError.
func (mr *MockDeployMsgerMockRecorder) SendError(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendError", reflect.TypeOf((*MockDeployMsger)(nil).SendError), arg0)
}

// SendMsg mocks base method.
func (m *MockDeployMsger) SendMsg(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendMsg", arg0)
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockDeployMsgerMockRecorder) SendMsg(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockDeployMsger)(nil).SendMsg), arg0)
}

// SendMsgWithContainerLog mocks base method.
func (m *MockDeployMsger) SendMsgWithContainerLog(arg0 string, arg1 []*types.Container) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendMsgWithContainerLog", arg0, arg1)
}

// SendMsgWithContainerLog indicates an expected call of SendMsgWithContainerLog.
func (mr *MockDeployMsgerMockRecorder) SendMsgWithContainerLog(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsgWithContainerLog", reflect.TypeOf((*MockDeployMsger)(nil).SendMsgWithContainerLog), arg0, arg1)
}

// SendProcessPercent mocks base method.
func (m *MockDeployMsger) SendProcessPercent(arg0 int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendProcessPercent", arg0)
}

// SendProcessPercent indicates an expected call of SendProcessPercent.
func (mr *MockDeployMsgerMockRecorder) SendProcessPercent(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendProcessPercent", reflect.TypeOf((*MockDeployMsger)(nil).SendProcessPercent), arg0)
}

// SendProtoMsg mocks base method.
func (m *MockDeployMsger) SendProtoMsg(arg0 contracts.WebsocketMessage) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendProtoMsg", arg0)
}

// SendProtoMsg indicates an expected call of SendProtoMsg.
func (mr *MockDeployMsgerMockRecorder) SendProtoMsg(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendProtoMsg", reflect.TypeOf((*MockDeployMsger)(nil).SendProtoMsg), arg0)
}
