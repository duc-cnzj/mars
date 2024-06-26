// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/v4/internal/contracts (interfaces: CommitInterface)
//
// Generated by this command:
//
//	mockgen -destination ../mock/mock_git_server_commit.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts CommitInterface
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockCommitInterface is a mock of CommitInterface interface.
type MockCommitInterface struct {
	ctrl     *gomock.Controller
	recorder *MockCommitInterfaceMockRecorder
}

// MockCommitInterfaceMockRecorder is the mock recorder for MockCommitInterface.
type MockCommitInterfaceMockRecorder struct {
	mock *MockCommitInterface
}

// NewMockCommitInterface creates a new mock instance.
func NewMockCommitInterface(ctrl *gomock.Controller) *MockCommitInterface {
	mock := &MockCommitInterface{ctrl: ctrl}
	mock.recorder = &MockCommitInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCommitInterface) EXPECT() *MockCommitInterfaceMockRecorder {
	return m.recorder
}

// GetAuthorEmail mocks base method.
func (m *MockCommitInterface) GetAuthorEmail() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuthorEmail")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetAuthorEmail indicates an expected call of GetAuthorEmail.
func (mr *MockCommitInterfaceMockRecorder) GetAuthorEmail() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuthorEmail", reflect.TypeOf((*MockCommitInterface)(nil).GetAuthorEmail))
}

// GetAuthorName mocks base method.
func (m *MockCommitInterface) GetAuthorName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuthorName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetAuthorName indicates an expected call of GetAuthorName.
func (mr *MockCommitInterfaceMockRecorder) GetAuthorName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuthorName", reflect.TypeOf((*MockCommitInterface)(nil).GetAuthorName))
}

// GetCommittedDate mocks base method.
func (m *MockCommitInterface) GetCommittedDate() *time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCommittedDate")
	ret0, _ := ret[0].(*time.Time)
	return ret0
}

// GetCommittedDate indicates an expected call of GetCommittedDate.
func (mr *MockCommitInterfaceMockRecorder) GetCommittedDate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommittedDate", reflect.TypeOf((*MockCommitInterface)(nil).GetCommittedDate))
}

// GetCommitterEmail mocks base method.
func (m *MockCommitInterface) GetCommitterEmail() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCommitterEmail")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCommitterEmail indicates an expected call of GetCommitterEmail.
func (mr *MockCommitInterfaceMockRecorder) GetCommitterEmail() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommitterEmail", reflect.TypeOf((*MockCommitInterface)(nil).GetCommitterEmail))
}

// GetCommitterName mocks base method.
func (m *MockCommitInterface) GetCommitterName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCommitterName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCommitterName indicates an expected call of GetCommitterName.
func (mr *MockCommitInterfaceMockRecorder) GetCommitterName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommitterName", reflect.TypeOf((*MockCommitInterface)(nil).GetCommitterName))
}

// GetCreatedAt mocks base method.
func (m *MockCommitInterface) GetCreatedAt() *time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCreatedAt")
	ret0, _ := ret[0].(*time.Time)
	return ret0
}

// GetCreatedAt indicates an expected call of GetCreatedAt.
func (mr *MockCommitInterfaceMockRecorder) GetCreatedAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCreatedAt", reflect.TypeOf((*MockCommitInterface)(nil).GetCreatedAt))
}

// GetID mocks base method.
func (m *MockCommitInterface) GetID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetID indicates an expected call of GetID.
func (mr *MockCommitInterfaceMockRecorder) GetID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockCommitInterface)(nil).GetID))
}

// GetMessage mocks base method.
func (m *MockCommitInterface) GetMessage() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMessage")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetMessage indicates an expected call of GetMessage.
func (mr *MockCommitInterfaceMockRecorder) GetMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMessage", reflect.TypeOf((*MockCommitInterface)(nil).GetMessage))
}

// GetProjectID mocks base method.
func (m *MockCommitInterface) GetProjectID() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProjectID")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetProjectID indicates an expected call of GetProjectID.
func (mr *MockCommitInterfaceMockRecorder) GetProjectID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProjectID", reflect.TypeOf((*MockCommitInterface)(nil).GetProjectID))
}

// GetShortID mocks base method.
func (m *MockCommitInterface) GetShortID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetShortID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetShortID indicates an expected call of GetShortID.
func (mr *MockCommitInterfaceMockRecorder) GetShortID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetShortID", reflect.TypeOf((*MockCommitInterface)(nil).GetShortID))
}

// GetTitle mocks base method.
func (m *MockCommitInterface) GetTitle() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTitle")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetTitle indicates an expected call of GetTitle.
func (mr *MockCommitInterfaceMockRecorder) GetTitle() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTitle", reflect.TypeOf((*MockCommitInterface)(nil).GetTitle))
}

// GetWebURL mocks base method.
func (m *MockCommitInterface) GetWebURL() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWebURL")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetWebURL indicates an expected call of GetWebURL.
func (mr *MockCommitInterfaceMockRecorder) GetWebURL() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWebURL", reflect.TypeOf((*MockCommitInterface)(nil).GetWebURL))
}
