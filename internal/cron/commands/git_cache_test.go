package commands

import (
	"errors"
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
	mm := make(map[string]struct{})
	for _, command := range cm.List() {
		mm[command.Name()] = struct{}{}
	}
	_, ok := mm["all_git_project_cache"]
	assert.True(t, ok)
	_, ok = mm["all_branch_cache"]
	assert.True(t, ok)
}

func mockGitServer(m *gomock.Controller, app *mock.MockApplicationInterface, cache bool) *mock.MockGitServer {
	gitS := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin: config.Plugin{
			Name: "test_git_server",
		},
		GitServerCached: cache,
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
	git := mockGitServer(m, app, true)
	app.EXPECT().Cache().Return(c).AnyTimes()
	c.EXPECT().Clear(plugins.CacheKeyAllProjects()).Times(0)
	git.EXPECT().AllProjects().Return(nil, errors.New("xx"))
	assert.Equal(t, "xx", allGitProjectCache().Error())
	c.EXPECT().SetWithTTL(plugins.CacheKeyAllProjects(), []byte("[]"), plugins.AllProjectsCacheSeconds).Times(1)
	git.EXPECT().AllProjects().Return(nil, nil)
	assert.Nil(t, allGitProjectCache())
}

func TestAllGitProjectCache1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	mockGitServer(m, app, false)
	assert.Nil(t, allGitProjectCache())
}

func TestAllBranchCache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	c := mock.NewMockCacheInterface(m)
	git := mockGitServer(m, app, true)
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

	c.EXPECT().SetWithTTL(gomock.Any(), []byte("[]"), plugins.AllBranchesCacheSeconds).Times(20)
	allBranchCache()
}

func TestAllBranchCache1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.GitProject{})
	mockGitServer(m, app, true)
	assert.Nil(t, allBranchCache())
}
