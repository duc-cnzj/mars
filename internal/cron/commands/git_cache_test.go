package commands

import (
	"errors"
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/cron"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGitCache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{GitServerCached: true})
	cm := cron.NewManager(nil, app)
	for _, callback := range cron.RegisteredCronJobs() {
		callback(cm, app)
	}
	assert.Len(t, cm.List(), 2)
}

func mockGitServer(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockGitServer {
	gitS := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin: config.Plugin{
			Name: "test_git_server",
		},
	}).AnyTimes()
	app.EXPECT().GetPluginByName("test_git_server").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	return gitS
}

func TestAllGitProjectCache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	git := mockGitServer(m, app)
	app.EXPECT().Cache().Return(c).AnyTimes()
	c.EXPECT().Clear(plugins.CacheKeyAllProjects()).Times(2)
	git.EXPECT().AllProjects().Return(nil, errors.New("xx"))
	assert.Equal(t, "xx", AllGitProjectCache().Error())
	git.EXPECT().AllProjects().Return(nil, nil)
	assert.Nil(t, AllGitProjectCache())
}

func TestAllBranchCache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	git := mockGitServer(m, app)
	app.EXPECT().Cache().Return(c).AnyTimes()
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.GitProject{})

	for i := 0; i < 20; i++ {
		git.EXPECT().AllBranches(fmt.Sprintf("%d", i)).Times(1).Return(nil, nil)
		db.Create(&models.GitProject{
			Name:         fmt.Sprintf("%d", i),
			GitProjectId: i,
			Enabled:      true,
		})
	}
	db.Create(&models.GitProject{
		Name:         fmt.Sprintf("%d", 21),
		GitProjectId: 21,
		Enabled:      false,
	})

	c.EXPECT().Clear(gomock.Any()).Times(20)
	AllBranchCache()
}
