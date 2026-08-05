package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/dustin/go-humanize"
)

// FromFile transform biz.File to proto FileModel.
func FromFile(f *biz.File) *types.FileModel {
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
		CreatedAt:      date.ToRFC3339(&f.CreatedAt),
		UpdatedAt:      date.ToRFC3339(&f.UpdatedAt),
		DeletedAt:      date.ToRFC3339(f.DeletedAt),
	}
}
