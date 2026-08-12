package data

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/biz"

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
	"go.opentelemetry.io/otel"
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
func (repo *projectRepo) Version(ctx context.Context, id int) (int, error) {
	get, err := repo.data.DB().Project.Query().Select(
		project.FieldID,
		project.FieldVersion,
	).Where(project.ID(id)).Only(ctx)
	if err != nil {
		return 0, errs.Wrap(err, "query project version")
	}
	return get.Version, nil
}

// List 分页查询项目列表；非 admin 只返回其可访问命名空间下的项目，防私有内容泄漏。
func (repo *projectRepo) List(ctx context.Context, input *biz.ListProjectInput) ([]*biz.Project, *pagination.Pagination, error) {
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
func (repo *projectRepo) Create(ctx context.Context, input *biz.CreateProjectInput) (*biz.Project, error) {
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
func (repo *projectRepo) UpdateProject(ctx context.Context, input *biz.UpdateProjectInput) (*biz.Project, error) {
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
func (repo *projectRepo) Show(ctx context.Context, id int) (*biz.Project, error) {
	_, span := otel.Tracer("").Start(ctx, "repo/project/Show")
	defer span.End()
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
func (repo *projectRepo) Delete(ctx context.Context, id int) error {
	return errs.Wrap(repo.data.DB().Project.DeleteOneID(id).Exec(ctx), "delete project")
}

// UpdateStatusByVersion 校验版本匹配后更新部署状态并递增版本号（乐观锁防并发覆盖）。
func (repo *projectRepo) UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (*biz.Project, error) {
	if _, err := repo.FindByVersion(ctx, id, version); err != nil {
		return nil, err
	}
	save, err := repo.data.DB().Project.UpdateOneID(id).SetDeployStatus(status).SetVersion(version + 1).Save(ctx)
	return toProject(save), errs.Wrap(err, "update project status by version")
}

// FindByVersion 按 ID 与版本号精确查找项目（用于乐观锁校验）。
func (repo *projectRepo) FindByVersion(ctx context.Context, id, version int) (*biz.Project, error) {
	first, err := repo.data.DB().Project.Query().Where(project.ID(id), project.Version(version)).First(ctx)
	return toProject(first), errs.Wrap(err, "find project by version")
}

// UpdateVersion 直接覆盖项目版本号。
func (repo *projectRepo) UpdateVersion(ctx context.Context, id int, version int) (*biz.Project, error) {
	save, err := repo.data.DB().Project.UpdateOneID(id).SetVersion(version).Save(ctx)
	return toProject(save), errs.Wrap(err, "update project version")
}

// UpdateDeployStatus 仅更新项目的部署状态。
func (repo *projectRepo) UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (*biz.Project, error) {
	save, err := repo.data.DB().Project.UpdateOneID(id).SetDeployStatus(status).Save(ctx)
	return toProject(save), errs.Wrap(err, "update deploy status")
}

// ListByDeployStatus 按部署状态集合过滤项目并携带 namespace 信息，
// cron FixDeployStatus 用 helm 实测状态修复失败/未知项目。
func (repo *projectRepo) ListByDeployStatus(ctx context.Context, statuses ...types.Deploy) ([]*biz.Project, error) {
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
func (repo *projectRepo) FindByName(ctx context.Context, name string, nsID int) (*biz.Project, error) {
	first, err := repo.data.DB().Project.Query().Where(project.Name(name), project.NamespaceID(nsID)).First(ctx)
	return toProject(first), errs.Wrap(err, "find project by name")
}

// FindProjectsByIDs 按主键批量取项目。endpoint 编排依赖项目的 Name 与 Manifest
// 来匹配集群内对象，这是纯数据读取，不包含任何编排逻辑。
func (repo *projectRepo) FindProjectsByIDs(ctx context.Context, ids ...int) ([]*biz.Project, error) {
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
