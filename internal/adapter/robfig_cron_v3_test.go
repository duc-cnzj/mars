package adapter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCronLogger_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	e := errors.New("xx")
	l.EXPECT().Errorf("[CRON]: %v", e)
	(&cronLogger{}).Error(e, "")
}

func TestCronLogger_Info(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Infof("[CRON]: %s, %v=%v", "aa", "a", "b")
	(&cronLogger{}).Info("aa", "a", "b")
}

func TestNewRobfigCronV3Runner(t *testing.T) {
	t.Parallel()
	runner := NewRobfigCronV3Runner().(*robfigCronV3Runner)
	assert.NotNil(t, runner.entryMap)
	assert.IsType(t, (*cron.Cron)(nil), runner.c)
}

func TestRobfigCronV3Runner_AddCommand(t *testing.T) {
	t.Parallel()
	runner := NewRobfigCronV3Runner().(*robfigCronV3Runner)
	assert.Error(t, runner.AddCommand("a", "", func() {}))
	assert.Nil(t, runner.AddCommand("a", "* * * * * *", func() {}))
	runner.Lock()
	defer runner.Unlock()
	assert.Equal(t, 1, len(runner.entryMap))
}

func TestRobfigCronV3Runner_Run(t *testing.T) {
	t.Parallel()
	runner := NewRobfigCronV3Runner().(*robfigCronV3Runner)
	assert.Nil(t, runner.Run(context.TODO()))
	<-runner.c.Stop().Done()
}

func TestRobfigCronV3Runner_Shutdown(t *testing.T) {
	t.Parallel()
	runner := NewRobfigCronV3Runner()
	err := runner.Shutdown(context.TODO())
	assert.Nil(t, err)
}

func TestRobfigCronV3Runner_Shutdown2(t *testing.T) {
	t.Parallel()
	runner := NewRobfigCronV3Runner()
	runner.AddCommand("test", "* * * * * *", func() {
		time.Sleep(100 * time.Second)
	})
	runner.Run(context.TODO())
	time.Sleep(1 * time.Second)
	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Error(t, runner.Shutdown(ctx))
}

func Test_formatString(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		num   int
		wants string
	}{
		{
			num:   4,
			wants: "[CRON]: %s, %v=%v, %v=%v",
		},
		{
			num:   0,
			wants: "[CRON]: %s",
		},
		{
			num:   2,
			wants: "[CRON]: %s, %v=%v",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.wants, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, formatString(tt.num))
		})
	}
}
