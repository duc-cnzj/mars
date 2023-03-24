package adapter

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"

	"github.com/robfig/cron/v3"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
)

type robfigCronV3Runner struct {
	sync.RWMutex

	c        *cron.Cron
	entryMap map[string]int64
}

// NewRobfigCronV3Runner return contracts.CronRunner
func NewRobfigCronV3Runner() contracts.CronRunner {
	return &robfigCronV3Runner{
		c: cron.New(
			cron.WithLocation(time.Local),
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(&cronLogger{}),
			),
		),
		entryMap: make(map[string]int64),
	}
}

// AddCommand add cron cmd.
func (c *robfigCronV3Runner) AddCommand(name string, expression string, fn func()) error {
	c.Lock()
	defer c.Unlock()
	id, err := c.c.AddFunc(expression, fn)
	if err != nil {
		return err
	}
	mlog.Infof("[CRON]: ADD '%s', spec: '%s', id: '%d'", name, expression, id)
	c.entryMap[name] = int64(id)
	return nil
}

// Run cron.
func (c *robfigCronV3Runner) Run(ctx context.Context) error {
	go func() {
		defer recovery.HandlePanic("[CRON]: robfig/cron/v3 Run")
		c.c.Run()
	}()
	return nil
}

// Shutdown cron.
func (c *robfigCronV3Runner) Shutdown(ctx context.Context) error {
	stopCtx := c.c.Stop()
	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type cronLogger struct{}

// Info msg.
func (c *cronLogger) Info(msg string, keysAndValues ...any) {
	mlog.Infof(formatString(len(keysAndValues)), append([]any{msg}, keysAndValues...)...)
}

// Error msg.
func (c *cronLogger) Error(err error, msg string, keysAndValues ...any) {
	mlog.Errorf("[CRON]: %v", err)
}

func formatString(numKeysAndValues int) string {
	var sb strings.Builder
	sb.WriteString("[CRON]: %s")
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
