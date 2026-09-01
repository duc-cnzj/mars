package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/errs"
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
	syncIn    struct {
		email string
		name  string
		roles []string
	}
	listIn  *ListUserInput
	result  *ListUserResult
	toggles []struct {
		email string
		admin bool
	}
	resetErr error
	resets   []string
}

func (f *fakeUserRepoForUserBiz) SyncLoginUser(ctx context.Context, email, name string, roles []string) error {
	f.syncIn.email = email
	f.syncIn.name = name
	f.syncIn.roles = roles
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

func (f *fakeUserRepoForUserBiz) ResetRolesOverride(ctx context.Context, email string) error {
	f.resets = append(f.resets, email)
	return f.resetErr
}

// TestUserBiz_List_Success 成功路径：直接透传 repo 查询。
func TestUserBiz_List_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{result: &ListUserResult{
		Items: []*User{{Email: "a@b.c"}},
		Pag:   pagination.NewPagination(1, 15, 1),
		Stats: UserStats{Total: 1, Admins: 1, Regular: 0},
	}}
	b := NewUserBiz(fake)

	out, err := b.List(context.TODO(), &ListUserInput{Page: 2, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), fake.listIn.Page)
	assert.Equal(t, int32(10), fake.listIn.PageSize)
	assert.Same(t, fake.result, out, "List 应原样透传 repo 结果")
}

// TestUserBiz_ToggleAdmin_Success 成功路径：超级管理员操作时透传 email/admin 到 repo。
func TestUserBiz_ToggleAdmin_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)
	ctx := superAdminCtx()

	assert.NoError(t, b.ToggleAdmin(ctx, "a@b.c", true))
	if assert.Len(t, fake.toggles, 1) {
		assert.Equal(t, "a@b.c", fake.toggles[0].email)
		assert.True(t, fake.toggles[0].admin)
	}
}

// TestUserBiz_ToggleAdmin_NonSuperAdminDenied 普通管理员不能修改他人权限：返回
// PermissionDenied，且不触达 repo。
func TestUserBiz_ToggleAdmin_NonSuperAdminDenied(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)
	ctx := SetUser(context.TODO(), &UserInfo{Email: "regular-admin@x.com", Roles: []string{MarsAdmin}})

	err := b.ToggleAdmin(ctx, "a@b.c", true)
	assert.Equal(t, errs.ErrorPermissionDenied, err, "普通管理员只能查看不能修改")
	assert.Empty(t, fake.toggles, "非超管不应触达 repo")
}

// TestUserBiz_ToggleAdmin_EmptyEmail 空邮箱是确定语义错误：返回 InvalidArgument，
// 不触达 repo。
func TestUserBiz_ToggleAdmin_EmptyEmail(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	err := b.ToggleAdmin(superAdminCtx(), "  ", true)
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "空邮箱应判定为参数不合法，got %v", err)
	assert.Empty(t, fake.toggles, "参数校验失败不应调用 repo")
}

// TestUserBiz_ToggleAdmin_RepoError 透传 repo 错误。
func TestUserBiz_ToggleAdmin_RepoError(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{toggleErr: errors.New("toggle boom")}
	b := NewUserBiz(fake)

	err := b.ToggleAdmin(superAdminCtx(), "a@b.c", false)
	assert.EqualError(t, err, "toggle boom")
}

// superAdminCtx 构造内置超级管理员身份的 ctx，供 ToggleAdmin 权限门卫放行。
func superAdminCtx() context.Context {
	return SetUser(context.TODO(), &UserInfo{Email: SuperAdminEmail, Roles: []string{MarsAdmin}})
}

// TestUserBiz_SyncLoginUser_Success 成功路径：邮箱 trim 后透传 email/name/roles 到 repo。
func TestUserBiz_SyncLoginUser_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	assert.NoError(t, b.SyncLoginUser(context.TODO(), "  a@b.c  ", "duc", []string{"mars_admin"}))
	assert.Equal(t, "a@b.c", fake.syncIn.email, "邮箱应 trim 后传给 repo")
	assert.Equal(t, "duc", fake.syncIn.name)
	assert.Equal(t, []string{"mars_admin"}, fake.syncIn.roles, "roles 应原样透传给 repo")
}

// TestUserBiz_SyncLoginUser_EmptyEmail 空邮箱是确定语义错误：返回 InvalidArgument，
// 不触达 repo。
func TestUserBiz_SyncLoginUser_EmptyEmail(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	err := b.SyncLoginUser(context.TODO(), "  ", "duc", []string{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "空邮箱应判定为参数不合法，got %v", err)
	assert.Empty(t, fake.syncIn.email, "参数校验失败不应调用 repo")
	assert.Empty(t, fake.syncIn.name)
}

// TestUserBiz_SyncLoginUser_RepoError 透传 repo 错误。
func TestUserBiz_SyncLoginUser_RepoError(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{syncErr: errors.New("sync boom")}
	b := NewUserBiz(fake)

	err := b.SyncLoginUser(context.TODO(), "a@b.c", "duc", []string{})
	assert.EqualError(t, err, "sync boom")
}

// TestUserBiz_ResetRolesOverride_Success 成功路径：超级管理员操作时把邮箱透传 repo。
func TestUserBiz_ResetRolesOverride_Success(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	assert.NoError(t, b.ResetRolesOverride(superAdminCtx(), "  a@b.c  "))
	if assert.Len(t, fake.resets, 1) {
		assert.Equal(t, "  a@b.c  ", fake.resets[0], "ResetRolesOverride 邮箱原样透传（trim 由 repo 归一处理）")
	}
}

// TestUserBiz_ResetRolesOverride_NonSuperAdminDenied 普通管理员不能解除接管：返回
// PermissionDenied，且不触达 repo。
func TestUserBiz_ResetRolesOverride_NonSuperAdminDenied(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)
	ctx := SetUser(context.TODO(), &UserInfo{Email: "regular-admin@x.com", Roles: []string{MarsAdmin}})

	err := b.ResetRolesOverride(ctx, "a@b.c")
	assert.Equal(t, errs.ErrorPermissionDenied, err, "解除接管同属角色管理，普通管理员只能查看")
	assert.Empty(t, fake.resets, "非超管不应触达 repo")
}

// TestUserBiz_ResetRolesOverride_EmptyEmail 空邮箱是确定语义错误：返回 InvalidArgument，
// 不触达 repo。
func TestUserBiz_ResetRolesOverride_EmptyEmail(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{}
	b := NewUserBiz(fake)

	err := b.ResetRolesOverride(superAdminCtx(), "  ")
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "空邮箱应判定为参数不合法，got %v", err)
	assert.Empty(t, fake.resets, "参数校验失败不应调用 repo")
}

// TestUserBiz_ResetRolesOverride_RepoError 透传 repo 错误。
func TestUserBiz_ResetRolesOverride_RepoError(t *testing.T) {
	fake := &fakeUserRepoForUserBiz{resetErr: errors.New("reset boom")}
	b := NewUserBiz(fake)

	err := b.ResetRolesOverride(superAdminCtx(), "a@b.c")
	assert.EqualError(t, err, "reset boom")
}
