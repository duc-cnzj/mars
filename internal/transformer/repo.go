package transformer

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
)

func FromRepo(repo *ent.Repo) *types.RepoModel {
	if repo == nil {
		return nil
	}
	return &types.RepoModel{
		Id:             int32(repo.ID),
		Name:           repo.Name,
		GitProjectId:   repo.GitProjectID,
		GitProjectName: repo.GitProjectName,
		Enabled:        repo.Enabled,
		MarsConfig:     repo.MarsConfig,
		CreatedAt:      date.ToHumanizeDatetimeString(&repo.CreatedAt),
		UpdatedAt:      date.ToHumanizeDatetimeString(&repo.UpdatedAt),
		DeletedAt:      date.ToHumanizeDatetimeString(repo.DeletedAt),
	}
}
