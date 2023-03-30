package cron

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type cronManager struct {
	runner contracts.CronRunner
	app    contracts.ApplicationInterface

	sync.RWMutex
	commands map[string]*command
}

func NewManager(runner contracts.CronRunner, app contracts.ApplicationInterface) contracts.CronManager {
	return &cronManager{commands: make(map[string]*command), runner: runner, app: app}
}

func (m *cronManager) NewCommand(name string, fn func() error) contracts.Command {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[name]; ok {
		panic(fmt.Sprintf("[CRON]: job %s already exists", name))
	}
	cmd := &command{expression: expression, name: name, fn: wrap(name, fn, func() contracts.Locker {
		return m.app.CacheLock()
	})}
	m.commands[name] = cmd
	return cmd
}

func (m *cronManager) Run(ctx context.Context) error {
	mlog.Info("[Server]: start cron.")
	for _, callback := range RegisteredCronJobs() {
		callback(m, m.app)
	}
	for _, command := range m.List() {
		if err := m.runner.AddCommand(command.Name(), command.Expression(), command.Func()); err != nil {
			return err
		}
	}

	return m.runner.Run(ctx)
}

func (m *cronManager) List() []contracts.Command {
	m.RLock()
	defer m.RUnlock()
	var cmds []contracts.Command
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

func (m *cronManager) Shutdown(ctx context.Context) error {
	mlog.Info("[Server]: shutdown cron manager.")
	return m.runner.Shutdown(ctx)
}

type sortCommand []contracts.Command

func (s sortCommand) Len() int {
	return len(s)
}

func (s sortCommand) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

func (s sortCommand) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
