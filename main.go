package main

import (
	_ "embed"

	"github.com/duc-cnzj/mars/cmd"
	_ "github.com/duc-cnzj/mars/internal/plugins/docker"
	_ "github.com/duc-cnzj/mars/internal/plugins/domain_resolver"
	_ "github.com/duc-cnzj/mars/internal/plugins/wssender"
)

//go:embed config_example.yaml
var configFile []byte

func main() {
	cmd.Execute(configFile)
}
