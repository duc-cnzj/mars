package transformer

import (
	"strings"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
)

func FromProject(project *ent.Project) *types.ProjectModel {
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
		DockerImage:       strings.Join(project.DockerImage, ","),
		PodSelectors:      project.PodSelectors,
		NamespaceId:       int32(project.NamespaceID),
		Atomic:            project.Atomic,
		EnvValues:         project.EnvValues,
		ExtraValues:       project.ExtraValues,
		FinalExtraValues:  strings.Join(project.FinalExtraValues, ","),
		DeployStatus:      project.DeployStatus,
		HumanizeCreatedAt: date.ToHumanizeDatetimeString(&project.CreatedAt),
		HumanizeUpdatedAt: date.ToHumanizeDatetimeString(&project.UpdatedAt),
		ConfigType:        project.ConfigType,
		GitCommitWebUrl:   project.GitCommitWebURL,
		GitCommitTitle:    project.GitCommitTitle,
		GitCommitAuthor:   project.GitCommitAuthor,
		GitCommitDate:     date.ToHumanizeDatetimeString(project.GitCommitDate),
		Version:           int32(project.Version),
		Namespace:         FromNamespace(project.Edges.Namespace),
		CreatedAt:         date.ToRFC3339DatetimeString(&project.CreatedAt),
		UpdatedAt:         date.ToRFC3339DatetimeString(&project.UpdatedAt),
		DeletedAt:         date.ToRFC3339DatetimeString(project.DeletedAt),
	}
}
