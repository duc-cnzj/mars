package bootstrappers

import (
	"fmt"
	"sync"

	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"

	"github.com/duc-cnzj/mars/internal/contracts"
)

type CronBootstrapper struct{}

func (c *CronBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cm := app.CronManager()
	if app.Config().GitServerCached {
		cm.NewCommand("all_git_project_cache", func() {
			plugins.GetGitServer().AllProjects()
		}).EveryFifteenSeconds()
		cm.NewCommand("all_branch_cache", func() {
			var enabledGitProjects []*models.GitProject
			app.DBManager().DB().Where("`enabled` = ?", true).Find(&enabledGitProjects)
			wg := &sync.WaitGroup{}
			for _, p := range enabledGitProjects {
				wg.Add(1)
				go func(pid int) {
					defer utils.HandlePanic("[CRON]: all_branch_cache")
					plugins.GetGitServer().AllBranches(fmt.Sprintf("%d", pid))
				}(p.GitProjectId)
			}
			wg.Wait()
		}).EveryFifteenSeconds()
	}
	//app.CronManager().NewCommand("panic", func() {
	//	time.Sleep(1000 * time.Microsecond)
	//	panic("err here")
	//}).EveryFiveSeconds()
	//app.CronManager().NewCommand("one", func() {
	//	time.Sleep(800 * time.Millisecond)
	//	mlog.Info("one", time.Now())
	//}).EverySecond()
	//app.CronManager().NewCommand("one-sleep", func() {
	//	mlog.Info("one-sleep", time.Now())
	//	time.Sleep(1500 * time.Millisecond)
	//}).EverySecond()
	//app.CronManager().NewCommand("two", func() {
	//	time.Sleep(1000 * time.Microsecond)
	//	mlog.Info("two: ", time.Now())
	//}).EveryFourSeconds()

	app.AddServer(app.CronManager())
	return nil
}
