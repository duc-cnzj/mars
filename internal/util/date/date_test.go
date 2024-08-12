package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToHumanizeDatetimeString(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "现在", ToHumanizeDatetimeString(&now))

	oneSecAgo := now.Add(-1 * time.Second)
	assert.Equal(t, "1 秒以前", ToHumanizeDatetimeString(&oneSecAgo))

	oneYearAgo := now.AddDate(-1, 0, 0)
	assert.Contains(t, "1 年以前", ToHumanizeDatetimeString(&oneYearAgo))

	nilTime := (*time.Time)(nil)
	assert.Equal(t, "", ToHumanizeDatetimeString(nilTime))
}

func TestToRFC3339DatetimeString(t *testing.T) {
	now := time.Now()
	assert.Equal(t, now.Format(time.RFC3339), ToRFC3339DatetimeString(&now))

	zeroTime := time.Time{}
	assert.Equal(t, "", ToRFC3339DatetimeString(&zeroTime))

	nilTime := (*time.Time)(nil)
	assert.Equal(t, "", ToRFC3339DatetimeString(nilTime))
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
