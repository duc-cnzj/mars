package biz

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"google.golang.org/protobuf/proto"
)

// HibernateWindow 是全局共享的休眠时间窗：管理员在后台维护若干命名时段，各仓库的
// 定时休眠配置按名字引用其中若干（repo.MarsConfig.Hibernate.WindowNames）。时间窗全局
// 共享、副本字段仓库各自配置——本类型只承载时段本身。
type HibernateWindow struct {
	ID          int
	Name        string
	Start       string
	End         string
	Description string
	// RepoNames 引用该窗的仓库名（升序）；仅 List 填充，供后台列表展示「哪些仓库用了这个窗」。
	RepoNames []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HibernateRepoOption 是可引用休眠时间窗的候选仓库：已配置休眠段（mars_config.hibernate
// 非空，无论是否启用）的仓库。未配休眠段的仓库无法按名引用窗，故不进候选。
type HibernateRepoOption struct {
	ID   int
	Name string
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
	// WindowRepoRefs 返回「窗名 → 引用该窗的仓库名（升序）」映射：单趟扫描全部带
	// mars_config 的仓库、按 window_names 分组。List 据此一次填全部窗的 RepoNames，
	// ReposReferencing 是它的单名薄封装。
	WindowRepoRefs(ctx context.Context) (map[string][]string, error)
	// ReposReferencing 返回 mars_config.hibernate.window_names 含 name 的仓库名（升序）。
	// 删除/改名时间窗前据此判"是否仍有下游关联"。实现为 WindowRepoRefs 的单名查表。
	ReposReferencing(ctx context.Context, name string) ([]string, error)
	// ProjectsReferencing 返回 project_hibernates.window_names 含 name 的项目 ID（升序）。
	// 项目从仓库窗列表里多选子集，故项目引用独立于仓库引用，删除/改名守卫须一并计入。
	ProjectsReferencing(ctx context.Context, name string) ([]int, error)
	// ProjectReferencingRepoIDs 返回「引用该窗的项目 ID → 所属仓库 ID」映射，供「从仓库移除窗」时
	// 按受影响仓库收窄项目子集漂移校验（见 SetWindowRepos）：只拦真正会失去该窗的仓库下的项目。
	ProjectReferencingRepoIDs(ctx context.Context, name string) (map[int]int, error)
	// Create 新建时间窗；名称重复由唯一索引兜底（返回 409）。
	Create(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	// Update 按 ID 更新时间窗；不存在返回 NotFound。
	Update(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	// UpdateCascade 原子改名：在同一事务内把窗口更新为 w，并把所有按 oldName 引用它的仓库
	// （mars_config.hibernate.window_names）与项目的（project_hibernates.window_names）改写为 w.Name，
	// 任一失败整体回滚；返回 (更新后的窗口, 被迁移的引用总数：仓库数 + 项目数)。
	// 不存在返回 404、名称重复返回 409。
	UpdateCascade(ctx context.Context, w *HibernateWindow, oldName string) (*HibernateWindow, int, error)
	// Delete 按 ID 删除时间窗（物理删除）；不存在不报错。
	Delete(ctx context.Context, id int) error
}

// HibernateWindowBiz 收口全局休眠时间窗的后台增删改查；对账按名称解析窗口走 Repo 端口。
type HibernateWindowBiz interface {
	// List 返回全部时间窗，并填充每窗引用的仓库名（RepoNames，供后台列表展示）。仓库引用
	// 单趟扫描后一次填全部窗。
	List(ctx context.Context) ([]*HibernateWindow, error)
	// Create 新建时间窗：校验名称与时段，名称重复返回 409。
	Create(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	// Update 更新时间窗：校验名称与时段，不存在返回 404，名称重复返回 409。改名时按 cascade 分流：
	// cascade=true 原子迁移引用旧名的仓库与项目（返回迁移引用数）；cascade=false 时旧名仍被引用则拒绝（400）。
	Update(ctx context.Context, w *HibernateWindow, cascade bool) (*HibernateWindow, int, error)
	// Delete 删除时间窗（物理删除，幂等）：仍有仓库或项目按名引用时拒绝（400），无关联时删除。
	Delete(ctx context.Context, id int) error
	// ListCandidateRepos 列出可引用时间窗的候选仓库（已配置休眠段，含禁用），按名升序。
	ListCandidateRepos(ctx context.Context) ([]*HibernateRepoOption, error)
	// SetWindowRepos 把时间窗 windowID 的引用仓库集合重设为 repoIDs（集合语义：缺则加入、多则移除），
	// 返回保存后引用该窗的仓库名（升序）。窗不存在返回 404；repoIDs 含未配置休眠段的仓库、
	// 移除会使某启用仓库失去最后一个窗、或移除会让受影响仓库下的项目选择越出仓库可用窗列表，返回 400。
	SetWindowRepos(ctx context.Context, windowID int, repoIDs []int) ([]string, error)
}

type hibernateWindowBiz struct {
	repo     HibernateWindowRepo
	repoRepo RepoRepo
}

// NewHibernateWindowBiz 构造全局休眠时间窗 biz。repoRepo 供候选仓库枚举与批量改写仓库引用
// （「一键配置窗到多仓库」）——与本包 NewRepoBiz 依赖 HibernateRepo 端口同一跨聚合模式。
func NewHibernateWindowBiz(repo HibernateWindowRepo, repoRepo RepoRepo) HibernateWindowBiz {
	return &hibernateWindowBiz{repo: repo, repoRepo: repoRepo}
}

// List 返回全部时间窗，并填充每窗引用的仓库名：先取窗列表，再单趟扫描仓库引用映射，
// 按窗名填 RepoNames（避免逐窗各扫一遍）。
func (b *hibernateWindowBiz) List(ctx context.Context) ([]*HibernateWindow, error) {
	windows, err := b.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if len(windows) == 0 {
		return windows, nil
	}
	refs, err := b.repo.WindowRepoRefs(ctx)
	if err != nil {
		return nil, err
	}
	for _, w := range windows {
		w.RepoNames = refs[w.Name]
	}
	return windows, nil
}

// Create 新建时间窗：先校验名称与时段，再落库（名称重复由唯一索引兜底返回 409）。
func (b *hibernateWindowBiz) Create(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
	if err := ValidateHibernateWindow(w.Name, w.Start, w.End); err != nil {
		return nil, err
	}
	return b.repo.Create(ctx, w)
}

// Update 更新时间窗：先校验名称与时段，再落库（不存在返回 404、名称重复返回 409）。
// 改名时按 cascade 分流——时间窗被仓库的 mars_config.hibernate.window_names 与项目的
// project_hibernates.window_names 按名引用，改名会让这些引用的对账按旧名未命中而静默跳过，
// 令其休眠悄然失效：
//   - cascade=true：原子把引用旧名的仓库与项目改写为新名（返回迁移引用数）；
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

// guardRenameRefs 校验「非级联改名」不击穿下游引用：旧名仍被仓库或项目引用则返回 400，
// 提示改用级联（cascade=true）或先手动迁移引用。
func (b *hibernateWindowBiz) guardRenameRefs(ctx context.Context, oldName string) error {
	repos, projects, err := b.referencingRefs(ctx, oldName)
	if err != nil {
		return err
	}
	if len(repos) == 0 && len(projects) == 0 {
		return nil
	}
	return errs.WrapInvalidArgument(
		fmt.Errorf("休眠时间窗 %q 仍被 %d 个仓库、%d 个项目引用（%s），改名会让这些引用静默失效；请以级联方式改名，或先手动把引用迁移到新名后再改",
			oldName, len(repos), len(projects), describeHibernateWindowRefs(repos, projects)),
		"更新休眠时间窗",
	)
}

// Delete 删除时间窗：先校验无下游关联（无仓库、无项目按名引用）再物理删除；窗口不存在时幂等成功。
// 存在引用则拒绝（400）——时间窗被仓库的 mars_config.hibernate.window_names 与项目的
// project_hibernates.window_names 按名引用，删掉会让这些引用的对账按名未命中而静默跳过，
// 令其休眠悄然失效，故必须先解除引用（同理，睡眠中的项目会因窗被删而对账判定为出窗并正常唤醒）。
func (b *hibernateWindowBiz) Delete(ctx context.Context, id int) error {
	w, err := b.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 窗口不存在：幂等成功，不触达删除（对齐 repo.Delete 的幂等语义）。
	if w == nil {
		return nil
	}
	repos, projects, err := b.referencingRefs(ctx, w.Name)
	if err != nil {
		return err
	}
	if len(repos) > 0 || len(projects) > 0 {
		return errs.WrapInvalidArgument(
			fmt.Errorf("休眠时间窗 %q 仍被 %d 个仓库、%d 个项目引用（%s），请先解除引用后再删除",
				w.Name, len(repos), len(projects), describeHibernateWindowRefs(repos, projects)),
			"删除休眠时间窗",
		)
	}
	return b.repo.Delete(ctx, id)
}

// ListCandidateRepos 列出可引用时间窗的候选仓库：已配置休眠段（cfg.Hibernate 非空，含禁用）的
// 仓库，按名升序。未配休眠段的仓库无法按名引用窗（window_names 只活在休眠段内），故不进候选。
func (b *hibernateWindowBiz) ListCandidateRepos(ctx context.Context) ([]*HibernateRepoOption, error) {
	all, err := b.repoRepo.All(ctx, &AllRepoRequest{})
	if err != nil {
		return nil, err
	}
	out := make([]*HibernateRepoOption, 0, len(all))
	for _, r := range all {
		if r.GetMarsConfig().Hibernate == nil {
			continue
		}
		out = append(out, &HibernateRepoOption{ID: r.ID, Name: r.Name})
	}
	slices.SortFunc(out, func(a, c *HibernateRepoOption) int { return strings.Compare(a.Name, c.Name) })
	return out, nil
}

// SetWindowRepos 把时间窗 windowID 的引用仓库集合重设为 repoIDs（集合语义：缺则加入、多则移除）：
//  1. 窗必须存在（不存在 404）；
//  2. repoIDs 每个都必须是候选仓库（未配休眠段的仓库不能引用窗，越界 400，不静默忽略）；
//  3. 若某候选仓库本次「失去窗 W」且其下仍有项目选定 W，则拒绝（400 点名仓库与项目）——项目选择的
//     子集不变式（项目选定的窗 ⊆ 仓库可用窗）不能被本操作从仓库侧击穿（对称 Toggle 的项目侧校验）；
//  4. 逐候选仓库把「是否引用本窗」对齐到目标（缺则加入、多则移除），变更后的配置过
//     ValidateHibernateConfig——守住「启用必须有生效窗」不变式：移除某启用仓库的最后一个窗即 400；
//  5. 有变更的仓库在一个事务内批量改写 mars_config（不触发 git 解析）；
//  6. 返回保存后引用该窗的仓库名（升序）。
func (b *hibernateWindowBiz) SetWindowRepos(ctx context.Context, windowID int, repoIDs []int) ([]string, error) {
	w, err := b.repo.GetByID(ctx, windowID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, errs.NotFound("休眠时间窗不存在")
	}
	all, err := b.repoRepo.All(ctx, &AllRepoRequest{})
	if err != nil {
		return nil, err
	}
	candidates := make(map[int]*Repo, len(all))
	for _, r := range all {
		if r.GetMarsConfig().Hibernate != nil {
			candidates[r.ID] = r
		}
	}
	want := make(map[int]bool, len(repoIDs))
	for _, id := range repoIDs {
		if _, ok := candidates[id]; !ok {
			return nil, errs.WrapInvalidArgument(
				fmt.Errorf("仓库 %d 未配置休眠段，不能引用时间窗", id),
				"配置休眠时间窗引用仓库",
			)
		}
		want[id] = true
	}
	// 受影响仓库：本次会「失去窗 W」的候选仓库（当前含 W 但不在目标集合内）。从这些仓库移除 W
	// 会让其下「已选 W 的项目」越出仓库可用窗列表，故在落库前先做子集守卫（拒绝则不产生半量写入）。
	affected := make(map[int]*Repo)
	for id, r := range candidates {
		if want[id] || !slices.Contains(r.GetMarsConfig().Hibernate.WindowNames, w.Name) {
			continue
		}
		affected[id] = r
	}
	if err := b.guardProjectWindowSubset(ctx, w.Name, affected); err != nil {
		return nil, err
	}
	updates := make(map[int]*mars.Config)
	for id, r := range candidates {
		cfg := r.GetMarsConfig()
		has := slices.Contains(cfg.Hibernate.WindowNames, w.Name)
		if has == want[id] {
			continue
		}
		next := withWindowNames(cfg, setWindowNamePresence(cfg.Hibernate.WindowNames, w.Name, want[id]))
		if err := ValidateHibernateConfig(next); err != nil {
			return nil, errs.InvalidArgument(fmt.Sprintf("仓库 %q：%v", r.Name, err))
		}
		updates[id] = next
	}
	if len(updates) > 0 {
		if err := b.repoRepo.BatchUpdateMarsConfig(ctx, updates); err != nil {
			return nil, err
		}
	}
	return b.repo.ReposReferencing(ctx, w.Name)
}

// guardProjectWindowSubset 校验「从仓库移除窗 W」不击穿项目选择的子集不变式（项目选定的窗必须
// 落在其仓库可用窗列表内）：若受影响仓库（本次会失去 W）下仍有项目选定 W，这些项目的选择将越出
// 仓库可用窗列表，拒绝并点名仓库与项目，提示先在项目侧取消对这些窗的休眠选择（对称 Toggle 的
// 子集校验）。只统计受影响仓库下的项目，避免误伤引用同名窗的其它仓库。affected 为空即无窗被移除。
func (b *hibernateWindowBiz) guardProjectWindowSubset(ctx context.Context, name string, affected map[int]*Repo) error {
	if len(affected) == 0 {
		return nil
	}
	refs, err := b.repo.ProjectReferencingRepoIDs(ctx, name)
	if err != nil {
		return err
	}
	blocked := make([]int, 0)
	for projectID, repoID := range refs {
		if _, ok := affected[repoID]; ok {
			blocked = append(blocked, projectID)
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	slices.Sort(blocked)
	repoNames := make([]string, 0, len(affected))
	for _, r := range affected {
		repoNames = append(repoNames, r.Name)
	}
	slices.Sort(repoNames)
	return errs.InvalidArgument(fmt.Sprintf(
		"时间窗 %q 仍被仓库 %s 下的项目 %s 引用：请先在项目详情里取消这些项目对该窗的休眠选择，再从仓库移除",
		name, summarizeHibernateWindowRefs(repoNames), summarizeHibernateWindowIDs(blocked),
	))
}

// withWindowNames 返回 cfg 的副本并把休眠段引用的窗名替换为 names（其余字段原样保留）。
// 必须克隆后再改：cfg 来自仓库实体、可能被调用方复用，就地改会污染调用方持有的配置。
func withWindowNames(cfg *mars.Config, names []string) *mars.Config {
	out := proto.Clone(cfg).(*mars.Config)
	out.Hibernate.WindowNames = names
	return out
}

// setWindowNamePresence 返回把 name 的「存在性」对齐到 present 后的新切片（保留其余元素顺序）：
// present=true 且原本不含则追加；present=false 则移除所有 name 出现。原切片不改动。
func setWindowNamePresence(names []string, name string, present bool) []string {
	out := make([]string, 0, len(names)+1)
	for _, n := range names {
		if n == name {
			continue
		}
		out = append(out, n)
	}
	if present {
		out = append(out, name)
	}
	return out
}

// referencingRefs 汇总按名引用该时间窗的下游：仓库名与项目 ID 两组。
func (b *hibernateWindowBiz) referencingRefs(ctx context.Context, name string) (repos []string, projects []int, err error) {
	repos, err = b.repo.ReposReferencing(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	projects, err = b.repo.ProjectsReferencing(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	return repos, projects, nil
}

// describeHibernateWindowRefs 把引用该时间窗的仓库名与项目 ID 压成报错用的短串：两类各限 5 个，
// 超出以「等 N 个」收尾，避免引用方很多时错误消息过长；空类省略。
func describeHibernateWindowRefs(repoNames []string, projectIDs []int) string {
	segs := make([]string, 0, 2)
	if len(repoNames) > 0 {
		segs = append(segs, fmt.Sprintf("仓库 %s", summarizeHibernateWindowRefs(repoNames)))
	}
	if len(projectIDs) > 0 {
		segs = append(segs, fmt.Sprintf("项目 %s", summarizeHibernateWindowIDs(projectIDs)))
	}
	return strings.Join(segs, "，")
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

// summarizeHibernateWindowIDs 把引用该时间窗的项目 ID 压成报错用的短串：最多列前 5 个，
// 超出则以「等 N 个」收尾。
func summarizeHibernateWindowIDs(ids []int) string {
	const limit = 5
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, strconv.Itoa(id))
	}
	if len(strs) > limit {
		return fmt.Sprintf("%s 等 %d 个", strings.Join(strs[:limit], ", "), len(strs))
	}
	return strings.Join(strs, ", ")
}
