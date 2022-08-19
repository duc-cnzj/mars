package cron

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/cron/commands"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type Manager struct {
	runner contracts.CronRunner
	app    contracts.ApplicationInterface

	sync.RWMutex
	commands map[string]*Command
}

func NewManager(runner contracts.CronRunner, app contracts.ApplicationInterface) *Manager {
	return &Manager{commands: make(map[string]*Command), runner: runner, app: app}
}

func (m *Manager) NewCommand(name string, fn func()) contracts.Command {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[name]; ok {
		panic(fmt.Sprintf("[CRON]: job %s already exists", name))
	}
	cmd := &Command{expression: expression, name: name, fn: Wrap(name, fn, m.app.DistributedLocks())}
	m.commands[name] = cmd
	return cmd
}

func (m *Manager) Run(ctx context.Context) error {
	mlog.Info("[Server]: start cron.")
	for _, callback := range commands.RegisteredCronJobs() {
		callback(m, m.app)
	}
	for _, command := range m.List() {
		if err := m.runner.AddCommand(command.Name(), command.Expression(), command.Func()); err != nil {
			return err
		}
	}

	return m.runner.Run(ctx)
}

func (m *Manager) List() []contracts.Command {
	m.RLock()
	defer m.RUnlock()
	var cmds []contracts.Command
	for _, c := range m.commands {
		cmds = append(cmds, &Command{
			name:       c.name,
			expression: c.expression,
			fn:         c.fn,
		})
	}
	sort.Sort(sortCommand(cmds))

	return cmds
}

func (m *Manager) Shutdown(ctx context.Context) error {
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
