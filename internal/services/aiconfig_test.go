package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// newAIConfigRouter 构造挂了 /api/ai/config 的路由，并返回鉴权/活跃度 mock 供用例设置期望。
func newAIConfigRouter(t *testing.T, cfg *config.Config) (*mux.Router, *biz.MockAuthBiz, *biz.MockAgentActivity) {
	t.Helper()
	ctrl := gomock.NewController(t)
	authBiz := biz.NewMockAuthBiz(ctrl)
	activity := biz.NewMockAgentActivity(ctrl)
	router := mux.NewRouter()
	newAIConfigHandler(cfg, authBiz, activity, mlog.NewForConfig(nil)).Register(router)
	return router, authBiz, activity
}

// TestAIConfigHandler_ReturnsModel 覆盖正常路径：携带合法 token 时回吐 JSON 模型配置。
func TestAIConfigHandler_ReturnsModel(t *testing.T) {
	router, authBiz, _ := newAIConfigRouter(t, &config.Config{AI: config.AIConfig{Model: "qwen-max"}})
	authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(&biz.UserInfo{Name: "alice"}, nil)
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"mars_admin"}, nil)

	req := httptest.NewRequest(http.MethodGet, aiConfigPath, nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"model":"qwen-max"}`, rec.Body.String())
}

// TestAIConfigHandler_Unauthorized 覆盖鉴权失败路径：未携带/非法 token 返回 401。
func TestAIConfigHandler_Unauthorized(t *testing.T) {
	router, authBiz, _ := newAIConfigRouter(t, &config.Config{})
	authBiz.EXPECT().VerifyToken(gomock.Any(), "").Return(nil, errors.New("invalid token"))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, aiConfigPath, nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestAIConfigHandler_UnconfiguredModel 覆盖配置缺失路径：ai.model 为空时返回 503，
// 不把空模型下发给 agent。
func TestAIConfigHandler_UnconfiguredModel(t *testing.T) {
	router, authBiz, _ := newAIConfigRouter(t, &config.Config{})
	authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(&biz.UserInfo{Name: "alice"}, nil)
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, aiConfigPath, nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "ai.model")
}

// TestAIConfigHandler_TouchesActivity 已登录用户拉配置时按邮箱派生的 agent 项目刷新活跃度。
func TestAIConfigHandler_TouchesActivity(t *testing.T) {
	router, authBiz, activity := newAIConfigRouter(t, &config.Config{AI: config.AIConfig{Model: "m"}})
	authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(&biz.UserInfo{Name: "alice", Email: "alice@mars.local"}, nil)
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), "alice@mars.local", gomock.Any()).Return(nil, nil)
	activity.EXPECT().Touch(gomock.Any(), biz.AgentProjectName("alice@mars.local")).Return(nil)

	req := httptest.NewRequest(http.MethodGet, aiConfigPath, nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestAIConfigHandler_TouchError 活跃度刷新失败只记日志，不阻断配置下发。
func TestAIConfigHandler_TouchError(t *testing.T) {
	router, authBiz, activity := newAIConfigRouter(t, &config.Config{AI: config.AIConfig{Model: "m"}})
	authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(&biz.UserInfo{Email: "alice@mars.local"}, nil)
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), "alice@mars.local", gomock.Any()).Return(nil, nil)
	activity.EXPECT().Touch(gomock.Any(), gomock.Any()).Return(errors.New("cache down"))

	req := httptest.NewRequest(http.MethodGet, aiConfigPath, nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"model":"m"}`, rec.Body.String())
}

// TestAIConfigHandler_MethodNotAllowed 覆盖方法限制：仅允许 GET，其余方法 405
// （且在路由层就被拒绝，不进入鉴权）。
func TestAIConfigHandler_MethodNotAllowed(t *testing.T) {
	router, _, _ := newAIConfigRouter(t, &config.Config{AI: config.AIConfig{Model: "m"}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, aiConfigPath, nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// errResponseWriter 是写响应必然失败的 ResponseWriter 替身，
// 用于确定性覆盖 json.Encode 的写失败分支。
type errResponseWriter struct {
	header http.Header
}

// Header 返回空 header 容器。
func (w *errResponseWriter) Header() http.Header { return w.header }

// Write 恒定失败，模拟客户端在响应写出前断开。
func (w *errResponseWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

// WriteHeader 是 no-op：失败发生在编码写出阶段。
func (w *errResponseWriter) WriteHeader(int) {}

// TestAIConfigHandler_EncodeError 覆盖写出失败的日志分支，确保不会 panic。
func TestAIConfigHandler_EncodeError(t *testing.T) {
	h := &aiConfigHandler{
		cfg:    &config.Config{AI: config.AIConfig{Model: "m"}},
		logger: mlog.NewForConfig(nil),
	}

	assert.NotPanics(t, func() {
		h.serve(&errResponseWriter{header: http.Header{}}, httptest.NewRequest(http.MethodGet, aiConfigPath, nil))
	})
}
