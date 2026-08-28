package biz

//go:generate go tool mockgen -destination ./mock_biz.go -package biz github.com/duc-cnzj/mars/v6/internal/biz AuthBiz,EndpointBiz,AccessTokenBiz,PictureBiz,GitBiz,ProjectBiz,K8sBiz,NamespaceBiz,RepoBiz,SettingsBiz,UserBiz

import "github.com/google/wire"

// WireBizSet 是 biz 层全部业务对象的 wire 装配集合，供 cmd/wire 注入依赖图。
var WireBizSet = wire.NewSet(
	NewEndpointBiz,
	NewAuthBiz,
	NewAccessTokenBiz,
	NewPictureBiz,
	NewFileBiz,
	NewGitBiz,
	NewChangelogBiz,
	NewProjectBiz,
	NewEventBiz,
	NewK8sBiz,
	NewHelmerBiz,
	NewRepoBiz,
	NewDeployBiz,
	NewNamespaceBiz,
	NewContainerBiz,
	NewMetricsBiz,
	NewSettingsBiz,
	NewUserBiz,
	NewAccessTokenManager,
	// AccessBiz 由 NewAccessBiz 直接构造：内部自绑 MustGetUser（见 context.go），
	// 无需传输层 wire.Value 注入 getUser 回调。
	NewAccessBiz,
)
