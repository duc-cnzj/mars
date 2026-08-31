package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/settings"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// settingsSvcMocks 聚合 settingsSvc 的下游 mock，由 newSettingsSvcWithMocks 统一构造。
// AccessBiz 用 nil repo 即可：settings 仅经 RequireAdmin 门禁，不触达实体加载。
type settingsSvcMocks struct {
	ctrl        *gomock.Controller
	settingsBiz *biz.MockSettingsBiz
}

func newSettingsSvcWithMocks(t *testing.T) (*settingsSvc, *settingsSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &settingsSvcMocks{
		ctrl:        ctrl,
		settingsBiz: biz.NewMockSettingsBiz(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewSettingsSvc(SettingsSvcDeps{
		SettingsBiz: mocks.settingsBiz,
		Logger:      logger,
	}).(*settingsSvc)
	if !ok {
		panic("NewSettingsSvc returned unexpected type")
	}
	return s, mocks
}

func TestNewSettingsSvc(t *testing.T) {
	svc, _ := newSettingsSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.settingsBiz)
	assert.NotNil(t, svc.logger)
}

// Test_settingsSvc_Get 成功路径：biz 配置分组映射到 proto 响应，masked 落位。
func Test_settingsSvc_Get(t *testing.T) {
	svc, mocks := newSettingsSvcWithMocks(t)
	mocks.settingsBiz.EXPECT().Get().Return(&biz.Settings{Groups: []*biz.ConfigGroup{
		{ID: "server", Items: []*biz.ConfigItem{
			{Key: "app_port", Value: "4000"},
			{Key: "admin_password", Value: "123456", Masked: true},
		}},
	}})

	resp, err := svc.Get(newAdminUserCtx(), &settings.GetRequest{})
	assert.NoError(t, err)
	if assert.Len(t, resp.Groups, 1) {
		g := resp.Groups[0]
		assert.Equal(t, "server", g.Id)
		if assert.Len(t, g.Items, 2) {
			assert.Equal(t, "app_port", g.Items[0].Key)
			assert.Equal(t, "4000", g.Items[0].Value)
			assert.False(t, g.Items[0].Masked)
			assert.Equal(t, "admin_password", g.Items[1].Key)
			assert.Equal(t, "123456", g.Items[1].Value)
			assert.True(t, g.Items[1].Masked)
		}
	}
}

// Test_settingsSvc_Get_Empty 空分组：响应为空不报错。
func Test_settingsSvc_Get_Empty(t *testing.T) {
	svc, mocks := newSettingsSvcWithMocks(t)
	mocks.settingsBiz.EXPECT().Get().Return(&biz.Settings{Groups: []*biz.ConfigGroup{}})

	resp, err := svc.Get(newAdminUserCtx(), &settings.GetRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.Groups)
}

// newOrdinaryAdminUserCtx 构造普通管理员上下文（mars_admin 角色但非超管固定邮箱）。
func newOrdinaryAdminUserCtx() context.Context {
	return biz.SetUser(context.TODO(), &biz.UserInfo{
		ID:    "9",
		Email: "ordinary-admin@mars.com",
		Name:  "ordinary-admin",
		Roles: []string{schematype.MarsAdmin},
	})
}

// Test_settingsSvc_Authorize 授权门禁：仅内置超级管理员放行（newAdminUserCtx 的
// adminEmail 即超管固定邮箱），普通管理员（mars_admin）与普通用户均拒绝。
func Test_settingsSvc_Authorize(t *testing.T) {
	svc, _ := newSettingsSvcWithMocks(t)

	ctx, err := svc.Authorize(newAdminUserCtx(), settings.Settings_Get_FullMethodName)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	_, err = svc.Authorize(newOrdinaryAdminUserCtx(), settings.Settings_Get_FullMethodName)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)

	_, err = svc.Authorize(newOtherUserCtx(), settings.Settings_Get_FullMethodName)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
}
