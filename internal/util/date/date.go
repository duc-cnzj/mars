package date

import (
	"fmt"
	"math"
	"time"

	"github.com/dustin/go-humanize"
)

// magnitudes 是 humanize 相对时间的中文展示刻度表。
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

// ToHumanizeDateTime 返回相对当前时间的可读中文描述（如"3 分钟前"）；t 为 nil 时返回空串。
func ToHumanizeDateTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return humanize.CustomRelTime(*t, time.Now(), "以前", "以后", magnitudes)
}

// ToRFC3339 将时间转换为 RFC3339 字符串（如 "2006-01-02T15:04:05Z07:00"）；nil 或零值返回空串。
func ToRFC3339(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// HumanDuration 将时长格式化为中文"x天x小时"描述，容忍 2 秒内的时钟偏差视为"0秒"。
func HumanDuration(d time.Duration) string {
	// 容忍 2 秒（不含）内的时钟偏差，偏差在这个量级可视为"几乎现在"。
	switch seconds := int(d.Seconds()); {
	case seconds <= -1:
		return "<invalid>"
	case seconds <= 0:
		return "0秒"
	case seconds < 60*2:
		return fmt.Sprintf("%d秒", seconds)
	}

	minutes := int(d / time.Minute)
	if minutes < 10 {
		s := int(d/time.Second) % 60
		if s == 0 {
			return fmt.Sprintf("%d分钟", minutes)
		}
		return fmt.Sprintf("%d分钟%d秒", minutes, s)
	}
	if minutes < 60*3 {
		return fmt.Sprintf("%d分钟", minutes)
	}

	hours := int(d / time.Hour)
	if hours < 8 {
		m := int(d/time.Minute) % 60
		if m == 0 {
			return fmt.Sprintf("%d小时", hours)
		}
		return fmt.Sprintf("%d小时%d分钟", hours, m)
	}
	if hours < 48 {
		return fmt.Sprintf("%d小时", hours)
	}
	if hours < 24*8 {
		h := hours % 24
		if h == 0 {
			return fmt.Sprintf("%d天", hours/24)
		}
		return fmt.Sprintf("%d天%d小时", hours/24, h)
	}
	if hours < 24*365*2 {
		return fmt.Sprintf("%d天", hours/24)
	}
	if hours < 24*365*8 {
		dy := int(hours/24) % 365
		if dy == 0 {
			return fmt.Sprintf("%d年", hours/24/365)
		}
		return fmt.Sprintf("%d年%d天", hours/24/365, dy)
	}
	return fmt.Sprintf("%d年", int(hours/24/365))
}
