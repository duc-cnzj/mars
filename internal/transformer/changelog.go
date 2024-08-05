package transformer

import (
	"strings"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
)

func FromChangeLog(c *repo.Changelog) *types.ChangelogModel {
	if c == nil {
		return nil
	}
	return &types.ChangelogModel{
		Id:               int32(c.ID),
		Version:          int32(c.Version),
		Username:         c.Username,
		Manifest:         c.Manifest,
		Config:           c.Config,
		ConfigChanged:    c.ConfigChanged,
		ProjectId:        int64(c.ProjectID),
		GitProjectId:     int64(c.GitProjectID),
		Project:          FromProject(c.Project),
		GitProject:       FromGitProject(c.GitProject),
		Date:             date.ToHumanizeDatetimeString(&c.CreatedAt),
		ConfigType:       c.ConfigType,
		GitBranch:        c.GitBranch,
		GitCommit:        c.GitCommit,
		DockerImage:      strings.Join(c.DockerImage, ","),
		EnvValues:        c.EnvValues,
		ExtraValues:      c.ExtraValues,
		FinalExtraValues: c.FinalExtraValues,
		GitCommitWebUrl:  c.GitCommitWebURL,
		GitCommitTitle:   c.GitCommitTitle,
		GitCommitAuthor:  c.GitCommitAuthor,
		GitCommitDate:    date.ToHumanizeDatetimeString(c.GitCommitDate),
		CreatedAt:        date.ToRFC3339DatetimeString(&c.CreatedAt),
		UpdatedAt:        date.ToRFC3339DatetimeString(&c.UpdatedAt),
		DeletedAt:        date.ToRFC3339DatetimeString(c.DeletedAt),
	}
}
