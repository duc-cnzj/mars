package cron

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/locker"
	"github.com/duc-cnzj/mars/v5/internal/metrics"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/rand"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/prometheus/client_golang/prometheus"
)

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

func NewManager(timer timer.Timer, runner Runner, locker locker.Locker, logger mlog.Logger) Manager {
	return &cronManager{
		timer:    timer,
		runner:   runner,
		Locker:   locker,
		logger:   logger.WithModule("cron/cronManager"),
		commands: make(map[string]*command),
	}
}

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

func (m *cronManager) Run(ctx context.Context) error {
	m.logger.Info("[Server]: start cron.")
	for _, cmd := range m.List() {
		if err := m.runner.AddCommand(cmd.Name(), cmd.Expression(), cmd.Func()); err != nil {
			return err
		}
	}

	return m.runner.Run(ctx)
}

func (m *cronManager) Shutdown(ctx context.Context) error {
	m.logger.Info("[Server]: shutdown cron manager.")
	return m.runner.Shutdown(ctx)
}

type sortCommand []Command

func (s sortCommand) Len() int {
	return len(s)
}

func (s sortCommand) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

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
