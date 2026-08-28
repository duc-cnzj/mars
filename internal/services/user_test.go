package services

import (
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/user"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// userSvcMocks 收口 userSvc 测试用 mock 集合。
type userSvcMocks struct {
	ctrl    *gomock.Controller
	userBiz *biz.MockUserBiz
}

// newUserSvcWithMocks 构造带 mock 的 userSvc。RequireAdmin 仅判定 ctx 内用户角色，
// 不需要 nsRepo/projBiz，故 AccessBiz 传 nil 依赖。
func newUserSvcWithMocks(t *testing.T) (*userSvc, *userSvcMocks) {
	ctrl := gomock.NewController(t)
	userBiz := biz.NewMockUserBiz(ctrl)
	svc := NewUserSvc(UserSvcDeps{
		UserBiz:   userBiz,
		AccessBiz: biz.NewAccessBiz(nil, nil),
		Logger:    mlog.NewForConfig(nil),
	}).(*userSvc)
	if svc == nil {
		panic("NewUserSvc returned unexpected type")
	}
	return svc, &userSvcMocks{ctrl: ctrl, userBiz: userBiz}
}

// TestNewUserSvc_Creation 构造后字段落位。
func TestNewUserSvc_Creation(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	assert.NotNil(t, svc.userBiz)
	assert.NotNil(t, svc.accessBiz)
	assert.NotNil(t, svc.logger)
	mocks.ctrl.Finish()
}

// Test_userSvc_Authorize 管理员放行，普通用户被拒（RequireAdmin 门卫）。
func Test_userSvc_Authorize(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	defer mocks.ctrl.Finish()

	_, err := svc.Authorize(newAdminUserCtx(), user.User_List_FullMethodName)
	assert.NoError(t, err)

	_, err = svc.Authorize(newOtherUserCtx(), user.User_List_FullMethodName)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

// Test_userSvc_List 成功路径：搜索/角色过滤入参透传，结果映射到响应（含统计）。
func Test_userSvc_List(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	defer mocks.ctrl.Finish()

	lastLogin := time.Now()
	mocks.userBiz.EXPECT().List(gomock.Any(), &biz.ListUserInput{
		Page:      2,
		PageSize:  10,
		Search:    "duc",
		AdminOnly: true,
	}).Return(&biz.ListUserResult{
		Items: []*biz.User{
			{ID: 1, Email: "duc@mars.dev", Name: "duc", Roles: []string{biz.MarsAdmin}},
			{ID: 2, Email: "u@mars.dev", Roles: []string{}, LastLogin: &lastLogin},
		},
		Pag:   pagination.NewPagination(2, 10, 2),
		Stats: biz.UserStats{Total: 3, Admins: 1, Regular: 2},
	}, nil)

	resp, err := svc.List(newAdminUserCtx(), &user.ListRequest{
		Page:     loPtr32(2),
		PageSize: loPtr32(10),
		Search:   "duc",
		Role:     "admin",
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), resp.Page)
	assert.Equal(t, int32(10), resp.PageSize)
	assert.Equal(t, int32(2), resp.Count)
	if assert.Len(t, resp.Items, 2) {
		assert.Equal(t, []string{"admin"}, resp.Items[0].Roles)
		assert.Equal(t, []string{"user"}, resp.Items[1].Roles)
		assert.NotNil(t, resp.Items[1].LastLogin)
	}
	if assert.NotNil(t, resp.Stats) {
		assert.Equal(t, int32(3), resp.Stats.Total)
		assert.Equal(t, int32(1), resp.Stats.Admins)
		assert.Equal(t, int32(2), resp.Stats.Regular)
	}
}

// Test_userSvc_List_Error 透传 biz 错误。
func Test_userSvc_List_Error(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	defer mocks.ctrl.Finish()

	mocks.userBiz.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("list boom"))
	_, err := svc.List(newAdminUserCtx(), &user.ListRequest{})
	assert.EqualError(t, err, "list boom")
}

// Test_userSvc_ToggleAdmin 成功路径透传。
func Test_userSvc_ToggleAdmin(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	defer mocks.ctrl.Finish()

	mocks.userBiz.EXPECT().ToggleAdmin(gomock.Any(), "a@b.c", true).Return(nil)
	_, err := svc.ToggleAdmin(newAdminUserCtx(), &user.ToggleAdminRequest{Email: "a@b.c", Admin: true})
	assert.NoError(t, err)
}

// Test_userSvc_ToggleAdmin_Error 透传 biz 错误（保留原始状态码）。
func Test_userSvc_ToggleAdmin_Error(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	defer mocks.ctrl.Finish()

	mocks.userBiz.EXPECT().ToggleAdmin(gomock.Any(), "a@b.c", false).
		Return(status.Error(codes.NotFound, "用户不存在"))
	_, err := svc.ToggleAdmin(newAdminUserCtx(), &user.ToggleAdminRequest{Email: "a@b.c"})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// Test_userSvc_Sync 成功路径：触发投影同步（内置管理员/命名空间成员 → users 表）。
func Test_userSvc_Sync(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	defer mocks.ctrl.Finish()

	mocks.userBiz.EXPECT().Sync(gomock.Any()).Return(nil)
	_, err := svc.Sync(newAdminUserCtx(), &user.SyncUsersRequest{})
	assert.NoError(t, err)
}

// Test_userSvc_Sync_Error 同步失败：透传 biz 错误（保留原始状态码）。
func Test_userSvc_Sync_Error(t *testing.T) {
	svc, mocks := newUserSvcWithMocks(t)
	defer mocks.ctrl.Finish()

	mocks.userBiz.EXPECT().Sync(gomock.Any()).
		Return(status.Error(codes.Internal, "sync boom"))
	_, err := svc.Sync(newAdminUserCtx(), &user.SyncUsersRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
}
