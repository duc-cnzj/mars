package transformer

import (
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/date"
	"github.com/dustin/go-humanize"
)

func FromFile(f *repo.File) *types.FileModel {
	if f == nil {
		return nil
	}
	return &types.FileModel{
		Id:             int32(f.ID),
		Path:           f.Path,
		Size:           int32(f.Size),
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
