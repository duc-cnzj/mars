package data

import (
	"context"
	"sort"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/hibernatewindow"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/repo"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
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
	logger mlog.Logger
	data   dataStore
}

// NewHibernateWindowRepo 构造全局休眠时间窗 repo。
func NewHibernateWindowRepo(logger mlog.Logger, data dataStore) biz.HibernateWindowRepo {
	return &hibernateWindowRepo{logger: logger.WithModule("repo/hibernate_window"), data: data}
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

// ReposReferencing 返回 mars_config.hibernate.window_name == name 的仓库名（升序）。
// 删除时间窗前判"是否仍有下游关联"用。mars_config 是 JSON 列，嵌套路径查询在 MySQL/SQLite
// 间方言不一，故取回后在 Go 内过滤（仓库为后台级、量小，扫描成本可接受）；
// 软删仓库由 SoftDeleteMixin 拦截器隐式过滤，不吃进来。
func (h *hibernateWindowRepo) ReposReferencing(ctx context.Context, name string) (out []string, err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/ReposReferencing")
	defer func() { endSpan(span, err) }()
	all, err := h.data.DB().Repo.Query().
		Where(repo.MarsConfigNotNil()).
		Select(repo.FieldName, repo.FieldMarsConfig).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query repos referencing hibernate window")
	}
	for _, e := range all {
		if e.MarsConfig.GetHibernate().GetWindowName() == name {
			out = append(out, e.Name)
		}
	}
	sort.Strings(out)
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

// Delete 按 ID 删除时间窗（物理删除）；不存在不报错。"无下游引用"的关联校验由 biz 层在调用前完成。
func (h *hibernateWindowRepo) Delete(ctx context.Context, id int) (err error) {
	ctx, span := tracer.Start(ctx, "hibernateWindowRepo/Delete")
	defer func() { endSpan(span, err) }()
	_, err = h.data.DB().HibernateWindow.Delete().
		Where(hibernatewindow.ID(id)).
		Exec(ctx)
	return errs.Wrap(err, "delete hibernate window")
}
