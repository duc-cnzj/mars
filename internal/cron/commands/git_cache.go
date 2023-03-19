package commands

import (
	"fmt"
	"sync"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
)

var AllGitProjectCache = func() error {
	var gitServer plugins.GitServer = plugins.GetGitServer()
	if cache, ok := gitServer.(plugins.GitCacheServer); ok {
		return cache.ReCacheAllProjects()
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
	gitServer := plugins.GetGitServer()
	if server, ok := gitServer.(plugins.GitCacheServer); ok {
		for i := 0; i < goroutineNum; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer recovery.HandlePanic("[CRON]: all_branch_cache")
				for gitProject := range ch {
					err := server.ReCacheAllBranches(fmt.Sprintf("%d", gitProject.GitProjectId))
					mlog.Debugf("[CRON]: fetch AllBranches: '%s' '%d', err: '%v'", gitProject.Name, gitProject.GitProjectId, err)
				}
			}()
		}
		for i := range enabledGitProjects {
			ch <- enabledGitProjects[i]
		}
		close(ch)
		wg.Wait()
	}

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
