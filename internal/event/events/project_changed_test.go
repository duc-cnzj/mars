package events

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandleProjectChanged(t *testing.T) {
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
		Namespace:    models.Namespace{Name: "aaa"},
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

	err := HandleProjectChanged(&ProjectChangedData{
		Project:  p,
		Manifest: "Manifest",
		Config:   "Config",
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
}
