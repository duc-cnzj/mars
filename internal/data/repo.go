package data

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/biz"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/repo"
	"github.com/duc-cnzj/mars/v6/internal/data/filters"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/duc-cnzj/mars/v6/internal/util/yaml"
	"github.com/samber/lo"
)

var _ biz.RepoRepo = (*repoImpl)(nil)

// repoImpl 是 biz.RepoRepo 的实现：封装 ent Repo 查询与 git 项目名/默认分支解析。
type repoImpl struct {
	data    dataStore
	gitRepo biz.GitRepo
}

// Get 按 ID 查询单个仓库。
func (r *repoImpl) Get(ctx context.Context, id int) (*biz.Repo, error) {
	get, err := r.data.DB().Repo.Get(ctx, id)
	return toRepo(get), errs.Wrap(err, "get repo")
}

// GetByName 按名称查询仓库，供 biz 层做名称唯一性校验（NotFound 由 errs.Wrap 归类）。
func (r *repoImpl) GetByName(ctx context.Context, name string) (*biz.Repo, error) {
	get, err := r.data.DB().Repo.Query().Where(repo.Name(name)).First(ctx)
	return toRepo(get), errs.Wrap(err, "get repo by name")
}

// NewRepo 构造仓库 repo 实现，注入 git 端口供创建/更新时解析项目名与默认分支。
func NewRepo(data dataStore, gitRepo biz.GitRepo) biz.RepoRepo {
	return &repoImpl{
		data:    data,
		gitRepo: gitRepo,
	}
}

// All 按启用状态与是否需要 git 仓库过滤返回全部仓库。
func (r *repoImpl) All(ctx context.Context, in *biz.AllRepoRequest) ([]*biz.Repo, error) {
	query := r.data.DB().Repo.Query().Where(
		filters.IfEnabled(in.Enabled),
		filters.IfBool(repo.FieldNeedGitRepo)(in.NeedGitRepo),
	)
	all, err := query.All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list repos")
	}
	return slice.Map(all, toRepo), nil
}

// List 分页查询仓库列表并返回分页信息。
func (r *repoImpl) List(ctx context.Context, in *biz.ListRepoRequest) ([]*biz.Repo, *pagination.Pagination, error) {
	query := r.data.DB().Repo.Query().
		Where(
			filters.IfOrderByIDDesc(in.OrderByIDDesc),
			filters.IfEnabled(in.Enabled),
			filters.IfNameLike(in.Name),
		)
	all, err := query.Clone().
		Select(
			repo.FieldID,
			repo.FieldName,
			repo.FieldEnabled,
			repo.FieldGitProjectID,
			repo.FieldGitProjectName,
			repo.FieldNeedGitRepo,
			repo.FieldDescription,
			repo.FieldCreatedAt,
			repo.FieldUpdatedAt,
		).
		Offset(pagination.GetPageOffset(in.Page, in.PageSize)).
		Limit(int(in.PageSize)).
		All(ctx)
	if err != nil {
		return nil, nil, errs.Wrap(err, "list repos")
	}
	count := query.Clone().CountX(ctx)

	return slice.Map(all, toRepo), pagination.NewPagination(in.Page, in.PageSize, count), nil
}

// Create 新建仓库：需要 git 仓库时先解析默认分支与项目名，并按需解析 IsSimpleEnv。
func (r *repoImpl) Create(ctx context.Context, in *biz.CreateRepoInput) (*biz.Repo, error) {
	var (
		projName      *string
		defaultBranch *string
		err           error
	)
	if in.NeedGitRepo {
		projName, defaultBranch, err = r.GetProjNameAndBranch(ctx, int(lo.FromPtr(in.GitProjectID)))
		if err != nil {
			return nil, err
		}
	}

	if in.MarsConfig != nil {
		in.MarsConfig.IsSimpleEnv = r.isSimpleEnv(ctx, in.MarsConfig)
	}

	cr := r.data.DB().Repo.Create().
		SetName(in.Name).
		SetNeedGitRepo(in.NeedGitRepo).
		SetNillableGitProjectName(projName).
		SetNillableDefaultBranch(defaultBranch).
		SetNillableGitProjectID(in.GitProjectID).
		SetEnabled(in.Enabled).
		SetDescription(in.Description).
		SetMarsConfig(in.MarsConfig)
	if !in.NeedGitRepo {
		cr.SetNillableGitProjectID(nil).
			SetNillableDefaultBranch(nil).
			SetNillableGitProjectName(nil)
	}
	save, err := cr.Save(ctx)
	return toRepo(save), errs.Wrap(err, "create repo")
}

