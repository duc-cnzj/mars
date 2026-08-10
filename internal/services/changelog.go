package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/changelog"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
)

var _ changelog.ChangelogServer = (*changelogSvc)(nil)

// changelogSvc 是 changelog.ChangelogServer 的 gRPC 实现：按项目查询最近变更记录，
// 经 access 校验访问权限，由 NewChangelogSvc 构造。
type changelogSvc struct {
	changelog.UnimplementedChangelogServer

	logger    mlog.Logger
	clBiz     biz.ChangelogBiz
	accessBiz biz.AccessBiz
}

// ChangelogSvcDeps 收口 NewChangelogSvc 的构造依赖，由 wire 按字段注入。
type ChangelogSvcDeps struct {
	ClBiz     biz.ChangelogBiz
	Logger    mlog.Logger
	AccessBiz biz.AccessBiz
}

// NewChangelogSvc 收口变更日志服务的构造依赖，由 wire 按字段注入。
func NewChangelogSvc(deps ChangelogSvcDeps) changelog.ChangelogServer {
	logger := deps.Logger.WithModule("services/changelog")
	return &changelogSvc{
		clBiz:     deps.ClBiz,
		logger:    logger,
		accessBiz: deps.AccessBiz,
	}
}

// FindLastChangelogsByProjectID 查询指定项目最近 5 条变更记录（可按 only_changed
// 过滤只返回版本有变化的），响应前做项目级访问控制。
func (c *changelogSvc) FindLastChangelogsByProjectID(ctx context.Context, request *changelog.FindLastChangelogsByProjectIDRequest) (*changelog.FindLastChangelogsByProjectIDResponse, error) {
	// 与 project.MemoryCpuAndEndpoints / metrics / endpoint 对齐：changelog 携带
	// 完整部署配置(Config)与环境变量(EnvValues)，属于私有命名空间的项目必须做
	// 命名空间级访问控制，否则任意登录用户可枚举 ProjectID 读到私有项目的密钥。
	if _, err := c.accessBiz.RequireProjectAccess(ctx, int(request.ProjectId)); err != nil {
		return nil, err
	}

	logs, err := c.clBiz.FindLastChangelogsByProjectID(ctx, &biz.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        request.OnlyChanged,
		ProjectID:          int(request.ProjectId),
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              5,
	})
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}

	return &changelog.FindLastChangelogsByProjectIDResponse{
		Items: slice.Map(logs, transformer.FromChangelog),
	}, nil
}
