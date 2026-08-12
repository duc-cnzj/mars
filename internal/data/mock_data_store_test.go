package data

// mock_data_store_test.go 手写 dataStore 接口的测试替身。
// dataStore 是 data 包内未导出接口，mockgen 无法为其生成 mock（mockgen 仅支持
// 导出接口），故按 gomock 手工规范维护。MockDataStore 只被 data 包测试消费
// （k8s/repo/helm_cluster/k8s_stream 四个测试文件），按「包内独占 mock 放
// _test.go」约定存放，不参与生产编译。

import (
	"context"
	"reflect"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"go.uber.org/mock/gomock"
)

// MockDataStore is a mock of dataStore interface.
type MockDataStore struct {
	ctrl     *gomock.Controller
	recorder *MockDataStoreMockRecorder
}

// MockDataStoreMockRecorder is the mock recorder for MockDataStore.
type MockDataStoreMockRecorder struct {
	mock *MockDataStore
}

// NewMockDataStore creates a new mock instance.
func NewMockDataStore(ctrl *gomock.Controller) *MockDataStore {
	mock := &MockDataStore{ctrl: ctrl}
	mock.recorder = &MockDataStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDataStore) EXPECT() *MockDataStoreMockRecorder {
	return m.recorder
}

// Config mocks base method.
func (m *MockDataStore) Config() *config.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Config")
	ret0, _ := ret[0].(*config.Config)
	return ret0
}

// Config indicates an expected call of Config.
func (mr *MockDataStoreMockRecorder) Config() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockDataStore)(nil).Config))
}

// WithTx mocks base method.
func (m *MockDataStore) WithTx(arg0 context.Context, arg1 func(*ent.Tx) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTx", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// WithTx indicates an expected call of WithTx.
func (mr *MockDataStoreMockRecorder) WithTx(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTx", reflect.TypeOf((*MockDataStore)(nil).WithTx), arg0, arg1)
}

// DB mocks base method.
func (m *MockDataStore) DB() *ent.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DB")
	ret0, _ := ret[0].(*ent.Client)
	return ret0
}

// DB indicates an expected call of DB.
func (mr *MockDataStoreMockRecorder) DB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DB", reflect.TypeOf((*MockDataStore)(nil).DB))
}

// K8s mocks base method.
func (m *MockDataStore) K8s() *K8sClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "K8s")
	ret0, _ := ret[0].(*K8sClient)
	return ret0
}

// K8s indicates an expected call of K8s.
func (mr *MockDataStoreMockRecorder) K8s() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "K8s", reflect.TypeOf((*MockDataStore)(nil).K8s))
}
