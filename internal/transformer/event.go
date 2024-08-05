package transformer

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
)

func FromEvent(e *repo.Event) *types.EventModel {
	if e == nil {
		return nil
	}
	var fID int64
	if e.FileID != nil {
		fID = int64(*e.FileID)
	}
	return &types.EventModel{
		Id:        int32(e.ID),
		Action:    e.Action,
		Username:  e.Username,
		Message:   e.Message,
		Old:       e.Old,
		New:       e.New,
		Duration:  e.Duration,
		FileId:    int32(fID),
		File:      FromFile(e.File),
		HasDiff:   e.Old != e.New,
		EventAt:   date.ToHumanizeDatetimeString(&e.CreatedAt),
		CreatedAt: date.ToRFC3339DatetimeString(&e.CreatedAt),
		UpdatedAt: date.ToRFC3339DatetimeString(&e.UpdatedAt),
		DeletedAt: date.ToRFC3339DatetimeString(e.DeletedAt),
	}
}
