package cron

import (
	"strconv"
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
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

func (c *command) Cron(expression string) contracts.Command {
	c.expression = expression
	return c
}

func (c *command) EverySecond() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "*")
	return c
}

func (c *command) EveryTwoSeconds() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "*/2")
	return c
}

func (c *command) EveryThreeSeconds() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "*/3")
	return c
}

func (c *command) EveryFourSeconds() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "*/4")
	return c
}

func (c *command) EveryFiveSeconds() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "*/5")
	return c
}

func (c *command) EveryTenSeconds() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "*/10")
	return c
}

func (c *command) EveryFifteenSeconds() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "*/15")
	return c
}

func (c *command) EveryThirtySeconds() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0,30")
	return c
}

func (c *command) EveryMinute() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*")
	return c
}

func (c *command) EveryTwoMinutes() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/2")
	return c
}

func (c *command) EveryThreeMinutes() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/3")
	return c
}

func (c *command) EveryFourMinutes() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/4")
	return c
}

func (c *command) EveryFiveMinutes() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/5")
	return c
}

func (c *command) EveryTenMinutes() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/10")
	return c
}

func (c *command) EveryFifteenMinutes() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "*/15")
	return c
}

func (c *command) EveryThirtyMinutes() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0,30")
	return c
}

func (c *command) Hourly() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	return c
}

func (c *command) HourlyAt(minutes []int) contracts.Command {
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

func (c *command) EveryTwoHours() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/2")
	return c
}

func (c *command) EveryThreeHours() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/3")
	return c
}

func (c *command) EveryFourHours() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/4")
	return c
}

func (c *command) EverySixHours() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "*/6")
	return c
}

func (c *command) Daily() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	return c
}

func (c *command) At(time string) contracts.Command {
	return c.DailyAt(time)
}

func (c *command) DailyAt(time string) contracts.Command {
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

func (c *command) Weekdays() contracts.Command {
	return c.Days([]int{MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY})
}

func (c *command) Weekends() contracts.Command {
	c.Days([]int{SATURDAY, SUNDAY})
	return c
}

func (c *command) Mondays() contracts.Command {
	c.Days([]int{MONDAY})
	return c
}

func (c *command) Tuesdays() contracts.Command {
	c.Days([]int{TUESDAY})
	return c
}

func (c *command) Wednesdays() contracts.Command {
	c.Days([]int{WEDNESDAY})
	return c
}

func (c *command) Thursdays() contracts.Command {
	c.Days([]int{THURSDAY})
	return c
}

func (c *command) Fridays() contracts.Command {
	c.Days([]int{FRIDAY})
	return c
}

func (c *command) Saturdays() contracts.Command {
	c.Days([]int{SATURDAY})
	return c
}

func (c *command) Sundays() contracts.Command {
	c.Days([]int{SUNDAY})
	return c
}

func (c *command) Weekly() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_WEEK, "1")
	return c
}

func (c *command) WeeklyOn(day int, time string) contracts.Command {
	if time == "" {
		time = "0:0"
	}
	c.DailyAt(time)
	c.Days([]int{day})
	return c
}

func (c *command) Monthly() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "1")
	return c
}

func (c *command) MonthlyOn(dayOfMonth string, time string) contracts.Command {
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

func (c *command) LastDayOfMonth(time string) contracts.Command {
	c.DailyAt(time)
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "L")
	return c
}

func (c *command) Quarterly() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "1")
	c.spliceIntoPosition(POS_MONTH, "1-12/3")
	return c
}

func (c *command) QuarterlyOn(dayOfQuarter string, time string) contracts.Command {
	if dayOfQuarter == "" {
		dayOfQuarter = "1"
	}
	c.DailyAt(time)
	c.spliceIntoPosition(POS_DAY_OF_MONTH, dayOfQuarter)
	c.spliceIntoPosition(POS_MONTH, "1-12/3")
	return c
}

func (c *command) Yearly() contracts.Command {
	c.spliceIntoPosition(POS_SECOND, "0")
	c.spliceIntoPosition(POS_MINUTE, "0")
	c.spliceIntoPosition(POS_HOUR, "0")
	c.spliceIntoPosition(POS_DAY_OF_MONTH, "1")
	c.spliceIntoPosition(POS_MONTH, "1")
	return c
}

func (c *command) YearlyOn(month string, dayOfMonth string, time string) contracts.Command {
	c.DailyAt(time)
	c.spliceIntoPosition(POS_DAY_OF_MONTH, dayOfMonth)
	c.spliceIntoPosition(POS_MONTH, month)
	return c
}

func (c *command) Days(days []int) contracts.Command {
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
