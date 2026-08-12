package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
)

// FromRepo 把 biz.Repo 转换为 proto RepoModel。
func FromRepo(repo *biz.Repo) *types.RepoModel {
	if repo == nil {
		return nil
	}
	return &types.RepoModel{
		Id:             int32(repo.ID),
		Name:           repo.Name,
		GitProjectId:   repo.GitProjectID,
		GitProjectName: repo.GitProjectName,
		Enabled:        repo.Enabled,
		MarsConfig:     repo.GetMarsConfig(),
		NeedGitRepo:    repo.NeedGitRepo,
		Description:    repo.Description,
		CreatedAt:      date.ToHumanizeDateTime(&repo.CreatedAt),
		UpdatedAt:      date.ToHumanizeDateTime(&repo.UpdatedAt),
		DeletedAt:      date.ToHumanizeDateTime(repo.DeletedAt),
	}
}
