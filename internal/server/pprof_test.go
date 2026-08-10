package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

// TestNewPprofRunner 覆盖构造器：WithModule + server 装配（真实 *http.Server，
// 含 pprofMux 构建的处理器）。
func TestNewPprofRunner(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockLogger := mlog.NewMockLogger(m)
	mockLogger.EXPECT().WithModule("server/pprofRunner").Return(mockLogger).Times(1)

	r := NewPprofRunner(mockLogger).(*pprofRunner)
	hs, ok := r.server.(*http.Server)
	assert.True(t, ok)
	assert.Equal(t, "localhost:6060", hs.Addr)
	assert.NotNil(t, hs.Handler)
}

// TestPprofRunnerRunAndShutdown 覆盖 Run/Shutdown 生命周期：注入 mock server，
// channel 同步等 ListenAndServe 真正执行后再 Shutdown，不绑真实端口、无 time.Sleep。
func TestPprofRunnerRunAndShutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewMockLogger(m)
	server := NewMockHttpServer(m)
	runner := &pprofRunner{logger: mockLogger, server: server}

	mockLogger.EXPECT().Info("[Server]: start pprofRunner runner.").Times(1)
	mockLogger.EXPECT().Info("Starting pprof server on localhost:6060.").Times(1)
	started := make(chan struct{})
	server.EXPECT().ListenAndServe().DoAndReturn(func() error {
		close(started)
		return nil
	}).Times(1)
	mockLogger.EXPECT().Info("[Server]: shutdown pprofRunner runner.").Times(1)
	server.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)

	assert.Nil(t, runner.Run(context.TODO()))
	<-started
	assert.Nil(t, runner.Shutdown(context.TODO()))
}

// TestPprofRunnerRunError 覆盖 Run 的失败路径：ListenAndServe 出错时记录 Error 日志。
func TestPprofRunnerRunError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockLogger := mlog.NewMockLogger(m)
	server := NewMockHttpServer(m)
	runner := &pprofRunner{logger: mockLogger, server: server}

	mockLogger.EXPECT().Info("[Server]: start pprofRunner runner.").Times(1)
	mockLogger.EXPECT().Info("Starting pprof server on localhost:6060.").Times(1)
	done := make(chan struct{})
	mockLogger.EXPECT().Error(gomock.Any()).Do(func(_ ...any) { close(done) }).Times(1)
	server.EXPECT().ListenAndServe().Return(assert.AnError).Times(1)

	assert.Nil(t, runner.Run(context.TODO()))
	// 超时守卫而非裸 <-done：变异（错误处理被删）时 Error 永不调用，裸等会挂死整个套件。
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("pprof runner 应记录 ListenAndServe 错误日志")
	}
}

// Test_pprofRunner_Shutdown 覆盖 Shutdown 的正常路径：底层 server 关闭成功。
func Test_pprofRunner_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockLogger := mlog.NewMockLogger(m)
	server := NewMockHttpServer(m)
	runner := &pprofRunner{logger: mockLogger, server: server}
	mockLogger.EXPECT().Info("[Server]: shutdown pprofRunner runner.").Times(1)
	server.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)
	assert.Nil(t, runner.Shutdown(context.TODO()))
}

// Test_pprofMux 直测 /debug/pprof 端点行为（不绑真实端口）：cmdline 返回 200 + 非空，
// 索引页 200，未注册路径 404。
func Test_pprofMux(t *testing.T) {
	h := pprofMux()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/debug/pprof/cmdline", nil))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Body.String())

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/debug/pprof/", nil))
	assert.Equal(t, http.StatusOK, rr.Code)

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/nope", nil))
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
