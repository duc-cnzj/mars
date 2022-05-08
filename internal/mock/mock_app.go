// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/duc-cnzj/mars/internal/contracts (interfaces: ApplicationInterface)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	config "github.com/duc-cnzj/mars/internal/config"
	contracts "github.com/duc-cnzj/mars/internal/contracts"
	singleflight "github.com/duc-cnzj/mars/internal/utils/singleflight"
	gomock "github.com/golang/mock/gomock"
)

// MockApplicationInterface is a mock of ApplicationInterface interface.
type MockApplicationInterface struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationInterfaceMockRecorder
}

// MockApplicationInterfaceMockRecorder is the mock recorder for MockApplicationInterface.
type MockApplicationInterfaceMockRecorder struct {
	mock *MockApplicationInterface
}

// NewMockApplicationInterface creates a new mock instance.
func NewMockApplicationInterface(ctrl *gomock.Controller) *MockApplicationInterface {
	mock := &MockApplicationInterface{ctrl: ctrl}
	mock.recorder = &MockApplicationInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApplicationInterface) EXPECT() *MockApplicationInterfaceMockRecorder {
	return m.recorder
}

// AddServer mocks base method.
func (m *MockApplicationInterface) AddServer(arg0 contracts.Server) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddServer", arg0)
}

// AddServer indicates an expected call of AddServer.
func (mr *MockApplicationInterfaceMockRecorder) AddServer(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddServer", reflect.TypeOf((*MockApplicationInterface)(nil).AddServer), arg0)
}

// Auth mocks base method.
func (m *MockApplicationInterface) Auth() contracts.AuthInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Auth")
	ret0, _ := ret[0].(contracts.AuthInterface)
	return ret0
}

// Auth indicates an expected call of Auth.
func (mr *MockApplicationInterfaceMockRecorder) Auth() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Auth", reflect.TypeOf((*MockApplicationInterface)(nil).Auth))
}

// BeforeServerRunHooks mocks base method.
func (m *MockApplicationInterface) BeforeServerRunHooks(arg0 contracts.Callback) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "BeforeServerRunHooks", arg0)
}

// BeforeServerRunHooks indicates an expected call of BeforeServerRunHooks.
func (mr *MockApplicationInterfaceMockRecorder) BeforeServerRunHooks(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeforeServerRunHooks", reflect.TypeOf((*MockApplicationInterface)(nil).BeforeServerRunHooks), arg0)
}

// Bootstrap mocks base method.
func (m *MockApplicationInterface) Bootstrap() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Bootstrap")
	ret0, _ := ret[0].(error)
	return ret0
}

// Bootstrap indicates an expected call of Bootstrap.
func (mr *MockApplicationInterfaceMockRecorder) Bootstrap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bootstrap", reflect.TypeOf((*MockApplicationInterface)(nil).Bootstrap))
}

// Cache mocks base method.
func (m *MockApplicationInterface) Cache() contracts.CacheInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Cache")
	ret0, _ := ret[0].(contracts.CacheInterface)
	return ret0
}

// Cache indicates an expected call of Cache.
func (mr *MockApplicationInterfaceMockRecorder) Cache() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Cache", reflect.TypeOf((*MockApplicationInterface)(nil).Cache))
}

// Config mocks base method.
func (m *MockApplicationInterface) Config() *config.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Config")
	ret0, _ := ret[0].(*config.Config)
	return ret0
}

// Config indicates an expected call of Config.
func (mr *MockApplicationInterfaceMockRecorder) Config() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockApplicationInterface)(nil).Config))
}

// DBManager mocks base method.
func (m *MockApplicationInterface) DBManager() contracts.DBManager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DBManager")
	ret0, _ := ret[0].(contracts.DBManager)
	return ret0
}

// DBManager indicates an expected call of DBManager.
func (mr *MockApplicationInterfaceMockRecorder) DBManager() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DBManager", reflect.TypeOf((*MockApplicationInterface)(nil).DBManager))
}

// Done mocks base method.
func (m *MockApplicationInterface) Done() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Done indicates an expected call of Done.
func (mr *MockApplicationInterfaceMockRecorder) Done() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Done", reflect.TypeOf((*MockApplicationInterface)(nil).Done))
}

