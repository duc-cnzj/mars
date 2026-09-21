package data

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"

	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/member"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
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
		UpdatedBy:        project.UpdatedBy,
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
// 下沉 SQL，stats 基于搜索命中全量（不带边的干净 query 计数，避免行侧 Order/Select 修饰符
// 渗进 COUNT，对齐 namespaceRepo.List 的计数口径），count 为分类过滤后总数（无过滤 = total）。
//
// 排序：updated_at {desc|asc} + id {desc|asc} 决胜键——datetime 整秒精度下同秒多条排序非全序，
// 加 id 保证 LIMIT/OFFSET 翻页不漂移/不重复。字段裁剪沿用旧 ListLiveness（短字段 + 边外键），
// 避免拉 config/override_values 等 longtext/JSON 大列。
func (repo *projectRepo) ListLivenessPage(ctx context.Context, query *biz.LivenessPageQuery) (page *biz.LivenessPageResult, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/ListLivenessPage")
	defer func() { endSpan(span, err) }()
	// 计数/统计用不带边的干净 query：search 命中全量上做 4 次 COUNT（total + 三分类），
	// 与分页行查询解耦——COUNT 会原样带上行查询的 Order/Select 修饰符，分开取才不被污染
	// （边的 eager load 是独立往返，本就不参与 COUNT）。
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
			project.FieldUpdatedBy,
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
						// ⚠️ member.DeletedAtIsNil() 必须显式带：成员被移出空间走 Member.Delete()
						// → SoftDeleteMixin 钩子转软删，而 HasMembersWith 的裸 sql.Selector 子查询
						// 不被 ent Interceptor 覆盖；与 namespaceRepo.List 同款谓词，漏掉即由
						// 全局项目列表泄漏私有空间内容（回归见
						// TestProjectRepoList_AccessFilter_RemovedMemberSoftDeleted）。
						namespace.HasMembersWith(member.DeletedAtIsNil(), member.Email(input.Email)),
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
		SetUpdatedBy(input.UpdatedBy).
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
		SetUpdatedBy(input.UpdatedBy).
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
//
// 本方法与下方 UpdateVersion / UpdateDeployStatus **刻意不同**：那两者是纯写路径（直接
// UpdateOneID），才必须自带 DeletedAtIsNil 谓词；本条属「读后写」家族，软删过滤由前置的
// FindByVersion 那次 SELECT 经软删拦截器完成（写软删行会在读阶段拿到 404，UPDATE 根本到不了），
// 故此处**不加**谓词并非漏写。副作用是留下与 RestoreDeleted 同类的 TOCTOU 窗口：读通过后、
// 写提交前若并发软删该行，UPDATE 仍会落到已软删行上——要根治得把存活条件写进 UPDATE 自身。
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
//
// 显式带 project.DeletedAtIsNil()：SoftDeleteMixin 的拦截器只作用于 SELECT、钩子只挂删除操作，
// UPDATE 路径没有任何自动软删过滤，谓词是承重代码——漏掉即让「写已软删行」静默成功。带上后，
// 写软删行命中 0 行，经 sqlgraph ensureExists 复查报 NotFound（404）而非静默改写。
func (repo *projectRepo) UpdateVersion(ctx context.Context, id int, version int) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/UpdateVersion")
	defer func() { endSpan(span, err) }()
	save, err := repo.data.DB().Project.UpdateOneID(id).
		Where(project.DeletedAtIsNil()).
		SetVersion(version).
		Save(ctx)
	return toProject(save), errs.Wrap(err, "update project version")
}

// UpdateDeployStatus 仅更新项目的部署状态。
//
// 显式带 project.DeletedAtIsNil()：原因同 UpdateVersion——UPDATE 不受软删拦截器/钩子约束，
// 缺谓词会把「更新已软删项目」变成静默成功。
func (repo *projectRepo) UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/UpdateDeployStatus")
	defer func() { endSpan(span, err) }()
	save, err := repo.data.DB().Project.UpdateOneID(id).
		Where(project.DeletedAtIsNil()).
		SetDeployStatus(status).
		Save(ctx)
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

