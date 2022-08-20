package commands

import (
	"fmt"
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/cron"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils/recovery"
)

var AllGitProjectCache = func() error {
	app.Cache().Clear(plugins.CacheKeyAllProjects())
	if _, err := plugins.GetGitServer().AllProjects(); err != nil {
		return err
	}
	return nil
}

var AllBranchCache = func() error {
	var (
		enabledGitProjects []*models.GitProject
		wg                 = &sync.WaitGroup{}
	)

	app.DB().Where("`enabled` = ?", true).Find(&enabledGitProjects)
	goroutineNum := len(enabledGitProjects)

	if len(enabledGitProjects) > 10 {
		goroutineNum = 8
	}

	ch := make(chan *models.GitProject, goroutineNum)
	for i := 0; i < goroutineNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer recovery.HandlePanic("[CRON]: all_branch_cache")
			for gitProject := range ch {
				app.Cache().Clear(plugins.CacheKeyAllBranches(gitProject.GitProjectId))
				_, err := plugins.GetGitServer().AllBranches(fmt.Sprintf("%d", gitProject.GitProjectId))
				mlog.Debugf("[CRON]: fetch AllBranches: '%s' '%d', err: '%v'", gitProject.Name, gitProject.GitProjectId, err)
			}
		}()
	}
	for i := range enabledGitProjects {
		ch <- enabledGitProjects[i]
	}
	close(ch)
	wg.Wait()
	return nil
}

func init() {
	cron.Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {
		if app.Config().GitServerCached {
			manager.NewCommand("all_git_project_cache", AllGitProjectCache).EveryFiveMinutes()
			manager.NewCommand("all_branch_cache", AllBranchCache).EveryTwoMinutes()
		}
	})
}