// EventDispatcher mocks base method.
func (m *MockApplicationInterface) EventDispatcher() contracts.DispatcherInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventDispatcher")
	ret0, _ := ret[0].(contracts.DispatcherInterface)
	return ret0
}

// EventDispatcher indicates an expected call of EventDispatcher.
func (mr *MockApplicationInterfaceMockRecorder) EventDispatcher() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventDispatcher", reflect.TypeOf((*MockApplicationInterface)(nil).EventDispatcher))
}

// GetPluginByName mocks base method.
func (m *MockApplicationInterface) GetPluginByName(arg0 string) contracts.PluginInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPluginByName", arg0)
	ret0, _ := ret[0].(contracts.PluginInterface)
	return ret0
}

// GetPluginByName indicates an expected call of GetPluginByName.
func (mr *MockApplicationInterfaceMockRecorder) GetPluginByName(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPluginByName", reflect.TypeOf((*MockApplicationInterface)(nil).GetPluginByName), arg0)
}

// GetPlugins mocks base method.
func (m *MockApplicationInterface) GetPlugins() map[string]contracts.PluginInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPlugins")
	ret0, _ := ret[0].(map[string]contracts.PluginInterface)
	return ret0
}

// GetPlugins indicates an expected call of GetPlugins.
func (mr *MockApplicationInterfaceMockRecorder) GetPlugins() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPlugins", reflect.TypeOf((*MockApplicationInterface)(nil).GetPlugins))
}

// IsDebug mocks base method.
func (m *MockApplicationInterface) IsDebug() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDebug")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDebug indicates an expected call of IsDebug.
func (mr *MockApplicationInterfaceMockRecorder) IsDebug() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDebug", reflect.TypeOf((*MockApplicationInterface)(nil).IsDebug))
}

// K8sClient mocks base method.
func (m *MockApplicationInterface) K8sClient() *contracts.K8sClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "K8sClient")
	ret0, _ := ret[0].(*contracts.K8sClient)
	return ret0
}

// K8sClient indicates an expected call of K8sClient.
func (mr *MockApplicationInterfaceMockRecorder) K8sClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "K8sClient", reflect.TypeOf((*MockApplicationInterface)(nil).K8sClient))
}

// Metrics mocks base method.
func (m *MockApplicationInterface) Metrics() contracts.Metrics {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Metrics")
	ret0, _ := ret[0].(contracts.Metrics)
	return ret0
}

// Metrics indicates an expected call of Metrics.
func (mr *MockApplicationInterfaceMockRecorder) Metrics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Metrics", reflect.TypeOf((*MockApplicationInterface)(nil).Metrics))
}

// Oidc mocks base method.
func (m *MockApplicationInterface) Oidc() contracts.OidcConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Oidc")
	ret0, _ := ret[0].(contracts.OidcConfig)
	return ret0
}

// Oidc indicates an expected call of Oidc.
func (mr *MockApplicationInterfaceMockRecorder) Oidc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Oidc", reflect.TypeOf((*MockApplicationInterface)(nil).Oidc))
}

// RegisterAfterShutdownFunc mocks base method.
func (m *MockApplicationInterface) RegisterAfterShutdownFunc(arg0 contracts.Callback) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterAfterShutdownFunc", arg0)
}

// RegisterAfterShutdownFunc indicates an expected call of RegisterAfterShutdownFunc.
func (mr *MockApplicationInterfaceMockRecorder) RegisterAfterShutdownFunc(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterAfterShutdownFunc", reflect.TypeOf((*MockApplicationInterface)(nil).RegisterAfterShutdownFunc), arg0)
}

// RegisterBeforeShutdownFunc mocks base method.
func (m *MockApplicationInterface) RegisterBeforeShutdownFunc(arg0 contracts.Callback) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterBeforeShutdownFunc", arg0)
}

// RegisterBeforeShutdownFunc indicates an expected call of RegisterBeforeShutdownFunc.
func (mr *MockApplicationInterfaceMockRecorder) RegisterBeforeShutdownFunc(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterBeforeShutdownFunc", reflect.TypeOf((*MockApplicationInterface)(nil).RegisterBeforeShutdownFunc), arg0)
}

// Run mocks base method.
func (m *MockApplicationInterface) Run() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockApplicationInterfaceMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockApplicationInterface)(nil).Run))
}

