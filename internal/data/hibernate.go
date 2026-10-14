package data

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/projecthibernate"
	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// toProjectHibernate 把 ent.ProjectHibernate 转换为 biz.ProjectHibernate（nil 安全）。
func toProjectHibernate(e *ent.ProjectHibernate) *biz.ProjectHibernate {
	if e == nil {
		return nil
	}
	return &biz.ProjectHibernate{
		ID:          e.ID,
		ProjectID:   e.ProjectID,
		Enabled:     e.Enabled,
		State:       biz.HibernateState(e.State),
		Snapshot:    e.Snapshot,
		WindowNames: e.WindowNames,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

var _ biz.HibernateRepo = (*hibernateRepo)(nil)

// hibernateRepo 是项目休眠运行态的持久化实现：按 project_id 读写开关/运行态/快照。
type hibernateRepo struct {
	data dataStore
}

// NewHibernateRepo 构造 hibernate repo。
func NewHibernateRepo(data dataStore) biz.HibernateRepo {
	return &hibernateRepo{data: data}
}

// Get 按项目 ID 查休眠记录；未配置返回 (nil, nil)，不把"未配置"当错误。
func (h *hibernateRepo) Get(ctx context.Context, projectID int) (out *biz.ProjectHibernate, err error) {
	ctx, span := tracer.Start(ctx, "hibernateRepo/Get")
	defer func() { endSpan(span, err) }()
	found, err := h.data.DB().ProjectHibernate.Query().
		Where(projecthibernate.ProjectID(projectID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, errs.Wrap(err, "query project hibernate")
	}
	return toProjectHibernate(found), nil
}

// ListByProjectIDs 按项目 ID 集合批量取休眠记录（key = project_id）；ids 为空返回空 map。
func (h *hibernateRepo) ListByProjectIDs(ctx context.Context, projectIDs ...int) (out map[int]*biz.ProjectHibernate, err error) {
	ctx, span := tracer.Start(ctx, "hibernateRepo/ListByProjectIDs")
	defer func() { endSpan(span, err) }()
	if len(projectIDs) == 0 {
		return map[int]*biz.ProjectHibernate{}, nil
	}
	all, err := h.data.DB().ProjectHibernate.Query().
		Where(projecthibernate.ProjectIDIn(projectIDs...)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list project hibernate by ids")
	}
	out = make(map[int]*biz.ProjectHibernate, len(all))
	for _, e := range all {
		out[e.ProjectID] = toProjectHibernate(e)
	}
	return out, nil
}

// ListEnabled 返回全部已开启项目级开关的休眠记录（休眠对账用）。对账据此单趟遍历全部启用
// 记录并按状态即时归位，开关已关的项目不在集合内。
func (h *hibernateRepo) ListEnabled(ctx context.Context) (out []*biz.ProjectHibernate, err error) {
	ctx, span := tracer.Start(ctx, "hibernateRepo/ListEnabled")
	defer func() { endSpan(span, err) }()
	all, err := h.data.DB().ProjectHibernate.Query().
		Where(projecthibernate.Enabled(true)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list enabled project hibernate")
	}
	return mapHibernate(all), nil
}

// Upsert 按 project_id 写入或更新休眠记录（唯一索引 project_id 保证幂等），返回落库后的行。
func (h *hibernateRepo) Upsert(ctx context.Context, in *biz.ProjectHibernate) (out *biz.ProjectHibernate, err error) {
	ctx, span := tracer.Start(ctx, "hibernateRepo/Upsert")
	defer func() { endSpan(span, err) }()
	err = h.data.DB().ProjectHibernate.Create().
		SetProjectID(in.ProjectID).
		SetEnabled(in.Enabled).
		SetState(string(in.State)).
		SetSnapshot(in.Snapshot).
		SetWindowNames(in.WindowNames).
		OnConflict(sql.ConflictColumns(projecthibernate.FieldProjectID)).
		UpdateNewValues().
		Exec(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "upsert project hibernate")
	}
	return h.Get(ctx, in.ProjectID)
}

// Delete 删除项目的休眠记录（项目删除时清理）。
func (h *hibernateRepo) Delete(ctx context.Context, projectID int) (err error) {
	ctx, span := tracer.Start(ctx, "hibernateRepo/Delete")
	defer func() { endSpan(span, err) }()
	_, err = h.data.DB().ProjectHibernate.Delete().
		Where(projecthibernate.ProjectID(projectID)).
		Exec(ctx)
	return errs.Wrap(err, "delete project hibernate")
}

// mapHibernate 批量转换 ent 记录为领域对象。
func mapHibernate(all []*ent.ProjectHibernate) []*biz.ProjectHibernate {
	out := make([]*biz.ProjectHibernate, 0, len(all))
	for _, e := range all {
		out = append(out, toProjectHibernate(e))
	}
	return out
}
