package utils

import (
	"strings"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/heroku/docker-registry-client/registry"
)

func ImageNotExists(repo, tag string) bool {
	for _, s := range Config().ImagePullSecrets {
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
