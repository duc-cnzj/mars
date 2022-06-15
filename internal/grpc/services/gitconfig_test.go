package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/gitconfig"
	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGitConfigSvc_Authorize(t *testing.T) {
	e := new(GitConfigSvc)
	ctx := context.TODO()
	ctx = auth.SetUser(ctx, &contracts.UserInfo{})
	_, err := e.Authorize(ctx, "")
	assert.ErrorIs(t, err, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error()))
	ctx = auth.SetUser(ctx, &contracts.UserInfo{
		Roles: []string{"admin"},
	})
	_, err = e.Authorize(ctx, "")
	assert.Nil(t, err)
}

func TestGitConfigSvc_GetDefaultChartValues(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	marsC := mars.Config{
		LocalChartPath: "",
	}
	marshal, _ := json.Marshal(&marsC)
	db.AutoMigrate(&models.GitProject{})
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  10,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	marsC2 := mars.Config{
		LocalChartPath: "1|master|charts",
	}
	marshal2, _ := json.Marshal(&marsC2)
	marsC3 := mars.Config{
		LocalChartPath: "charts",
	}
	marshal3, _ := json.Marshal(&marsC3)
	db.Create(&models.GitProject{
		DefaultBranch: "dev2",
		Name:          "app2",
		GitProjectId:  11,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal2),
	})
	db.Create(&models.GitProject{
		DefaultBranch: "dev3",
		Name:          "app3",
		GitProjectId:  12,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal3),
	})
	_, err := new(GitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 10,
		Branch:       "dev",
	})
	assert.Nil(t, err)
	gits := mockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("1", "master", "charts/values.yaml").Times(1).Return("xxx", nil)
	res, _ := new(GitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 11,
		Branch:       "dev",
	})
	assert.Equal(t, "xxx", res.Value)
	gits.EXPECT().GetFileContentWithBranch("12", "dev", "charts/values.yaml").Times(1).Return("aaa", nil)
	res, _ = new(GitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 12,
		Branch:       "dev",
	})
	assert.Equal(t, "aaa", res.Value)
}

func TestGitConfigSvc_GlobalConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})
	config, err := new(GitConfigSvc).GlobalConfig(context.TODO(), &gitconfig.GlobalConfigRequest{
		GitProjectId: 11,
	})
	assert.Nil(t, err)
	assert.Equal(t, (&mars.Config{}).String(), config.Config.String())
	assert.Equal(t, false, config.Enabled)
	marsC := mars.Config{
		ConfigFile:       "f",
		ConfigFileValues: "x",
		ConfigField:      "ff",
		IsSimpleEnv:      true,
		ConfigFileType:   "yaml",
		LocalChartPath:   "1|master|charts",
		Branches:         []string{"dev"},
		ValuesYaml:       "aaa",
		Elements: []*mars.Element{{
			Path:         "p",
			Type:         0,
			Default:      "xx",
			Description:  "desc",
			SelectValues: nil,
		}},
	}
	marshal, _ := json.Marshal(&marsC)
	db.Create(&models.GitProject{
		ID:            0,
		DefaultBranch: "dev",
		Name:          "name",
		GitProjectId:  11,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})

	config, err = new(GitConfigSvc).GlobalConfig(context.TODO(), &gitconfig.GlobalConfigRequest{
		GitProjectId: 11,
	})
	assert.Nil(t, err)
	assert.Equal(t, marsC.String(), config.Config.String())
	assert.Equal(t, true, config.Enabled)
}

