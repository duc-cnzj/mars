package services

import (
	"context"
	"errors"
	"testing"

	hibernatepb "github.com/duc-cnzj/mars/api/v6/proto/hibernate"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// hibernateSvcMocks 聚合 hibernateSvc 的下游 mock，由 newHibernateSvcWithMocks 统一构造。
// AccessBiz 用 nil repo 即可：hibernate 仅经 RequireAdmin 门禁，不触达实体加载。
type hibernateSvcMocks struct {
	ctrl   *gomock.Controller
	window *biz.MockHibernateWindowBiz
}

func newHibernateSvcWithMocks(t *testing.T) (*hibernateSvc, *hibernateSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &hibernateSvcMocks{ctrl: ctrl, window: biz.NewMockHibernateWindowBiz(ctrl)}
	s, ok := NewHibernateSvc(HibernateSvcDeps{
		HibernateWindowBiz: mocks.window,
		AccessBiz:          biz.NewAccessBiz(nil, nil),
		Logger:             mlog.NewForConfig(nil),
	}).(*hibernateSvc)
	if !ok {
		panic("NewHibernateSvc returned unexpected type")
	}
	return s, mocks
}

// TestNewHibernateSvc 覆盖构造后字段装配。
func TestNewHibernateSvc(t *testing.T) {
	svc, _ := newHibernateSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.hibernateWindowBiz)
	assert.NotNil(t, svc.accessBiz)
	assert.NotNil(t, svc.logger)
}

// Test_hibernateSvc_Authorize 覆盖分级门禁：
//   - ListWindows 登录即可读（admin 与非 admin 均放行）——项目卡片被普通用户可见，
//     需按 window_name 展示时段；
//   - CreateWindow 这类维护操作仍限管理员，非 admin 拒绝。
func Test_hibernateSvc_Authorize(t *testing.T) {
	svc, _ := newHibernateSvcWithMocks(t)

	// ListWindows：admin 与非 admin 均放行。
	for name, ctx := range map[string]context.Context{
		"admin": newAdminUserCtx(),
		"plain": newOtherUserCtx(),
	} {
		got, err := svc.Authorize(ctx, hibernatepb.Hibernate_ListWindows_FullMethodName)
		assert.NoError(t, err, name)
		assert.NotNil(t, got, name)
	}

	// CreateWindow：admin 放行、非 admin 拒绝。
	ctx, err := svc.Authorize(newAdminUserCtx(), hibernatepb.Hibernate_CreateWindow_FullMethodName)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	_, err = svc.Authorize(newOtherUserCtx(), hibernatepb.Hibernate_CreateWindow_FullMethodName)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}

// Test_hibernateSvc_ListWindows 覆盖成功映射（含 nil 安全）与错误收敛。
func Test_hibernateSvc_ListWindows(t *testing.T) {
	svc, mocks := newHibernateSvcWithMocks(t)
	mocks.window.EXPECT().List(gomock.Any()).Return([]*biz.HibernateWindow{
		{ID: 1, Name: "夜间", Start: "23:00", End: "06:00", Description: "过夜"},
		{ID: 2, Name: "午休", Start: "12:00", End: "13:00"},
	}, nil)

	resp, err := svc.ListWindows(newAdminUserCtx(), &hibernatepb.ListWindowsRequest{})
	assert.NoError(t, err)
	if assert.Len(t, resp.Items, 2) {
		assert.Equal(t, int32(1), resp.Items[0].Id)
		assert.Equal(t, "夜间", resp.Items[0].Name)
		assert.Equal(t, "23:00", resp.Items[0].Start)
		assert.Equal(t, "06:00", resp.Items[0].End)
		assert.Equal(t, "过夜", resp.Items[0].Description)
		assert.Equal(t, "午休", resp.Items[1].Name)
		assert.Empty(t, resp.Items[1].Description)
	}

	// 错误路径。
	mocks.window.EXPECT().List(gomock.Any()).Return(nil, errors.New("db down"))
	_, err = svc.ListWindows(newAdminUserCtx(), &hibernatepb.ListWindowsRequest{})
	assert.Error(t, err)
}

// Test_hibernateSvc_CreateWindow 覆盖成功映射与错误收敛。
func Test_hibernateSvc_CreateWindow(t *testing.T) {
	svc, mocks := newHibernateSvcWithMocks(t)
	mocks.window.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, w *biz.HibernateWindow) (*biz.HibernateWindow, error) {
			assert.Equal(t, "夜间", w.Name)
			assert.Equal(t, "23:00", w.Start)
			assert.Equal(t, "06:00", w.End)
			assert.Equal(t, "过夜", w.Description)
			w.ID = 7
			return w, nil
		})

	resp, err := svc.CreateWindow(newAdminUserCtx(), &hibernatepb.CreateWindowRequest{
		Name: "夜间", Start: "23:00", End: "06:00", Description: "过夜",
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(7), resp.Window.Id)
	assert.Equal(t, "夜间", resp.Window.Name)

	mocks.window.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
	_, err = svc.CreateWindow(newAdminUserCtx(), &hibernatepb.CreateWindowRequest{Name: "夜间", Start: "23:00", End: "06:00"})
	assert.Error(t, err)
}

// Test_hibernateSvc_UpdateWindow 覆盖成功映射（ID 透传）与错误收敛。
func Test_hibernateSvc_UpdateWindow(t *testing.T) {
	svc, mocks := newHibernateSvcWithMocks(t)
	mocks.window.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, w *biz.HibernateWindow) (*biz.HibernateWindow, error) {
			assert.Equal(t, 3, w.ID)
			assert.Equal(t, "午休", w.Name)
			return w, nil
		})

	resp, err := svc.UpdateWindow(newAdminUserCtx(), &hibernatepb.UpdateWindowRequest{
		Id: 3, Name: "午休", Start: "12:00", End: "13:00",
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(3), resp.Window.Id)

	mocks.window.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
	_, err = svc.UpdateWindow(newAdminUserCtx(), &hibernatepb.UpdateWindowRequest{Id: 3, Name: "午休", Start: "12:00", End: "13:00"})
	assert.Error(t, err)
}

// Test_hibernateSvc_DeleteWindow 覆盖成功（ID 透传）、被引用拒删（400 码透传）与错误收敛。
func Test_hibernateSvc_DeleteWindow(t *testing.T) {
	svc, mocks := newHibernateSvcWithMocks(t)
	mocks.window.EXPECT().Delete(gomock.Any(), 3).Return(nil)

	resp, err := svc.DeleteWindow(newAdminUserCtx(), &hibernatepb.DeleteWindowRequest{Id: 3})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// 被引用拒删：biz 以 400 拒绝，services 透传且不改码（客户端据此提示先解除引用）。
	mocks.window.EXPECT().Delete(gomock.Any(), 3).
		Return(errs.WrapInvalidArgument(errors.New("仍被引用"), "删除休眠时间窗"))
	_, err = svc.DeleteWindow(newAdminUserCtx(), &hibernatepb.DeleteWindowRequest{Id: 3})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	mocks.window.EXPECT().Delete(gomock.Any(), 3).Return(errors.New("db down"))
	_, err = svc.DeleteWindow(newAdminUserCtx(), &hibernatepb.DeleteWindowRequest{Id: 3})
	assert.Error(t, err)
}

// Test_toHibernateWindowProto_Nil 覆盖 nil 安全转换。
func Test_toHibernateWindowProto_Nil(t *testing.T) {
	assert.Nil(t, toHibernateWindowProto(nil))
}
