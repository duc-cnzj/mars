// Package plugins 是插件装配清单：单一事实来源。
//
// 新增插件在此加一行 blank import 并注明注册名，main 只需 import 本包。
// 每个插件包的 init() 负责自注册到 app 注册表——装配层开放封闭：
// 新增插件不改任何现有代码（main.go 与 app 包均不动）。
package plugins

import (
	_ "github.com/duc-cnzj/mars/v6/internal/plugins/domainmanager"    // 注册：default_domain_manager / manual_domain_manager / cert-manager_domain_manager / sync_secret_domain_manager
	_ "github.com/duc-cnzj/mars/v6/internal/plugins/gitserver/gitlab" // 注册：gitlab
	_ "github.com/duc-cnzj/mars/v6/internal/plugins/picture"          // 注册：picture_bing / picture_cartoon
	_ "github.com/duc-cnzj/mars/v6/internal/plugins/wssender/memory"  // 注册：ws_sender_memory
	_ "github.com/duc-cnzj/mars/v6/internal/plugins/wssender/nsq"     // 注册：ws_sender_nsq
	_ "github.com/duc-cnzj/mars/v6/internal/plugins/wssender/redis"   // 注册：ws_sender_redis
)
