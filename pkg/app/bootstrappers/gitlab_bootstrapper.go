package bootstrappers

import (
	"github.com/duc-cnzj/mars/pkg/contracts"
	"github.com/xanzy/go-gitlab"
)

type GitlabBootstrapper struct{}

func (g *GitlabBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	client, err := gitlab.NewClient(app.Config().GitlabToken, gitlab.WithBaseURL(app.Config().GitlabBaseURL))
	if err != nil {
		return err
	}

	app.SetGitlabClient(client)

	return nil
}
