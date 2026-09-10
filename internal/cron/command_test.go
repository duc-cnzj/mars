package cron

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newCommand() *command {
	return &command{expression: expression}
}

func TestCommand_At(t *testing.T) {
	var tests = []struct {
		time  string
		wants string
	}{
		{
			time:  "02:00",
			wants: "0 00 02 * * *",
		},
		{
			time:  "01",
			wants: "0 0 01 * * *",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.time, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, newCommand().At(tt.time).Expression())
		})
	}
}

func TestCommand_Cron(t *testing.T) {
	var tests = []struct {
		expr string
	}{
		{
			expr: "* * * * * *",
		},
		{
			expr: "0 0 0 0 0 0",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expr, newCommand().Cron(tt.expr).Expression())
		})
	}
}

func TestCommand_Daily(t *testing.T) {
	assert.Equal(t, "0 0 0 * * *", newCommand().Daily().Expression())
}

func TestCommand_DailyAt(t *testing.T) {
	var tests = []struct {
		time  string
		wants string
	}{
		{
			time:  "02:58",
			wants: "0 58 02 * * *",
		},
		{
			time:  "01",
			wants: "0 0 01 * * *",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.time, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, newCommand().DailyAt(tt.time).Expression())
		})
	}
}

func TestCommand_Days(t *testing.T) {
	var tests = []struct {
		days  []int
		wants string
	}{
		{
			days:  []int{MONDAY, SUNDAY},
			wants: "0 * * * * 1,0",
		},
		{
			days:  []int{FRIDAY},
			wants: "0 * * * * 5",
		},
	}
	for i, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, newCommand().Days(tt.days).Expression())
		})
	}
}

func TestCommand_EveryFifteenMinutes(t *testing.T) {
	assert.Equal(t, "0 */15 * * * *", newCommand().EveryFifteenMinutes().Expression())
}

func TestCommand_EveryFifteenSeconds(t *testing.T) {
	assert.Equal(t, "*/15 * * * * *", newCommand().EveryFifteenSeconds().Expression())

}

func TestCommand_EveryFiveMinutes(t *testing.T) {
	assert.Equal(t, "0 */5 * * * *", newCommand().EveryFiveMinutes().Expression())
}

func TestCommand_EveryFiveSeconds(t *testing.T) {
	assert.Equal(t, "*/5 * * * * *", newCommand().EveryFiveSeconds().Expression())

}

func TestCommand_EveryFourHours(t *testing.T) {
	assert.Equal(t, "0 0 */4 * * *", newCommand().EveryFourHours().Expression())

}

func TestCommand_EveryFourMinutes(t *testing.T) {
	assert.Equal(t, "0 */4 * * * *", newCommand().EveryFourMinutes().Expression())
}

func TestCommand_EveryFourSeconds(t *testing.T) {
	assert.Equal(t, "*/4 * * * * *", newCommand().EveryFourSeconds().Expression())
}

func TestCommand_EveryMinute(t *testing.T) {
	assert.Equal(t, "0 * * * * *", newCommand().EveryMinute().Expression())
}

func TestCommand_EverySecond(t *testing.T) {
	assert.Equal(t, "* * * * * *", newCommand().EverySecond().Expression())
}

func TestCommand_EverySixHours(t *testing.T) {
	assert.Equal(t, "0 0 */6 * * *", newCommand().EverySixHours().Expression())
}

func TestCommand_EveryTenMinutes(t *testing.T) {
	assert.Equal(t, "0 */10 * * * *", newCommand().EveryTenMinutes().Expression())
}

func TestCommand_EveryTenSeconds(t *testing.T) {
	assert.Equal(t, "*/10 * * * * *", newCommand().EveryTenSeconds().Expression())
}

func TestCommand_EveryThirtyMinutes(t *testing.T) {
	assert.Equal(t, "0 0,30 * * * *", newCommand().EveryThirtyMinutes().Expression())
}

func TestCommand_EveryThirtySeconds(t *testing.T) {
	assert.Equal(t, "0,30 * * * * *", newCommand().EveryThirtySeconds().Expression())
}

func TestCommand_EveryThreeHours(t *testing.T) {
	assert.Equal(t, "0 0 */3 * * *", newCommand().EveryThreeHours().Expression())
}

func TestCommand_EveryThreeMinutes(t *testing.T) {
	assert.Equal(t, "0 */3 * * * *", newCommand().EveryThreeMinutes().Expression())
}

func TestCommand_EveryThreeSeconds(t *testing.T) {
	assert.Equal(t, "*/3 * * * * *", newCommand().EveryThreeSeconds().Expression())
}

func TestCommand_EveryTwoHours(t *testing.T) {
	assert.Equal(t, "0 0 */2 * * *", newCommand().EveryTwoHours().Expression())
}

func TestCommand_EveryTwoMinutes(t *testing.T) {
	assert.Equal(t, "0 */2 * * * *", newCommand().EveryTwoMinutes().Expression())
}

func TestCommand_EveryTwoSeconds(t *testing.T) {
	assert.Equal(t, "*/2 * * * * *", newCommand().EveryTwoSeconds().Expression())
}

func TestCommand_Expression(t *testing.T) {
	cmd := newCommand()
	assert.Equal(t, cmd.expression, cmd.Expression())
}

func TestCommand_Fridays(t *testing.T) {
	assert.Equal(t, "0 * * * * 5", newCommand().Fridays().Expression())
}

