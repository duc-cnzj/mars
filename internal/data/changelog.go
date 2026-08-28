package data

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/biz"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/changelog"
	"github.com/duc-cnzj/mars/v6/internal/data/filters"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
)

// toChangeLog 把 ent.Changelog 转换为 biz.Changelog（nil 安全）。
func toChangeLog(c *ent.Changelog) *biz.Changelog {
	if c == nil {
		return nil
	}
	return &biz.Changelog{
		ID:               c.ID,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		DeletedAt:        c.DeletedAt,
		Version:          c.Version,
		Username:         c.Username,
		Config:           c.Config,
		GitBranch:        c.GitBranch,
		GitCommit:        c.GitCommit,
		DockerImage:      c.DockerImage,
		EnvValues:        c.EnvValues,
		ExtraValues:      c.ExtraValues,
		FinalExtraValues: c.FinalExtraValues,
		GitCommitWebURL:  c.GitCommitWebURL,
		GitCommitTitle:   c.GitCommitTitle,
		GitCommitAuthor:  c.GitCommitAuthor,
		GitCommitDate:    c.GitCommitDate,
		ConfigChanged:    c.ConfigChanged,
		ProjectID:        c.ProjectID,
		Project:          toProject(c.Edges.Project),
	}
}

var _ biz.ChangelogRepo = (*changelogRepo)(nil)

// changelogRepo 是 changelog 的持久化实现：负责变更记录的创建与按项目查询。
type changelogRepo struct {
	logger mlog.Logger
	data   dataStore
}

// NewChangelogRepo 构造 changelog repo。
func NewChangelogRepo(logger mlog.Logger, data dataStore) biz.ChangelogRepo {
	return &changelogRepo{logger: logger.WithModule("repo/changelog"), data: data}
}

// Create 创建一条变更记录并返回转换后的领域对象。
func (c *changelogRepo) Create(ctx context.Context, input *biz.CreateChangeLogInput) (out *biz.Changelog, err error) {
	ctx, span := tracer.Start(ctx, "changelogRepo/Create")
	defer func() { endSpan(span, err) }()
	db := c.data.DB()
	save, err := db.Changelog.Create().
		SetVersion(input.Version).
		SetUsername(input.Username).
		SetConfig(input.Config).
		SetGitBranch(input.GitBranch).
		SetGitCommit(input.GitCommit).
		SetDockerImage(input.DockerImage).
		SetEnvValues(input.EnvValues).
		SetExtraValues(input.ExtraValues).
		SetFinalExtraValues(input.FinalExtraValues).
		SetGitCommitWebURL(input.GitCommitWebURL).
		SetGitCommitTitle(input.GitCommitTitle).
		SetGitCommitAuthor(input.GitCommitAuthor).
		SetNillableGitCommitDate(input.GitCommitDate).
		SetConfigChanged(input.ConfigChanged).
		SetProjectID(input.ProjectID).
		Save(ctx)
	return toChangeLog(save), errs.Wrap(err, "create changelog")
}

// FindLastChangelogsByProjectID 按项目查询最近变更记录，支持 config_changed 过滤与 version 降序。
func (c *changelogRepo) FindLastChangelogsByProjectID(ctx context.Context, input *biz.FindLastChangelogsByProjectIDChangeLogInput) (out []*biz.Changelog, err error) {
	ctx, span := tracer.Start(ctx, "changelogRepo/FindLastChangelogsByProjectID")
	defer func() { endSpan(span, err) }()
	db := c.data.DB()
	var onlyChanged *bool
	if input.OnlyChanged {
		onlyChanged = lo.ToPtr(true)
	}
	all, err := db.Changelog.Query().
		Where(
			filters.IfBool(changelog.FieldConfigChanged)(onlyChanged),
			changelog.ProjectID(input.ProjectID),
			filters.IfOrderByDesc(changelog.FieldVersion)(input.OrderByVersionDesc),
		).
		Limit(input.Limit).
		All(ctx)
	return slice.Map(all, toChangeLog), errs.Wrap(err, "find last changelogs")
}

// FindLastChangeByProjectID 查询项目最近的一条变更记录（按 ID 降序）。
func (c *changelogRepo) FindLastChangeByProjectID(ctx context.Context, projectID int) (out *biz.Changelog, err error) {
	ctx, span := tracer.Start(ctx, "changelogRepo/FindLastChangeByProjectID")
	defer func() { endSpan(span, err) }()
	db := c.data.DB()
	first, err := db.Changelog.Query().Where(changelog.ProjectID(projectID)).Order(ent.Desc(changelog.FieldID)).First(ctx)
	return toChangeLog(first), errs.Wrap(err, "find last change")
}

// CountByProjectIDs 按项目 ID 集合聚合各项目变更记录数（GROUP BY project_id），
// 供项目活跃度清单汇总 deploy_count。ids 为空返回空 map。
func (c *changelogRepo) CountByProjectIDs(ctx context.Context, ids ...int) (out map[int]int, err error) {
	ctx, span := tracer.Start(ctx, "changelogRepo/CountByProjectIDs")
	defer func() { endSpan(span, err) }()
	if len(ids) == 0 {
		return map[int]int{}, nil
	}
	var rows []struct {
		ProjectID int `json:"project_id"`
		Count     int `json:"count"`
	}
	err = c.data.DB().Changelog.Query().
		Where(changelog.ProjectIDIn(ids...)).
		GroupBy(changelog.FieldProjectID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, errs.Wrap(err, "count changelog by project ids")
	}
	out = make(map[int]int, len(rows))
	for _, row := range rows {
		out[row.ProjectID] = row.Count
	}
	return out, nil
}
