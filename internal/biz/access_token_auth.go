package biz

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

type accessTokenManager struct {
	accessTokenBiz AccessTokenBiz
	timer          timer.Timer
	logger         mlog.Logger
}

// NewAccessTokenManager 构造 access token 管理器（TokenManager 实现）。
func NewAccessTokenManager(accessTokenBiz AccessTokenBiz, timer timer.Timer, logger mlog.Logger) TokenManager {
	return &accessTokenManager{accessTokenBiz: accessTokenBiz, timer: timer, logger: logger}
}

// VerifyAndTouch 按 token 查询 access token 并回写最近使用时间，返回对应用户信息；
// token 不存在时返回 false。
func (m *accessTokenManager) VerifyAndTouch(ctx context.Context, token string, now time.Time) (*UserInfo, bool) {
	at, err := m.accessTokenBiz.FindByToken(ctx, token)
	if err != nil {
		return nil, false
	}
	if err := m.accessTokenBiz.TouchLastUsedAt(ctx, token, now); err != nil {
		m.logger.Warningf("failed to touch last used at for token: %v", err)
	}
	return &at.UserInfo, true
}
