package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/settings"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

var _ settings.SettingsServer = (*settingsSvc)(nil)

// settingsSvc 是 settings.SettingsServer 的 gRPC 实现：只读聚合平台配置，
// 供管理员后台系统设置页展示，经 Authorize admin 门禁后仅管理员可调用。
type settingsSvc struct {
	settings.UnimplementedSettingsServer

	settingsBiz biz.SettingsBiz
	accessBiz   biz.AccessBiz
	logger      mlog.Logger
}

// SettingsSvcDeps 收口 NewSettingsSvc 的构造依赖，由 wire 按字段注入。
type SettingsSvcDeps struct {
	SettingsBiz biz.SettingsBiz
	AccessBiz   biz.AccessBiz
	Logger      mlog.Logger
}

// NewSettingsSvc 构造配置聚合服务，依赖由 wire 注入。
func NewSettingsSvc(deps SettingsSvcDeps) settings.SettingsServer {
	return &settingsSvc{
		settingsBiz: deps.SettingsBiz,
		accessBiz:   deps.AccessBiz,
		logger:      deps.Logger.WithModule("services/settings"),
	}
}

// Authorize 是服务级授权门禁：配置聚合仅管理员后台可见，无 allowlist 全员要求 admin。
func (s *settingsSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return s.accessBiz.RequireAdmin(ctx, fullMethodName)
}

// Get 返回平台配置分组视图（只读），biz 分组映射到 proto 响应。
func (s *settingsSvc) Get(ctx context.Context, request *settings.GetRequest) (*settings.GetResponse, error) {
	all := s.settingsBiz.Get()
	resp := &settings.GetResponse{Groups: make([]*settings.ConfigGroup, 0, len(all.Groups))}
	for _, g := range all.Groups {
		group := &settings.ConfigGroup{Id: g.ID, Items: make([]*settings.ConfigItem, 0, len(g.Items))}
		for _, item := range g.Items {
			group.Items = append(group.Items, &settings.ConfigItem{
				Key:    item.Key,
				Value:  item.Value,
				Masked: item.Masked,
			})
		}
		resp.Groups = append(resp.Groups, group)
	}
	return resp, nil
}
