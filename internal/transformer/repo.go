package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"google.golang.org/protobuf/proto"
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
		AiAgent:        repo.AIAgent,
		CreatedAt:      date.ToHumanizeDateTime(&repo.CreatedAt),
		UpdatedAt:      date.ToHumanizeDateTime(&repo.UpdatedAt),
		DeletedAt:      date.ToHumanizeDateTime(repo.DeletedAt),
	}
}

// WithoutHibernate 返回 m 的副本并抹掉 mars_config.hibernate（不改原对象）。
// 导出/导入暂不支持定时休眠：导出文件不带该字段，导入侧也不认它，避免把仓库级休眠配置
// （副本字段/时间窗引用）跨环境搬运后留下悬空引用。
// 必须克隆后再改：m.MarsConfig 常引用 biz 层共享的配置指针，就地置空会污染调用方。
func WithoutHibernate(m *types.RepoModel) *types.RepoModel {
	if m == nil || m.MarsConfig == nil || m.MarsConfig.Hibernate == nil {
		return m
	}
	out := proto.Clone(m).(*types.RepoModel)
	out.MarsConfig.Hibernate = nil
	return out
}
