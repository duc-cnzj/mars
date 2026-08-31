package data

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"

	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/member"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/filters"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
)

// toProject 把 ent.Project 转换为 biz.Project（nil 安全），并顺带转换关联的 Namespace/Repo。
func toProject(project *ent.Project) *biz.Project {
	if project == nil {
		return nil
	}
	return &biz.Project{
		ID:               project.ID,
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
		DeletedAt:        project.DeletedAt,
		Name:             project.Name,
		GitProjectID:     project.GitProjectID,
		GitBranch:        project.GitBranch,
		GitCommit:        project.GitCommit,
		Config:           project.Config,
		OverrideValues:   project.OverrideValues,
		DockerImage:      project.DockerImage,
		PodSelectors:     project.PodSelectors,
		Atomic:           project.Atomic,
		DeployStatus:     project.DeployStatus,
		EnvValues:        project.EnvValues,
		ExtraValues:      project.ExtraValues,
		FinalExtraValues: project.FinalExtraValues,
		Version:          project.Version,
		ConfigType:       project.ConfigType,
		GitCommitWebURL:  project.GitCommitWebURL,
		GitCommitTitle:   project.GitCommitTitle,
		GitCommitAuthor:  project.GitCommitAuthor,
		GitCommitDate:    project.GitCommitDate,
		NamespaceID:      project.NamespaceID,
		RepoID:           project.RepoID,
		Namespace:        toNamespace(project.Edges.Namespace),
		Repo:             toRepo(project.Edges.Repo),
		Manifest:         project.Manifest,
	}
}

var _ biz.ProjectRepo = (*projectRepo)(nil)

// projectRepo 是 biz.ProjectRepo 的实现：封装 ent Project 查询与命名空间过滤。
type projectRepo struct {
	logger mlog.Logger

	externalIp string
	data       dataStore
}

// NewProjectRepo 构造项目 repo 实现，注入日志与 dataStore。
func NewProjectRepo(logger mlog.Logger, data dataStore) biz.ProjectRepo {
	return &projectRepo{
		logger:     logger.WithModule("repo/project"),
		externalIp: data.Config().ExternalIp,
		data:       data,
	}
}

// Version 查询项目当前版本号。
func (repo *projectRepo) Version(ctx context.Context, id int) (version int, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/Version")
	defer func() { endSpan(span, err) }()
	get, err := repo.data.DB().Project.Query().Select(
		project.FieldID,
		project.FieldVersion,
	).Where(project.ID(id)).Only(ctx)
	if err != nil {
		return 0, errs.Wrap(err, "query project version")
	}
	return get.Version, nil
}

// projectSearchPred 项目治理关键词搜索谓词：匹配项目名或命名空间名（模糊，不分大小写）。
func projectSearchPred(search string) func(*sql.Selector) {
	return project.Or(
		project.NameContainsFold(search),
		project.HasNamespaceWith(namespace.NameContainsFold(search)),
	)
}

// projectLivenessPred 项目活跃度分类 SQL 谓词：分类键是普通列 updated_at，直接按边界比较。
// 边界由分类基准 now 推导（活跃=updated_at > now-31d；僵尸=<= now-90d；低活跃=两者之间），
// 与 biz.classifyLiveness 的 int(now.Sub(ts).Hours()/24) 阈值数学等价（datetime 整秒存储 +
// DSN loc=Local 使 SQL 边界参数与 Go time.Now() 同墙钟，见 internal/config/config.go DSN）。
// 非法 liveness 值返回恒假谓词，复现旧逻辑「无行命中非法分类」的空列表语义。
func projectLivenessPred(liveness string, now time.Time) func(*sql.Selector) {
	active, zombie := livenessBoundaries(now)
	switch liveness {
	case "active":
		return project.UpdatedAtGT(active)
	case "zombie":
		return project.UpdatedAtLTE(zombie)
	case "dormant":
		return project.And(project.UpdatedAtLTE(active), project.UpdatedAtGT(zombie))
	default:
		// 非法 liveness 值 → 恒假谓词（FALSE），复现旧逻辑「无行命中非法分类」的空列表语义。
		return func(s *sql.Selector) { s.Where(sql.False()) }
	}
}

