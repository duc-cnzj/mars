package services

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/changelog"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestChangelogSvc_Show(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := mock.NewMockApplicationInterface(ctrl)
	instance.SetInstance(app)

	db, f := testutil.SetGormDB(ctrl, app)
	defer f()

	c := new(ChangelogSvc)
	_, err := c.Show(context.TODO(), &changelog.ShowRequest{
		ProjectId:   int64(9999),
		OnlyChanged: false,
	})
	assert.Error(t, err)
	db.AutoMigrate(&models.Changelog{}, &models.Project{}, &models.GitProject{}, &models.Namespace{})
	p1 := &models.Project{Name: "p1", Namespace: models.Namespace{Name: "aaa"}}
	assert.Nil(t, db.Create(p1).Error)
	gitp1 := &models.GitProject{
		GitProjectId: 100,
	}
	assert.Nil(t, db.Create(gitp1).Error)
	gitp2 := &models.GitProject{
		GitProjectId: 101,
	}
	assert.Nil(t, db.Create(gitp2).Error)
	parse, _ := time.Parse("2006-01-02 00:00:00", "2022-06-17 00:00:00")
	var testData = []models.Changelog{
		{
			Version:       1,
			Username:      "duc",
			ConfigChanged: false,
			ProjectID:     p1.ID,
			GitProjectID:  gitp2.ID,
		},
		{
			Version:       2,
			Username:      "ducb",
			ConfigChanged: false,
			ProjectID:     p1.ID,
			GitProjectID:  gitp2.ID,
		},
		{
			Version:       3,
			Username:      "duc3",
			ConfigChanged: true,
			ProjectID:     p1.ID,
			GitProjectID:  gitp2.ID,
		},
		{
			Version:       4,
			Username:      "duc4",
			ConfigChanged: true,
			ProjectID:     p1.ID,
			GitProjectID:  gitp2.ID,
		},
		{
			Version:       5,
			Username:      "duc5",
			ConfigChanged: true,
			ProjectID:     p1.ID,
			GitProjectID:  gitp2.ID,
		},
		{
			ID:              0,
			Version:         6,
			Username:        "duc6",
			Config:          "config6",
			GitCommitWebUrl: "url",
			GitCommitTitle:  "title",
			GitCommitAuthor: "duc",
			GitCommitDate:   &parse,
			ConfigChanged:   true,
			ProjectID:       p1.ID,
			GitProjectID:    gitp1.ID,
			Project:         models.Project{},
			GitProject:      models.GitProject{},
		},
	}
	for _, datum := range testData {
		assert.Nil(t, db.Create(&datum).Error)
	}
	//"ID", "Version", "Username", "Config", "ConfigChanged", "ProjectID", "GitProjectID"
	show, err := c.Show(context.TODO(), &changelog.ShowRequest{
		ProjectId:   int64(p1.ID),
		OnlyChanged: false,
	})
	assert.Nil(t, err)
	assert.Len(t, show.Items, 5)

	assert.Equal(t, "title", show.Items[0].GitCommitTitle)
	assert.Equal(t, date.ToHumanizeDatetimeString(&parse), show.Items[0].GitCommitDate)
	assert.Equal(t, "url", show.Items[0].GitCommitWebUrl)
	assert.NotEqual(t, "很久以前", show.Items[0].Date)
	assert.Equal(t, "duc", show.Items[0].GitCommitAuthor)

	assert.Equal(t, "duc6", show.Items[0].Username)
	assert.Equal(t, "config6", show.Items[0].Config)
	assert.Equal(t, true, show.Items[0].ConfigChanged)

	assert.Equal(t, int64(gitp1.ID), show.Items[0].GitProjectId)
	assert.Equal(t, int64(p1.ID), show.Items[0].ProjectId)
	assert.Equal(t, int64(6), show.Items[0].Version)
	assert.Equal(t, int64(5), show.Items[1].Version)
	assert.Equal(t, int64(4), show.Items[2].Version)
	assert.Equal(t, int64(3), show.Items[3].Version)
	assert.Equal(t, int64(2), show.Items[4].Version)
	show, _ = c.Show(context.TODO(), &changelog.ShowRequest{
		ProjectId:   int64(p1.ID),
		OnlyChanged: true,
	})
	assert.Len(t, show.Items, 4)
}
