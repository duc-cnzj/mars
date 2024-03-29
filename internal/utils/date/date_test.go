package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToHumanizeDatetimeString(t *testing.T) {
	nowFunc = func() time.Time {
		return time.Time{}
	}
	defer func() {
		nowFunc = time.Now
	}()

	tt := nowFunc()
	assert.Equal(t, "现在", ToHumanizeDatetimeString(&tt))
	t1 := nowFunc().Add(-1 * time.Second)
	assert.Equal(t, "1 秒以前", ToHumanizeDatetimeString(&t1))
	t2 := nowFunc().Add(-1 * time.Minute)
	assert.Equal(t, "1 分钟以前", ToHumanizeDatetimeString(&t2))
	t3 := nowFunc().Add(1 * time.Minute)
	assert.Equal(t, "1 分钟以后", ToHumanizeDatetimeString(&t3))
	t4 := nowFunc().Add(1 * time.Hour)
	assert.Equal(t, "1 小时以后", ToHumanizeDatetimeString(&t4))
	t5 := nowFunc().Add(-1 * time.Hour)
	assert.Equal(t, "1 小时以前", ToHumanizeDatetimeString(&t5))
	t6 := nowFunc().Add(-1 * time.Hour * 24 * 365)
	assert.Equal(t, "1 年以前", ToHumanizeDatetimeString(&t6))
	t7 := nowFunc().Add(1 * time.Hour * 24 * 365)
	assert.Equal(t, "1 年以后", ToHumanizeDatetimeString(&t7))
	t8 := nowFunc().Add(2 * time.Hour * 24 * 365)
	assert.Equal(t, "2 年以后", ToHumanizeDatetimeString(&t8))
	t9 := nowFunc().Add(-2 * time.Hour * 24 * 365)
	assert.Equal(t, "2 年以前", ToHumanizeDatetimeString(&t9))
	t10 := nowFunc().Add(-200 * time.Hour * 24 * 365)
	assert.Equal(t, "很久以前", ToHumanizeDatetimeString(&t10))
	t11 := nowFunc().Add(-2 * time.Hour * 24 * 30)
	assert.Equal(t, "2 个月以前", ToHumanizeDatetimeString(&t11))
	t12 := nowFunc().Add(2 * time.Hour * 24 * 30)
	assert.Equal(t, "2 个月以后", ToHumanizeDatetimeString(&t12))

	assert.Empty(t, ToHumanizeDatetimeString(nil))
}

func TestToRFC3339DatetimeString(t *testing.T) {
	_, err := time.Parse(time.RFC3339, ToRFC3339DatetimeString(&time.Time{}))
	assert.Nil(t, err)
	datetimeString := ToRFC3339DatetimeString(nil)
	assert.Equal(t, (time.Time{}).Format(time.RFC3339), datetimeString)
}

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{d: -1 * time.Second, want: "0秒"},
		{d: -100 * time.Second, want: "<invalid>"},
		{d: time.Second, want: "1秒"},
		{d: 70 * time.Second, want: "70秒"},
		{d: 190 * time.Second, want: "3分钟10秒"},
		{d: 9 * time.Minute, want: "9分钟"},
		{d: 70 * time.Minute, want: "70分钟"},
		{d: 7 * time.Hour, want: "7小时"},
		{d: 7*time.Hour + 10*time.Minute, want: "7小时10分钟"},
		{d: 47 * time.Hour, want: "47小时"},
		{d: 7 * 24 * time.Hour, want: "7天"},
		{d: 49 * time.Hour, want: "2天1小时"},
		{d: (8*24 + 2) * time.Hour, want: "8天"},
		{d: (367 * 24) * time.Hour, want: "367天"},
		{d: (365*2*24 + 25) * time.Hour, want: "2年1天"},
		{d: (365*7*24 + 2) * time.Hour, want: "7年"},
		{d: (365*8*24 + 2) * time.Hour, want: "8年"},
	}
	for _, tt := range tests {
		t.Run(tt.d.String(), func(t *testing.T) {
			if got := HumanDuration(tt.d); got != tt.want {
				t.Errorf("HumanDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
