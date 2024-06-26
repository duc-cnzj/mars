package cron

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWrap_Recovery(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(false)
	l := mock.NewMockLocker(m)
	l.EXPECT().ID().Return("1").AnyTimes()
	l.EXPECT().RenewalAcquire(lockKey("duc"), defaultLockSeconds, defaultRenewSeconds).Times(1).Return(func() {}, true)
	wrap("duc", func() error {
		panic("err")
	}, func() contracts.Locker {
		return l
	})()
	assert.True(t, true)
}

func TestWrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLocker(m)
	l.EXPECT().ID().Return("1").AnyTimes()
	l.EXPECT().RenewalAcquire(lockKey("duc"), defaultLockSeconds, defaultRenewSeconds).Times(1).Return(nil, false)
	called := false
	wrap("duc", func() error {
		called = true
		return nil
	}, func() contracts.Locker {
		return l
	})()
	assert.False(t, called)
}

func TestWrap_FuncErrorReturn(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLocker(m)
	l.EXPECT().ID().Return("1").AnyTimes()
	called := 0
	l.EXPECT().RenewalAcquire(lockKey("duc"), defaultLockSeconds, defaultRenewSeconds).Times(1).Return(func() {
		called++
	}, true)
	wrap("duc", func() error {
		called++
		return errors.New("xxx")
	}, func() contracts.Locker {
		return l
	})()

	assert.Equal(t, 2, called)
}
