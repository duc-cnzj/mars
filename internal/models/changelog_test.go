package models

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/utils/date"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestChangelog_ProtoTransform(t *testing.T) {
	m := &Changelog{
		ID:            1,
		Version:       2,
		Username:      "abc",
		Manifest:      "xxx",
		Config:        "config",
		ConfigChanged: true,
		ProjectID:     10,
		GitProjectID:  10,
		CreatedAt:     time.Now().Add(15 * time.Minute),
		UpdatedAt:     time.Now().Add(30 * time.Minute),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Now().Add(-10 * time.Second),
			Valid: true,
		},
		Project:    Project{},
		GitProject: GitProject{},
	}
	assert.Equal(t, &types.ChangelogModel{
		Id:            int64(m.ID),
		Version:       int64(m.Version),
		Username:      m.Username,
		Manifest:      m.Manifest,
		Config:        m.Config,
		ConfigChanged: m.ConfigChanged,
		ProjectId:     int64(m.ProjectID),
		GitProjectId:  int64(m.GitProjectID),
		Project:       m.Project.ProtoTransform(),
		GitProject:    m.GitProject.ProtoTransform(),
		Date:          date.ToHumanizeDatetimeString(&m.CreatedAt),
		CreatedAt:     date.ToRFC3339DatetimeString(&m.CreatedAt),
		UpdatedAt:     date.ToRFC3339DatetimeString(&m.UpdatedAt),
		DeletedAt:     date.ToRFC3339DatetimeString(&m.DeletedAt.Time),
	}, m.ProtoTransform())
}
