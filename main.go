package main

import (
	_ "embed"
	"math/rand"
	"time"

	"github.com/duc-cnzj/mars/cmd"

	_ "github.com/duc-cnzj/mars/plugins/domain_manager"
	_ "github.com/duc-cnzj/mars/plugins/git_server/github"
	_ "github.com/duc-cnzj/mars/plugins/git_server/gitlab"
	_ "github.com/duc-cnzj/mars/plugins/picture"
	_ "github.com/duc-cnzj/mars/plugins/wssender/memory"
	_ "github.com/duc-cnzj/mars/plugins/wssender/nsq"
	_ "github.com/duc-cnzj/mars/plugins/wssender/redis"
)

//go:embed config_example.yaml
var configFile []byte

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	cmd.Execute(configFile)
}
