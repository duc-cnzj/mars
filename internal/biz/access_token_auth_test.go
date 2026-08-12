package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

// fakeAccessTokenBizForManager 记录 FindByToken/TouchLastUsedAt 调用并返回罐头数据/错误。
type fakeAccessTokenBizForManager struct {
	AccessTokenBiz
	findErr, touchErr       error
	findCalled, touchCalled bool
	user                    *UserInfo
}

func (f *fakeAccessTokenBizForManager) FindByToken(ctx context.Context, token string) (*AccessToken, error) {
	f.findCalled = true
	if f.findErr != nil {
		return nil, f.findErr
	}
	return &AccessToken{Token: token, UserInfo: *f.user}, nil
}

func (f *fakeAccessTokenBizForManager) TouchLastUsedAt(ctx context.Context, token string, t time.Time) error {
	f.touchCalled = true
	return f.touchErr
}

func newTokenManagerForTest(biz AccessTokenBiz) TokenManager {
	return NewAccessTokenManager(biz, timer.NewReal(), mlog.NewForConfig(nil))
}

func TestAccessTokenManager_VerifyAndTouch_Valid(t *testing.T) {
	f := &fakeAccessTokenBizForManager{user: &UserInfo{Name: "duc", Email: "duc@x.io"}}
	m := newTokenManagerForTest(f)
	now := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	got, ok := m.VerifyAndTouch(context.TODO(), "tok", now)
	assert.True(t, ok)
	assert.True(t, f.findCalled)
	assert.True(t, f.touchCalled)
	assert.Equal(t, "duc", got.Name)
}

func TestAccessTokenManager_VerifyAndTouch_FindError(t *testing.T) {
	f := &fakeAccessTokenBizForManager{findErr: errors.New("db down"), user: &UserInfo{}}
	m := newTokenManagerForTest(f)
	got, ok := m.VerifyAndTouch(context.TODO(), "tok", time.Now())
	assert.False(t, ok)
	assert.True(t, f.findCalled)
	assert.False(t, f.touchCalled)
	assert.Nil(t, got)
}

func TestAccessTokenManager_VerifyAndTouch_TouchError(t *testing.T) {
	// TouchLastUsedAt 失败只告警不回退，仍返回用户。
	f := &fakeAccessTokenBizForManager{touchErr: errors.New("touch down"), user: &UserInfo{Name: "duc"}}
	m := newTokenManagerForTest(f)
	got, ok := m.VerifyAndTouch(context.TODO(), "tok", time.Now())
	assert.True(t, ok)
	assert.True(t, f.findCalled)
	assert.True(t, f.touchCalled)
	assert.Equal(t, "duc", got.Name)
}
