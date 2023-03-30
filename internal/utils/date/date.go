package date

import (
	"fmt"
	"math"
	"time"

	"github.com/dustin/go-humanize"
)

var nowFunc = time.Now

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
	if t == nil {
		return ""
	}
	return humanize.CustomRelTime(*t, nowFunc(), "以前", "以后", magnitudes)
}

// ToRFC3339DatetimeString "2006-01-02T15:04:05Z07:00"
func ToRFC3339DatetimeString(t *time.Time) string {
	if t == nil {
		t = &time.Time{}
	}
	return t.Format(time.RFC3339)
}

func HumanDuration(d time.Duration) string {
	// Allow deviation no more than 2 seconds(excluded) to tolerate machine time
	// inconsistence, it can be considered as almost now.
	if seconds := int(d.Seconds()); seconds < -1 {
		return "<invalid>"
	} else if seconds < 0 {
		return "0秒"
	} else if seconds < 60*2 {
		return fmt.Sprintf("%d秒", seconds)
	}
	minutes := int(d / time.Minute)
	if minutes < 10 {
		s := int(d/time.Second) % 60
		if s == 0 {
			return fmt.Sprintf("%d分钟", minutes)
		}
		return fmt.Sprintf("%d分钟%d秒", minutes, s)
	} else if minutes < 60*3 {
		return fmt.Sprintf("%d分钟", minutes)
	}
	hours := int(d / time.Hour)
	if hours < 8 {
		m := int(d/time.Minute) % 60
		if m == 0 {
			return fmt.Sprintf("%d小时", hours)
		}
		return fmt.Sprintf("%d小时%d分钟", hours, m)
	} else if hours < 48 {
		return fmt.Sprintf("%d小时", hours)
	} else if hours < 24*8 {
		h := hours % 24
		if h == 0 {
			return fmt.Sprintf("%d天", hours/24)
		}
		return fmt.Sprintf("%d天%d小时", hours/24, h)
	} else if hours < 24*365*2 {
		return fmt.Sprintf("%d天", hours/24)
	} else if hours < 24*365*8 {
		dy := int(hours/24) % 365
		if dy == 0 {
			return fmt.Sprintf("%d年", hours/24/365)
		}
		return fmt.Sprintf("%d年%d天", hours/24/365, dy)
	}
	return fmt.Sprintf("%d年", int(hours/24/365))
}
