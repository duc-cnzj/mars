package cron

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/robfig/cron/v3"
)

// Runner 抽象本地 cron 调度器：注册命令、启动、关闭。
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

// NewRobfigCronV3Runner 构造基于 robfig/cron 的本地 Runner 实现。
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

// AddCommand 注册一个定时任务，name 为任务名、expression 为 cron 表达式。
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

// Run 在后台 goroutine 启动调度器并立即返回。
func (c *robfigCronV3Runner) Run(ctx context.Context) error {
	go func() {
		defer c.logger.HandlePanic("[CRON]: robfig/cron/v3 Run")
		c.c.Run()
	}()
	return nil
}

// Shutdown 停止调度器并等待其退出。
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

// Info 打印 robfig/cron 的 Info 级消息。
func (c *cronLogger) Info(msg string, keysAndValues ...any) {
	c.logger.Infof(formatString(len(keysAndValues)), append([]any{msg}, keysAndValues...)...)
}

// Error 打印 robfig/cron 的错误消息。
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
