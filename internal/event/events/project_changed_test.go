package events

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHandleProjectChanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(ctrl)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.Changelog{}, &models.GitProject{})
	db.Create(&models.Changelog{
		Version:       10,
		Manifest:      "",
		Config:        "cfg100",
		ConfigChanged: false,
		ProjectID:     0,
		GitProjectID:  100,
	})
	db.Create(&models.Changelog{
		Version:       5,
		Manifest:      "",
		Config:        "cfg99",
		ConfigChanged: false,
		ProjectID:     0,
		GitProjectID:  99,
	})
	db.Create(&models.GitProject{
		GitProjectId: 100,
	})

	HandleProjectChanged(&ProjectChangedData{
		Project: &models.Project{
			ID:           888,
			GitProjectId: 100,
		},
		Manifest: "Manifest",
		Config:   "Config",
		Username: "duc",
	}, EventProjectChanged)
	clog := models.Changelog{}
	db.Last(&clog)
	assert.Equal(t, uint8(11), clog.Version)
	assert.Equal(t, "duc", clog.Username)
	assert.Equal(t, "Config", clog.Config)
	assert.Equal(t, 888, clog.ProjectID)
	assert.Equal(t, true, clog.ConfigChanged)
	assert.Equal(t, 100, clog.GitProjectID)
}
