package docker

import (
	"strings"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/heroku/docker-registry-client/registry"
)

var name = "docker_default"

func init() {
	p := &dockerPlugin{}
	plugins.RegisterPlugin(p.Name(), p)
}

type dockerPlugin struct{}

func (d *dockerPlugin) Name() string {
	return name
}

func (d *dockerPlugin) Initialize(args map[string]interface{}) error {
	mlog.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

func (d *dockerPlugin) Destroy() error {
	mlog.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

func (d *dockerPlugin) ImageNotExists(repo, tag string) bool {
	for _, s := range app.Config().ImagePullSecrets {
		server := s.Server
		server = strings.TrimPrefix(strings.TrimPrefix(server, "https://"), "http://")
		if strings.HasPrefix(repo, server) {
			hub, err := registry.New("https://"+server, s.Username, s.Password)
			if err != nil {
				mlog.Error(err)
				return false
			}

			if tags, err := hub.Tags(strings.TrimPrefix(strings.TrimPrefix(repo, server), "/")); err == nil {
				for _, t := range tags {
					if t == tag {
						return false
					}
				}
				return true
			}
		}
	}

	return false
}
