package cron

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/robfig/cron/v3"
)

type Runner interface {
	AddCommand(name string, expression string, fn func()) error
	Run(context.Context) error
	Shutdown(context.Context) error
}

var _ Runner = (*robfigCronV3Runner)(nil)

type robfigCronV3Runner struct {
	sync.RWMutex
	logger   mlog.Logger
	c        *cron.Cron
	entryMap map[string]int64
}

// NewRobfigCronV3Runner return contracts.Runner
func NewRobfigCronV3Runner(logger mlog.Logger) Runner {
	return &robfigCronV3Runner{
		logger: logger.WithModule("cron/robfigCronV3Runner"),
		c: cron.New(
			cron.WithLocation(time.Local),
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(&cronLogger{
					logger: logger,
				}),
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
	c.logger.Infof("[CRON]: ADD '%s', spec: '%s', id: '%d'", name, expression, id)
	c.entryMap[name] = int64(id)
	return nil
}

// Run cron.
func (c *robfigCronV3Runner) Run(ctx context.Context) error {
	go func() {
		defer c.logger.HandlePanic("[CRON]: robfig/cron/v3 Run")
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

type cronLogger struct {
	logger mlog.Logger
}

// Info msg.
func (c *cronLogger) Info(msg string, keysAndValues ...any) {
	c.logger.Infof(formatString(len(keysAndValues)), append([]any{msg}, keysAndValues...)...)
}

// Error msg.
func (c *cronLogger) Error(err error, msg string, keysAndValues ...any) {
	c.logger.Errorf("[CRON]: %v", err)
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
