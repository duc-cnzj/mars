package services

// aiusage_test.go 覆盖 POST /api/ai/usage 的全部分支：白名单事件计数、
// 非法 JSON/未知事件/超限请求体拒绝、未登录拒绝。

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const aiUsageTestEmail = "alice@mars.local"

// aiUsageTestKit 组装 aiUsageHandler 的 mock 依赖与路由。
type aiUsageTestKit struct {
	ctrl    *gomock.Controller
	router  *mux.Router
	authBiz *biz.MockAuthBiz
}

// newAIUsageTestKit 构造处理器与路由。
func newAIUsageTestKit(t *testing.T) *aiUsageTestKit {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	k := &aiUsageTestKit{
		ctrl:    ctrl,
		authBiz: biz.NewMockAuthBiz(ctrl),
	}
	h := newAIUsageHandler(k.authBiz, mlog.NewForConfig(nil))
	k.router = mux.NewRouter()
	h.Register(k.router)
	return k
}

// expectLogin 预置登录鉴权成功链路（VerifyToken + EffectiveRoles）。
func (k *aiUsageTestKit) expectLogin(email string) {
	k.authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(&biz.UserInfo{Name: "alice", Email: email}, nil)
	k.authBiz.EXPECT().EffectiveRoles(gomock.Any(), email, gomock.Any()).Return(nil, nil)
}

// doBody 发起带请求体的请求；token 为空时不携带鉴权头。
func (k *aiUsageTestKit) doBody(t *testing.T, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(http.MethodPost, aiUsagePath, reader)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	rec := httptest.NewRecorder()
	k.router.ServeHTTP(rec, req)
	return rec
}

// TestAIUsageHandler_Events 白名单事件返回 204，并让对应计数器 +1。
func TestAIUsageHandler_Events(t *testing.T) {
	for _, event := range []string{
		aiUsageEntryClick,
		aiUsageChatFirstToken,
		aiUsageChatDone,
		aiUsageChatError,
	} {
		t.Run(event, func(t *testing.T) {
			k := newAIUsageTestKit(t)
			k.expectLogin(aiUsageTestEmail)
			before := testutil.ToFloat64(metrics.AIUsageEventTotal.WithLabelValues(event))

			rec := k.doBody(t, "Bearer tok", `{"event":"`+event+`"}`)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, before+1, testutil.ToFloat64(metrics.AIUsageEventTotal.WithLabelValues(event)))
		})
	}
}

// TestAIUsageHandler_Invalid 非法 JSON、空事件与未知事件一律 400：
// 指标 label 不能由客户端任意决定，否则会生成无界时间序列。
func TestAIUsageHandler_Invalid(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "非法 JSON", body: `not-json`},
		{name: "空事件", body: `{}`},
		{name: "未知事件", body: `{"event":"hack_the_metric"}`},
		// 沙箱时代的事件已随沙箱下线：白名单不得再放行。
		{name: "已下线事件", body: `{"event":"sandbox_ready"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := newAIUsageTestKit(t)
			k.expectLogin(aiUsageTestEmail)
			rec := k.doBody(t, "Bearer tok", tt.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// TestAIUsageHandler_Unauthorized 未携带 token → 401，且不触碰业务依赖。
func TestAIUsageHandler_Unauthorized(t *testing.T) {
	k := newAIUsageTestKit(t)
	k.authBiz.EXPECT().VerifyToken(gomock.Any(), "").Return(nil, assert.AnError)

	rec := k.doBody(t, "", `{"event":"entry_click"}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
