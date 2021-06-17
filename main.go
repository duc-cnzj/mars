package main

import (
	_ "embed"

	"github.com/DuC-cnZj/mars/cmd"
)

//go:embed config_example.yaml
var configFile []byte

func main() {
	cmd.Execute(configFile)
}
