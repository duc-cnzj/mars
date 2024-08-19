package main

import (
	"github.com/duc-cnzj/mars/v4/cmd"
	"github.com/duc-cnzj/mars/v4/internal/logo"

	_ "embed"

	_ "github.com/duc-cnzj/mars/v4/internal/plugins/domainmanager"
	_ "github.com/duc-cnzj/mars/v4/internal/plugins/gitserver/github"
	_ "github.com/duc-cnzj/mars/v4/internal/plugins/gitserver/gitlab"
	_ "github.com/duc-cnzj/mars/v4/internal/plugins/picture"
	_ "github.com/duc-cnzj/mars/v4/internal/plugins/wssender/memory"
	_ "github.com/duc-cnzj/mars/v4/internal/plugins/wssender/nsq"
	_ "github.com/duc-cnzj/mars/v4/internal/plugins/wssender/redis"
)

//go:embed config_example.yaml
var configFile []byte

func main() {
	cmd.Execute(configFile, logo.WithAuthor())
}
