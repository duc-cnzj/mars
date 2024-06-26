// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: PodFileCopier)
//
// Generated by this command:
//
//	mockgen -destination ../mock/mock_pod_copier.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts PodFileCopier
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	contracts "github.com/duc-cnzj/mars/v4/internal/contracts"
	gomock "go.uber.org/mock/gomock"
	kubernetes "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
)

// MockPodFileCopier is a mock of PodFileCopier interface.
type MockPodFileCopier struct {
	ctrl     *gomock.Controller
	recorder *MockPodFileCopierMockRecorder
}

// MockPodFileCopierMockRecorder is the mock recorder for MockPodFileCopier.
type MockPodFileCopierMockRecorder struct {
	mock *MockPodFileCopier
}

// NewMockPodFileCopier creates a new mock instance.
func NewMockPodFileCopier(ctrl *gomock.Controller) *MockPodFileCopier {
	mock := &MockPodFileCopier{ctrl: ctrl}
	mock.recorder = &MockPodFileCopierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPodFileCopier) EXPECT() *MockPodFileCopierMockRecorder {
	return m.recorder
}

// Copy mocks base method.
func (m *MockPodFileCopier) Copy(arg0, arg1, arg2, arg3, arg4 string, arg5 kubernetes.Interface, arg6 *rest.Config) (*contracts.CopyFileToPodResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Copy", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(*contracts.CopyFileToPodResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Copy indicates an expected call of Copy.
func (mr *MockPodFileCopierMockRecorder) Copy(arg0, arg1, arg2, arg3, arg4, arg5, arg6 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Copy", reflect.TypeOf((*MockPodFileCopier)(nil).Copy), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}
