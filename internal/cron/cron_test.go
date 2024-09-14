package cron

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/util/timer"

	"github.com/duc-cnzj/mars/v5/internal/locker"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestManager_List(t *testing.T) {
	m := NewManager(nil, nil, nil, mlog.NewForConfig(nil))
	m.NewCommand("a", func() error {
		return nil
	})
	m.NewCommand("c", func() error {
		return nil
	})
	m.NewCommand("b", func() error {
		return nil
	})
	l := m.List()
	assert.Len(t, l, 3)
	assert.Equal(t, "a", l[0].Name())
	assert.Equal(t, "b", l[1].Name())
	assert.Equal(t, "c", l[2].Name())
}

func TestManager_NewCommand(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := NewMockRunner(m)
	cm := NewManager(timer.NewReal(), runner, nil, mlog.NewForConfig(nil))
	cmd := cm.NewCommand("duc", func() error {
		return nil
	})
	assert.Implements(t, (*Command)(nil), cmd)

	assert.Panics(t, func() {
		cm.NewCommand("duc", func() error {
			return nil
		})
	})
	called := false
	l := locker.NewMockLocker(m)
	l.EXPECT().ID().Return("1").AnyTimes()
	l.EXPECT().RenewalAcquire(lockKey("aaaa"), defaultLockSeconds, defaultRenewSeconds).Return(func() {}, true)
	NewManager(timer.NewReal(), nil, l, mlog.NewForConfig(nil)).NewCommand("aaaa", func() error {
		called = true
		return nil
	}).Func()()
	assert.True(t, called)
}

func TestManager_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := NewMockRunner(m)
	cm := NewManager(nil, runner, nil, mlog.NewForConfig(nil))
	cm.NewCommand("duc", func() error { return nil }).EveryTwoSeconds()
	ctx := context.TODO()
	runner.EXPECT().Run(ctx).Times(1)
	runner.EXPECT().AddCommand("duc", "*/2 * * * * *", gomock.Any()).Times(1)
	cm.Run(ctx)
}

func TestManager_Run_err(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := NewMockRunner(m)

	cm := NewManager(nil, runner, nil, mlog.NewForConfig(nil))
	runner.EXPECT().AddCommand("a", expression, gomock.Any()).Times(1).Return(errors.New("xxx"))
	cm.NewCommand("a", func() error {
		return nil
	})
	assert.Equal(t, "xxx", cm.Run(context.TODO()).Error())
}

func TestManager_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := NewMockRunner(m)
	ctx := context.TODO()
	runner.EXPECT().Shutdown(ctx).Times(1)
	cm := NewManager(nil, runner, nil, mlog.NewForConfig(nil))
	cm.Shutdown(ctx)
}

func TestNewManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := NewMockRunner(m)
	l := locker.NewMockLocker(m)
	manager := NewManager(nil, runner, l, mlog.NewForConfig(nil))
	assert.NotNil(t, manager.(*cronManager).commands)
	assert.Implements(t, (*Manager)(nil), manager)
	assert.NotNil(t, manager.(*cronManager).Locker)
	assert.NotNil(t, manager.(*cronManager).logger)
}

func Test_sortCommand(t *testing.T) {
	cmds := []Command{
		&command{
			name: "c",
		},
		&command{
			name: "a",
		},
		&command{
			name: "b",
		},
	}

	sort.Sort(sortCommand(cmds))
	assert.Equal(t, "a", cmds[0].Name())
	assert.Equal(t, "b", cmds[1].Name())
	assert.Equal(t, "c", cmds[2].Name())
}

func Test_cronManager_wrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := NewMockRunner(m)
	l := locker.NewMockLocker(m)
	called := false
	l.EXPECT().ID().Return("1").AnyTimes()
	l.EXPECT().RenewalAcquire(gomock.Any(), gomock.Any(), gomock.Any()).Return(func() {
		called = true
	}, true).Times(2)
	manager := NewManager(timer.NewReal(), runner, l, mlog.NewForConfig(nil)).(*cronManager)
	manager.wrap("a", func() error {
		return errors.New("x")
	})()
	assert.True(t, called)

	assert.NotPanics(t, func() {
		manager.wrap("a", func() error {
			panic("err")
		})()
	})
}
