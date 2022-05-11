package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/changelog"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestChangelogSvc_Show(t *testing.T) {
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
	db.AutoMigrate(&models.Changelog{})
	c := new(ChangelogSvc)
	seedChangelog(db)
	//"ID", "Version", "Username", "Config", "ConfigChanged", "ProjectID", "GitProjectID"
	show, err := c.Show(context.TODO(), &changelog.ShowRequest{
		ProjectId:   1,
		OnlyChanged: false,
	})
	assert.Nil(t, err)
	assert.Len(t, show.Items, 5)
	assert.Equal(t, "duc6", show.Items[0].Username)
	assert.Equal(t, "config6", show.Items[0].Config)
	assert.Equal(t, true, show.Items[0].ConfigChanged)
	assert.Equal(t, int64(100), show.Items[0].GitProjectId)
	assert.Equal(t, int64(1), show.Items[0].ProjectId)
	assert.Equal(t, int64(6), show.Items[0].Version)
	assert.Equal(t, int64(5), show.Items[1].Version)
	assert.Equal(t, int64(4), show.Items[2].Version)
	assert.Equal(t, int64(3), show.Items[3].Version)
	assert.Equal(t, int64(2), show.Items[4].Version)
	show, _ = c.Show(context.TODO(), &changelog.ShowRequest{
		ProjectId:   1,
		OnlyChanged: true,
	})
	assert.Len(t, show.Items, 4)
}

func seedChangelog(db *gorm.DB) {
	var testData = []models.Changelog{
		{
			Version:       1,
			Username:      "duc",
			ConfigChanged: false,
			ProjectID:     1,
		},
		{
			Version:       2,
			Username:      "ducb",
			ConfigChanged: false,
			ProjectID:     1,
		},
		{
			Version:       3,
			Username:      "duc3",
			ConfigChanged: true,
			ProjectID:     1,
		},
		{
			Version:       4,
			Username:      "duc4",
			ConfigChanged: true,
			ProjectID:     1,
		},
		{
			Version:       5,
			Username:      "duc5",
			ConfigChanged: true,
			ProjectID:     1,
		},
		{
			Version:       6,
			Username:      "duc6",
			Config:        "config6",
			ConfigChanged: true,
			ProjectID:     1,
			GitProjectID:  100,
		},
	}
	for _, datum := range testData {
		db.Create(&datum)
	}
}
