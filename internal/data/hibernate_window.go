package data

import (
	"context"
	"slices"
	"sort"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/hibernatewindow"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/projecthibernate"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/repo"
	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// toHibernateWindow 把 ent.HibernateWindow 转换为 biz.HibernateWindow（nil 安全）。
func toHibernateWindow(e *ent.HibernateWindow) *biz.HibernateWindow {
	if e == nil {
		return nil
	}
	return &biz.HibernateWindow{
		ID:          e.ID,
		Name:        e.Name,
		Start:       e.Start,
		End:         e.End,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

var _ biz.HibernateWindowRepo = (*hibernateWindowRepo)(nil)

// hibernateWindowRepo 是全局休眠时间窗的持久化实现：管理员在后台维护命名时段，各仓库按名字引用。
type hibernateWindowRepo struct {
	data dataStore
}

// NewHibernateWindowRepo 构造全局休眠时间窗 repo。
func NewHibernateWindowRepo(data dataStore) biz.HibernateWindowRepo {
	return &hibernateWindowRepo{data: data}
}

// List 返回全部时间窗（按名称升序）。
func (h *hibernateWindowRepo) List(ctx context.Context) (out []*biz.HibernateWindow, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/List")
	defer func() { endSpan(span, err) }()
	all, err := h.data.DB().HibernateWindow.Query().
		Order(ent.Asc(hibernatewindow.FieldName)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list hibernate windows")
	}
	out = make([]*biz.HibernateWindow, 0, len(all))
	for _, e := range all {
		out = append(out, toHibernateWindow(e))
	}
	return out, nil
}

// GetByName 按名称查时间窗；未命中返回 (nil, nil)，不把"不存在"当错误。
func (h *hibernateWindowRepo) GetByName(ctx context.Context, name string) (out *biz.HibernateWindow, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/GetByName")
	defer func() { endSpan(span, err) }()
	found, err := h.data.DB().HibernateWindow.Query().
		Where(hibernatewindow.Name(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, errs.Wrap(err, "query hibernate window by name")
	}
	return toHibernateWindow(found), nil
}

// GetByID 按 ID 查时间窗；未命中返回 (nil, nil)（不把"不存在"当错误，删除路径据此幂等）。
func (h *hibernateWindowRepo) GetByID(ctx context.Context, id int) (out *biz.HibernateWindow, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/GetByID")
	defer func() { endSpan(span, err) }()
	found, err := h.data.DB().HibernateWindow.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, errs.Wrap(err, "query hibernate window by id")
	}
	return toHibernateWindow(found), nil
}

// WindowRepoRefs 返回「窗名 → 引用该窗的仓库名（升序）」映射：单趟扫描全部带 mars_config
// 的仓库，按各仓的 hibernate.window_names 分组。mars_config 是 JSON 列，嵌套数组包含查询在
// MySQL/SQLite 间方言不一，故取回后在 Go 内分组（仓库为后台级、量小，扫描成本可接受）；
// 软删仓库由 SoftDeleteMixin 拦截器隐式过滤，不吃进来。
func (h *hibernateWindowRepo) WindowRepoRefs(ctx context.Context) (out map[string][]string, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/WindowRepoRefs")
	defer func() { endSpan(span, err) }()
	all, err := h.data.DB().Repo.Query().
		Where(repo.MarsConfigNotNil()).
		Select(repo.FieldName, repo.FieldMarsConfig).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query repos referencing hibernate window")
	}
	out = make(map[string][]string)
	for _, e := range all {
		for _, name := range e.MarsConfig.GetHibernate().GetWindowNames() {
			out[name] = append(out[name], e.Name)
		}
	}
	for name := range out {
		sort.Strings(out[name])
	}
	return out, nil
}

// ReposReferencing 返回 mars_config.hibernate.window_names 含 name 的仓库名（升序）。
// 删除/改名时间窗前判"是否仍有下游关联"用。实现为 WindowRepoRefs 的单名查表（未命中返回 nil）。
func (h *hibernateWindowRepo) ReposReferencing(ctx context.Context, name string) (out []string, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/ReposReferencing")
	defer func() { endSpan(span, err) }()
	refs, err := h.WindowRepoRefs(ctx)
	if err != nil {
		return nil, err
	}
	return refs[name], nil
}

// ProjectsReferencing 返回 project_hibernates.window_names 含 name 的项目 ID（升序）。
// 删除/改名时间窗前判"项目是否仍有下游关联"用。window_names 是 JSON 数组列，跨方言包含查询
// 不便，故取回后在 Go 内过滤（项目引用行为可控）。本表无软删 mixin，全部行都参与判定。
func (h *hibernateWindowRepo) ProjectsReferencing(ctx context.Context, name string) (out []int, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/ProjectsReferencing")
	defer func() { endSpan(span, err) }()
	all, err := h.data.DB().ProjectHibernate.Query().
		Select(projecthibernate.FieldProjectID, projecthibernate.FieldWindowNames).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query projects referencing hibernate window")
	}
	for _, e := range all {
		if slices.Contains(e.WindowNames, name) {
			out = append(out, e.ProjectID)
		}
	}
	sort.Ints(out)
	return out, nil
}

// ProjectReferencingRepoIDs 返回「引用该窗的项目 ID → 所属仓库 ID」映射，供「从仓库移除窗」时
// 按受影响仓库收窄漂移校验：只拦真正会失去该窗的仓库下的项目，不误伤引用同名窗的其它仓库。
// 先单趟扫 project_hibernates 命中窗名收集项目 ID，再按 ID 批量取仓库归属——Project 查询经
// SoftDeleteMixin 拦截器过滤软删项目，故其悬留的休眠行不计入；项目无仓库时 repo_id 为 0。
func (h *hibernateWindowRepo) ProjectReferencingRepoIDs(ctx context.Context, name string) (out map[int]int, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/ProjectReferencingRepoIDs")
	defer func() { endSpan(span, err) }()
	rows, err := h.data.DB().ProjectHibernate.Query().
		Select(projecthibernate.FieldProjectID, projecthibernate.FieldWindowNames).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query projects referencing hibernate window")
	}
	ids := make([]int, 0, len(rows))
	for _, e := range rows {
		if slices.Contains(e.WindowNames, name) {
			ids = append(ids, e.ProjectID)
		}
	}
	out = make(map[int]int, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	projs, err := h.data.DB().Project.Query().
		Where(project.IDIn(ids...)).
		Select(project.FieldID, project.FieldRepoID).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query projects' repo for hibernate window refs")
	}
	for _, p := range projs {
		out[p.ID] = p.RepoID
	}
	return out, nil
}

