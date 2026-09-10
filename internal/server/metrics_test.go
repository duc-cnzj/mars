package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestNewMetricsRunner 覆盖构造器：WithModule + 端口 + server 装配（真实 *http.Server，
// 含 metricsHandler 构建的处理器）。
func TestNewMetricsRunner(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockLogger := mlog.NewMockLogger(m)
	mockLogger.EXPECT().WithModule("server/metricsRunner").Return(mockLogger).Times(1)

	r := NewMetricsRunner("9999", mockLogger, prometheus.NewRegistry()).(*metricsRunner)
	assert.Equal(t, "9999", r.port)
	hs, ok := r.s.(*http.Server)
	assert.True(t, ok)
	assert.Equal(t, ":9999", hs.Addr)
	assert.NotNil(t, hs.Handler)
}

// TestMetricsRunnerRunAndShutdown 覆盖 Run/Shutdown 生命周期：注入 mock server，
// channel 同步等 ListenAndServe 真正执行后再 Shutdown，不绑真实端口、无 time.Sleep。
func TestMetricsRunnerRunAndShutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewMockLogger(m)
	server := NewMockHttpServer(m)
	runner := &metricsRunner{port: "8080", logger: mockLogger, s: server}

	mockLogger.EXPECT().Infof("[Server]: metrics running at :%s/metrics", "8080").Times(1)
	started := make(chan struct{})
	server.EXPECT().ListenAndServe().DoAndReturn(func() error {
		close(started)
		return nil
	}).Times(1)
	server.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)

	assert.Nil(t, runner.Run(context.TODO()))
	<-started
	assert.Nil(t, runner.Shutdown(context.TODO()))
}

// TestMetricsRunnerRunError 覆盖 Run 的失败路径：ListenAndServe 出错时记录 Error 日志
// （此前静默吞错，端口冲突时指标服务静默下线无任何日志——本测试为其兜底，删除错误
// 处理回归裸 `m.s.ListenAndServe()` 时 Error EXPECT 缺失即炸）。
func TestMetricsRunnerRunError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewMockLogger(m)
	server := NewMockHttpServer(m)
	runner := &metricsRunner{port: "8080", logger: mockLogger, s: server}

	mockLogger.EXPECT().Infof("[Server]: metrics running at :%s/metrics", "8080").Times(1)
	done := make(chan struct{})
	mockLogger.EXPECT().Error(gomock.Any()).Do(func(_ ...any) { close(done) }).Times(1)
	server.EXPECT().ListenAndServe().Return(assert.AnError).Times(1)

	assert.Nil(t, runner.Run(context.TODO()))
	// 超时守卫而非裸 <-done：变异（错误处理被删）时 Error 永不调用，裸等会挂死整个套件。
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("metrics runner 应记录 ListenAndServe 错误日志")
	}
}

// TestMetricsRunnerShutdownError 覆盖 Shutdown 的失败路径：底层 server 关闭出错时上抛。
func TestMetricsRunnerShutdownError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	server := NewMockHttpServer(m)
	server.EXPECT().Shutdown(gomock.Any()).Return(assert.AnError).Times(1)
	runner := &metricsRunner{s: server}
	assert.Error(t, runner.Shutdown(context.TODO()))
}

// Test_metricsHandler 直测 /metrics 端点行为（不绑真实端口）：注册测试指标后
// GET /metrics 返回 200 且响应体含指标名，未注册路径 404。
func Test_metricsHandler(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGauge(prometheus.GaugeOpts{Name: "audit_test_metric", Help: "test metric"}))
	h := metricsHandler(reg)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/metrics", nil))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "audit_test_metric")

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/other", nil))
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
