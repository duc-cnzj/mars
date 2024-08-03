package transformer

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent"
)

func FromRepo(repo *ent.Repo) *types.RepoModel {
	if repo == nil {
		return nil
	}
	return &types.RepoModel{
		Id:           int64(repo.ID),
		GitProjectId: repo.GitProjectID,
		Enabled:      repo.Enabled,
		MarsConfig:   repo.MarsConfig,
	}
}
