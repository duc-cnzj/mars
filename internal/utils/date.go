package utils

import (
	"math"
	"time"

	"github.com/dustin/go-humanize"
)

var magnitudes = []humanize.RelTimeMagnitude{
	{time.Second, "现在", time.Second},
	{2 * time.Second, "1 秒%s", 1},
	{time.Minute, "%d 秒%s", time.Second},
	{2 * time.Minute, "1 分钟%s", 1},
	{time.Hour, "%d 分钟%s", time.Minute},
	{2 * time.Hour, "1 小时%s", 1},
	{humanize.Day, "%d 小时%s", time.Hour},
	{2 * humanize.Day, "1 天%s", 1},
	{humanize.Week, "%d 天%s", humanize.Day},
	{2 * humanize.Week, "1 周%s", 1},
	{humanize.Month, "%d 周%s", humanize.Week},
	{2 * humanize.Month, "1 个月%s", 1},
	{humanize.Year, "%d 个月%s", humanize.Month},
	{18 * humanize.Month, "1 年%s", 1},
	{2 * humanize.Year, "2 年%s", 1},
	{humanize.LongTime, "%d 年%s", humanize.Year},
	{math.MaxInt64, "很久%s", 1},
}

func ToHumanizeDatetimeString(t *time.Time) string {
	return humanize.CustomRelTime(*t, time.Now(), "以前", "从现在起", magnitudes)
}
