package main

import (
	"github.com/duc-cnzj/mars/v6/cmd"
	"github.com/duc-cnzj/mars/v6/internal/logo"

	_ "embed"

	// 空导入 internal/plugins：触发各插件子包 init() 注册内置插件实现（域名管理/图片/代码库/WebSocket 发送等），供应用装配。
	_ "github.com/duc-cnzj/mars/v6/internal/plugins"
)

// configFile 是随二进制嵌入的默认配置文件内容，作为命令启动时的默认配置。
//
//go:embed config_example.yaml
var configFile []byte

// main 是 mars 的入口函数，加载默认配置并启动命令行应用。
func main() {
	cmd.Execute(configFile, logo.WithAuthor())
}
