package models

import (
	"time"

	"github.com/duc-cnzj/mars/v4/internal/utils/date"

	"github.com/duc-cnzj/mars-client/v4/types"
	"gorm.io/gorm"
)

type Event struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Action   uint8  `json:"action" gorm:"type:tinyint;not null;default:0;index;"`
	Username string `json:"username" gorm:"size:255;not null;default:'';comment:用户名称"`
	Message  string `json:"message" gorm:"size:255;not null;default:'';"`

	Old      string `json:"old" gorm:"type:longtext;"`
	New      string `json:"new" gorm:"type:longtext;"`
	Duration string `json:"duration" gorm:"not null;default:''"`

	FileID *int `json:"file_id" gorm:"nullable;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	File *File
}

func (e *Event) ProtoTransform() *types.EventModel {
	var f *types.FileModel
	if e.File != nil {
		f = e.File.ProtoTransform()
	}
	var fID int64
	if e.FileID != nil {
		fID = int64(*e.FileID)
	}
	return &types.EventModel{
		Id:        int64(e.ID),
		Action:    types.EventActionType(e.Action),
		Username:  e.Username,
		Message:   e.Message,
		Old:       e.Old,
		New:       e.New,
		Duration:  e.Duration,
		FileId:    fID,
		File:      f,
		HasDiff:   e.Old != e.New,
		EventAt:   date.ToHumanizeDatetimeString(&e.CreatedAt),
		CreatedAt: date.ToRFC3339DatetimeString(&e.CreatedAt),
		UpdatedAt: date.ToRFC3339DatetimeString(&e.UpdatedAt),
		DeletedAt: date.ToRFC3339DatetimeString(&e.DeletedAt.Time),
	}
}
