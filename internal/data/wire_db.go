package data

import (
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/google/wire"
)

// WireDB 提供 *dataImpl 实体与 Data/dataStore 及多个窄端口的绑定：
// wire 按类型同一性匹配（不做可赋值推断），构造器返回具体类型 *dataImpl，
// 故必须显式 Bind 才能解析到 Data（启动门面）、dataStore（repo 存储端口）、
// MinioGetter/DBGetter（cmd 装配期惰性取数窄端口）与 biz.AuthConfigProvider
// （AuthBiz 配置取数窄接口，dataImpl 是天然实现）。
var WireDB = wire.NewSet(
	NewData,
	wire.Bind(new(Data), new(*dataImpl)),
	wire.Bind(new(dataStore), new(*dataImpl)),
	wire.Bind(new(MinioGetter), new(*dataImpl)),
	wire.Bind(new(DBGetter), new(*dataImpl)),
	wire.Bind(new(biz.AuthConfigProvider), new(*dataImpl)),
)
