package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// llmTestEmail 是 LLM 代理准入校验使用的登录用户邮箱。
const llmTestEmail = "alice@mars.local"

// llmProxyTestKit 组装 LLM 代理的 mock 依赖与路由。
type llmProxyTestKit struct {
	router   *mux.Router
	authBiz  *biz.MockAuthBiz
	repoBiz  *biz.MockRepoBiz
	nsBiz    *biz.MockNamespaceBiz
	projBiz  *biz.MockProjectBiz
	activity *biz.MockAgentActivity
	handler  *llmProxyHandler
}

// newLLMProxyRouter 构造挂了 LLM 代理路由的 kit，供用例设置鉴权/准入/活跃度期望。
func newLLMProxyRouter(t *testing.T, cfg *config.Config) *llmProxyTestKit {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	k := &llmProxyTestKit{
		authBiz:  biz.NewMockAuthBiz(ctrl),
		repoBiz:  biz.NewMockRepoBiz(ctrl),
		nsBiz:    biz.NewMockNamespaceBiz(ctrl),
		projBiz:  biz.NewMockProjectBiz(ctrl),
		activity: biz.NewMockAgentActivity(ctrl),
	}
	k.handler = newLLMProxyHandler(
		cfg, k.authBiz, k.repoBiz, k.nsBiz, k.projBiz, k.activity, mlog.NewForConfig(nil))
	k.router = mux.NewRouter()
	k.handler.Register(k.router)
	return k
}

// allowAuth 设置一次合法 token 的鉴权期望（VerifyToken + EffectiveRoles）。
func (k *llmProxyTestKit) allowAuth() {
	k.authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").
		Return(&biz.UserInfo{Name: "alice", Email: llmTestEmail}, nil)
	k.authBiz.EXPECT().EffectiveRoles(gomock.Any(), llmTestEmail, gomock.Any()).Return(nil, nil)
}

// allowAgent 预置「AI 仓库启用 + 空间存在 + 用户已部署 agent 项目」的准入链路，
// 并预置活跃度刷新；touchErr 用于覆盖刷新失败分支。
func (k *llmProxyTestKit) allowAgent(touchErr error) {
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(&biz.Repo{ID: 7, Enabled: true}, nil)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
		Return(&biz.Namespace{ID: 3}, nil)
	k.projBiz.EXPECT().FindByName(gomock.Any(), biz.AgentProjectName(llmTestEmail), 3).
		Return(&biz.Project{ID: 11}, nil)
	k.activity.EXPECT().Touch(gomock.Any(), biz.AgentProjectName(llmTestEmail)).Return(touchErr)
}

// do 发起一次携带合法 token 的代理请求。
func (k *llmProxyTestKit) do(t *testing.T, method, path string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	k.router.ServeHTTP(rec, req)
	return rec
}

