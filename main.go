package main

import (
	_ "embed"

	_ "github.com/duc-cnzj/mars/plugins/domain_resolver"
	_ "github.com/duc-cnzj/mars/plugins/git_server/github"
	_ "github.com/duc-cnzj/mars/plugins/git_server/gitlab"
	_ "github.com/duc-cnzj/mars/plugins/picture"
	_ "github.com/duc-cnzj/mars/plugins/wssender"

	"github.com/duc-cnzj/mars/cmd"
)

//go:embed config_example.yaml
var configFile []byte

func main() {
	cmd.Execute(configFile)
}