// ListLivenessPage 分页查询活跃度聚合所需项目（真 SQL 分页）：分类过滤/排序/统计/分页全部
// 下沉 SQL，stats 基于搜索命中全量（无 edges 的干净 query 计数，避免 JOIN/边查询放大，对齐
// namespaceRepo.List 的计数口径），count 为分类过滤后总数（无过滤 = total）。
//
// 排序：updated_at {desc|asc} + id {desc|asc} 决胜键——datetime 整秒精度下同秒多条排序非全序，
// 加 id 保证 LIMIT/OFFSET 翻页不漂移/不重复。字段裁剪沿用旧 ListLiveness（短字段 + 边外键），
// 避免拉 config/override_values 等 longtext/JSON 大列。
func (repo *projectRepo) ListLivenessPage(ctx context.Context, query *biz.LivenessPageQuery) (page *biz.LivenessPageResult, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/ListLivenessPage")
	defer func() { endSpan(span, err) }()
	// 计数/统计用无 edges 干净 query：search 命中全量上做 4 次 COUNT（total + 三分类），
	// 与分页行查询解耦，避免 WithNamespace/WithRepo 的 JOIN 放大计数。
	base := repo.data.DB().Project.Query()
	if query.Search != "" {
		base = base.Where(projectSearchPred(query.Search))
	}
	total, err := base.Clone().Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count project liveness total")
	}
	active, err := base.Clone().Where(projectLivenessPred("active", query.Now)).Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count project liveness active")
	}
	dormant, err := base.Clone().Where(projectLivenessPred("dormant", query.Now)).Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count project liveness dormant")
	}
	zombie, err := base.Clone().Where(projectLivenessPred("zombie", query.Now)).Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count project liveness zombie")
	}
	count := total
	if query.Liveness != "" {
		filtered, err := base.Clone().Where(projectLivenessPred(query.Liveness, query.Now)).Count(ctx)
		if err != nil {
			return nil, errs.Wrap(err, "count project liveness filtered")
		}
		count = filtered
	}
	// 分页行：先叠加搜索 + 分类过滤（*ProjectQuery 上 Where），再 Select 裁剪短字段、
	// 排序（updated_at + id 决胜键）与 Offset/Limit；边（仓库/命名空间）沿用旧 ListLiveness。
	order := sql.OrderDesc()
	if query.Sort == "asc" {
		order = sql.OrderAsc()
	}
	rows := repo.data.DB().Project.Query().WithNamespace().WithRepo()
	if query.Search != "" {
		rows = rows.Where(projectSearchPred(query.Search))
	}
	if query.Liveness != "" {
		rows = rows.Where(projectLivenessPred(query.Liveness, query.Now))
	}
	all, err := rows.
		Select(
			project.FieldID,
			project.FieldName,
			project.FieldUpdatedAt,
			project.FieldDeployStatus,
			project.FieldGitBranch,
			project.FieldGitCommit,
			project.FieldGitCommitTitle,
			project.FieldGitCommitAuthor,
			project.FieldGitCommitDate,
			project.FieldNamespaceID,
			project.FieldRepoID,
		).
		Order(
			sql.OrderByField(project.FieldUpdatedAt, order).ToFunc(),
			sql.OrderByField(project.FieldID, order).ToFunc(),
		).
		Offset(pagination.GetPageOffset(query.Page, query.PageSize)).
		Limit(int(query.PageSize)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list project liveness page")
	}
	return &biz.LivenessPageResult{
		Projects: slice.Map(all, toProject),
		Count:    count,
		Stats:    biz.LivenessStats{Total: total, Active: active, Dormant: dormant, Zombie: zombie},
	}, nil
}

