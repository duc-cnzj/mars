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
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// HibernateState 是项目定时休眠的运行态。
type HibernateState string

const (
	// HibernateStateAwake 清醒（未休眠）。零值语义，未配置休眠的项目一律视为清醒。
	HibernateStateAwake HibernateState = "awake"
	// HibernateStateSleeping 已休眠：副本字段被置 0 重部署，等待窗口结束后唤醒。
	HibernateStateSleeping HibernateState = "sleeping"
)

// ProjectHibernate 是项目定时休眠的运行态（1:1 项目）：项目级开关 + 运行态 + 休眠前
// 副本字段的值快照。副本字段映射与时间窗是仓库级配置（repo.MarsConfig.Hibernate），不在此。
type ProjectHibernate struct {
	ID        int
	ProjectID int
	Enabled   bool
	State     HibernateState
	// Snapshot 休眠前副本字段的值，唤醒时据此恢复。
	Snapshot []*websocket_pb.ExtraValue
	// WindowNames 项目从仓库窗列表里多选的子集（引用 repo.MarsConfig.Hibernate.WindowNames）。
	// 启用（Enabled）时必须非空——空选择不再是「继承仓库全量」，而是「未选」，Toggle 会拒；
	// 关闭时保留上次选择，供重新开启时回显。
	WindowNames []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// normalizeWindowNames 归一化时间窗名列表：逐个 trim、丢弃空串、按首次出现顺序去重。
// 返回切片元素均非空；无有效元素时返回 nil。仓库配置与项目选择共用同一口径。
func normalizeWindowNames(names []string) []string {
	seen := make(map[string]bool, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// normalizeRepoHibernateWindowNames 就地把仓库配置里引用的窗名归一化（trim + 去重 + 丢空），令仓库侧
// 引用与时间窗的规范名对齐：项目侧 Toggle 已在写入前归一化，仓库侧此前遗漏——经 API/导入存入带空白
// 的窗名（如 " 夜间 "）会让按规范名的一切判定（列表「引用仓库」展示、删窗/改名守卫、SetWindowRepos
// 候选比对）全部落空，留下悬空引用并令休眠静默失效。cfg 或 Hibernate 为 nil 时不动（禁用态无引用）。
func normalizeRepoHibernateWindowNames(cfg *mars.Config) {
	if cfg == nil || cfg.Hibernate == nil {
		return
	}
	cfg.Hibernate.WindowNames = normalizeWindowNames(cfg.Hibernate.WindowNames)
}

// removedWindowNames 返回「oldNames 有而 nextNames 没有」的窗名（按 oldNames 归一化后的顺序）。
// 两侧都经 normalizeWindowNames 归一化（trim + 去重），故比较的是规范名。用于仓库编辑表单移除窗
// 时判定「本次移除了哪些窗」。
func removedWindowNames(oldNames, nextNames []string) []string {
	keep := make(map[string]bool, len(nextNames))
	for _, n := range normalizeWindowNames(nextNames) {
		keep[n] = true
	}
	var out []string
	for _, n := range normalizeWindowNames(oldNames) {
		if !keep[n] {
			out = append(out, n)
		}
	}
	return out
}

// intersectsWindowNames 判断 selected 与 names 是否存在交集（用于「某项目是否选用了被移除的窗」）。
func intersectsWindowNames(selected, names []string) bool {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	for _, n := range selected {
		if set[n] {
			return true
		}
	}
	return false
}

// InHibernateWindow 判定 now 是否落在 [start, end) 休眠窗口内（服务端本地时区，HH:MM）。
// start > end 表示跨零点（如 23:00→06:00）。任一时刻无法解析、或 start == end 时视为
// 无窗口，返回 false。
func InHibernateWindow(now time.Time, start, end string) bool {
	from, ok1 := parseClockMinutes(start)
	to, ok2 := parseClockMinutes(end)
	if !ok1 || !ok2 || from == to {
		return false
	}
	cur := now.Hour()*60 + now.Minute()
	if from < to {
		return cur >= from && cur < to
	}
	return cur >= from || cur < to
}

// parseClockMinutes 把 "HH:MM" 解析为「当天第几分钟」，非法返回 ok=false。
func parseClockMinutes(s string) (int, bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, false
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, false
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

// SleepExtraValues 返回把副本字段置为 "0" 后的额外配置项（新切片，不改原切片也不共享
// 元素指针）：命中的字段改值为 "0"，原切片未出现的字段追加 {Path, "0"}。
func SleepExtraValues(cur []*websocket_pb.ExtraValue, replicaPaths []string) []*websocket_pb.ExtraValue {
	want := make(map[string]bool, len(replicaPaths))
	for _, p := range replicaPaths {
		want[p] = true
	}
	out := make([]*websocket_pb.ExtraValue, 0, len(cur)+len(replicaPaths))
	seen := make(map[string]bool, len(replicaPaths))
	for _, v := range cur {
		if want[v.Path] {
			seen[v.Path] = true
			out = append(out, &websocket_pb.ExtraValue{Path: v.Path, Value: "0", Description: v.Description})
			continue
		}
		out = append(out, &websocket_pb.ExtraValue{Path: v.Path, Value: v.Value, Description: v.Description})
	}
	for _, p := range replicaPaths {
		if !seen[p] {
			out = append(out, &websocket_pb.ExtraValue{Path: p, Value: "0"})
		}
	}
	return out
}

// SnapshotReplicaValues 从全量元素值里截取副本字段的值作为休眠快照，未出现的字段跳过。
func SnapshotReplicaValues(all []*websocket_pb.ExtraValue, replicaPaths []string) []*websocket_pb.ExtraValue {
	byPath := make(map[string]*websocket_pb.ExtraValue, len(all))
	for _, v := range all {
		byPath[v.Path] = v
	}
	out := make([]*websocket_pb.ExtraValue, 0, len(replicaPaths))
	for _, p := range replicaPaths {
		if v, ok := byPath[p]; ok {
			out = append(out, &websocket_pb.ExtraValue{Path: p, Value: v.Value})
		}
	}
	return out
}

// WakeExtraValues 用快照值回填副本字段（在 cur 基础上覆盖，不新增快照之外的字段）。
func WakeExtraValues(cur, snapshot []*websocket_pb.ExtraValue) []*websocket_pb.ExtraValue {
	want := make(map[string]string, len(snapshot))
	for _, v := range snapshot {
		want[v.Path] = v.Value
	}
	out := make([]*websocket_pb.ExtraValue, 0, len(cur))
	for _, v := range cur {
		if value, ok := want[v.Path]; ok {
			out = append(out, &websocket_pb.ExtraValue{Path: v.Path, Value: value, Description: v.Description})
			continue
		}
		out = append(out, &websocket_pb.ExtraValue{Path: v.Path, Value: v.Value, Description: v.Description})
	}
	return out
}

// ReplicaValuesAllZero 判断副本字段在 cur 中是否全部为 "0"（休眠落 0 态的标志值）。
// 唤醒侧据此区分「项目仍处休眠态」（可安全按快照回填）与「用户已手动重部署改过副本数」
// （应放弃自动唤醒、保留人的意图）：只要有一个副本字段非 "0"（含字段缺失），即视为已被外部改动。
// replicaPaths 为空时返回 true（无字段可判，按「无改动」处理，交回正常唤醒流程）。
func ReplicaValuesAllZero(cur []*websocket_pb.ExtraValue, replicaPaths []string) bool {
	byPath := make(map[string]string, len(cur))
	for _, v := range cur {
		byPath[v.Path] = v.Value
	}
	for _, p := range replicaPaths {
		if byPath[p] != "0" {
			return false
		}
	}
	return true
}

// ValidateHibernateConfig 校验仓库级休眠配置：启用时要求副本字段非空、每个字段都命中
// 已定义的自定义字段、且字段类型能承载 "0"，并至少引用了一个全局休眠时间窗（window_names 非空）。
// 时间窗本身的有效性（HH:MM、启止不同）由后台时间窗管理校验；本函数只保证引用非空——
// 保持纯函数不查库（导入路径不经过本函数；删除时间窗的关联校验在后台删窗路径，不在此；
// 其余途径产生的悬空引用由对账按未命中跳过兜底）。禁用时不校验。
func ValidateHibernateConfig(cfg *mars.Config) error {
	if cfg == nil || cfg.Hibernate == nil || !cfg.Hibernate.Enabled {
		return nil
	}
	h := cfg.Hibernate
	if len(h.ReplicaPaths) == 0 {
		return errs.WrapInvalidArgument(errors.New("启用定时休眠时，副本字段不能为空"), "校验仓库休眠配置")
	}
	byPath := make(map[string]*mars.Element, len(cfg.Elements))
	for _, e := range cfg.Elements {
		byPath[e.Path] = e
	}
	for _, p := range h.ReplicaPaths {
		element, ok := byPath[p]
		if !ok {
			return errs.WrapInvalidArgument(fmt.Errorf("副本字段 %s 不在已定义的自定义字段中", p), "校验仓库休眠配置")
		}
		if !elementSupportsZero(element) {
			return errs.WrapInvalidArgument(
				fmt.Errorf("副本字段 %s 的类型无法承载 0（需 InputNumber，或白名单含 \"0\" 的 Select/Radio/Number* 字段）", p),
				"校验仓库休眠配置",
			)
		}
	}
	if len(normalizeWindowNames(h.WindowNames)) == 0 {
		return errs.WrapInvalidArgument(errors.New("启用定时休眠时，至少选择一个全局休眠时间窗"), "校验仓库休眠配置")
	}
	return nil
}

// elementSupportsZero 判定自定义字段类型能否承载值 "0"：InputNumber 可承载任意整数；
// Select/Radio/NumberSelect/NumberRadio 走白名单校验（见 deploy.ElementsLoader.typedValue），
// 故白名单须含 "0"；其余类型（字符串/多行/开关）语义上不是副本数，不支持。
func elementSupportsZero(e *mars.Element) bool {
	switch e.Type {
	case mars.ElementType_ElementTypeInputNumber:
		return true
	case mars.ElementType_ElementTypeSelect,
		mars.ElementType_ElementTypeRadio,
		mars.ElementType_ElementTypeNumberSelect,
		mars.ElementType_ElementTypeNumberRadio:
		return slices.Contains(e.SelectValues, "0")
	default:
		return false
	}
}

// HibernateRepo 是项目休眠运行态的持久化端口。
type HibernateRepo interface {
	// Get 按项目 ID 查休眠记录；未配置返回 (nil, nil)。
	Get(ctx context.Context, projectID int) (*ProjectHibernate, error)
	// ListByProjectIDs 按项目 ID 集合批量取休眠记录（key = project_id）。
	ListByProjectIDs(ctx context.Context, projectIDs ...int) (map[int]*ProjectHibernate, error)
	// ListEnabled 返回全部已开启项目级开关的休眠记录（休眠对账用）。对账据此单趟遍历
	// 全部启用记录，故开关已关的项目天然不在集合内——即「关闭开关即放弃接管」的总闸。
	ListEnabled(ctx context.Context) ([]*ProjectHibernate, error)
	// Upsert 按 project_id 写入或更新休眠记录。
	Upsert(ctx context.Context, h *ProjectHibernate) (*ProjectHibernate, error)
	// Delete 删除项目的休眠记录（项目删除时清理）。
	Delete(ctx context.Context, projectID int) error
}

// HibernateDeployer 执行项目休眠/唤醒的实际重部署。实现落在 internal/deploy——只有
// deploy 能调用 ApplyProject（deploy → biz 单向，biz 反向 import 即成环）。
type HibernateDeployer interface {
	// Redeploy 以给定额外配置项重部署项目：副本字段的目标值由调用方（对账）算好放进
	// extraValues，本方法只负责把它经 helm 重部署落库。
	Redeploy(ctx context.Context, proj *Project, extraValues []*websocket_pb.ExtraValue) error
}

// HibernateBiz 收口项目定时休眠的开关与运行态查询；休眠/唤醒的重部署与周期对账分别在
// deploy / cronjob 层，本层不触碰部署与调度。
type HibernateBiz interface {
	// ListByProjectIDs 批量查询项目休眠运行态（key = project_id）。
	ListByProjectIDs(ctx context.Context, projectIDs ...int) (map[int]*ProjectHibernate, error)
	// Toggle 打开/关闭项目级休眠开关。打开时要求所属仓库已启用休眠能力且项目至少选定一个时间窗，
	// 任一不满足即 400（空选择不再表示「继承仓库全量」，而是「未选」，须提示先选）；
	// windowNames 是项目从仓库时间窗里多选的子集，每个必须属于仓库可用窗列表（否则 400）。
	// 打开时若既有休眠态已名不副实（副本字段非全 0），重置为清醒并清空快照。
	// 仅切换开关，不触发重部署（休眠/唤醒由对账在窗口边界推进）。
	Toggle(ctx context.Context, projectID int, enabled bool, windowNames []string) (*ProjectHibernate, error)
}

type hibernateBiz struct {
	repo     HibernateRepo
	projRepo ProjectRepo
	repoRepo RepoRepo
}

// NewHibernateBiz 构造休眠 biz：repo 持久化运行态，projRepo/repoRepo 用于开关校验。
func NewHibernateBiz(repo HibernateRepo, projRepo ProjectRepo, repoRepo RepoRepo) HibernateBiz {
	return &hibernateBiz{repo: repo, projRepo: projRepo, repoRepo: repoRepo}
}

// ListByProjectIDs 批量查询项目休眠运行态（透传 repo）。
func (h *hibernateBiz) ListByProjectIDs(ctx context.Context, projectIDs ...int) (map[int]*ProjectHibernate, error) {
	if len(projectIDs) == 0 {
		return map[int]*ProjectHibernate{}, nil
	}
	return h.repo.ListByProjectIDs(ctx, projectIDs...)
}

// Toggle 打开/关闭项目级休眠开关：校验项目存在、打开时所属仓库已启用休眠能力、且项目至少选定
// 一个时间窗（空选择即 400）、选定的窗口子集落在仓库可用窗列表内，再按 project_id upsert
// （已存在则保留运行态与快照，只改开关与窗口选择）。打开时若既有休眠态已名不副实（副本字段非
// 全 0），重置为清醒并清空快照，交回对账重新判定。
func (h *hibernateBiz) Toggle(ctx context.Context, projectID int, enabled bool, windowNames []string) (*ProjectHibernate, error) {
	proj, err := h.projRepo.Show(ctx, projectID)
	if err != nil {
		return nil, err
	}
	sel := normalizeWindowNames(windowNames)
	current, err := h.repo.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	state := HibernateStateAwake
	var snapshot []*websocket_pb.ExtraValue
	if current != nil {
		state = current.State
		snapshot = current.Snapshot
	}
	if enabled {
		repo, rerr := h.repoRepo.Get(ctx, proj.RepoID)
		if rerr != nil {
			return nil, rerr
		}
		cfg := repo.GetMarsConfig()
		if cfg.Hibernate == nil || !cfg.Hibernate.Enabled ||
			len(cfg.Hibernate.ReplicaPaths) == 0 || len(normalizeWindowNames(cfg.Hibernate.WindowNames)) == 0 {
			return nil, errs.WrapInvalidArgument(
				fmt.Errorf("仓库 %s 未启用定时休眠（需在后台配置副本字段与时间窗）", repo.Name),
				"切换项目休眠开关",
			)
		}
		// 启用必须显式选择至少一个时间窗：空选择 = 未选（不继承仓库全量），拒绝开启并提示先选。
		// 清空时间窗不走这里——它由前端以 enabled=false + 空窗提交，直接落到函数末尾的 Upsert，
		// 语义为「不再休眠该项目」（项目级休眠必须有生效窗，无窗即不停机）；故本闸门只在
		// 真正开启（enabled=true）时生效。
		if len(sel) == 0 {
			return nil, errs.WrapInvalidArgument(
				errors.New("启用定时休眠时，至少选择一个时间窗"),
				"切换项目休眠开关",
			)
		}
		if err := validateWindowSubset(sel, cfg.Hibernate.WindowNames); err != nil {
			return nil, err
		}
		// 重新开启时校正陈旧休眠态：仅当副本字段确实仍全为 "0"（真在休眠）才沿用 sleeping + 快照。
		// 关闭期间若被手动改过副本数，旧快照已失效、状态名不副实——重置为 awake 并清空快照，
		// 交回对账按当前值重新判定（窗口内重新快照并休眠、出窗保持清醒），
		// 避免出现「标记休眠中、实际有副本」的错位。
		if state == HibernateStateSleeping && !ReplicaValuesAllZero(proj.FinalExtraValues, cfg.Hibernate.ReplicaPaths) {
			state = HibernateStateAwake
			snapshot = nil
		}
	}
	return h.repo.Upsert(ctx, &ProjectHibernate{
		ProjectID:   projectID,
		Enabled:     enabled,
		State:       state,
		Snapshot:    snapshot,
		WindowNames: sel,
	})
}

// validateWindowSubset 校验项目选定的窗口名子集是否全部落在仓库可用窗列表内。空选择的
// 拦截在 Toggle 调用前完成；命中首个不在列表内的名字即返回 400。
func validateWindowSubset(selected, available []string) error {
	allowed := make(map[string]bool, len(available))
	for _, n := range normalizeWindowNames(available) {
		allowed[n] = true
	}
	for _, n := range selected {
		if !allowed[n] {
			return errs.WrapInvalidArgument(
				fmt.Errorf("时间窗 %s 不在仓库可用窗列表中", n),
				"切换项目休眠开关",
			)
		}
	}
	return nil
}