// TestLLMProxy_ForwardsToUpstream 覆盖正常路径：路径前缀被剥离、query 保留、
// 沙箱令牌被替换为上游 key，响应原样回传。
func TestLLMProxy_ForwardsToUpstream(t *testing.T) {
	var gotPath, gotQuery, gotAuth, gotCookie, gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		gotCookie = r.Header.Get("Cookie")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{
		Model:   "deepseek-chat",
		BaseURL: upstream.URL + "/v1",
		APIKey:  "sk-deepseek",
	}})
	k.allowAuth()
	k.allowAgent(nil)

	req := httptest.NewRequest(http.MethodPost, llmProxyPath+"/chat/completions?x=1", strings.NewReader(`{"model":"deepseek-chat"}`))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Cookie", "mars_session=secret")
	rec := httptest.NewRecorder()
	k.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"ok":true}`, rec.Body.String())
	assert.Equal(t, "yes", rec.Header().Get("X-Upstream"))
	assert.Equal(t, "/v1/chat/completions", gotPath)
	assert.Equal(t, "x=1", gotQuery)
	assert.Equal(t, "Bearer sk-deepseek", gotAuth)
	assert.Empty(t, gotCookie, "mars 会话 Cookie 不得透传给模型服务")
	assert.Equal(t, `{"model":"deepseek-chat"}`, gotBody)
}

// TestLLMProxy_NoAPIKeyStripsAuthorization 覆盖上游无鉴权的场景：
// 不配置 api_key 时必须删除 Authorization，绝不把沙箱令牌透传给上游。
func TestLLMProxy_NoAPIKeyStripsAuthorization(t *testing.T) {
	var authHeader []string
	var gotCookie, gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Values("Authorization")
		gotCookie = r.Header.Get("Cookie")
		gotPath = r.URL.Path
	}))
	defer upstream.Close()

	// BaseURL 不带路径前缀：验证 joinURLPath 的空 base 分支。
	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: upstream.URL}})
	k.allowAuth()
	k.allowAgent(nil)

	req := httptest.NewRequest(http.MethodGet, llmProxyPath+"/models", nil)
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Cookie", "mars_session=secret")
	rec := httptest.NewRecorder()
	k.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, authHeader)
	assert.Empty(t, gotCookie, "mars 会话 Cookie 不得透传给模型服务")
	assert.Equal(t, "/models", gotPath)
}

// TestLLMProxy_StreamsSSE 覆盖流式响应：上游逐帧写出的 SSE 内容完整回传给客户端。
func TestLLMProxy_StreamsSSE(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("data: {\"delta\":\"你\"}\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer upstream.Close()

	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{
		BaseURL: upstream.URL + "/v1",
		APIKey:  "sk",
	}})
	k.allowAuth()
	k.allowAgent(nil)

	rec := k.do(t, http.MethodPost, llmProxyPath+"/chat/completions", nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "data: {\"delta\":\"你\"}")
	assert.Contains(t, rec.Body.String(), "data: [DONE]")
}

// TestLLMProxy_Unauthorized 覆盖鉴权失败路径：未携带/非法 token 返回 401，
// 请求不会到达上游。
func TestLLMProxy_Unauthorized(t *testing.T) {
	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: "http://127.0.0.1:1"}})
	k.authBiz.EXPECT().VerifyToken(gomock.Any(), "").Return(nil, errors.New("invalid token"))

	rec := httptest.NewRecorder()
	k.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, llmProxyPath+"/models", nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestLLMProxy_MissingEmail 鉴权通过但用户没有 email：无法派生 agent 项目，403。
func TestLLMProxy_MissingEmail(t *testing.T) {
	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: "http://127.0.0.1:1"}})
	k.authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(&biz.UserInfo{Name: "alice"}, nil)
	k.authBiz.EXPECT().EffectiveRoles(gomock.Any(), "", gomock.Any()).Return(nil, nil)

	rec := k.do(t, http.MethodGet, llmProxyPath+"/models", nil)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "email")
}

// TestLLMProxy_AgentNotReady 覆盖 AI 功能准入：各查询失败/未启用都拒绝转发。
func TestLLMProxy_AgentNotReady(t *testing.T) {
	notFound := errs.WrapNotFound(errors.New("not found"), "lookup")

	cases := map[string]struct {
		setup      func(*llmProxyTestKit)
		wantStatus int
		wantBody   string
	}{
		"未配置 AI 仓库": {
			setup: func(k *llmProxyTestKit) {
				k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(nil, notFound)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "尚未配置",
		},
		"仓库查询失败": {
			setup: func(k *llmProxyTestKit) {
				k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "查询 AI Agent 仓库失败",
		},
		"仓库已禁用": {
			setup: func(k *llmProxyTestKit) {
				k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(&biz.Repo{ID: 7}, nil)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "已禁用",
		},
		"空间不存在": {
			setup: func(k *llmProxyTestKit) {
				k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(&biz.Repo{ID: 7, Enabled: true}, nil)
				k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, notFound)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "空间尚未创建",
		},
		"空间查询失败": {
			setup: func(k *llmProxyTestKit) {
				k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(&biz.Repo{ID: 7, Enabled: true}, nil)
				k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "查询 AI Agent 空间失败",
		},
		"项目不存在": {
			setup: func(k *llmProxyTestKit) {
				k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(&biz.Repo{ID: 7, Enabled: true}, nil)
				k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
					Return(&biz.Namespace{ID: 3}, nil)
				k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).Return(nil, notFound)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "/api/ai/agent/enable",
		},
		"项目查询失败": {
			setup: func(k *llmProxyTestKit) {
				k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(&biz.Repo{ID: 7, Enabled: true}, nil)
				k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
					Return(&biz.Namespace{ID: 3}, nil)
				k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "查询 AI Agent 项目失败",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: "http://127.0.0.1:1"}})
			k.allowAuth()
			tc.setup(k)

			rec := k.do(t, http.MethodGet, llmProxyPath+"/models", nil)

			assert.Equal(t, tc.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.wantBody)
		})
	}
}

// TestLLMProxy_UnconfiguredUpstream 覆盖 ai.base_url 缺失/非法：准入通过后
// 返回 503，让 agent 拿到明确原因而不是连接错误。
func TestLLMProxy_UnconfiguredUpstream(t *testing.T) {
	cases := map[string]string{
		"空地址":    "",
		"缺少协议":   "api.deepseek.com/v1",
		"非法 URL": "http://[::1",
	}
	for name, baseURL := range cases {
		t.Run(name, func(t *testing.T) {
			k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: baseURL}})
			k.allowAuth()
			k.allowAgent(nil)

			rec := k.do(t, http.MethodGet, llmProxyPath+"/models", nil)

			assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
			assert.Contains(t, rec.Body.String(), "ai.base_url")
		})
	}
}

// TestLLMProxy_UpstreamUnreachable 覆盖上游不可达：返回 502 并记录错误。
func TestLLMProxy_UpstreamUnreachable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := upstream.URL
	upstream.Close()

	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: baseURL, APIKey: "sk"}})
	k.allowAuth()
	k.allowAgent(nil)

	rec := k.do(t, http.MethodGet, llmProxyPath+"/models", nil)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "LLM 上游不可用")
}

// TestLLMProxy_ClientCanceled 覆盖客户端主动断开：上游等待时取消请求，
// ErrorHandler 不写 502（避免把取消当服务端故障）。
func TestLLMProxy_ClientCanceled(t *testing.T) {
	started := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer upstream.Close()

	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: upstream.URL, APIKey: "sk"}})
	k.allowAuth()
	k.allowAgent(nil)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-started
		cancel()
	}()

	req := httptest.NewRequest(http.MethodGet, llmProxyPath+"/models", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	k.router.ServeHTTP(rec, req)

	assert.NotContains(t, rec.Body.String(), "LLM 上游不可用")
}

// TestLLMProxy_TouchesActivity 携带 email 的请求在转发前刷新 agent 活跃度。
func TestLLMProxy_TouchesActivity(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer upstream.Close()

	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: upstream.URL}})
	k.allowAuth()
	k.allowAgent(nil)

	rec := k.do(t, http.MethodGet, llmProxyPath+"/models", nil)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestLLMProxy_TouchError 活跃度刷新失败只记日志，不阻断 LLM 代理。
func TestLLMProxy_TouchError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer upstream.Close()

	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: upstream.URL}})
	k.allowAuth()
	k.allowAgent(errors.New("cache down"))

	rec := k.do(t, http.MethodGet, llmProxyPath+"/models", nil)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestLLMProxy_GuardWithoutUser 直接调用 guard（不经过鉴权中间件）：
// 拿不到用户一律 403，且不触达准入依赖。
func TestLLMProxy_GuardWithoutUser(t *testing.T) {
	k := newLLMProxyRouter(t, &config.Config{AI: config.AIConfig{BaseURL: "http://127.0.0.1:1"}})
	called := false
	h := k.handler.guard(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, llmProxyPath+"/models", nil))

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, called, "未注入用户的请求不得进入代理")
}

// TestJoinURLPath 覆盖路径拼接：空 base、带/不带斜杠的组合都只有一个分隔符。
func TestJoinURLPath(t *testing.T) {
	assert.Equal(t, "/v1/chat/completions", joinURLPath("/v1", "/chat/completions"))
	assert.Equal(t, "/v1/chat/completions", joinURLPath("/v1/", "/chat/completions"))
	assert.Equal(t, "/chat/completions", joinURLPath("", "/chat/completions"))
}
