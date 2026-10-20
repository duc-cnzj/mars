package data

import (
	"context"
	"fmt"

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
	"google.golang.org/protobuf/proto"
)

var _ biz.RepoRepo = (*repoImpl)(nil)

// aiAgentSlotKey 是 AI Agent 仓库唯一槽位的固定值：ai_agent=true 时写入，否则为 NULL。
// 可空唯一列天然保证「至多一行 true」——MySQL/SQLite 的唯一索引允许多个 NULL，
// 并发把两个仓库设为 AI Agent 时由唯一索引兜底（ent.ConstraintError → 409）。
const aiAgentSlotKey = "ai"

// aiAgentSlotKeyFor 返回 ai_agent 对应的唯一槽位值（false 时为 nil）。
func aiAgentSlotKeyFor(aiAgent bool) *string {
	if !aiAgent {
		return nil
	}
	return lo.ToPtr(aiAgentSlotKey)
}

// clearOtherAIAgentSlots 清除除 keepID 外的 AI Agent 槽位。keepID 传 0 表示新建场景，
// 清除全部旧槽位。必须含软删行：软删行仍占用唯一索引，漏清会让新 AI 仓库写入撞唯一约束。
// 批量 Update 不经过软删查询拦截器（拦截器只作用于 Query），故软删行天然命中。
func clearOtherAIAgentSlots(ctx context.Context, tx *ent.Tx, keepID int) error {
	up := tx.Repo.Update().Where(repo.AiAgentKeyNotNil())
	if keepID > 0 {
		up = up.Where(repo.IDNEQ(keepID))
	}
	_, err := up.SetAiAgent(false).ClearAiAgentKey().Save(ctx)
	return err
}

// repoImpl 是 biz.RepoRepo 的实现：封装 ent Repo 查询与 git 项目名/默认分支解析。
type repoImpl struct {
	data    dataStore
	gitRepo biz.GitRepo
}

// Get 按 ID 查询单个仓库。
func (r *repoImpl) Get(ctx context.Context, id int) (out *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/Get")
	defer func() { endSpan(span, err) }()
	get, err := r.data.DB().Repo.Get(ctx, id)
	return toRepo(get), errs.Wrap(err, "get repo")
}

// GetByName 按名称查询仓库，供 biz 层做名称唯一性校验（NotFound 由 errs.Wrap 归类）。
func (r *repoImpl) GetByName(ctx context.Context, name string) (out *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/GetByName")
	defer func() { endSpan(span, err) }()
	get, err := r.data.DB().Repo.Query().Where(repo.Name(name)).First(ctx)
	return toRepo(get), errs.Wrap(err, "get repo by name")
}

// GetAIAgent 返回全局唯一的 AI Agent 仓库；未配置时返回 ent.NotFound（由 errs.Wrap 归类 404）。
func (r *repoImpl) GetAIAgent(ctx context.Context) (out *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/GetAIAgent")
	defer func() { endSpan(span, err) }()
	first, err := r.data.DB().Repo.Query().Where(repo.AiAgentEQ(true)).First(ctx)
	return toRepo(first), errs.Wrap(err, "get ai agent repo")
}

// NewRepo 构造仓库 repo 实现，注入 git 端口供创建/更新时解析项目名与默认分支。
func NewRepo(data dataStore, gitRepo biz.GitRepo) biz.RepoRepo {
	return &repoImpl{
		data:    data,
		gitRepo: gitRepo,
	}
}