// ListAllProjectBriefs 查询全部项目的精简投影（仅 Name/PodSelectors + 关联 Namespace.Name），
// 供空间资源聚合做 pod→项目归属映射。不投影 config/override_values/env_values/extra_values/
// manifest 等 longtext/JSON 大列——全列拉取在项目多时是显著的传输与解码成本；toProject 复制
// 未选字段为零值，消费端仅读上述三字段，安全。
func (repo *projectRepo) ListAllProjectBriefs(ctx context.Context) (projects []*biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/ListAll")
	defer func() { endSpan(span, err) }()
	all, err := repo.data.DB().Project.Query().
		Select(project.FieldID, project.FieldName, project.FieldPodSelectors).
		WithNamespace(func(q *ent.NamespaceQuery) { q.Select(namespace.FieldName) }).
		Order(ent.Desc(project.FieldID)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list all projects")
	}
	return slice.Map(all, toProject), nil
}

// List 分页查询项目列表；非 admin 只返回其可访问命名空间下的项目，防私有内容泄漏。
func (repo *projectRepo) List(ctx context.Context, input *biz.ListProjectInput) (projects []*biz.Project, pag *pagination.Pagination, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/List")
	defer func() { endSpan(span, err) }()
	query := repo.data.DB().Project.Query().
		WithNamespace().
		Where(filters.IfOrderByDesc("id")(input.OrderByIDDesc))
	// 与 namespaceRepo.List 的访问谓词保持一致：非 admin 只能看到其可访问
	// 命名空间（公开/创建者/成员且私有）下的项目，否则私有命名空间的内容会
	// 通过全局项目列表泄漏（project.List 是唯一无 namespace 过滤的项目入口）。
	if !input.IsAdmin {
		query = query.Where(
			project.HasNamespaceWith(
				namespace.Or(
					namespace.And(
						namespace.HasMembersWith(member.Email(input.Email)),
						namespace.Private(true),
					),
					namespace.Private(false),
					namespace.CreatorEmail(input.Email),
				),
			),
		)
	}
	all := query.Clone().
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		AllX(ctx)
	count := query.Clone().CountX(ctx)
	return slice.Map(all, toProject), pagination.NewPagination(input.Page, input.PageSize, count), nil
}

// Create 新建项目并落库。
func (repo *projectRepo) Create(ctx context.Context, input *biz.CreateProjectInput) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/Create")
	defer func() { endSpan(span, err) }()
	save, err := repo.data.DB().Project.Create().
		SetName(input.Name).
		SetCreator(input.Creator).
		SetGitProjectID(input.GitProjectID).
		SetGitBranch(input.GitBranch).
		SetGitCommit(input.GitCommit).
		SetConfig(input.Config).
		SetNillableAtomic(input.Atomic).
		SetDeployStatus(input.DeployStatus).
		SetConfigType(input.ConfigType).
		SetNamespaceID(input.NamespaceID).
		SetPodSelectors(input.PodSelectors).
		SetRepoID(input.RepoID).
		Save(ctx)
	return toProject(save), errs.Wrap(err, "create project")
}

// UpdateProject 更新项目配置/部署信息。
func (repo *projectRepo) UpdateProject(ctx context.Context, input *biz.UpdateProjectInput) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/UpdateProject")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Project.Query().Where(project.ID(input.ID)).First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "update project")
	}
	save, err := first.Update().
		SetGitBranch(input.GitBranch).
		SetGitCommit(input.GitCommit).
		SetConfig(input.Config).
		SetNillableAtomic(input.Atomic).
		SetConfigType(input.ConfigType).
		SetManifest(input.Manifest).
		SetPodSelectors(input.PodSelectors).
		SetDockerImage(input.DockerImage).
		SetGitCommitTitle(input.GitCommitTitle).
		SetGitCommitWebURL(input.GitCommitWebURL).
		SetGitCommitAuthor(input.GitCommitAuthor).
		SetNillableGitCommitDate(input.GitCommitDate).
		SetExtraValues(input.ExtraValues).
		SetFinalExtraValues(input.FinalExtraValues).
		SetEnvValues(input.EnvValues).
		SetOverrideValues(input.OverrideValues).
		Save(ctx)
	return toProject(save), errs.Wrap(err, "update project")
}

