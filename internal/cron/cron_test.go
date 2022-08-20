package cron

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestManager_List(t *testing.T) {
	m := NewManager(nil, nil)
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
	mk := gomock.NewController(t)
	defer mk.Finish()
	app := testutil.MockApp(mk)
	m := NewManager(nil, nil)
	cmd := m.NewCommand("duc", func() error {
		return nil
	})
	assert.Implements(t, (*contracts.Command)(nil), cmd)

	assert.Panics(t, func() {
		m.NewCommand("duc", func() error {
			return nil
		})
	})
	called := false
	l := mock.NewMockLocker(mk)
	app.EXPECT().CacheLock().Return(l).AnyTimes()
	l.EXPECT().ID().Return("1").AnyTimes()
	l.EXPECT().RenewalAcquire(lockKey("aaaa"), defaultLockSeconds, defaultRenewSeconds).Return(func() {}, true)
	NewManager(nil, app).NewCommand("aaaa", func() error {
		called = true
		return nil
	}).Func()()
	assert.True(t, called)
}

func TestManager_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	runner := mock.NewMockCronRunner(m)
	called := false
	Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {
		called = true
	})
	cm := NewManager(runner, app)
	cm.NewCommand("duc", func() error {
		return nil
	}).EveryTwoSeconds()
	ctx := context.TODO()
	runner.EXPECT().Run(ctx).Times(1)
	runner.EXPECT().AddCommand("duc", "*/2 * * * * *", gomock.Any()).Times(1)
	cm.Run(ctx)
	assert.True(t, called)
}

func TestManager_Run_err(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	runner := mock.NewMockCronRunner(m)

	cm := NewManager(runner, app)
	runner.EXPECT().AddCommand("a", expression, gomock.Any()).Times(1).Return(errors.New("xxx"))
	cm.NewCommand("a", func() error {
		return nil
	})
	assert.Equal(t, "xxx", cm.Run(context.TODO()).Error())
}

func TestManager_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := mock.NewMockCronRunner(m)
	ctx := context.TODO()
	runner.EXPECT().Shutdown(ctx).Times(1)
	cm := NewManager(runner, nil)
	cm.Shutdown(ctx)
}

func TestNewManager(t *testing.T) {
	manager := NewManager(nil, nil)
	assert.NotNil(t, manager.commands)
	assert.Implements(t, (*contracts.CronManager)(nil), manager)
}

func Test_sortCommand(t *testing.T) {
	cmds := []contracts.Command{
		&Command{
			name: "c",
		},
		&Command{
			name: "a",
		},
		&Command{
			name: "b",
		},
	}

	sort.Sort(sortCommand(cmds))
	assert.Equal(t, "a", cmds[0].Name())
	assert.Equal(t, "b", cmds[1].Name())
	assert.Equal(t, "c", cmds[2].Name())
}