// All 按启用状态与是否需要 git 仓库过滤返回全部仓库。
func (r *repoImpl) All(ctx context.Context, in *biz.AllRepoRequest) (repos []*biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/All")
	defer func() { endSpan(span, err) }()
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
func (r *repoImpl) List(ctx context.Context, in *biz.ListRepoRequest) (repos []*biz.Repo, pag *pagination.Pagination, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/List")
	defer func() { endSpan(span, err) }()
	query := r.data.DB().Repo.Query().
		Where(
			filters.IfOrderByIDDesc(in.OrderByIDDesc),
			filters.IfEnabled(in.Enabled),
			filters.IfAIAgent(in.AIAgent),
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
			repo.FieldAiAgent,
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
func (r *repoImpl) Create(ctx context.Context, in *biz.CreateRepoInput) (get *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/Create")
	defer func() { endSpan(span, err) }()
	projName, defaultBranch, err := r.resolveRepoFields(ctx, in.NeedGitRepo, in.GitProjectID, in.MarsConfig)
	if err != nil {
		return nil, err
	}

	var save *ent.Repo
	err = r.data.WithTx(ctx, func(tx *ent.Tx) error {
		// AI Agent 仓库全局唯一：先清掉旧槽位（含软删行），再插入新行，避免撞唯一索引。
		if in.AIAgent {
			if txErr := clearOtherAIAgentSlots(ctx, tx, 0); txErr != nil {
				return txErr
			}
		}
		cr := tx.Repo.Create().
			SetName(in.Name).
			SetNeedGitRepo(in.NeedGitRepo).
			SetNillableGitProjectName(projName).
			SetNillableDefaultBranch(defaultBranch).
			SetNillableGitProjectID(in.GitProjectID).
			SetEnabled(in.Enabled).
			SetDescription(in.Description).
			SetMarsConfig(in.MarsConfig).
			SetAiAgent(in.AIAgent).
			SetNillableAiAgentKey(aiAgentSlotKeyFor(in.AIAgent))
		if !in.NeedGitRepo {
			cr.SetNillableGitProjectID(nil).
				SetNillableDefaultBranch(nil).
				SetNillableGitProjectName(nil)
		}
		s, txErr := cr.Save(ctx)
		save = s
		return txErr
	})
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

// resolveRepoFields 解析仓库的派生字段：需要 git 仓库时取项目名与默认分支，并计算 IsSimpleEnv。
// 纯外部查询（git API / yaml 解析），不触碰 DB，供 Create/Update/Import 复用。
func (r *repoImpl) resolveRepoFields(ctx context.Context, needGitRepo bool, gitProjectID *int32, marsConfig *mars.Config) (projName, defaultBranch *string, err error) {
	if needGitRepo {
		projName, defaultBranch, err = r.GetProjNameAndBranch(ctx, int(lo.FromPtr(gitProjectID)))
		if err != nil {
			return nil, nil, err
		}
	}
	if marsConfig != nil {
		marsConfig.IsSimpleEnv = r.isSimpleEnv(ctx, marsConfig)
	}
	return projName, defaultBranch, nil
}

// Show 查询单个仓库并预加载其关联项目。
func (r *repoImpl) Show(ctx context.Context, id int) (out *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/Show")
	defer func() { endSpan(span, err) }()
	get, err := r.data.DB().Repo.Query().Where(repo.ID(id)).WithProjects().Only(ctx)
	return toRepo(get), errs.Wrap(err, "show repo")
}

// Update 更新仓库：需要 git 仓库时重新解析默认分支与项目名。
func (r *repoImpl) Update(ctx context.Context, in *biz.UpdateRepoInput) (get *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/Update")
	defer func() { endSpan(span, err) }()
	projName, defaultBranch, err := r.resolveRepoFields(ctx, in.NeedGitRepo, in.GitProjectID, in.MarsConfig)
	if err != nil {
		return nil, err
	}

	var save *ent.Repo
	err = r.data.WithTx(ctx, func(tx *ent.Tx) error {
		// 设为 AI Agent 仓库时先清掉其它仓库（含软删行）的槽位，事务内保证唯一。
		if in.AIAgent != nil && *in.AIAgent {
			if txErr := clearOtherAIAgentSlots(ctx, tx, int(in.ID)); txErr != nil {
				return txErr
			}
		}
		up := tx.Repo.
			UpdateOneID(int(in.ID)).
			SetName(in.Name).
			SetNeedGitRepo(in.NeedGitRepo).
			SetNillableGitProjectID(in.GitProjectID).
			SetNillableGitProjectName(projName).
			SetNillableDefaultBranch(defaultBranch).
			SetDescription(in.Description).
			SetMarsConfig(in.MarsConfig)
		// nil 表示调用方未指定该字段：保持现有 AI 标记与唯一槽位不变。
		if in.AIAgent != nil {
			up.SetAiAgent(*in.AIAgent)
			if *in.AIAgent {
				up.SetAiAgentKey(aiAgentSlotKey)
			} else {
				up.ClearAiAgentKey()
			}
		}
		if !in.NeedGitRepo {
			up.ClearGitProjectID().ClearGitProjectName().ClearDefaultBranch()
		}
		s, txErr := up.Save(ctx)
		save = s
		return txErr
	})
	return toRepo(save), errs.Wrap(err, "update repo")
}

// Delete 按 ID 删除仓库。
func (r *repoImpl) Delete(ctx context.Context, id int) (err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/Delete")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(r.data.DB().Repo.DeleteOneID(id).Exec(ctx), "delete repo")
}

// GetProjNameAndBranch 从 git 端口取项目名与默认分支，供创建/更新前填充仓库字段。
func (r *repoImpl) GetProjNameAndBranch(ctx context.Context, projID int) (projName, defaultBranch *string, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/GetProjNameAndBranch")
	defer func() { endSpan(span, err) }()
	project, err := r.gitRepo.GetByProjectID(ctx, projID)
	if err != nil {
		return nil, nil, err
	}
	defaultBranch = lo.ToPtr(project.DefaultBranch)
	projName = lo.ToPtr(project.Name)
	return projName, defaultBranch, nil
}

// ToggleEnabled 切换仓库的启用状态。
func (r *repoImpl) ToggleEnabled(ctx context.Context, id int, enabled bool) (get *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/ToggleEnabled")
	defer func() { endSpan(span, err) }()
	save, err := r.data.DB().Repo.UpdateOneID(id).SetEnabled(enabled).Save(ctx)
	return toRepo(save), errs.Wrap(err, "toggle repo enabled")
}

// BatchUpdateMarsConfig 在一个事务内批量覆盖若干仓库的 mars_config（轻量：直接写配置列，
// 不走 resolveRepoFields 的 git 解析）。任一仓库不存在或写失败整体回滚；id 缺失由 errs.Wrap
// 归为 404。供「一键配置休眠时间窗到多仓库」这类只改配置的批量场景。
func (r *repoImpl) BatchUpdateMarsConfig(ctx context.Context, configs map[int]*mars.Config) (err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/BatchUpdateMarsConfig")
	defer func() { endSpan(span, err) }()
	err = r.data.WithTx(ctx, func(tx *ent.Tx) error {
		for id, cfg := range configs {
			if _, txErr := tx.Repo.UpdateOneID(id).SetMarsConfig(cfg).Save(ctx); txErr != nil {
				return txErr
			}
		}
		return nil
	})
	return errs.Wrap(err, "batch update repo mars config")
}

// Clone 复制仓库：读取原仓库字段后用新名字重新创建。
func (r *repoImpl) Clone(ctx context.Context, input *biz.CloneRepoInput) (out *biz.Repo, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/Clone")
	defer func() { endSpan(span, err) }()
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

// configWithHibernate 返回 cfg 的副本并把 hibernate 字段替换为 h（其余字段原样保留）。
// 导入暂不支持定时休眠：新建置空（h=nil）、更新沿用库内既有值，避免导入把管理员已配好的
// 休眠能力（副本字段/时间窗引用）清掉或带入悬空引用。
// 必须克隆后再改：cfg 来自导入请求、可能被调用方复用，就地改会污染输入。
func configWithHibernate(cfg *mars.Config, h *mars.Hibernate) *mars.Config {
	if cfg == nil {
		if h == nil {
			return nil
		}
		return &mars.Config{Hibernate: h}
	}
	out := proto.Clone(cfg).(*mars.Config)
	out.Hibernate = h
	return out
}

// planImportItem 是单条导入的计划：判定创建/更新并预解析派生字段，外部查询在事务外完成。
type planImportItem struct {
	create        bool
	id            int
	item          *biz.ImportRepoItem
	projName      *string
	defaultBranch *string
	// preservedHibernate 更新场景下沿用的库内既有休眠配置（导入暂不支持 hibernate，不改它）。
	// 新建场景为 nil。
	preservedHibernate *mars.Hibernate
}

// planImport 为导入条目逐条做计划：按 name 判定创建/更新，并预解析派生字段
// （git 项目名/默认分支/IsSimpleEnv，均为外部查询）。只读不落库，供 Import 与 PreviewImport 共用。
// 幂等依赖 name 唯一：库内同名行大于 1 时视为脏数据拒绝导入，避免 First 静默只覆盖一行留下残行。
func (r *repoImpl) planImport(ctx context.Context, items []*biz.ImportRepoItem) ([]*planImportItem, error) {
	plans := make([]*planImportItem, 0, len(items))
	for _, item := range items {
		p := &planImportItem{item: item}
		existing, qErr := r.data.DB().Repo.Query().Where(repo.Name(item.Name)).All(ctx)
		if qErr != nil {
			return nil, errs.Wrap(qErr, "get repo by name")
		}
		switch {
		case len(existing) > 1:
			return nil, errs.WrapInvalidArgument(fmt.Errorf("repo 名称 %q 在数据库中重复（%d 行），请先清理后重试导入", item.Name, len(existing)), "import repo")
		case len(existing) == 1:
			p.id = existing[0].ID
			// 导入暂不支持定时休眠：更新时沿用库内既有 hibernate，避免导出让仓库失去休眠能力。
			p.preservedHibernate = existing[0].MarsConfig.GetHibernate()
		default:
			p.create = true
		}
		projName, defaultBranch, err := r.resolveRepoFields(ctx, item.NeedGitRepo, item.GitProjectID, item.MarsConfig)
		if err != nil {
			return nil, err
		}
		p.projName, p.defaultBranch = projName, defaultBranch
		plans = append(plans, p)
	}
	return plans, nil
}

// Import 批量导入仓库：按名称幂等覆盖（已存在走更新、不存在走创建）。
// 先对全部条目做计划并解析派生字段（外部查询在事务外完成），
// 再在单个事务内整体应用——任一条目失败整体回滚，避免批量导入半途失败留下部分变更。
func (r *repoImpl) Import(ctx context.Context, items []*biz.ImportRepoItem) (created, updated int, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/Import")
	defer func() { endSpan(span, err) }()
	plans, err := r.planImport(ctx, items)
	if err != nil {
		return 0, 0, err
	}

	// 事务内整体应用：全部成功才提交，任一出错回滚（created/updated 仅提交成功后有效）。
	if err := r.data.WithTx(ctx, func(tx *ent.Tx) error {
		for _, p := range plans {
			if p.create {
				// 导入中的 AI Agent 仓库同样要顶替目标环境旧槽位（唯一索引兜底）。
				if p.item.AIAgent {
					if err := clearOtherAIAgentSlots(ctx, tx, 0); err != nil {
						return err
					}
				}
				if _, err := tx.Repo.Create().
					SetName(p.item.Name).
					SetNeedGitRepo(p.item.NeedGitRepo).
					SetNillableGitProjectID(p.item.GitProjectID).
					SetNillableGitProjectName(p.projName).
					SetNillableDefaultBranch(p.defaultBranch).
					SetEnabled(p.item.Enabled).
					SetDescription(p.item.Description).
					SetMarsConfig(configWithHibernate(p.item.MarsConfig, nil)).
					SetAiAgent(p.item.AIAgent).
					SetNillableAiAgentKey(aiAgentSlotKeyFor(p.item.AIAgent)).
					Save(ctx); err != nil {
					return err
				}
				created++
				continue
			}
			if p.item.AIAgent {
				if err := clearOtherAIAgentSlots(ctx, tx, p.id); err != nil {
					return err
				}
			}
			up := tx.Repo.UpdateOneID(p.id).
				SetName(p.item.Name).
				SetNeedGitRepo(p.item.NeedGitRepo).
				SetNillableGitProjectID(p.item.GitProjectID).
				SetNillableGitProjectName(p.projName).
				SetNillableDefaultBranch(p.defaultBranch).
				SetDescription(p.item.Description).
				SetMarsConfig(configWithHibernate(p.item.MarsConfig, p.preservedHibernate)).
				SetEnabled(p.item.Enabled).
				SetAiAgent(p.item.AIAgent)
			if p.item.AIAgent {
				up.SetAiAgentKey(aiAgentSlotKey)
			} else {
				up.ClearAiAgentKey()
			}
			if !p.item.NeedGitRepo {
				up.ClearGitProjectID().ClearGitProjectName().ClearDefaultBranch()
			}
			if _, err := up.Save(ctx); err != nil {
				return err
			}
			updated++
		}
		return nil
	}); err != nil {
		return 0, 0, errs.Wrap(err, "import repo")
	}
	return created, updated, nil
}

// PreviewImport 干跑导入：复用 planImport 的计划逻辑但不落库，返回「将新建/将更新」的计数（导入前预览用）。
func (r *repoImpl) PreviewImport(ctx context.Context, items []*biz.ImportRepoItem) (created, updated int, err error) {
	ctx, span := tracer.Start(ctx, "repoImpl/PreviewImport")
	defer func() { endSpan(span, err) }()
	plans, err := r.planImport(ctx, items)
	if err != nil {
		return 0, 0, err
	}
	for _, p := range plans {
		if p.create {
			created++
		} else {
			updated++
		}
	}
	return created, updated, nil
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
		AIAgent:        data.AiAgent,
		Projects:       slice.Map(data.Edges.Projects, toProject),
	}
}
