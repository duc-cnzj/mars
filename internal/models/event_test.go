package models

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEvent_ProtoTransform(t *testing.T) {
	fID := 10
	m := Event{
		ID:        1,
		Action:    2,
		Username:  "duc",
		Message:   "ok",
		Old:       "old",
		New:       "new",
		Duration:  "15s ago",
		FileID:    &fID,
		CreatedAt: time.Now().Add(15 * time.Minute),
		UpdatedAt: time.Now().Add(30 * time.Minute),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Now().Add(-10 * time.Second),
			Valid: true,
		},
		File: nil,
	}
	assert.Equal(t, &types.EventModel{
		Id:        int64(m.ID),
		Action:    types.EventActionType(m.Action),
		Username:  m.Username,
		Message:   m.Message,
		Old:       m.Old,
		New:       m.New,
		HasDiff:   m.Old != m.New,
		Duration:  m.Duration,
		FileId:    int64(fID),
		File:      nil,
		EventAt:   date.ToHumanizeDatetimeString(&m.CreatedAt),
		CreatedAt: date.ToRFC3339DatetimeString(&m.CreatedAt),
		UpdatedAt: date.ToRFC3339DatetimeString(&m.UpdatedAt),
		DeletedAt: date.ToRFC3339DatetimeString(&m.DeletedAt.Time),
	}, m.ProtoTransform())
	m = Event{
		ID:        1,
		Action:    2,
		Username:  "duc",
		Message:   "ok",
		Old:       "old",
		New:       "new",
		Duration:  "15s ago",
		FileID:    &fID,
		CreatedAt: time.Now().Add(15 * time.Minute),
		UpdatedAt: time.Now().Add(30 * time.Minute),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Now().Add(-10 * time.Second),
			Valid: true,
		},
		File: &File{
			ID:            fID,
			Path:          "/filepath",
			Size:          1000,
			Username:      "duc",
			Namespace:     "devops",
			Pod:           "pod",
			Container:     "container",
			ContainerPath: "path",
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
			DeletedAt:     gorm.DeletedAt{},
		},
	}
	assert.Equal(t, &types.EventModel{
		Id:        int64(m.ID),
		Action:    types.EventActionType(m.Action),
		Username:  m.Username,
		Message:   m.Message,
		Old:       m.Old,
		New:       m.New,
		HasDiff:   m.Old != m.New,
		Duration:  m.Duration,
		FileId:    int64(fID),
		File:      m.File.ProtoTransform(),
		EventAt:   date.ToHumanizeDatetimeString(&m.CreatedAt),
		CreatedAt: date.ToRFC3339DatetimeString(&m.CreatedAt),
		UpdatedAt: date.ToRFC3339DatetimeString(&m.UpdatedAt),
		DeletedAt: date.ToRFC3339DatetimeString(&m.DeletedAt.Time),
	}, m.ProtoTransform())
}
