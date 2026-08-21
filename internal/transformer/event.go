package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/samber/lo"
)

// FromEvent 把 biz.Event 转换为 proto EventModel。
func FromEvent(e *biz.Event) *types.EventModel {
	if e == nil {
		return nil
	}

	return &types.EventModel{
		Id:            int32(e.ID),
		Action:        e.Action,
		Username:      e.Username,
		Message:       e.Message,
		Old:           e.Old,
		New:           e.New,
		Duration:      e.Duration,
		FileId:        int32(lo.FromPtr(e.FileID)),
		File:          FromFile(e.File),
		HasDiff:       e.HasDiff,
		EventAt:       date.ToHumanizeDateTime(&e.CreatedAt),
		CreatedAt:     date.ToRFC3339(&e.CreatedAt),
		UpdatedAt:     date.ToRFC3339(&e.UpdatedAt),
		DeletedAt:     date.ToRFC3339(e.DeletedAt),
		OperatorEmail: e.OperatorEmail,
	}
}
