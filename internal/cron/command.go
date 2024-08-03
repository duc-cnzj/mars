package cron

import (
	"strconv"
	"strings"
)

// second minute hour `day of the month` month `day of the week`
const expression = "* * * * * *"

const (
	POS_SECOND = iota
	POS_MINUTE
	POS_HOUR
	POS_DAY_OF_MONTH
	POS_MONTH
	POS_DAY_OF_WEEK
)

const (
	SUNDAY = iota
	MONDAY
	TUESDAY
	WEDNESDAY
	THURSDAY
	FRIDAY
	SATURDAY
)

type Command interface {
	//Name 名称
	Name() string
	//Func 方法
	Func() func()
	//Expression cron 表达式: "* * * * * *"
	Expression() string
	//Cron 自定义 expression "* * * * * *"
	Cron(expression string) Command
	//EverySecond 每 1 秒
	EverySecond() Command
	//EveryTwoSeconds 每 2 秒
	EveryTwoSeconds() Command
	//EveryThreeSeconds 每 3 秒
	EveryThreeSeconds() Command
	//EveryFourSeconds 每 4 秒
	EveryFourSeconds() Command
	//EveryFiveSeconds 每 5 秒
	EveryFiveSeconds() Command
	//EveryTenSeconds 每 10 秒
	EveryTenSeconds() Command
	//EveryFifteenSeconds 每 15 秒
	EveryFifteenSeconds() Command
	//EveryThirtySeconds 每 30 秒
	EveryThirtySeconds() Command
	//EveryMinute 每分钟
	EveryMinute() Command
	//EveryTwoMinutes 每 2 分钟
	EveryTwoMinutes() Command
	//EveryThreeMinutes 每 3 分钟
	EveryThreeMinutes() Command
	//EveryFourMinutes 每 4 分钟
	EveryFourMinutes() Command
	//EveryFiveMinutes 每 5 分钟
	EveryFiveMinutes() Command
	//EveryTenMinutes 每 10 分钟
	EveryTenMinutes() Command
	//EveryFifteenMinutes 每 15 分钟
	EveryFifteenMinutes() Command
	//EveryThirtyMinutes 每 30 分钟
	EveryThirtyMinutes() Command
	//Hourly 每小时
	Hourly() Command
	//HourlyAt 每小时的第几分钟
	HourlyAt([]int) Command
	//EveryTwoHours 每 2 小时
	EveryTwoHours() Command
	//EveryThreeHours 每 3 小时
	EveryThreeHours() Command
	//EveryFourHours 每 4 小时
	EveryFourHours() Command
	//EverySixHours 每 6 小时
	EverySixHours() Command
	//Daily 每天
	Daily() Command
	//DailyAt 每天几点(time: "2:00")
	DailyAt(time string) Command
	//At alias of DailyAt
	At(string) Command
	//Weekdays 工作日 1-5
	Weekdays() Command
	//Weekends 周末
	Weekends() Command
	//Mondays 周一
	Mondays() Command
	//Tuesdays 周二
	Tuesdays() Command
	//Wednesdays 周三
	Wednesdays() Command
	//Thursdays 周四
	Thursdays() Command
	//Fridays 周五
	Fridays() Command
	//Saturdays 周六
	Saturdays() Command
	//Sundays 周日
	Sundays() Command
	//Weekly 每周一
	Weekly() Command
	//WeeklyOn 周日几(day) 几点(time: "0:0")
	WeeklyOn(day int, time string) Command
	Monthly() Command
	// MonthlyOn dayOfMonth: 1, time: "0:0"
	MonthlyOn(dayOfMonth string, time string) Command
	//LastDayOfMonth 每月最后一天
	LastDayOfMonth(time string) Command
	// Quarterly 每季度执行
	Quarterly() Command
	// QuarterlyOn 每季度的第几天，几点(time: "0:0")执行
	QuarterlyOn(dayOfQuarter string, time string) Command
	//Yearly 每年
	Yearly() Command
	//YearlyOn 每年几月(month) 哪天(dayOfMonth) 时间(time: "0:0")
	YearlyOn(month string, dayOfMonth string, time string) Command
	//Days 天(0-6: 周日-周六)
	Days([]int) Command
}

var _ Command = (*command)(nil)

type command struct {
	name       string
	expression string

	fn func()
}

func (c *command) Func() func() {
	return c.fn
}

func (c *command) Expression() string {
	return c.expression
}

func (c *command) Name() string {
	return c.name
}

func (c *command) Cron(expression string) Command {
	c.expression = expression
	return c
}

func (c *command) EverySecond() Command {
	c.spliceIntoPosition(POS_SECOND, "*")
	return c
}

func (c *command) EveryTwoSeconds() Command {
	c.spliceIntoPosition(POS_SECOND, "*/2")
	return c
}

func (c *command) EveryThreeSeconds() Command {
	c.spliceIntoPosition(POS_SECOND, "*/3")
	return c
}

func (c *command) EveryFourSeconds() Command {
	c.spliceIntoPosition(POS_SECOND, "*/4")
	return c
}

