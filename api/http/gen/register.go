package gen

// blank-import 注册所有 mars 服务包，GlobalRegistry 才能解析消息的 Go 类型。
// 必须保持 package gen：Generate/goType 依赖这些包 init 后写入
// protoregistry.GlobalFiles/GlobalTypes 的副作用，挪出本包就不生效。
import (
	_ "github.com/duc-cnzj/mars/api/v6/proto/auth"
	_ "github.com/duc-cnzj/mars/api/v6/proto/changelog"
	_ "github.com/duc-cnzj/mars/api/v6/proto/cluster"
	_ "github.com/duc-cnzj/mars/api/v6/proto/container"
	_ "github.com/duc-cnzj/mars/api/v6/proto/endpoint"
	_ "github.com/duc-cnzj/mars/api/v6/proto/event"
	_ "github.com/duc-cnzj/mars/api/v6/proto/file"
	_ "github.com/duc-cnzj/mars/api/v6/proto/git"
	_ "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	_ "github.com/duc-cnzj/mars/api/v6/proto/namespace"
	_ "github.com/duc-cnzj/mars/api/v6/proto/picture"
	_ "github.com/duc-cnzj/mars/api/v6/proto/project"
	_ "github.com/duc-cnzj/mars/api/v6/proto/repo"
	_ "github.com/duc-cnzj/mars/api/v6/proto/settings"
	_ "github.com/duc-cnzj/mars/api/v6/proto/token"
	_ "github.com/duc-cnzj/mars/api/v6/proto/user"
	_ "github.com/duc-cnzj/mars/api/v6/proto/version"
)