// SetAuth mocks base method.
func (m *MockApplicationInterface) SetAuth(arg0 contracts.AuthInterface) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAuth", arg0)
}

// SetAuth indicates an expected call of SetAuth.
func (mr *MockApplicationInterfaceMockRecorder) SetAuth(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAuth", reflect.TypeOf((*MockApplicationInterface)(nil).SetAuth), arg0)
}

// SetCache mocks base method.
func (m *MockApplicationInterface) SetCache(arg0 contracts.CacheInterface) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCache", arg0)
}

// SetCache indicates an expected call of SetCache.
func (mr *MockApplicationInterfaceMockRecorder) SetCache(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCache", reflect.TypeOf((*MockApplicationInterface)(nil).SetCache), arg0)
}

// SetEventDispatcher mocks base method.
func (m *MockApplicationInterface) SetEventDispatcher(arg0 contracts.DispatcherInterface) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetEventDispatcher", arg0)
}

// SetEventDispatcher indicates an expected call of SetEventDispatcher.
func (mr *MockApplicationInterfaceMockRecorder) SetEventDispatcher(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetEventDispatcher", reflect.TypeOf((*MockApplicationInterface)(nil).SetEventDispatcher), arg0)
}

// SetK8sClient mocks base method.
func (m *MockApplicationInterface) SetK8sClient(arg0 *contracts.K8sClient) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetK8sClient", arg0)
}

// SetK8sClient indicates an expected call of SetK8sClient.
func (mr *MockApplicationInterfaceMockRecorder) SetK8sClient(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetK8sClient", reflect.TypeOf((*MockApplicationInterface)(nil).SetK8sClient), arg0)
}

// SetMetrics mocks base method.
func (m *MockApplicationInterface) SetMetrics(arg0 contracts.Metrics) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetMetrics", arg0)
}

// SetMetrics indicates an expected call of SetMetrics.
func (mr *MockApplicationInterfaceMockRecorder) SetMetrics(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMetrics", reflect.TypeOf((*MockApplicationInterface)(nil).SetMetrics), arg0)
}

// SetOidc mocks base method.
func (m *MockApplicationInterface) SetOidc(arg0 contracts.OidcConfig) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetOidc", arg0)
}

// SetOidc indicates an expected call of SetOidc.
func (mr *MockApplicationInterfaceMockRecorder) SetOidc(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetOidc", reflect.TypeOf((*MockApplicationInterface)(nil).SetOidc), arg0)
}

// SetPlugins mocks base method.
func (m *MockApplicationInterface) SetPlugins(arg0 map[string]contracts.PluginInterface) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPlugins", arg0)
}

// SetPlugins indicates an expected call of SetPlugins.
func (mr *MockApplicationInterfaceMockRecorder) SetPlugins(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPlugins", reflect.TypeOf((*MockApplicationInterface)(nil).SetPlugins), arg0)
}

// SetUploader mocks base method.
func (m *MockApplicationInterface) SetUploader(arg0 contracts.Uploader) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUploader", arg0)
}

// SetUploader indicates an expected call of SetUploader.
func (mr *MockApplicationInterfaceMockRecorder) SetUploader(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUploader", reflect.TypeOf((*MockApplicationInterface)(nil).SetUploader), arg0)
}

// Shutdown mocks base method.
func (m *MockApplicationInterface) Shutdown() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockApplicationInterfaceMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockApplicationInterface)(nil).Shutdown))
}

// Singleflight mocks base method.
func (m *MockApplicationInterface) Singleflight() *singleflight.Group {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Singleflight")
	ret0, _ := ret[0].(*singleflight.Group)
	return ret0
}

// Singleflight indicates an expected call of Singleflight.
func (mr *MockApplicationInterfaceMockRecorder) Singleflight() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Singleflight", reflect.TypeOf((*MockApplicationInterface)(nil).Singleflight))
}

// Uploader mocks base method.
func (m *MockApplicationInterface) Uploader() contracts.Uploader {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Uploader")
	ret0, _ := ret[0].(contracts.Uploader)
	return ret0
}

// Uploader indicates an expected call of Uploader.
func (mr *MockApplicationInterfaceMockRecorder) Uploader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Uploader", reflect.TypeOf((*MockApplicationInterface)(nil).Uploader))
}