// Create 新建时间窗；名称重复由唯一索引兜底（errs.Wrap 归类为 409）。
func (h *hibernateWindowRepo) Create(ctx context.Context, w *biz.HibernateWindow) (out *biz.HibernateWindow, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/Create")
	defer func() { endSpan(span, err) }()
	created, err := h.data.DB().HibernateWindow.Create().
		SetName(w.Name).
		SetStart(w.Start).
		SetEnd(w.End).
		SetDescription(w.Description).
		Save(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "create hibernate window")
	}
	return toHibernateWindow(created), nil
}

// Update 按 ID 更新时间窗；不存在返回 NotFound（errs.Wrap 归类为 404）。
func (h *hibernateWindowRepo) Update(ctx context.Context, w *biz.HibernateWindow) (out *biz.HibernateWindow, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/Update")
	defer func() { endSpan(span, err) }()
	updated, err := h.data.DB().HibernateWindow.UpdateOneID(w.ID).
		SetName(w.Name).
		SetStart(w.Start).
		SetEnd(w.End).
		SetDescription(w.Description).
		Save(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "update hibernate window")
	}
	return toHibernateWindow(updated), nil
}

// UpdateCascade 原子改名：在同一事务内把时间窗更新为 w，并把所有按 oldName 引用它的仓库
// （mars_config.hibernate.window_names）与项目（project_hibernates.window_names）改写为 w.Name。
// 任一失败整体回滚——避免「窗口已改名但引用未迁移」或反之的半迁移状态。返回 (更新后的窗口,
// 被迁移的引用总数：仓库数 + 项目数)；不存在/名称重复由 errs.Wrap 归为 404 / 409。
// 软删仓库由 SoftDeleteMixin 拦截器隐式过滤，与其语义一致（不迁移）。
func (h *hibernateWindowRepo) UpdateCascade(ctx context.Context, w *biz.HibernateWindow, oldName string) (out *biz.HibernateWindow, migrated int, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/UpdateCascade")
	defer func() { endSpan(span, err) }()
	var updated *ent.HibernateWindow
	err = h.data.WithTx(ctx, func(tx *ent.Tx) error {
		u, txErr := tx.HibernateWindow.UpdateOneID(w.ID).
			SetName(w.Name).
			SetStart(w.Start).
			SetEnd(w.End).
			SetDescription(w.Description).
			Save(ctx)
		if txErr != nil {
			return txErr
		}
		updated = u
		repos, txErr := tx.Repo.Query().
			Where(repo.MarsConfigNotNil()).
			All(ctx)
		if txErr != nil {
			return txErr
		}
		for _, e := range repos {
			hb := e.MarsConfig.GetHibernate()
			// 名称未命中的仓库（含 hibernate 为 nil）跳过；命中的 hb 必非 nil。
			if !slices.Contains(hb.GetWindowNames(), oldName) {
				continue
			}
			hb.WindowNames = replaceWindowName(hb.WindowNames, oldName, w.Name)
			if _, txErr := tx.Repo.UpdateOneID(e.ID).SetMarsConfig(e.MarsConfig).Save(ctx); txErr != nil {
				return txErr
			}
			migrated++
		}
		// 项目级选择同样按名引用，须一并改写，否则项目对账会按旧名未命中而失效。
		phs, txErr := tx.ProjectHibernate.Query().
			Select(projecthibernate.FieldID, projecthibernate.FieldWindowNames).
			All(ctx)
		if txErr != nil {
			return txErr
		}
		for _, p := range phs {
			if !slices.Contains(p.WindowNames, oldName) {
				continue
			}
			if _, txErr := tx.ProjectHibernate.UpdateOneID(p.ID).
				SetWindowNames(replaceWindowName(p.WindowNames, oldName, w.Name)).
				Save(ctx); txErr != nil {
				return txErr
			}
			migrated++
		}
		return nil
	})
	if err != nil {
		return nil, 0, errs.Wrap(err, "update hibernate window with cascade")
	}
	return toHibernateWindow(updated), migrated, nil
}

// replaceWindowName 把 names 里的 oldName 逐个替换为 newName，保留顺序与其余元素（新切片）。
func replaceWindowName(names []string, oldName, newName string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		if n == oldName {
			n = newName
		}
		out[i] = n
	}
	return out
}

// Delete 按 ID 删除时间窗（物理删除）；不存在不报错。"无下游引用"的关联校验由 biz 层在调用前完成。
func (h *hibernateWindowRepo) Delete(ctx context.Context, id int) (err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/Delete")
	defer func() { endSpan(span, err) }()
	_, err = h.data.DB().HibernateWindow.Delete().
		Where(hibernatewindow.ID(id)).
		Exec(ctx)
	return errs.Wrap(err, "delete hibernate window")
}
