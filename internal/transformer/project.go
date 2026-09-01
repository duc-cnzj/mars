package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
)

// FromProject 把 biz.Project 转换为 proto ProjectModel。
func FromProject(project *biz.Project) *types.ProjectModel {
	if project == nil {
		return nil
	}
	return &types.ProjectModel{
		Id:                int32(project.ID),
		Name:              project.Name,
		GitProjectId:      int32(project.GitProjectID),
		GitBranch:         project.GitBranch,
		GitCommit:         project.GitCommit,
		Config:            project.Config,
		OverrideValues:    project.OverrideValues,
		DockerImage:       project.DockerImage,
		PodSelectors:      project.PodSelectors,
		NamespaceId:       int32(project.NamespaceID),
		Atomic:            project.Atomic,
		EnvValues:         project.EnvValues,
		ExtraValues:       project.ExtraValues,
		FinalExtraValues:  project.FinalExtraValues,
		DeployStatus:      project.DeployStatus,
		HumanizeCreatedAt: date.ToHumanizeDateTime(&project.CreatedAt),
		HumanizeUpdatedAt: date.ToHumanizeDateTime(&project.UpdatedAt),
		UpdatedBy:         project.UpdatedBy,
		ConfigType:        project.ConfigType,
		GitCommitWebUrl:   project.GitCommitWebURL,
		GitCommitTitle:    project.GitCommitTitle,
		GitCommitAuthor:   project.GitCommitAuthor,
		GitCommitDate:     date.ToHumanizeDateTime(project.GitCommitDate),
		Version:           int32(project.Version),
		RepoId:            int32(project.RepoID),
		Repo:              FromRepo(project.Repo),
		Namespace:         FromNamespace(project.Namespace),
		CreatedAt:         date.ToRFC3339(&project.CreatedAt),
		UpdatedAt:         date.ToRFC3339(&project.UpdatedAt),
		DeletedAt:         date.ToRFC3339(project.DeletedAt),
	}
}
