package bootstrappers

import (
	"github.com/DuC-cnZj/mars/pkg/contracts"
	"github.com/xanzy/go-gitlab"
)

type GitlabBootstrapper struct{}

func (g *GitlabBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	client, err := gitlab.NewClient("TKGtsB1zcYRuqFvyawst", gitlab.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		return err
	}

	app.SetGitlabClient(client)

	return nil
}
