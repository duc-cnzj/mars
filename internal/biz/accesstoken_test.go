package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeAccessTokenRepoForAccessTokenBiz 记录各写操作是否被调用，输入校验测试中 repo 不被调用（调用即 panic）。
// findErr/expired 控制 FindByToken 的返回，用于覆盖 Lease 的错误与过期分支。
// now 是注入的时钟基准，零值时回退 time.Now()（旧测试兼容）。
type fakeAccessTokenRepoForAccessTokenBiz struct {
	AccessTokenRepo
	grantCalled, leaseFindCalled, leaseUpdateCalled bool
	revokeCalled, touchCalled                       bool
	listCalled                                      bool
	findErr                                         error
	expired                                         bool
	now                                             time.Time
}

func (f *fakeAccessTokenRepoForAccessTokenBiz) Grant(ctx context.Context, input *GrantAccessTokenInput) (*AccessToken, error) {
	f.grantCalled = true
	return &AccessToken{ID: 1, Usage: input.Usage}, nil
}

func (f *fakeAccessTokenRepoForAccessTokenBiz) List(ctx context.Context, input *ListAccessTokenInput) ([]*AccessToken, *pagination.Pagination, error) {
	f.listCalled = true
	return []*AccessToken{{ID: 1}}, nil, nil
}

func (f *fakeAccessTokenRepoForAccessTokenBiz) FindByToken(ctx context.Context, token string) (*AccessToken, error) {
	f.leaseFindCalled = true
	if f.findErr != nil {
		return nil, f.findErr
	}
	base := f.now
	if base.IsZero() {
		base = time.Now()
	}
	expiredAt := base.Add(time.Hour)
	if f.expired {
		expiredAt = base.Add(-time.Hour)
	}
	return &AccessToken{ID: 1, Token: token, ExpiredAt: expiredAt}, nil
}

func (f *fakeAccessTokenRepoForAccessTokenBiz) UpdateExpiresAt(ctx context.Context, token string, t time.Time) (*AccessToken, error) {
	f.leaseUpdateCalled = true
	return &AccessToken{ID: 1, Token: token, ExpiredAt: t}, nil
}

func (f *fakeAccessTokenRepoForAccessTokenBiz) Revoke(ctx context.Context, token string) error {
	f.revokeCalled = true
	return nil
}

func (f *fakeAccessTokenRepoForAccessTokenBiz) TouchLastUsedAt(ctx context.Context, token string, t time.Time) error {
	f.touchCalled = true
	return nil
}

func newAccessTokenBizForTest(repo AccessTokenRepo) AccessTokenBiz {
	return NewAccessTokenBiz(mlog.NewForConfig(nil), timer.NewReal(), repo)
}

// fixedClock 是注入固定时刻的 timer.Timer 实现，使过期判定测试确定性化、无 time.Now 竞态。
type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func (f fixedClock) Since(t time.Time) time.Duration { return f.now.Sub(t) }

// newAccessTokenBizWithClock 构造注入固定时钟的 access token biz。
func newAccessTokenBizWithClock(now time.Time, repo AccessTokenRepo) AccessTokenBiz {
	return NewAccessTokenBiz(mlog.NewForConfig(nil), fixedClock{now: now}, repo)
}

func TestAccessTokenBiz_Grant_NilInput(t *testing.T) {
	b := newAccessTokenBizForTest(&fakeAccessTokenRepoForAccessTokenBiz{})
	got, err := b.Grant(context.TODO(), nil)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "grant input 或 user 不能为空", status.Convert(err).Message())
}

func TestAccessTokenBiz_Grant_NilUser(t *testing.T) {
	b := newAccessTokenBizForTest(&fakeAccessTokenRepoForAccessTokenBiz{})
	got, err := b.Grant(context.TODO(), &GrantAccessTokenInput{User: nil, ExpireSeconds: 3600})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "grant input 或 user 不能为空", status.Convert(err).Message())
}

