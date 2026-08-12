package data

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAccessTokenRepo 构造基于 sqlite 的 accessTokenRepo。
func newAccessTokenRepo(t *testing.T) (*accessTokenRepo, *ent.Client) {
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	repo := NewAccessTokenRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}), timer.NewReal())
	return repo.(*accessTokenRepo), entdb
}

// TestAccessTokenRepo_List 覆盖分页/邮箱过滤/软删除包含三种查询路径。
func TestAccessTokenRepo_List(t *testing.T) {
	repo, entdb := newAccessTokenRepo(t)
	ctx := context.TODO()

	for i := 0; i < 3; i++ {
		_, err := repo.Grant(ctx, &biz.GrantAccessTokenInput{
			User:          &biz.UserInfo{Name: "u", Email: "a@b.c"},
			Usage:         "test",
			ExpireSeconds: 3600,
		})
		assert.NoError(t, err)
	}
	// 软删除一条，验证 WithSoftDelete 分支
	deleted := entdb.AccessToken.Query().FirstX(ctx)
	entdb.AccessToken.DeleteOneID(deleted.ID).Exec(ctx)

	t.Run("default excludes soft deleted", func(t *testing.T) {
		list, pag, err := repo.List(ctx, &biz.ListAccessTokenInput{Page: 1, PageSize: 10, Email: "a@b.c"})
		assert.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, int32(2), pag.Count)
		assert.Equal(t, int32(1), pag.Page)
		assert.Equal(t, int32(10), pag.PageSize)
	})

	t.Run("with soft delete includes deleted", func(t *testing.T) {
		list, _, err := repo.List(ctx, &biz.ListAccessTokenInput{Page: 1, PageSize: 10, WithSoftDelete: true, Email: "a@b.c"})
		assert.NoError(t, err)
		assert.Len(t, list, 3)
	})

	t.Run("email filter mismatch returns empty", func(t *testing.T) {
		list, pag, err := repo.List(ctx, &biz.ListAccessTokenInput{Page: 1, PageSize: 10, Email: "nobody@x.y"})
		assert.NoError(t, err)
		assert.Empty(t, list)
		assert.Equal(t, int32(0), pag.Count)
	})
}

// TestAccessTokenRepo_List_ErrorBranch 覆盖 List 在 DB 故障时返回错误而非 panic
// （AllX 已改为 All + errs.Wrap，错误分支必须被测试承重）。
func TestAccessTokenRepo_List_ErrorBranch(t *testing.T) {
	closed := NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}})
	repo := NewAccessTokenRepo(closed, timer.NewReal()).(*accessTokenRepo)
	_, _, err := repo.List(context.TODO(), &biz.ListAccessTokenInput{Page: 1, PageSize: 10})
	assert.Error(t, err)
}

// TestAccessTokenRepo_Grant_FindByToken_UpdateExpiresAt 覆盖生成/查询/续期链路。
func TestAccessTokenRepo_Grant_FindByToken_UpdateExpiresAt(t *testing.T) {
	repo, _ := newAccessTokenRepo(t)
	ctx := context.TODO()

	granted, err := repo.Grant(ctx, &biz.GrantAccessTokenInput{
		User:          &biz.UserInfo{Name: "duc", Email: "duc@mars.dev"},
		Usage:         "cli",
		ExpireSeconds: 60,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, granted.Token)
	assert.Equal(t, "duc@mars.dev", granted.Email)

	found, err := repo.FindByToken(ctx, granted.Token)
	assert.NoError(t, err)
	assert.Equal(t, granted.ID, found.ID)
	assert.Equal(t, "duc", found.UserInfo.Name)

	newExpire := time.Now().Add(24 * time.Hour)
	updated, err := repo.UpdateExpiresAt(ctx, granted.Token, newExpire)
	assert.NoError(t, err)
	assert.WithinDuration(t, newExpire, updated.ExpiredAt, time.Second)

	_, err = repo.FindByToken(ctx, "missing-token")
	assert.Error(t, err)

	_, err = repo.UpdateExpiresAt(ctx, "missing-token", newExpire)
	assert.Error(t, err)
}

// TestAccessTokenRepo_Revoke_TouchLastUsedAt 覆盖撤销与最近使用时间回写。
func TestAccessTokenRepo_Revoke_TouchLastUsedAt(t *testing.T) {
	repo, _ := newAccessTokenRepo(t)
	ctx := context.TODO()

	granted, err := repo.Grant(ctx, &biz.GrantAccessTokenInput{
		User:          &biz.UserInfo{Email: "a@b.c"},
		Usage:         "web",
		ExpireSeconds: 60,
	})
	assert.NoError(t, err)

	now := time.Now()
	assert.NoError(t, repo.TouchLastUsedAt(ctx, granted.Token, now))

	after, _ := repo.FindByToken(ctx, granted.Token)
	assert.WithinDuration(t, now, *after.LastUsedAt, time.Second)

	assert.NoError(t, repo.Revoke(ctx, granted.Token))

	// 软删除后 FindByToken 默认不可见
	_, err = repo.FindByToken(ctx, granted.Token)
	assert.Error(t, err)
	// SkipSoftDelete 下仍可见
	visible, err := repo.FindByToken(mixin.SkipSoftDelete(ctx), granted.Token)
	assert.NoError(t, err)
	assert.Equal(t, granted.ID, visible.ID)
}

// TestToAccessToken 覆盖 nil 与实体两种转换。
func TestToAccessToken(t *testing.T) {
	assert.Nil(t, toAccessToken(nil))
	tok := &ent.AccessToken{ID: 1, Token: "t", Email: "e"}
	assert.Equal(t, 1, toAccessToken(tok).ID)
	assert.Equal(t, "t", toAccessToken(tok).Token)
}
