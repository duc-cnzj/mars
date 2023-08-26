package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cache2 "github.com/duc-cnzj/mars/v4/internal/cache"

	"github.com/duc-cnzj/mars-client/v4/gitconfig"
	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGitConfigSvc_Authorize(t *testing.T) {
	e := new(gitConfigSvc)
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
	_, err := new(gitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 10,
		Branch:       "dev",
	})
	assert.Nil(t, err)
	gits := mockGitServer(m, app)
	gits.EXPECT().GetFileContentWithBranch("1", "master", "charts/values.yaml").Times(1).Return("", errors.New("xxx"))
	_, err = new(gitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 11,
		Branch:       "dev",
	})
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, s.Code())
	assert.Equal(t, "xxx", s.Message())
	gits.EXPECT().GetFileContentWithBranch("1", "master", "charts/values.yaml").Times(1).Return("xxx", nil)
	res, _ := new(gitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 11,
		Branch:       "dev",
	})
	assert.Equal(t, "xxx", res.Value)
	gits.EXPECT().GetFileContentWithBranch("12", "dev", "charts/values.yaml").Times(1).Return("aaa", nil)
	res, _ = new(gitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 12,
		Branch:       "dev",
	})
	assert.Equal(t, "aaa", res.Value)

	gits.EXPECT().GetFileContentWithBranch("12", "master", "charts/values.yaml").Times(1).Return("", errors.New("bbb"))

	_, err = new(gitConfigSvc).GetDefaultChartValues(context.TODO(), &gitconfig.DefaultChartValuesRequest{
		GitProjectId: 12,
		Branch:       "",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	assert.Equal(t, "bbb", fromError.Message())
}

func TestGitConfigSvc_GlobalConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	db.AutoMigrate(&models.GitProject{})
	config, err := new(gitConfigSvc).GlobalConfig(context.TODO(), &gitconfig.GlobalConfigRequest{
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

	config, err = new(gitConfigSvc).GlobalConfig(context.TODO(), &gitconfig.GlobalConfigRequest{
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
	show, err := new(gitConfigSvc).Show(context.TODO(), &gitconfig.ShowRequest{
		GitProjectId: 11,
		Branch:       "dev",
	})
	assert.Nil(t, err)
	assert.Equal(t, marsC.String(), show.Config.String())
	gits := mockGitServer(m, app)
	p := mock.NewMockProjectInterface(m)
	gits.EXPECT().GetProject(gomock.Any()).Return(p, nil)
	p.EXPECT().GetDefaultBranch().Return("abc")
	show, err = new(gitConfigSvc).Show(context.TODO(), &gitconfig.ShowRequest{
		GitProjectId: 12,
		Branch:       "",
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc", show.Branch)

	gits.EXPECT().GetFileContentWithBranch("199", "dev199", ".mars.yaml").Return("", errors.New("xxx"))
	show, err = new(gitConfigSvc).Show(context.TODO(), &gitconfig.ShowRequest{
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
	testutil.AssertAuditLogFired(m, app)
	_, err := new(gitConfigSvc).ToggleGlobalStatus(adminCtx(), &gitconfig.ToggleGlobalStatusRequest{
		GitProjectId: 11,
		Enabled:      false,
	})
	assert.Nil(t, err)
	db.First(p)
	assert.False(t, p.GlobalEnabled)
	_, err = new(gitConfigSvc).ToggleGlobalStatus(adminCtx(), &gitconfig.ToggleGlobalStatusRequest{
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
	d := testutil.AssertAuditLogFired(m, app)
	update, err := new(gitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config:       mc,
	})
	assert.Nil(t, err)
	assert.Equal(t, mc.String(), update.Config.String())
	mc.Branches = nil
	d.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(2)
	update, _ = new(gitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config:       mc,
	})
	assert.Equal(t, []string{"*"}, update.Config.Branches)

	cache := mock.NewMockCacheInterface(m)
	app.EXPECT().Cache().Return(cache).AnyTimes()
	mc.DisplayName = "app"
	cache.EXPECT().Clear(cache2.NewKey("ProjectOptions")).Times(2)
	d.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(2)
	update, _ = new(gitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config:       mc,
	})
	assert.Equal(t, "app", update.Config.DisplayName)
	_, err = new(gitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 9999999,
		Config:       mc,
	})
	assert.Equal(t, "record not found", err.Error())
	update, _ = new(gitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config: &mars.Config{
			ConfigField:      "",
			ConfigFileValues: " aa ",
			ValuesYaml:       " bb ",
			IsSimpleEnv:      false,
		},
	})
	assert.Equal(t, " aa", update.Config.ConfigFileValues)
	assert.Equal(t, " bb", update.Config.ValuesYaml)
	assert.Equal(t, true, update.Config.IsSimpleEnv)
	update, _ = new(gitConfigSvc).Update(adminCtx(), &gitconfig.UpdateRequest{
		GitProjectId: 11,
		Config: &mars.Config{
			ConfigField: "xxxx",
			IsSimpleEnv: true,
		},
	})
	assert.Equal(t, true, update.Config.IsSimpleEnv)
}

func TestConfigDefaultNotRequired(t *testing.T) {
	var tests = []struct {
		input      *mars.Element
		wantsError bool
	}{
		{
			input: &mars.Element{
				Path:    "xx",
				Default: "xx",
			},
			wantsError: false,
		},
		{
			input: &mars.Element{
				Path:    "xx",
				Default: "",
			},
			wantsError: false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			if tt.wantsError {
				assert.Error(t, tt.input.Validate())
			} else {
				assert.NoError(t, tt.input.Validate())
			}
		})
	}
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
