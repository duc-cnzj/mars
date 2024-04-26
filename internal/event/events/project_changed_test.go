package events

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandleProjectChanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	commitDate, _ := time.Parse("2006-01-02 15:04:05", "2022-06-17 10:00:00")
	db.AutoMigrate(&models.Changelog{}, &models.GitProject{}, &models.Project{})
	p := &models.Project{
		Name:             "app",
		GitProjectId:     100,
		GitBranch:        "dev",
		GitCommit:        "commit",
		Config:           "cfg",
		OverrideValues:   "xxx",
		DockerImage:      "app:v1",
		PodSelectors:     "app=xx",
		Atomic:           true,
		DeployStatus:     0,
		EnvValues:        "env_vars",
		ExtraValues:      "extra_vars",
		FinalExtraValues: "final_extra_values",
		ConfigType:       "go",
		Manifest:         "manifest",
		GitCommitWebUrl:  "web_url",
		GitCommitTitle:   "title",
		GitCommitAuthor:  "duc",
		GitCommitDate:    &commitDate,
		Namespace:        models.Namespace{Name: "aaa"},
	}
	db.Create(p)
	gp := &models.GitProject{
		GitProjectId: 100,
	}
	db.Create(gp)
	db.Create(&models.Changelog{
		Version:       10,
		Manifest:      "",
		Config:        "cfg100",
		ConfigChanged: false,
		ProjectID:     p.ID,
		GitProjectID:  gp.ID,
	})
	db.Create(&models.Changelog{
		Version:       5,
		Manifest:      "",
		Config:        "cfg99",
		ConfigChanged: false,
		ProjectID:     0,
		GitProjectID:  0,
	})

	p.Config = "Config"
	err := HandleProjectChanged(&ProjectChangedData{
		Project:  p,
		Username: "duc",
	}, EventProjectChanged)
	assert.Nil(t, err)
	clog := models.Changelog{}
	db.Last(&clog)
	assert.Equal(t, int(11), int(clog.Version))
	assert.Equal(t, "duc", clog.Username)
	assert.Equal(t, "Config", clog.Config)
	assert.Equal(t, p.ID, clog.ProjectID)
	assert.Equal(t, true, clog.ConfigChanged)
	assert.Equal(t, gp.ID, clog.GitProjectID)

	assert.Equal(t, clog.Manifest, p.Manifest)
	assert.Equal(t, clog.ConfigType, p.ConfigType)
	assert.Equal(t, clog.GitBranch, p.GitBranch)
	assert.Equal(t, clog.GitCommit, p.GitCommit)
	assert.Equal(t, clog.DockerImage, p.DockerImage)
	assert.Equal(t, clog.EnvValues, p.EnvValues)
	assert.Equal(t, clog.ExtraValues, p.ExtraValues)
	assert.Equal(t, clog.FinalExtraValues, p.FinalExtraValues)
	assert.Equal(t, clog.GitCommitWebUrl, p.GitCommitWebUrl)
	assert.Equal(t, clog.GitCommitTitle, p.GitCommitTitle)
	assert.Equal(t, clog.GitCommitAuthor, p.GitCommitAuthor)
	assert.Equal(t, clog.GitCommitDate, p.GitCommitDate)
}

func TestHandleProjectChanged_ConfigChanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	db.AutoMigrate(&models.Changelog{}, &models.GitProject{}, &models.Project{})
	p := &models.Project{
		Name:         "app",
		GitProjectId: 100,
		GitCommit:    "commit",
		Config:       "cfg",
		Namespace:    models.Namespace{Name: "aaa"},
	}
	db.Create(p)
	gp := &models.GitProject{
		GitProjectId: 100,
	}
	db.Create(gp)
	db.Create(&models.Changelog{
		Version:       10,
		Config:        "cfg",
		GitCommit:     "commit",
		ConfigChanged: false,
		ProjectID:     p.ID,
		GitProjectID:  gp.ID,
	})

	err := HandleProjectChanged(&ProjectChangedData{
		Project:  p,
		Username: "duc",
	}, EventProjectChanged)
	assert.Nil(t, err)
	clog := models.Changelog{}
	db.Last(&clog)
	assert.False(t, clog.ConfigChanged)
	assert.Nil(t, db.Model(&p).UpdateColumn("git_commit", "commit2").Error)
	err = HandleProjectChanged(&ProjectChangedData{
		Project:  p,
		Username: "duc",
	}, EventProjectChanged)
	assert.Nil(t, err)
	clog2 := models.Changelog{}
	db.Last(&clog2)
	assert.True(t, clog2.ConfigChanged)

	err = HandleProjectChanged(&ProjectChangedData{
		Project:  p,
		Username: "duc",
	}, EventProjectChanged)
	assert.Nil(t, err)
	clog3 := models.Changelog{}
	db.Last(&clog3)
	assert.False(t, clog3.ConfigChanged)

	assert.Nil(t, db.Model(&p).UpdateColumn("config", "cfg2").Error)
	err = HandleProjectChanged(&ProjectChangedData{
		Project:  p,
		Username: "duc",
	}, EventProjectChanged)
	assert.Nil(t, err)
	clog4 := models.Changelog{}
	db.Last(&clog4)
	assert.True(t, clog4.ConfigChanged)
}
