package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// HibernateWindow 是全局共享的休眠时间窗：管理员在后台维护若干命名时段，各仓库的
// 定时休眠配置按名字引用其中一个（repo.MarsConfig.Hibernate.WindowName）。时间窗全局
// 共享、副本字段仓库各自配置——本类型只承载时段本身。
type HibernateWindow struct {
	ID          int
	Name        string
	Start       string
	End         string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ValidateHibernateWindow 校验全局休眠时间窗：名称非空、启止时间可解析为 HH:MM 且不相同。
// 启用仓库休眠后，对账会按名称解析该窗，故名称必须可解析定位、时段必须有效。
func ValidateHibernateWindow(name, start, end string) error {
	if strings.TrimSpace(name) == "" {
		return errs.WrapInvalidArgument(errors.New("时间窗名称不能为空"), "校验休眠时间窗")
	}
	if _, ok := parseClockMinutes(start); !ok {
		return errs.WrapInvalidArgument(fmt.Errorf("休眠开始时间 %q 非法，应为 HH:MM", start), "校验休眠时间窗")
	}
	if _, ok := parseClockMinutes(end); !ok {
		return errs.WrapInvalidArgument(fmt.Errorf("休眠结束时间 %q 非法，应为 HH:MM", end), "校验休眠时间窗")
	}
	if start == end {
		return errs.WrapInvalidArgument(errors.New("休眠开始与结束时间不能相同"), "校验休眠时间窗")
	}
	return nil
}

// HibernateWindowRepo 是全局休眠时间窗的持久化端口。
type HibernateWindowRepo interface {
	// List 返回全部时间窗（按名称升序）。
	List(ctx context.Context) ([]*HibernateWindow, error)
	// GetByName 按名称查时间窗；未命中返回 (nil, nil)，不把"不存在"当错误。
	GetByName(ctx context.Context, name string) (*HibernateWindow, error)
	// GetByID 按 ID 查时间窗；未命中返回 (nil, nil)（删除路径据此判"窗口是否存在"以实现幂等）。
	GetByID(ctx context.Context, id int) (*HibernateWindow, error)
	// ReposReferencing 返回按 mars_config.hibernate.window_name == name 引用该时间窗的仓库名（升序）。
	// 删除前据此判"是否仍有下游关联"：有引用则不允许删除。
	ReposReferencing(ctx context.Context, name string) ([]string, error)
	// Create 新建时间窗；名称重复由唯一索引兜底（返回 409）。
	Create(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	// Update 按 ID 更新时间窗；不存在返回 NotFound。
	Update(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	// UpdateCascade 原子改名：在同一事务内把窗口更新为 w，并把所有按 oldName 引用它的仓库的
	// window_name 改写为 w.Name，任一失败整体回滚；返回 (更新后的窗口, 被迁移引用的仓库数)。
	// 不存在返回 404、名称重复返回 409。
	UpdateCascade(ctx context.Context, w *HibernateWindow, oldName string) (*HibernateWindow, int, error)
	// Delete 按 ID 删除时间窗（物理删除）；不存在不报错。
	Delete(ctx context.Context, id int) error
}

// HibernateWindowBiz 收口全局休眠时间窗的后台增删改查；对账按名称解析窗口走 Repo 端口。
type HibernateWindowBiz interface {
	// List 返回全部时间窗。
	List(ctx context.Context) ([]*HibernateWindow, error)
	// Create 新建时间窗：校验名称与时段，名称重复返回 409。
	Create(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	// Update 更新时间窗：校验名称与时段，不存在返回 404，名称重复返回 409。改名时按 cascade 分流：
	// cascade=true 原子迁移引用旧名的仓库（返回迁移仓库数）；cascade=false 时旧名仍被引用则拒绝（400）。
	Update(ctx context.Context, w *HibernateWindow, cascade bool) (*HibernateWindow, int, error)
	// Delete 删除时间窗（物理删除，幂等）：仍有仓库按名引用时拒绝（400），无关联时删除。
	Delete(ctx context.Context, id int) error
}

type hibernateWindowBiz struct {
	repo HibernateWindowRepo
}

// NewHibernateWindowBiz 构造全局休眠时间窗 biz。
func NewHibernateWindowBiz(repo HibernateWindowRepo) HibernateWindowBiz {
	return &hibernateWindowBiz{repo: repo}
}

// List 返回全部时间窗（透传 repo）。
func (b *hibernateWindowBiz) List(ctx context.Context) ([]*HibernateWindow, error) {
	return b.repo.List(ctx)
}

// Create 新建时间窗：先校验名称与时段，再落库（名称重复由唯一索引兜底返回 409）。
func (b *hibernateWindowBiz) Create(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
	if err := ValidateHibernateWindow(w.Name, w.Start, w.End); err != nil {
		return nil, err
	}
	return b.repo.Create(ctx, w)
}

// Update 更新时间窗：先校验名称与时段，再落库（不存在返回 404、名称重复返回 409）。
// 改名时按 cascade 分流——时间窗被仓库的 mars_config.hibernate.window_name 按名引用，
// 改名会让这些仓库的对账按旧名未命中而静默跳过，令其休眠悄然失效：
//   - cascade=true：原子把引用旧名的仓库改写为新名（返回迁移仓库数）；
//   - cascade=false：旧名仍被引用则拒绝（400），提示改以级联方式确认（对称 Delete 的引用守卫）。
func (b *hibernateWindowBiz) Update(ctx context.Context, w *HibernateWindow, cascade bool) (*HibernateWindow, int, error) {
	if err := ValidateHibernateWindow(w.Name, w.Start, w.End); err != nil {
		return nil, 0, err
	}
	existing, err := b.repo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, 0, err
	}
	// 窗口不存在或名称未变：无引用需迁移，走普通更新（不存在交由 repo.Update 落 404）。
	if existing == nil || existing.Name == w.Name {
		updated, err := b.repo.Update(ctx, w)
		return updated, 0, err
	}
	// 改名且级联：由 repo 在同一事务内改名 + 迁移引用，返回迁移仓库数。
	if cascade {
		return b.repo.UpdateCascade(ctx, w, existing.Name)
	}
	if err := b.guardRenameRefs(ctx, existing.Name); err != nil {
		return nil, 0, err
	}
	updated, err := b.repo.Update(ctx, w)
	return updated, 0, err
}

// guardRenameRefs 校验「非级联改名」不击穿下游引用：旧名仍被仓库引用则返回 400，
// 提示改用级联（cascade=true）或先手动迁移引用。
func (b *hibernateWindowBiz) guardRenameRefs(ctx context.Context, oldName string) error {
	refs, err := b.repo.ReposReferencing(ctx, oldName)
	if err != nil {
		return err
	}
	if len(refs) == 0 {
		return nil
	}
	return errs.WrapInvalidArgument(
		fmt.Errorf("休眠时间窗 %q 仍被 %d 个仓库引用（%s），改名会让这些仓库的休眠静默失效；请以级联方式改名，或先手动把引用迁移到新名后再改",
			oldName, len(refs), summarizeHibernateWindowRefs(refs)),
		"更新休眠时间窗",
	)
}

// Delete 删除时间窗：先校验无下游关联（无仓库按名引用）再物理删除；窗口不存在时幂等成功。
// 存在引用则拒绝（400）——时间窗被仓库的 mars_config.hibernate.window_name 按名引用，
// 删掉会让这些仓库的对账按名未命中而静默跳过，令其休眠悄然失效，故必须先解除引用。
func (b *hibernateWindowBiz) Delete(ctx context.Context, id int) error {
	w, err := b.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 窗口不存在：幂等成功，不触达删除（对齐 repo.Delete 的幂等语义）。
	if w == nil {
		return nil
	}
	refs, err := b.repo.ReposReferencing(ctx, w.Name)
	if err != nil {
		return err
	}
	if len(refs) > 0 {
		return errs.WrapInvalidArgument(
			fmt.Errorf("休眠时间窗 %q 仍被 %d 个仓库引用（%s），请先在仓库配置中解除引用后再删除",
				w.Name, len(refs), summarizeHibernateWindowRefs(refs)),
			"删除休眠时间窗",
		)
	}
	return b.repo.Delete(ctx, id)
}

// summarizeHibernateWindowRefs 把引用该时间窗的仓库名压成报错用的短串：最多列前 5 个，
// 超出则以「等 N 个」收尾，避免引用方很多时错误消息过长。
func summarizeHibernateWindowRefs(names []string) string {
	const limit = 5
	if len(names) > limit {
		return fmt.Sprintf("%s 等 %d 个", strings.Join(names[:limit], ", "), len(names))
	}
	return strings.Join(names, ", ")
}
