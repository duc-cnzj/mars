package utils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBranchPass(t *testing.T) {
	cfg := &mars.Config{
		Branches: []string{"master"},
	}
	assert.True(t, BranchPass(cfg, "master"))
	assert.False(t, BranchPass(cfg, "dev"))
	cfg = &mars.Config{
		Branches: []string{"*"},
	}
	assert.True(t, BranchPass(cfg, "master"))
	cfg = &mars.Config{
		Branches: []string{"dev-*"},
	}
	assert.True(t, BranchPass(cfg, "dev-aaa"))
	assert.False(t, BranchPass(cfg, "nodev-aaa"))
	cfg = &mars.Config{}
	assert.True(t, BranchPass(cfg, "dev-aaa"))
	assert.True(t, BranchPass(cfg, "ccc"))
	cfg = &mars.Config{Branches: []string{"*-dev"}}
	assert.True(t, BranchPass(cfg, "a-dev"))
	assert.True(t, BranchPass(cfg, "b-dev"))
}

func TestGetProjectMarsConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	db.AutoMigrate(&models.GitProject{})
	mc := mars.Config{
		ConfigFile:       "cf",
		ConfigFileValues: "vv",
		ConfigField:      "f",
	}
	marshal, _ := json.Marshal(&mc)
	db.Create(&models.GitProject{
		GitProjectId:  99,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	cfg, _ := GetProjectMarsConfig(99, "dev")
	assert.Equal(t, &mc, cfg)
	db.Create(&models.GitProject{
		GitProjectId:  199,
		GlobalEnabled: false,
	})
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{
		Name: "test_git_server",
		Args: nil,
	}}).AnyTimes()
	gs := mock.NewMockGitServer(ctrl)
	gs.EXPECT().Initialize(gomock.Any()).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	app.EXPECT().GetPluginByName("test_git_server").Return(gs)
	pid := 199
	gs.EXPECT().GetFileContentWithBranch(fmt.Sprintf("%v", pid), "dev", ".mars.yaml").Return(string(marshal), nil)
	cfg, _ = GetProjectMarsConfig(pid, "dev")
	assert.Equal(t, &mc, cfg)
}

func TestIsRemoteChart(t *testing.T) {
	assert.False(t, IsRemoteChart(&mars.Config{LocalChartPath: "abc|branch|path"}))
	assert.True(t, IsRemoteChart(&mars.Config{LocalChartPath: "1|branch|path"}))
	assert.False(t, IsRemoteChart(&mars.Config{LocalChartPath: "pid"}))
}

func TestIsRemoteConfigFile(t *testing.T) {
	assert.False(t, IsRemoteConfigFile(&mars.Config{ConfigFile: "abc|branch|path"}))
	assert.True(t, IsRemoteConfigFile(&mars.Config{ConfigFile: "1|branch|path"}))
	assert.False(t, IsRemoteConfigFile(&mars.Config{ConfigFile: "pid"}))
}

func TestParseInputConfig(t *testing.T) {
	inputConfig, err := ParseInputConfig(nil, "")
	assert.Nil(t, err)
	assert.Empty(t, inputConfig)
	v, _ := ParseInputConfig(&mars.Config{
		IsSimpleEnv: false,
		ConfigField: "conf->config",
	}, `{"name": "duc", "age": 18}`)
	wants := `conf:
  config:
    age: 18
    name: duc
`
	assert.Equal(t, wants, v)
	v, _ = ParseInputConfig(&mars.Config{
		IsSimpleEnv: true,
		ConfigField: "conf->config",
	}, "name: duc\nage: 18")
	wants = `conf:
  config: |-
    name: duc
    age: 18
`
	assert.Equal(t, wants, v)
}

func Test_intPid(t *testing.T) {
	assert.True(t, intPid("1"))
	assert.True(t, intPid("-1"))
	assert.True(t, intPid("10"))
	assert.False(t, intPid("abc"))
	assert.False(t, intPid("1_a"))
}

func TestGetProjectName(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gitS := mock.NewMockGitServer(m)
	app.EXPECT().Config().Return(&config.Config{GitServerPlugin: config.Plugin{Name: "gits"}}).AnyTimes()
	app.EXPECT().GetPluginByName("gits").Return(gitS).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gitS.EXPECT().Initialize(gomock.All()).AnyTimes()
	p := mock.NewMockProjectInterface(m)
	gitS.EXPECT().GetProject("1").Return(p, nil).AnyTimes()
	p.EXPECT().GetName().Return("app").AnyTimes()
	assert.Equal(t, "app", GetProjectName("1", &mars.Config{}))
	assert.Equal(t, "app-2", GetProjectName("1", &mars.Config{DisplayName: "app-2"}))
}