// isSimpleEnv 判断配置是否属简单环境：本地 values 解析失败时回退远端 chart 的 values.yaml。
func (r *repoImpl) isSimpleEnv(ctx context.Context, config *mars.Config) bool {
	if config == nil || config.ConfigField == "" || config.LocalChartPath == "" {
		return true
	}
	isSimple, err := yaml.IsSimpleEnv(config.ConfigField, config.ValuesYaml)
	if err == nil {
		return isSimple
	}
	// 本地解析失败回退远端 values.yaml；远端也失败时降级为 false（非简单）。
	// 该字段仅用于前端表单渲染，失败不阻断 repo 创建，错误在此有意吞掉。
	yamlData, _ := r.gitRepo.GetChartValuesYaml(ctx, config.LocalChartPath)
	isSimple, _ = yaml.IsSimpleEnv(config.ConfigField, yamlData)
	return isSimple
}

// Show 查询单个仓库并预加载其关联项目。
func (r *repoImpl) Show(ctx context.Context, id int) (*biz.Repo, error) {
	get, err := r.data.DB().Repo.Query().Where(repo.ID(id)).WithProjects().Only(ctx)
	return toRepo(get), errs.Wrap(err, "show repo")
}

// Update 更新仓库：需要 git 仓库时重新解析默认分支与项目名。
func (r *repoImpl) Update(ctx context.Context, in *biz.UpdateRepoInput) (*biz.Repo, error) {
	var (
		projName      *string
		defaultBranch *string
		err           error
	)

	if in.NeedGitRepo {
		projName, defaultBranch, err = r.GetProjNameAndBranch(ctx, int(*in.GitProjectID))
		if err != nil {
			return nil, err
		}
	}

	if in.MarsConfig != nil {
		in.MarsConfig.IsSimpleEnv = r.isSimpleEnv(ctx, in.MarsConfig)
	}

	up := r.data.DB().Repo.
		UpdateOneID(int(in.ID)).
		SetName(in.Name).
		SetNeedGitRepo(in.NeedGitRepo).
		SetNillableGitProjectID(in.GitProjectID).
		SetNillableGitProjectName(projName).
		SetNillableDefaultBranch(defaultBranch).
		SetDescription(in.Description).
		SetMarsConfig(in.MarsConfig)
	if !in.NeedGitRepo {
		up.ClearGitProjectID().ClearGitProjectName().ClearDefaultBranch()
	}
	save, err := up.Save(ctx)
	return toRepo(save), errs.Wrap(err, "update repo")
}

// Delete 按 ID 删除仓库。
func (r *repoImpl) Delete(ctx context.Context, id int) error {
	return errs.Wrap(r.data.DB().Repo.DeleteOneID(id).Exec(ctx), "delete repo")
}

// GetProjNameAndBranch 从 git 端口取项目名与默认分支，供创建/更新前填充仓库字段。
func (r *repoImpl) GetProjNameAndBranch(ctx context.Context, projID int) (*string, *string, error) {
	var (
		defaultBranch *string
		projName      *string
	)

	project, err := r.gitRepo.GetByProjectID(ctx, projID)
	if err != nil {
		return nil, nil, err
	}
	defaultBranch = lo.ToPtr(project.DefaultBranch)
	projName = lo.ToPtr(project.Name)
	return projName, defaultBranch, nil
}

// ToggleEnabled 切换仓库的启用状态。
func (r *repoImpl) ToggleEnabled(ctx context.Context, id int, enabled bool) (*biz.Repo, error) {
	save, err := r.data.DB().Repo.UpdateOneID(id).SetEnabled(enabled).Save(ctx)
	return toRepo(save), errs.Wrap(err, "toggle repo enabled")
}

// Clone 复制仓库：读取原仓库字段后用新名字重新创建。
func (r *repoImpl) Clone(ctx context.Context, input *biz.CloneRepoInput) (*biz.Repo, error) {
	get, err := r.data.DB().Repo.Get(ctx, input.ID)
	if err != nil {
		return nil, errs.Wrap(err, "get repo")
	}
	return r.Create(ctx, &biz.CreateRepoInput{
		Name:         input.Name,
		Enabled:      get.Enabled,
		NeedGitRepo:  get.NeedGitRepo,
		GitProjectID: lo.ToPtr(get.GitProjectID),
		MarsConfig:   get.MarsConfig,
		Description:  get.Description,
	})
}

// toRepo 把 ent.Repo 转换为 biz.Repo（nil 安全），并顺带转换关联的 Projects。
func toRepo(data *ent.Repo) *biz.Repo {
	if data == nil {
		return nil
	}
	return &biz.Repo{
		ID:             data.ID,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
		DeletedAt:      data.DeletedAt,
		Name:           data.Name,
		DefaultBranch:  data.DefaultBranch,
		GitProjectName: data.GitProjectName,
		GitProjectID:   data.GitProjectID,
		Enabled:        data.Enabled,
		NeedGitRepo:    data.NeedGitRepo,
		MarsConfig:     data.MarsConfig,
		Description:    data.Description,
		Projects:       slice.Map(data.Edges.Projects, toProject),
	}
}
