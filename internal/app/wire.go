package app

import "github.com/google/wire"

//go:generate go tool mockgen -destination ./mock_types.go -package app github.com/duc-cnzj/mars/v6/internal/app PluginManager,Picture,App,PubSub,WsSender,GitServer,DomainManager,HttpHandler

// WireApp 是 app 包的 wire provider 集合。cron/event 用例已拆出独立包
// （internal/cronjob、internal/eventhandler），组合根只保留插件管理装配。
var WireApp = wire.NewSet(NewPluginManager)