func TestGitConfigSvc_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	marsC := mars.Config{
		LocalChartPath: "aaa",
	}
	marshal, _ := json.Marshal(&marsC)
	db.AutoMigrate(&models.GitProject{})
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  11,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	db.Create(&models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  12,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	})
	show, err := new(GitConfigSvc).Show(context.TODO(), &gitconfig.ShowRequest{
		GitProjectId: 11,
		Branch:       "dev",
	})
	assert.Nil(t, err)
	assert.Equal(t, marsC.String(), show.Config.String())
	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject(gomock.Any()).Return(p, nil)
	p.EXPECT().GetDefaultBranch().Return("abc")
	show, err = new(GitConfigSvc).Show(context.TODO(), &gitconfig.ShowRequest{
		GitProjectId: 12,
		Branch:       "",
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc", show.Branch)

	gits.EXPECT().GetFileContentWithBranch("199", "dev199", ".mars.yaml").Return("", errors.New("xxx"))
	show, err = new(GitConfigSvc).Show(context.TODO(), &gitconfig.ShowRequest{
		GitProjectId: 199,
		Branch:       "dev199",
	})
	assert.Nil(t, err)
	assert.Equal(t, "dev199", show.Branch)
	assert.Equal(t, (&mars.Config{}).String(), show.Config.String())
}

func TestGitConfigSvc_ToggleGlobalStatus(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	marsC := mars.Config{
		LocalChartPath: "aaa",
	}
	marshal, _ := json.Marshal(&marsC)
	db.AutoMigrate(&models.GitProject{})
	p := &models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  11,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	}
	db.Create(p)
	assertAuditLogFired(m, app)
	_, err := new(GitConfigSvc).ToggleGlobalStatus(adminCtx(), &gitconfig.ToggleGlobalStatusRequest{
		GitProjectId: 11,
		Enabled:      false,
	})
	assert.Nil(t, err)
	db.First(p)
	assert.False(t, p.GlobalEnabled)
	_, err = new(GitConfigSvc).ToggleGlobalStatus(adminCtx(), &gitconfig.ToggleGlobalStatusRequest{
		GitProjectId: 12,
		Enabled:      false,
	})
	assert.Nil(t, err)
	newp := &models.GitProject{}
	db.Where("`git_project_id` = ?", 12).First(&newp)
	assert.NotZero(t, newp.ID)
}

func TestGitConfigSvc_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	marsC := mars.Config{
		LocalChartPath: "aaa",
	}
	marshal, _ := json.Marshal(&marsC)
	db.AutoMigrate(&models.GitProject{})
	p := &models.GitProject{
		DefaultBranch: "dev",
		Name:          "app",
		GitProjectId:  11,
		Enabled:       true,
		GlobalEnabled: true,
		GlobalConfig:  string(marshal),
	}
	db.Create(p)
	mc := &mars.Config{
		ConfigFile:       "xx",
		ConfigFileValues: "a",
		ConfigField:      "env",
		IsSimpleEnv:      true,
		ConfigFileType:   "",
		LocalChartPath:   "",
		Branches:         []string{"dev"},
		ValuesYaml:       "",
		Elements:         nil,
	}
	d := assertAuditLogFired(m, app)
	update, err := new(GitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config:       mc,
	})
	assert.Nil(t, err)
	assert.Equal(t, mc.String(), update.Config.String())
	mc.Branches = nil
	d.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(1)
	update, _ = new(GitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config:       mc,
	})
	assert.Equal(t, []string{"*"}, update.Config.Branches)

	cache := mock.NewMockCacheInterface(m)
	app.EXPECT().Cache().Return(cache)
	mc.DisplayName = "app"
	cache.EXPECT().Clear("ProjectOptions").Times(1)
	d.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(1)
	update, _ = new(GitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config:       mc,
	})
	assert.Equal(t, "app", update.Config.DisplayName)
	_, err = new(GitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 9999999,
		Config:       mc,
	})
	assert.Equal(t, "record not found", err.Error())
}

func Test_getDefaultBranch(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	gits := mockGitServer(m, app)
	gits.EXPECT().GetProject(gomock.Any()).Return(nil, errors.New(""))
	branch, err := getDefaultBranch(1)
	assert.Empty(t, branch)
	assert.Error(t, err)
	p := mock.NewMockProjectInterface(m)
	p.EXPECT().GetDefaultBranch().Return("abc")
	gits.EXPECT().GetProject(gomock.Any()).Return(p, nil)
	branch, _ = getDefaultBranch(1)
	assert.Equal(t, "abc", branch)
}
