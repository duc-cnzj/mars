package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHumanDuration(t *testing.T) {
}

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
}

func TestToRFC3339DatetimeString(t *testing.T) {
	_, err := time.Parse(time.RFC3339, ToRFC3339DatetimeString(&time.Time{}))
	assert.Nil(t, err)
}
