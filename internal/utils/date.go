package utils

import (
	"time"

	"github.com/dustin/go-humanize"
)

var magnitudes = []humanize.RelTimeMagnitude{
	{time.Second, "现在", time.Second},
	{2 * time.Second, "1 秒%s", 1},
	{time.Minute, "%d 秒%s", time.Second},
	{humanize.Day - time.Second, "%d 分钟%s", time.Minute},
	{humanize.Day, "%d 小时%s", time.Hour},
	{2 * humanize.Day, "1 天%s", 1},
	{humanize.Week, "%d 天%s", humanize.Day},
	{2 * humanize.Week, "1 周%s", 1},
	{6 * humanize.Month, "%d 周%s", humanize.Week},
	{humanize.Year, "%d 月%s", humanize.Month},
}

func ToHumanizeDatetimeString(t *time.Time) string {
	return humanize.CustomRelTime(*t, time.Now(), "以前", "从现在起", magnitudes)
}
