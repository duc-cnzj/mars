package application

import "github.com/google/wire"

//go:generate mockgen -destination ./mock_types.go -package application github.com/duc-cnzj/mars/v4/internal/application PluginManger,Picture,Project,App,WsServer
var WireApp = wire.NewSet(NewPluginManager)
