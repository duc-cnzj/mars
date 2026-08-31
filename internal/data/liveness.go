package data

import (
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
)

// livenessBoundaries 由分类基准时间 now 推导活跃/僵尸阈值边界（天数取自
// biz.ActiveLivenessDays/ZombieLivenessDays，与 biz.classifyLiveness 单一事实来源一致）：
// 活跃 = updated_at > active；僵尸 = updated_at <= zombie；低活跃 = 两者之间。
// 注意活跃边界取 (ActiveLivenessDays+1) 天：classifyLiveness 以 int(now.Sub(ts).Hours()/24)
// 向下取整天数，elapsed < 31 天才判活跃（ts > now-31d），而非 ts > now-30d——差一秒即错分类。
// now 由 biz 捕获一次传入，SQL 分类过滤与 biz 行级标记共用同一基准，杜绝边界竞态。
func livenessBoundaries(now time.Time) (active, zombie time.Time) {
	return now.Add(-time.Duration(biz.ActiveLivenessDays+1) * 24 * time.Hour),
		now.Add(-time.Duration(biz.ZombieLivenessDays) * 24 * time.Hour)
}
