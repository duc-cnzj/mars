package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// 部署趋势窗口约束：默认 30 天、上限 90 天（服务端权威值，前端面板默认 30 与之对齐）。
const (
	DeployTrendDefaultDays = 30
	DeployTrendMaxDays     = 90
)

// DeployDailyCount 单日部署计数。date 为服务端本地时区 YYYY-MM-DD，
// 调用方原样展示，不做时区换算（避免客户端二次折算切错天界）。
type DeployDailyCount struct {
	Date  string
	Count int
}

// ChangelogBiz 收口项目变更记录业务：查询最近记录、创建记录与按天聚合部署计数。
type ChangelogBiz interface {
	// FindLastChangelogsByProjectID 查询项目最近一批变更记录。
	FindLastChangelogsByProjectID(ctx context.Context, input *FindLastChangelogsByProjectIDChangeLogInput) ([]*Changelog, error)
	// Create 创建一条变更记录。
	Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error)
	// FindLastChangeByProjectID 查询项目最近一条变更记录。
	FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error)
	// DeployDailyCounts 近 days 天每日部署次数（升序、无部署的天补 0、长度恒等于 days）。
	// 天界 = 服务端本地时区（time.Local），数据源 changelog 与项目活跃度 deploy_count 同口径。
	// days 需 ≥1，调用方负责默认/上限收敛（见 DeployTrendDefaultDays/DeployTrendMaxDays）。
	DeployDailyCounts(ctx context.Context, days int) ([]*DeployDailyCount, error)
}

type changelogBiz struct {
	clRepo ChangelogRepo
}

// NewChangelogBiz 构造 changelog biz。
func NewChangelogBiz(clRepo ChangelogRepo) ChangelogBiz {
	return &changelogBiz{clRepo: clRepo}
}

// FindLastChangelogsByProjectID 查询项目的最近变更记录列表（透传 repo）。
func (c *changelogBiz) FindLastChangelogsByProjectID(ctx context.Context, input *FindLastChangelogsByProjectIDChangeLogInput) ([]*Changelog, error) {
	return c.clRepo.FindLastChangelogsByProjectID(ctx, input)
}

// Create 校验输入后创建变更记录。
func (c *changelogBiz) Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error) {
	if input == nil || input.ProjectID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("changelog 不能为空或 project id 不能小于等于 0"), "create changelog")
	}
	return c.clRepo.Create(ctx, input)
}

// FindLastChangeByProjectID 查询项目的最近一条变更记录（透传 repo）。
func (c *changelogBiz) FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error) {
	return c.clRepo.FindLastChangeByProjectID(ctx, projectID)
}

// DeployDailyCounts 近 days 天每日部署次数：repo 只回窗口内 created_at 时间戳，本层
// 在 Go 内按服务端本地时区（time.Local）分桶 + 零填充，杜绝 SQL 会话时区分桶切错天界，
// 也不回表拉 changelog 的 config/git_commit_title 长文本。
func (c *changelogBiz) DeployDailyCounts(ctx context.Context, days int) ([]*DeployDailyCount, error) {
	loc := time.Local
	now := time.Now().In(loc)
	// 窗口 [今天-(days-1) 00:00, 明天 00:00)：末位即今天，保证含当日
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(days - 1))
	endDay := startDay.AddDate(0, 0, days)
	created, err := c.clRepo.SelectCreatedAtBetween(ctx, startDay, endDay)
	if err != nil {
		return nil, err
	}
	byDay := make(map[string]int, days)
	for _, ts := range created {
		t := ts.In(loc)
		byDay[dayKey(t)]++
	}
	out := make([]*DeployDailyCount, 0, days)
	for d := 0; d < days; d++ {
		day := startDay.AddDate(0, 0, d)
		out = append(out, &DeployDailyCount{Date: dayKey(day), Count: byDay[dayKey(day)]})
	}
	return out, nil
}

// dayKey 把时间折叠为本地时区 YYYY-MM-DD 桶键。
func dayKey(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

// ChangelogRepo 是项目变更记录仓库端口。
type ChangelogRepo interface {
	// FindLastChangelogsByProjectID 查询项目最近一批变更记录。
	FindLastChangelogsByProjectID(ctx context.Context, input *FindLastChangelogsByProjectIDChangeLogInput) ([]*Changelog, error)
	// Create 创建一条变更记录。
	Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error)
	// FindLastChangeByProjectID 查询项目最近一条变更记录。
	FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error)
	// CountByProjectIDs 按项目 ID 集合聚合各项目变更记录数（GROUP BY project_id）。
	CountByProjectIDs(ctx context.Context, ids ...int) (map[int]int, error)
	// SelectCreatedAtBetween 取窗口 [since, until) 内全部 changelog 的 created_at（仅返回该列，
	// 不回表读 config/git_commit_title 长文本；软删由拦截器统一过滤）。供部署趋势按天聚合。
	SelectCreatedAtBetween(ctx context.Context, since, until time.Time) ([]time.Time, error)
}