// FindDeletedByName 按名称 + 命名空间查询**已软删**的项目，供超管恢复被误删的项目使用。
// 同名项目可能被反复"删除→重建→再删除"，故按 deleted_at 倒序取最近一次删除的那行。
//
// id 是必须的第二决胜键：deleted_at 是 MySQL datetime（秒级），"删→同名重建→再删"落在同一秒
// 时两行时间戳相同，仅按它排序取哪行**不确定**——取错会让旧记录转在册，随后新记录被同名在册
// 检查永久挡死。id 越大者创建越晚，同秒内也就是后删的那条，故同样倒序。「字段 + 主键」双键定序
// 与 namespace.go 的 renumberFavoriteSortOrders（(sort_order, id) 稳定序）同理，那边写在单个
// Order(f1, f2) 调用内，此处分两次 Order——语义等价（ent 的 Order 是 append 而非覆盖）。
// 必须绕过软删拦截器：拦截器会给查询补 deleted_at IS NULL，与「只取软删行」自相矛盾。
func (repo *projectRepo) FindDeletedByName(ctx context.Context, name string, nsID int) (proj *biz.Project, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/FindDeletedByName")
	defer func() { endSpan(span, err) }()
	first, err := repo.data.DB().Project.Query().
		Where(
			project.Name(name),
			project.NamespaceID(nsID),
			project.DeletedAtNotNil(),
		).
		Order(ent.Desc(project.FieldDeletedAt)).
		Order(ent.Desc(project.FieldID)).
		First(mixin.SkipSoftDelete(ctx))
	return toProject(first), errs.Wrap(err, "find deleted project by name")
}

// ListAdminDeletedPage 分页查询「可恢复的已删除项目」（超管恢复页的列表），并携带命名空间边
// ——恢复请求按「空间名 + 项目名」定位，行内缺了空间边就是一条无法恢复的记录。
//
// 「可恢复」由两个条件定义：项目已软删 + 所属空间仍存活。后者与 projectRepo.RestoreDeleted 的
// 前置校验（空间存活，否则 400）严格对应——列表即「Restore 会接受什么」，不多不少。
//
// 刻意**不**加 deleted_with_namespace=false：该条件被「空间存活」严格蕴含。级联标记只在
// namespaceRepo.Delete 软删空间的同一事务内置 true，也只在 namespaceRepo.RestoreDeleted 恢复
// 空间的同一事务内清 false，故「标记为 true 且空间存活」不可达；而「空间已删」的行本就已被
// 排除。多写一个恒不改变结果集的谓词属于冗余条件，会让读者误以为它承担了过滤职责。
// （注意其中一类行确实只有本条件能挡：先单独删项目、后删其空间时，项目的级联标记仍是 false
// ——namespaceRepo.Delete 只标记存活项目——但它所属空间已删，同样被「空间存活」排除。）
//
// HasNamespaceWith 产出的是裸 sql.Selector 子查询，SoftDeleteMixin 的拦截器进不去，故命名空间
// 侧的 DeletedAtIsNil 必须显式写出来（此处本就走 SkipSoftDelete，更不能指望拦截器兜底）。
//
// 排序：deleted_at 倒序 + id 决胜键。「最近删除」是页面的核心语义，而 deleted_at 是 MySQL
// datetime（秒级），同秒多条时仅按它排序非全序，LIMIT/OFFSET 翻页会漂移/重复；id 倒序即同秒内
// 后删的那条在前，与 FindDeletedByName 的双键定序同理。
//
// 必须绕过软删拦截器：拦截器会给查询补 deleted_at IS NULL，与「只取软删行」自相矛盾。
func (repo *projectRepo) ListAdminDeletedPage(ctx context.Context, query *biz.ProjectDeletedListPageQuery) (page *biz.ProjectDeletedListPageResult, err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/ListAdminDeletedPage")
	defer func() { endSpan(span, err) }()
	ctx = mixin.SkipSoftDelete(ctx)
	base := repo.data.DB().Project.Query().
		Where(
			project.DeletedAtNotNil(),
			project.HasNamespaceWith(namespace.DeletedAtIsNil()),
		)
	if query.Search != "" {
		base = base.Where(projectSearchPred(query.Search))
	}
	// 计数用不带边的干净 query（对齐 ListLivenessPage 口径）：COUNT 会原样带上行查询的
	// Order/Select 修饰符，故与分页行分开取；WithNamespace 的 eager load 是独立往返、
	// 本就不参与 COUNT，挂上也不会放大。
	count, err := base.Clone().Count(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "count deleted projects")
	}
	// 字段裁剪：只取列表要展示的列 + 空间边外键，避免拉 config/override_values/manifest 等
	// longtext/JSON 大列；未选字段经 toProject 复制为零值，消费端不读它们。
	all, err := base.Clone().
		WithNamespace().
		Select(
			project.FieldID,
			project.FieldName,
			project.FieldNamespaceID,
			project.FieldUpdatedBy,
			project.FieldCreatedAt,
			project.FieldUpdatedAt,
			project.FieldDeletedAt,
		).
		Order(ent.Desc(project.FieldDeletedAt)).
		Order(ent.Desc(project.FieldID)).
		Offset(pagination.GetPageOffset(query.Page, query.PageSize)).
		Limit(int(query.PageSize)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list deleted projects page")
	}
	return &biz.ProjectDeletedListPageResult{Projects: slice.Map(all, toProject), Count: count}, nil
}

