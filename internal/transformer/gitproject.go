package transformer

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
)

func FromGitProject(g *ent.GitProject) *types.GitProjectModel {
	if g == nil {
		return nil
	}
	return &types.GitProjectModel{
		Id:            int32(g.ID),
		DefaultBranch: g.DefaultBranch,
		Name:          g.Name,
		GitProjectId:  int32(g.GitProjectID),
		Enabled:       g.Enabled,
		GlobalEnabled: g.GlobalEnabled,
		GlobalConfig:  g.GlobalConfig,
		CreatedAt:     date.ToRFC3339DatetimeString(&g.CreatedAt),
		UpdatedAt:     date.ToRFC3339DatetimeString(&g.UpdatedAt),
		DeletedAt:     date.ToRFC3339DatetimeString(g.DeletedAt),
	}
}
