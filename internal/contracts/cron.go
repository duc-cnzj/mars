package contracts

//go:generate mockgen -destination ../mock/mock_cron_runner.go -package mock github.com/duc-cnzj/mars/internal/contracts CronRunner

import "context"

type CronRunner interface {
	AddCommand(name string, expression string, fn func()) error
	Run(context.Context) error
	Shutdown(context.Context) error
}

type CronManager interface {
	NewCommand(name string, fn func() error) Command
	Run(context.Context) error
	Shutdown(context.Context) error

	List() []Command
}

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