// RestoreDeleted 恢复软删的项目：清空 deleted_at，并把 deploy_status 重置为 StatusUnknown。
//
// 重置 deploy_status 的原因：删除项目时 helm release 已被物理卸载，库里残留的"已部署"
// 是对不存在资源的错误描述；StatusUnknown 正是 ReleaseStatus 对不存在 release 的返回值，
// 与"记录已恢复、请重新部署"的语义一致。
//
// 恢复前校验所属命名空间未被软删：把项目复活到已删空间下等于制造脏数据——列表按空间聚合、
// 重新部署都需要一个存在的环境。此时应引导调用方先恢复空间（连带恢复整批级联项目）。
//
// 本路径**不写** deleted_with_namespace：该标记是"随空间级联删除"的批次标识，不变量为「为
// true ⟺ 项目正处于随空间级联软删的状态」——而下面的空间存活校验已把标记必为 true 的行全部
// 挡在 400 之外（标记为 true 只可能由 namespaceRepo.Delete 在软删空间的**同一事务**写入），
// 故能走到 UPDATE 的行该列必然已是 false，写回是恒等操作。清标记由两级恢复中真正需要它的那条
// 承担：namespaceRepo.RestoreDeleted 的批量 UPDATE（它才是会碰到 true 的地方）。
// （本列上线前的存量行保持列默认值 false，见迁移文件
// 20260921012959_add_project_deleted_with_namespace.sql。）
func (repo *projectRepo) RestoreDeleted(ctx context.Context, id int) (err error) {
	ctx, span := tracer.Start(ctx, "projectRepo/RestoreDeleted")
	defer func() { endSpan(span, err) }()
	// 目标行本身是软删行，而本函数后续无论 SELECT 还是清 deleted_at 的 UPDATE 都与软删拦截器
	// 补的 deleted_at IS NULL 相矛盾（SELECT 取不到会误报 NotFound，UPDATE 命中 0 行且静默
	// "成功"），故整段统一绕过拦截器。
	ctx = mixin.SkipSoftDelete(ctx)
	proj, err := repo.data.DB().Project.Query().Where(project.ID(id)).Only(ctx)
	if err != nil {
		return errs.Wrap(err, "restore project")
	}
	if proj.DeletedAt == nil {
		// 未软删的项目无需恢复：显式报错优于静默成功，避免调用方把「什么都没做」当完成。
		return errs.WrapInvalidArgument(fmt.Errorf("项目 %d 未被删除，无需恢复", id), "restore project")
	}
	// 空间存活校验与项目写入放进同一事务：两者不平摊在两条独立 autocommit 语句上，校验通过后
	// 不再有「语句间隙」被并发删除钻空子。
	//
	// 注意本事务只把窗口**收窄到事务自身范围**，并未根除：MySQL 默认 REPEATABLE READ 下，本事务
	// 的 SELECT 建立快照后，并发事务仍可软删该空间并提交——我们的 UPDATE 落在 projects 行、
	// 对方的落在 namespaces 行，互不冲突，写入照样落地，仍会产出「项目存活、所属空间软删」的
	// 孤儿。要根除需把空间存活条件写进 UPDATE 自身的 WHERE（EXISTS 子查询）或对该空间行加锁。
	// 这一残留窗口与 cron 侧的 nil 命名空间守卫（cron_tasks.go 的 FixDeployStatus）配套兜底：
	// 真出现孤儿时定时任务不会 panic，只是跳过它。
	return errs.Wrap(repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		ns, err := tx.Namespace.Query().Where(namespace.ID(proj.NamespaceID)).Only(ctx)
		if err != nil {
			return err
		}
		if ns.DeletedAt != nil {
			// 这里只构造领域语义错误、不再包裹：外层 errs.Wrap 会保留本错误已带的 400 码
			// （wrapErr 经 status.Convert 识别到确定码后原样保留），再包一层只会让日志链上
			// 出现两条重复的 "restore project" 堆栈，客户端可见 message 二者相同。
			return errs.InvalidArgument(
				fmt.Sprintf("项目所属空间 %s 已被删除，请先恢复空间（连带恢复空间下整批项目）", ns.Name),
			)
		}
		return tx.Project.UpdateOneID(id).
			SetDeployStatus(types.Deploy_StatusUnknown).
			ClearDeletedAt().
			Exec(ctx)
	}), "restore project")
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