func TestAccessTokenBiz_Grant_InvalidExpireSeconds(t *testing.T) {
	b := newAccessTokenBizForTest(&fakeAccessTokenRepoForAccessTokenBiz{})
	got, err := b.Grant(context.TODO(), &GrantAccessTokenInput{User: &UserInfo{Name: "duc"}, ExpireSeconds: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "expireSeconds 必须大于 0", status.Convert(err).Message())
}

func TestAccessTokenBiz_Grant_Valid(t *testing.T) {
	f := &fakeAccessTokenRepoForAccessTokenBiz{}
	b := newAccessTokenBizForTest(f)
	got, err := b.Grant(context.TODO(), &GrantAccessTokenInput{User: &UserInfo{Name: "duc"}, ExpireSeconds: 3600})
	assert.NoError(t, err)
	assert.True(t, f.grantCalled)
	assert.Equal(t, 1, got.ID)
}

func TestAccessTokenBiz_List_Passthrough(t *testing.T) {
	f := &fakeAccessTokenRepoForAccessTokenBiz{}
	b := newAccessTokenBizForTest(f)
	got, pag, err := b.List(context.TODO(), &ListAccessTokenInput{})
	assert.NoError(t, err)
	assert.True(t, f.listCalled)
	assert.Len(t, got, 1)
	assert.Nil(t, pag)
}

func TestAccessTokenBiz_FindByToken_Passthrough(t *testing.T) {
	now := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	f := &fakeAccessTokenRepoForAccessTokenBiz{now: now}
	b := newAccessTokenBizForTest(f)
	got, err := b.FindByToken(context.TODO(), "tok")
	assert.NoError(t, err)
	assert.True(t, f.leaseFindCalled)
	assert.Equal(t, "tok", got.Token)
}

func TestAccessTokenBiz_Lease_EmptyToken(t *testing.T) {
	b := newAccessTokenBizForTest(&fakeAccessTokenRepoForAccessTokenBiz{})
	got, err := b.Lease(context.TODO(), "", 3600)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "token 不能为空", status.Convert(err).Message())
}

func TestAccessTokenBiz_Lease_Valid(t *testing.T) {
	now := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	f := &fakeAccessTokenRepoForAccessTokenBiz{now: now}
	b := newAccessTokenBizWithClock(now, f)
	got, err := b.Lease(context.TODO(), "tok", 3600)
	assert.NoError(t, err)
	assert.True(t, f.leaseFindCalled)
	assert.True(t, f.leaseUpdateCalled)
	assert.Equal(t, 1, got.ID)
}

func TestAccessTokenBiz_Revoke_EmptyToken(t *testing.T) {
	b := newAccessTokenBizForTest(&fakeAccessTokenRepoForAccessTokenBiz{})
	err := b.Revoke(context.TODO(), "")
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "token 不能为空", status.Convert(err).Message())
}

func TestAccessTokenBiz_Revoke_Valid(t *testing.T) {
	f := &fakeAccessTokenRepoForAccessTokenBiz{}
	b := newAccessTokenBizForTest(f)
	assert.NoError(t, b.Revoke(context.TODO(), "tok"))
	assert.True(t, f.revokeCalled)
}

func TestAccessTokenBiz_TouchLastUsedAt_EmptyToken(t *testing.T) {
	b := newAccessTokenBizForTest(&fakeAccessTokenRepoForAccessTokenBiz{})
	err := b.TouchLastUsedAt(context.TODO(), "", time.Now())
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "token 不能为空", status.Convert(err).Message())
}

func TestAccessTokenBiz_TouchLastUsedAt_Valid(t *testing.T) {
	f := &fakeAccessTokenRepoForAccessTokenBiz{}
	b := newAccessTokenBizForTest(f)
	assert.NoError(t, b.TouchLastUsedAt(context.TODO(), "tok", time.Now()))
	assert.True(t, f.touchCalled)
}

func TestAccessTokenBiz_Lease_FindByTokenError(t *testing.T) {
	f := &fakeAccessTokenRepoForAccessTokenBiz{findErr: errors.New("db down")}
	b := newAccessTokenBizForTest(f)
	got, err := b.Lease(context.TODO(), "tok", 3600)
	assert.Nil(t, got)
	assert.True(t, f.leaseFindCalled)
	assert.False(t, f.leaseUpdateCalled)
	assert.ErrorContains(t, err, "db down")
}

func TestAccessTokenBiz_Lease_Expired(t *testing.T) {
	now := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	f := &fakeAccessTokenRepoForAccessTokenBiz{expired: true, now: now}
	b := newAccessTokenBizWithClock(now, f)
	got, err := b.Lease(context.TODO(), "tok", 3600)
	assert.Nil(t, got)
	assert.True(t, f.leaseFindCalled)
	assert.False(t, f.leaseUpdateCalled)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "token 已经过期", status.Convert(err).Message())
}

// TestAccessToken_IsExpired 直接测模型过期判定方法：now 与 ExpiredAt 的三种大小关系。
func TestAccessToken_IsExpired(t *testing.T) {
	now := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name      string
		expiredAt time.Time
		want      bool
	}{
		{"已过期", now.Add(-time.Minute), true},
		{"未过期", now.Add(time.Minute), false},
		{"边界相等视为未过期", now, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			at := &AccessToken{ExpiredAt: c.expiredAt}
			assert.Equal(t, c.want, at.IsExpired(now))
		})
	}
}
