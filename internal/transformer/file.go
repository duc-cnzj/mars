package transformer

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
	"github.com/dustin/go-humanize"
)

func FromFile(f *ent.File) *types.FileModel {
	if f == nil {
		return nil
	}
	return &types.FileModel{
		Id:             int64(f.ID),
		Path:           f.Path,
		Size:           int64(f.Size),
		Username:       f.Username,
		Namespace:      f.Namespace,
		Pod:            f.Pod,
		Container:      f.Container,
		Container_Path: f.ContainerPath,
		HumanizeSize:   humanize.Bytes(f.Size),
		CreatedAt:      date.ToRFC3339DatetimeString(&f.CreatedAt),
		UpdatedAt:      date.ToRFC3339DatetimeString(&f.UpdatedAt),
		DeletedAt:      date.ToRFC3339DatetimeString(f.DeletedAt),
	}
}
