// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/api/v4/metrics (interfaces: Metrics_StreamTopPodServer)
//
// Generated by this command:
//
//	mockgen -destination ./mock_svc.go -package services github.com/duc-cnzj/mars/api/v4/metrics Metrics_StreamTopPodServer
//

// Package services is a generated GoMock package.
package services

import (
	context "context"
	reflect "reflect"

	metrics "github.com/duc-cnzj/mars/api/v4/metrics"
	gomock "go.uber.org/mock/gomock"
	metadata "google.golang.org/grpc/metadata"
)

// MockMetrics_StreamTopPodServer is a mock of Metrics_StreamTopPodServer interface.
type MockMetrics_StreamTopPodServer struct {
	ctrl     *gomock.Controller
	recorder *MockMetrics_StreamTopPodServerMockRecorder
}

// MockMetrics_StreamTopPodServerMockRecorder is the mock recorder for MockMetrics_StreamTopPodServer.
type MockMetrics_StreamTopPodServerMockRecorder struct {
	mock *MockMetrics_StreamTopPodServer
}

// NewMockMetrics_StreamTopPodServer creates a new mock instance.
func NewMockMetrics_StreamTopPodServer(ctrl *gomock.Controller) *MockMetrics_StreamTopPodServer {
	mock := &MockMetrics_StreamTopPodServer{ctrl: ctrl}
	mock.recorder = &MockMetrics_StreamTopPodServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetrics_StreamTopPodServer) EXPECT() *MockMetrics_StreamTopPodServerMockRecorder {
	return m.recorder
}

// Context mocks base method.
func (m *MockMetrics_StreamTopPodServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockMetrics_StreamTopPodServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockMetrics_StreamTopPodServer)(nil).Context))
}

// RecvMsg mocks base method.
func (m *MockMetrics_StreamTopPodServer) RecvMsg(arg0 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockMetrics_StreamTopPodServerMockRecorder) RecvMsg(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockMetrics_StreamTopPodServer)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockMetrics_StreamTopPodServer) Send(arg0 *metrics.TopPodResponse) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockMetrics_StreamTopPodServerMockRecorder) Send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockMetrics_StreamTopPodServer)(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockMetrics_StreamTopPodServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockMetrics_StreamTopPodServerMockRecorder) SendHeader(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockMetrics_StreamTopPodServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m *MockMetrics_StreamTopPodServer) SendMsg(arg0 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockMetrics_StreamTopPodServerMockRecorder) SendMsg(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockMetrics_StreamTopPodServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method.
func (m *MockMetrics_StreamTopPodServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockMetrics_StreamTopPodServerMockRecorder) SetHeader(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockMetrics_StreamTopPodServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockMetrics_StreamTopPodServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockMetrics_StreamTopPodServerMockRecorder) SetTrailer(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockMetrics_StreamTopPodServer)(nil).SetTrailer), arg0)
}