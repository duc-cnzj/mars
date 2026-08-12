package transformer

import (
	"strings"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
)

// FromChangelog 把 biz.Changelog 转换为 proto ChangelogModel。
func FromChangelog(c *biz.Changelog) *types.ChangelogModel {
	if c == nil {
		return nil
	}
	return &types.ChangelogModel{
		Id:               int32(c.ID),
		Version:          int32(c.Version),
		Username:         c.Username,
		Config:           c.Config,
		ConfigChanged:    c.ConfigChanged,
		ProjectId:        int64(c.ProjectID),
		Project:          FromProject(c.Project),
		Date:             date.ToHumanizeDateTime(&c.CreatedAt),
		GitBranch:        c.GitBranch,
		GitCommit:        c.GitCommit,
		DockerImage:      strings.Join(c.DockerImage, ","),
		EnvValues:        c.EnvValues,
		ExtraValues:      c.ExtraValues,
		FinalExtraValues: c.FinalExtraValues,
		GitCommitWebUrl:  c.GitCommitWebURL,
		GitCommitTitle:   c.GitCommitTitle,
		GitCommitAuthor:  c.GitCommitAuthor,
		GitCommitDate:    date.ToHumanizeDateTime(c.GitCommitDate),
		CreatedAt:        date.ToRFC3339(&c.CreatedAt),
		UpdatedAt:        date.ToRFC3339(&c.UpdatedAt),
		DeletedAt:        date.ToRFC3339(c.DeletedAt),
	}
}
