package application

import "github.com/google/wire"

//go:generate mockgen -destination ./mock_types.go -package application github.com/duc-cnzj/mars/v5/internal/application PluginManger,Picture,Project,App,WsHttpServer,PubSub,WsSender,GitServer,Commit,DomainManager,Pipeline,Branch
var WireApp = wire.NewSet(NewPluginManager)
