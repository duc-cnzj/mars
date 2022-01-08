package utils

import (
	"math"
	"time"

	"github.com/dustin/go-humanize"
)

var magnitudes = []humanize.RelTimeMagnitude{
	{D: time.Second, Format: "现在", DivBy: time.Second},
	{D: 2 * time.Second, Format: "1 秒%s", DivBy: 1},
	{D: time.Minute, Format: "%d 秒%s", DivBy: time.Second},
	{D: 2 * time.Minute, Format: "1 分钟%s", DivBy: 1},
	{D: time.Hour, Format: "%d 分钟%s", DivBy: time.Minute},
	{D: 2 * time.Hour, Format: "1 小时%s", DivBy: 1},
	{D: humanize.Day, Format: "%d 小时%s", DivBy: time.Hour},
	{D: 2 * humanize.Day, Format: "1 天%s", DivBy: 1},
	{D: humanize.Week, Format: "%d 天%s", DivBy: humanize.Day},
	{D: 2 * humanize.Week, Format: "1 周%s", DivBy: 1},
	{D: humanize.Month, Format: "%d 周%s", DivBy: humanize.Week},
	{D: 2 * humanize.Month, Format: "1 个月%s", DivBy: 1},
	{D: humanize.Year, Format: "%d 个月%s", DivBy: humanize.Month},
	{D: 18 * humanize.Month, Format: "1 年%s", DivBy: 1},
	{D: 2 * humanize.Year, Format: "2 年%s", DivBy: 1},
	{D: humanize.LongTime, Format: "%d 年%s", DivBy: humanize.Year},
	{D: math.MaxInt64, Format: "很久%s", DivBy: 1},
}

func ToHumanizeDatetimeString(t *time.Time) string {
	return humanize.CustomRelTime(*t, time.Now(), "以前", "从现在起", magnitudes)
}

func ToRFC3339DatetimeString(t *time.Time) string {
	return t.Format(time.RFC3339)
}