func TestCommand_Func(t *testing.T) {
	assert.Nil(t, newCommand().Func())
	cmd := newCommand()
	i := 0
	fn := func() { i++ }
	cmd.fn = fn
	cmd.Func()()
	assert.Equal(t, 1, i)
}

func TestCommand_Hourly(t *testing.T) {
	assert.Equal(t, "0 0 * * * *", newCommand().Hourly().Expression())
}

func TestCommand_HourlyAt(t *testing.T) {
	var tests = []struct {
		time  []int
		wants string
	}{
		{
			time:  []int{1, 2, 3},
			wants: "0 1,2,3 * * * *",
		},
		{
			time:  []int{10, 20, 50},
			wants: "0 10,20,50 * * * *",
		},
		{
			time:  []int{},
			wants: "0 0 * * * *",
		},
	}
	for i, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, newCommand().HourlyAt(tt.time).Expression())
		})
	}
}

func TestCommand_LastDayOfMonth(t *testing.T) {
	assert.Equal(t, "0 00 15 L * *", newCommand().LastDayOfMonth("15:00").Expression())
}

func TestCommand_Mondays(t *testing.T) {
	assert.Equal(t, "0 * * * * 1", newCommand().Mondays().Expression())
}

func TestCommand_Monthly(t *testing.T) {
	assert.Equal(t, "0 0 0 1 * *", newCommand().Monthly().Expression())
}

func TestCommand_MonthlyOn(t *testing.T) {
	var tests = []struct {
		dom   string
		time  string
		wants string
	}{
		{
			dom:   "3",
			time:  "2:00",
			wants: "0 00 2 3 * *",
		},
		{
			dom:   "",
			time:  "",
			wants: "0 0 0 1 * *",
		},
	}
	for i, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, newCommand().MonthlyOn(tt.dom, tt.time).Expression())
		})
	}
}

func TestCommand_Name(t *testing.T) {
	cmd := &command{name: "duc"}
	assert.Equal(t, "duc", cmd.Name())
}

func TestCommand_Quarterly(t *testing.T) {
	assert.Equal(t, "0 0 0 1 1-12/3 *", newCommand().Quarterly().Expression())
}

func TestCommand_QuarterlyOn(t *testing.T) {
	var tests = []struct {
		doq   string
		time  string
		wants string
	}{
		{
			doq:   "3",
			time:  "2:00",
			wants: "0 00 2 3 1-12/3 *",
		},
		{
			doq:   "",
			time:  "",
			wants: "0 0 0 1 1-12/3 *",
		},
	}
	for i, test := range tests {
		tt := test
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, newCommand().QuarterlyOn(tt.doq, tt.time).Expression())
		})
	}
}

func TestCommand_Saturdays(t *testing.T) {
	assert.Equal(t, "0 * * * * 6", newCommand().Saturdays().Expression())
}

func TestCommand_Sundays(t *testing.T) {
	assert.Equal(t, "0 * * * * 0", newCommand().Sundays().Expression())

}

func TestCommand_Thursdays(t *testing.T) {
	assert.Equal(t, "0 * * * * 4", newCommand().Thursdays().Expression())
}

func TestCommand_Tuesdays(t *testing.T) {
	assert.Equal(t, "0 * * * * 2", newCommand().Tuesdays().Expression())
}

func TestCommand_Wednesdays(t *testing.T) {
	assert.Equal(t, "0 * * * * 3", newCommand().Wednesdays().Expression())
}

func TestCommand_Weekdays(t *testing.T) {
	assert.Equal(t, "0 * * * * 1,2,3,4,5", newCommand().Weekdays().Expression())
}

func TestCommand_Weekends(t *testing.T) {
	assert.Equal(t, "0 * * * * 6,0", newCommand().Weekends().Expression())
}

func TestCommand_Weekly(t *testing.T) {
	assert.Equal(t, "0 0 0 * * 1", newCommand().Weekly().Expression())
}

func TestCommand_WeeklyOn(t *testing.T) {
	assert.Equal(t, "0 00 19 * * 0", newCommand().WeeklyOn(SUNDAY, "19:00").Expression())
	assert.Equal(t, "0 0 0 * * 0", newCommand().WeeklyOn(SUNDAY, "").Expression())
}

func TestCommand_Yearly(t *testing.T) {
	assert.Equal(t, "0 0 0 1 1 *", newCommand().Yearly().Expression())
}

func TestCommand_YearlyOn(t *testing.T) {
	assert.Equal(t, "0 33 3 4 3 *", newCommand().YearlyOn("3", "4", "3:33").Expression())
}

func TestCommand_spliceIntoPosition(t *testing.T) {
	cmd := newCommand()
	cmd.spliceIntoPosition(POS_SECOND, "1")
	cmd.spliceIntoPosition(POS_MINUTE, "2")
	cmd.spliceIntoPosition(POS_HOUR, "3")
	cmd.spliceIntoPosition(POS_DAY_OF_MONTH, "4")
	cmd.spliceIntoPosition(POS_MONTH, "5")
	cmd.spliceIntoPosition(POS_DAY_OF_WEEK, "6")
	assert.Equal(t, "1 2 3 4 5 6", cmd.Expression())
}

func TestMixture(t *testing.T) {
	assert.Equal(t, "0 0 */6 * * 3", newCommand().EverySixHours().Wednesdays().Expression())
}
