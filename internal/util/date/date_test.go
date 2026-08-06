package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToHumanizeDateTime(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "现在", ToHumanizeDateTime(&now))

	oneSecAgo := now.Add(-1 * time.Second)
	assert.Equal(t, "1 秒以前", ToHumanizeDateTime(&oneSecAgo))

	oneYearAgo := now.AddDate(-1, 0, 0)
	assert.Equal(t, "1 年以前", ToHumanizeDateTime(&oneYearAgo))

	nilTime := (*time.Time)(nil)
	assert.Equal(t, "", ToHumanizeDateTime(nilTime))
}

func TestToRFC3339(t *testing.T) {
	now := time.Now()
	assert.Equal(t, now.Format(time.RFC3339), ToRFC3339(&now))

	zeroTime := time.Time{}
	assert.Equal(t, "", ToRFC3339(&zeroTime))

	nilTime := (*time.Time)(nil)
	assert.Equal(t, "", ToRFC3339(nilTime))
}

func TestHumanDuration(t *testing.T) {
	assert.Equal(t, "0秒", HumanDuration(0))
	assert.Equal(t, "1秒", HumanDuration(time.Second))
	assert.Equal(t, "60秒", HumanDuration(1*time.Minute))
	assert.Equal(t, "2分钟", HumanDuration(2*time.Minute))
	assert.Equal(t, "11分钟", HumanDuration(11*time.Minute))
	assert.Equal(t, "60分钟", HumanDuration(time.Hour))
	assert.Equal(t, "3小时", HumanDuration(3*time.Hour))
	assert.Equal(t, "3天", HumanDuration(3*24*time.Hour))
	assert.Equal(t, "3年", HumanDuration(3*365*24*time.Hour))
	assert.Equal(t, "<invalid>", HumanDuration(-1*time.Second))
}
func TestHumanDurationWithYearsAndDays(t *testing.T) {
	twoYearsAndThreeDays := 2*365*24*time.Hour + 3*24*time.Hour
	assert.Equal(t, "2年3天", HumanDuration(twoYearsAndThreeDays))

	eightYears := 8 * 365 * 24 * time.Hour
	assert.Equal(t, "8年", HumanDuration(eightYears))
}

func TestHumanDurationWithSeconds(t *testing.T) {
	twoSecondsDuration := 2 * time.Second
	assert.Equal(t, "2秒", HumanDuration(twoSecondsDuration))
}

func TestHumanDurationWithMinutesAndSeconds(t *testing.T) {
	nineMinutesAndThirtySecondsDuration := 9*time.Minute + 30*time.Second
	assert.Equal(t, "9分钟30秒", HumanDuration(nineMinutesAndThirtySecondsDuration))
}

func TestHumanDurationWithHoursAndMinutes(t *testing.T) {
	sevenHoursAndFortyFiveMinutes := 7*time.Hour + 45*time.Minute
	assert.Equal(t, "7小时45分钟", HumanDuration(sevenHoursAndFortyFiveMinutes))
}

func TestHumanDurationWithDaysAndHours(t *testing.T) {
	sevenDaysAndSixHours := 7*24*time.Hour + 6*time.Hour
	assert.Equal(t, "7天6小时", HumanDuration(sevenDaysAndSixHours))
}
func TestHumanDurationWithFortyEightHours(t *testing.T) {
	fortyEightHours := 47 * time.Hour
	assert.Equal(t, "47小时", HumanDuration(fortyEightHours))
}

func TestHumanDurationWithTwoYears(t *testing.T) {
	twoYears := (2 * 365 * 24 * time.Hour) - 1*time.Hour
	assert.Equal(t, "729天", HumanDuration(twoYears))
}
