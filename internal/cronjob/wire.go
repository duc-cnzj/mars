package cronjob

import "github.com/google/wire"

// WireCronJob 是 cronjob 包的 wire provider 集合。
// 惰性插件闭包（getCerts）由组合根（cmd）注入。
var WireCronJob = wire.NewSet(NewTasks)
