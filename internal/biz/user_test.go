package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeUserRepoForUserBiz 内嵌 UserRepo 接口，仅覆写被测编排涉及的四个方法，
// 记录调用入参，其余未用方法调用即 panic（暴露未预期的下游交互）。
type fakeUserRepoForUserBiz struct {
	UserRepo
	syncErr   error
	listErr   error
	toggleErr error
	synced    bool
	syncIn    struct {
		email string
		name  string
	}
	listIn  *ListUserInput
	result  *ListUserResult
	toggles []struct {
		email string
		admin bool
	}
}

func (f *fakeUserRepoForUserBiz) EnsureSynced(ctx context.Context) error {
	f.synced = true
	return f.syncErr
}

func (f *fakeUserRepoForUserBiz) SyncLoginUser(ctx context.Context, email, name string) error {
	f.syncIn.email = email
	f.syncIn.name = name
	return f.syncErr
}

func (f *fakeUserRepoForUserBiz) List(ctx context.Context, input *ListUserInput) (*ListUserResult, error) {
	f.listIn = input
	return f.result, f.listErr
}

func (f *fakeUserRepoForUserBiz) ToggleAdmin(ctx context.Context, email string, admin bool) error {
	f.toggles = append(f.toggles, struct {
		email string
		admin bool
	}{email, admin})
	return f.toggleErr
}

// TestUserBiz_List_Success 成功路径：直接透传 repo 查询，不做隐式同步。
func TestUserBiz_List_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{result: &ListUserResult{
		Items: []*User{{Email: "a@b.c"}},
		Pag:   pagination.NewPagination(1, 15, 1),
		Stats: UserStats{Total: 1, Admins: 1, Regular: 0},
	}}
	b := NewUserBiz(fake)

	out, err := b.List(context.TODO(), &ListUserInput{Page: 2, PageSize: 10})
	assert.NoError(t, err)
	assert.False(t, fake.synced, "List 不应隐式同步（同步由「同步用户」按钮显式触发）")
	assert.Equal(t, int32(2), fake.listIn.Page)
	assert.Equal(t, int32(10), fake.listIn.PageSize)
	assert.Same(t, fake.result, out, "List 应原样透传 repo 结果")
}

// TestUserBiz_Sync_Success 成功路径：Sync 触发 repo.EnsureSynced 全量同步。
func TestUserBiz_Sync_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	assert.NoError(t, b.Sync(context.TODO()))
	assert.True(t, fake.synced, "Sync 应触发 repo.EnsureSynced")
}

// TestUserBiz_Sync_Error 同步失败：透传 repo 错误。
func TestUserBiz_Sync_Error(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{syncErr: errors.New("sync boom")}
	b := NewUserBiz(fake)

	err := b.Sync(context.TODO())
	assert.EqualError(t, err, "sync boom")
}

// TestUserBiz_ToggleAdmin_Success 成功路径：透传 email/admin 到 repo。
func TestUserBiz_ToggleAdmin_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	assert.NoError(t, b.ToggleAdmin(context.TODO(), "a@b.c", true))
	if assert.Len(t, fake.toggles, 1) {
		assert.Equal(t, "a@b.c", fake.toggles[0].email)
		assert.True(t, fake.toggles[0].admin)
	}
}

// TestUserBiz_ToggleAdmin_EmptyEmail 空邮箱是确定语义错误：返回 InvalidArgument，
// 不触达 repo。
func TestUserBiz_ToggleAdmin_EmptyEmail(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	err := b.ToggleAdmin(context.TODO(), "  ", true)
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "空邮箱应判定为参数不合法，got %v", err)
	assert.Empty(t, fake.toggles, "参数校验失败不应调用 repo")
}

// TestUserBiz_ToggleAdmin_RepoError 透传 repo 错误。
func TestUserBiz_ToggleAdmin_RepoError(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{toggleErr: errors.New("toggle boom")}
	b := NewUserBiz(fake)

	err := b.ToggleAdmin(context.TODO(), "a@b.c", false)
	assert.EqualError(t, err, "toggle boom")
}

// TestUserBiz_SyncLoginUser_Success 成功路径：邮箱 trim 后透传 email/name 到 repo。
func TestUserBiz_SyncLoginUser_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	assert.NoError(t, b.SyncLoginUser(context.TODO(), "  a@b.c  ", "duc"))
	assert.Equal(t, "a@b.c", fake.syncIn.email, "邮箱应 trim 后传给 repo")
	assert.Equal(t, "duc", fake.syncIn.name)
}

// TestUserBiz_SyncLoginUser_EmptyEmail 空邮箱是确定语义错误：返回 InvalidArgument，
// 不触达 repo。
func TestUserBiz_SyncLoginUser_EmptyEmail(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	err := b.SyncLoginUser(context.TODO(), "  ", "duc")
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "空邮箱应判定为参数不合法，got %v", err)
	assert.Empty(t, fake.syncIn.email, "参数校验失败不应调用 repo")
	assert.Empty(t, fake.syncIn.name)
}

// TestUserBiz_SyncLoginUser_RepoError 透传 repo 错误。
func TestUserBiz_SyncLoginUser_RepoError(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{syncErr: errors.New("sync boom")}
	b := NewUserBiz(fake)

	err := b.SyncLoginUser(context.TODO(), "a@b.c", "duc")
	assert.EqualError(t, err, "sync boom")
}
