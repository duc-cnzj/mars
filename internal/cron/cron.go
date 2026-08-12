package cron

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/prometheus/client_golang/prometheus"
)

// Manager 收口 cron 任务的生命周期：注册命令、启动调度、关闭回收。
type Manager interface {
	NewCommand(name string, fn func() error) Command
	Run(context.Context) error
	Shutdown(context.Context) error
	List() []Command
}

var _ Manager = (*cronManager)(nil)

type cronManager struct {
	timer  timer.Timer
	runner Runner
	Locker locker.Locker
	logger mlog.Logger
	sync.RWMutex
	commands map[string]*command
}

// NewManager 构造 cronManager，注入 timer/runner/locker/logger。
func NewManager(timer timer.Timer, runner Runner, locker locker.Locker, logger mlog.Logger) Manager {
	return &cronManager{
		timer:    timer,
		runner:   runner,
		Locker:   locker,
		logger:   logger.WithModule("cron/cronManager"),
		commands: make(map[string]*command),
	}
}

// List 返回全部命令的拷贝（按名字排序），避免调用方直接改内部命令。
func (m *cronManager) List() []Command {
	m.RLock()
	defer m.RUnlock()
	var cmds []Command
	for _, c := range m.commands {
		cmds = append(cmds, &command{
			name:       c.name,
			expression: c.expression,
			fn:         c.fn,
		})
	}
	sort.Sort(sortCommand(cmds))

	return cmds
}

// NewCommand 注册并返回新命令；任务名重复时 panic（装配期错误尽早暴露）。
func (m *cronManager) NewCommand(name string, fn func() error) Command {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[name]; ok {
		panic(fmt.Sprintf("[CRON]: job %s already exists", name))
	}
	cmd := &command{expression: expression, name: name, fn: m.wrap(name, fn)}
	m.commands[name] = cmd
	return cmd
}

// Run 把所有命令注册进 runner 并启动，任一注册失败立即返回。
func (m *cronManager) Run(ctx context.Context) error {
	m.logger.Info("[Server]: start cron.")
	for _, cmd := range m.List() {
		if err := m.runner.AddCommand(cmd.Name(), cmd.Expression(), cmd.Func()); err != nil {
			return err
		}
	}

	return m.runner.Run(ctx)
}

// Shutdown 关闭 cron runner。
func (m *cronManager) Shutdown(ctx context.Context) error {
	m.logger.Info("[Server]: shutdown cron manager.")
	return m.runner.Shutdown(ctx)
}

type sortCommand []Command

// Len 返回命令个数，sort.Interface 实现。
func (s sortCommand) Len() int {
	return len(s)
}

// Less 按名字升序比较，sort.Interface 实现。
func (s sortCommand) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

// Swap 交换两个命令位置，sort.Interface 实现。
func (s sortCommand) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

const (
	defaultLockSeconds  int64 = 30
	defaultRenewSeconds int64 = 20
)

func (m *cronManager) wrap(name string, fn func() error) func() {
	label := prometheus.Labels{"cron_name": name}
	return func() {
		defer m.logger.HandlePanicWithCallback("[CRON]: "+name, func(err error) {
			metrics.CronPanicCount.With(label).Inc()
		})

		time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)
		releaseFn, acquired := m.Locker.RenewalAcquire(lockKey(name), defaultLockSeconds, defaultRenewSeconds)
		if acquired {
			now := m.timer.Now()
			defer func(t time.Time) {
				m.logger.Infof("[CRON-DONE: %s]: '%s' done, use %s.", m.Locker.ID(), name, m.timer.Since(t))
				metrics.CronDuration.With(label).Observe(m.timer.Since(t).Seconds())
				metrics.CronCommandCount.With(label).Inc()
			}(now)
			m.logger.Infof("[CRON-START: %s]: '%s' start at %s.", m.Locker.ID(), name, now.Format("2006-01-02 15:04:05.000"))
			defer releaseFn()

			if err := fn(); err != nil {
				m.logger.Errorf("[CRON]: '%s' err: '%v'", name, err)
				metrics.CronErrorCount.With(label).Inc()
			}
		}
	}
}

func lockKey(name string) string {
	return fmt.Sprintf("cron-%s", name)
}