func (c *command) EveryFiveSeconds() Command {
	c.spliceIntoPosition(POS_SECOND, "*/5")
	return c
}

func (c *command) EveryTenSeconds() Command {
	c.spliceIntoPosition(POS_SECOND, "*/10")
	return c
}

func (c *command) EveryFifteenSeconds() Command {
	c.spliceIntoPosition(POS_SECOND, "*/15")
	return c
}

func (c *command) EveryThirtySeconds() Command {
	c.spliceIntoPosition(POS_SECOND, "0,30")
	return c
}

func (c *command) EveryMinute() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*")
	return c
}

func (c *command) EveryTwoMinutes() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/2")
	return c
}

func (c *command) EveryThreeMinutes() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/3")
	return c
}

func (c *command) EveryFourMinutes() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/4")
	return c
}

func (c *command) EveryFiveMinutes() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/5")
	return c
}

func (c *command) EveryTenMinutes() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/10")
	return c
}

func (c *command) EveryFifteenMinutes() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/15")
	return c
}

func (c *command) EveryThirtyMinutes() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0,30")
	return c
}

func (c *command) Hourly() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	return c
}

func (c *command) HourlyAt(minutes []int) Command {
	var minsStr []string
	for _, day := range minutes {
		minsStr = append(minsStr, strconv.Itoa(day))
	}
	if len(minutes) == 0 {
		minsStr = []string{"0"}
	}
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, strings.Join(minsStr, ","))
	return c
}

func (c *command) EveryTwoHours() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/2")
	return c
}

func (c *command) EveryThreeHours() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/3")
	return c
}

func (c *command) EveryFourHours() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/4")
	return c
}

func (c *command) EverySixHours() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/6")
	return c
}

func (c *command) Daily() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	return c
}

func (c *command) At(time string) Command {
	return c.DailyAt(time)
}

func (c *command) DailyAt(time string) Command {
	hour, minute := "0", "0"
	if time != "" {
		split := strings.Split(time, ":")
		if len(split) == 2 {
			minute = split[1]
		}
		hour = split[0]
	}
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_HOUR, hour)
	c.spliceIntoPosition(POS_MINUTE, minute)
	return c
}

func (c *command) Weekdays() Command {
	return c.Days([]int{MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY})
}

func (c *command) Weekends() Command {
	c.Days([]int{SATURDAY, SUNDAY})
	return c
}

func (c *command) Mondays() Command {
	c.Days([]int{MONDAY})
	return c
}

func (c *command) Tuesdays() Command {
	c.Days([]int{TUESDAY})
	return c
}

func (c *command) Wednesdays() Command {
	c.Days([]int{WEDNESDAY})
	return c
}

func (c *command) Thursdays() Command {
	c.Days([]int{THURSDAY})
	return c
}

func (c *command) Fridays() Command {
	c.Days([]int{FRIDAY})
	return c
}

func (c *command) Saturdays() Command {
	c.Days([]int{SATURDAY})
	return c
}

func (c *command) Sundays() Command {
	c.Days([]int{SUNDAY})
	return c
}

func (c *command) Weekly() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_WEEK, "1")
	return c
}

func (c *command) WeeklyOn(day int, time string) Command {
	if time == "" {
		time = "0:0"
	}
	c.DailyAt(time)
	c.Days([]int{day})
	return c
}

func (c *command) Monthly() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "1")
	return c
}

func (c *command) MonthlyOn(dayOfMonth string, time string) Command {
	if dayOfMonth == "" {
		dayOfMonth = "1"
	}
	if time == "" {
		time = "0:0"
	}
	c.DailyAt(time)
	c.spliceIntoPosition(POS_DAY_OF_MONTH, dayOfMonth)
	return c
}

func (c *command) LastDayOfMonth(time string) Command {
	c.DailyAt(time)
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "L")
	return c
}

func (c *command) Quarterly() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "1")
	c.spliceIntoPosition(POS_MONTH, "1-12/3")
	return c
}

func (c *command) QuarterlyOn(dayOfQuarter string, time string) Command {
	if dayOfQuarter == "" {
		dayOfQuarter = "1"
	}
	c.DailyAt(time)
	c.spliceIntoPosition(POS_DAY_OF_MONTH, dayOfQuarter)
	c.spliceIntoPosition(POS_MONTH, "1-12/3")
	return c
}

func (c *command) Yearly() Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "1")
	c.spliceIntoPosition(POS_MONTH, "1")
	return c
}

func (c *command) YearlyOn(month string, dayOfMonth string, time string) Command {
	c.DailyAt(time)
	c.spliceIntoPosition(POS_DAY_OF_MONTH, dayOfMonth)
	c.spliceIntoPosition(POS_MONTH, month)
	return c
}

func (c *command) Days(days []int) Command {
	var daysStr []string
	for _, day := range days {
		daysStr = append(daysStr, strconv.Itoa(day))
	}
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_DAY_OF_WEEK, strings.Join(daysStr, ","))
	return c
}

func (c *command) spliceIntoPosition(pos int, val string) {
	split := strings.Split(c.expression, " ")
	split[pos] = val
	c.expression = strings.Join(split, " ")
}