// Show 查询单个项目并预加载关联仓库与命名空间。
func (repo *projectRepo) Show(ctx context.Context, id int) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/Show")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Project.
		Query().
		WithRepo().
		WithNamespace().
		Where(project.ID(id)).
		First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query project")
	}
	return toProject(first), nil
}

// Delete 按 ID 删除项目。
func (repo *projectRepo) Delete(ctx context.Context, id int) (err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/Delete")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(repo.data.DB().Project.DeleteOneID(id).Exec(ctx), "delete project")
}

// UpdateStatusByVersion 校验版本匹配后更新部署状态并递增版本号（乐观锁防并发覆盖）。
func (repo *projectRepo) UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/UpdateStatusByVersion")
	defer func() { endSpan(span, err) }()
	if _, err := repo.FindByVersion(ctx, id, version); err != nil {
		return nil, err
	}
	save, err := repo.data.DB().Project.UpdateOneID(id).SetDeployStatus(status).SetVersion(version + 1).Save(ctx)
	return toProject(save), errs.Wrap(err, "update project status by version")
}

// FindByVersion 按 ID 与版本号精确查找项目（用于乐观锁校验）。
func (repo *projectRepo) FindByVersion(ctx context.Context, id, version int) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/FindByVersion")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Project.Query().Where(project.ID(id), project.Version(version)).First(ctx)
	return toProject(first), errs.Wrap(err, "find project by version")
}

// UpdateVersion 直接覆盖项目版本号。
func (repo *projectRepo) UpdateVersion(ctx context.Context, id int, version int) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/UpdateVersion")
	defer func() { endSpan(span, err) }()
	save, err := repo.data.DB().Project.UpdateOneID(id).SetVersion(version).Save(ctx)
	return toProject(save), errs.Wrap(err, "update project version")
}

// UpdateDeployStatus 仅更新项目的部署状态。
func (repo *projectRepo) UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/UpdateDeployStatus")
	defer func() { endSpan(span, err) }()
	save, err := repo.data.DB().Project.UpdateOneID(id).SetDeployStatus(status).Save(ctx)
	return toProject(save), errs.Wrap(err, "update deploy status")
}

// ListByDeployStatus 按部署状态集合过滤项目并携带 namespace 信息，
// cron FixDeployStatus 用 helm 实测状态修复失败/未知项目。
func (repo *projectRepo) ListByDeployStatus(ctx context.Context, statuses ...types.Deploy) (projects []*biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/ListByDeployStatus")
	defer func() { endSpan(span, err) }()
	all, err := repo.data.DB().Project.Query().
		WithNamespace(func(query *ent.NamespaceQuery) {
			query.Select(namespace.FieldID, namespace.FieldName)
		}).
		Where(project.DeployStatusIn(statuses...)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list project by deploy status")
	}
	return slice.Map(all, toProject), nil
}

// FindByName 按名称与命名空间 ID 查找项目。
func (repo *projectRepo) FindByName(ctx context.Context, name string, nsID int) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/FindByName")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Project.Query().Where(project.Name(name), project.NamespaceID(nsID)).First(ctx)
	return toProject(first), errs.Wrap(err, "find project by name")
}

// FindProjectsByIDs 按主键批量取项目。endpoint 编排依赖项目的 Name 与 Manifest
// 来匹配集群内对象，这是纯数据读取，不包含任何编排逻辑。
func (repo *projectRepo) FindProjectsByIDs(ctx context.Context, ids ...int) (projects []*biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/FindProjectsByIDs")
	defer func() { endSpan(span, err) }()
	if len(ids) == 0 {
		return nil, nil
	}
	all, err := repo.data.DB().Project.Query().
		Where(project.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "find projects by ids")
	}
	return slice.Map(all, toProject), nil
}
