package adapter

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils/recovery"
)

type RobfigCronV3Runner struct {
	sync.RWMutex

	c        *cron.Cron
	entryMap map[string]int64
}

func NewRobfigCronV3Runner() *RobfigCronV3Runner {
	return &RobfigCronV3Runner{
		c: cron.New(
			cron.WithLocation(time.Local),
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(&CronLogger{}),
			),
		),
		entryMap: make(map[string]int64),
	}
}

func (c *RobfigCronV3Runner) AddCommand(name string, expression string, fn func()) error {
	c.Lock()
	defer c.Unlock()
	id, err := c.c.AddFunc(expression, fn)
	if err != nil {
		return err
	}
	mlog.Debugf("[CRON]: ADD '%s', spec: '%s', id: '%d'", name, expression, id)
	c.entryMap[name] = int64(id)
	return nil
}

func (c *RobfigCronV3Runner) Run(ctx context.Context) error {
	go func() {
		defer recovery.HandlePanic("[CRON]: robfig/cron/v3 Run")
		c.c.Run()
	}()
	return nil
}

func (c *RobfigCronV3Runner) Shutdown(ctx context.Context) error {
	stopCtx := c.c.Stop()
	select {
	case <-stopCtx.Done():
		return stopCtx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

type CronLogger struct{}

func (c *CronLogger) Info(msg string, keysAndValues ...any) {
	mlog.Infof(formatString(len(keysAndValues)), append([]any{msg}, keysAndValues...)...)
}

func (c *CronLogger) Error(err error, msg string, keysAndValues ...any) {
	mlog.Errorf("[CRON]: %v", err)
}

func formatString(numKeysAndValues int) string {
	var sb strings.Builder
	sb.WriteString("[Cron]: %s")
	if numKeysAndValues > 0 {
		sb.WriteString(", ")
	}
	for i := 0; i < numKeysAndValues/2; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("%v=%v")
	}
	return sb.String()
}
