package main

import (
	_ "embed"

	"github.com/duc-cnzj/mars/cmd"

	_ "github.com/duc-cnzj/mars/plugins/domain_manager"
	_ "github.com/duc-cnzj/mars/plugins/git_server/github"
	_ "github.com/duc-cnzj/mars/plugins/git_server/gitlab"
	_ "github.com/duc-cnzj/mars/plugins/picture"
	_ "github.com/duc-cnzj/mars/plugins/wssender"
)

//go:embed config_example.yaml
var configFile []byte

func main() {
	cmd.Execute(configFile)
}
