package transformer

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
)

func FromNamespace(ns *repo.Namespace) *types.NamespaceModel {
	if ns == nil {
		return nil
	}
	return &types.NamespaceModel{
		Id:          int32(ns.ID),
		Name:        ns.Name,
		Projects:    serialize.Serialize(ns.Projects, FromProject),
		Description: ns.Description,
		CreatedAt:   date.ToRFC3339DatetimeString(&ns.CreatedAt),
		UpdatedAt:   date.ToRFC3339DatetimeString(&ns.UpdatedAt),
		DeletedAt:   date.ToRFC3339DatetimeString(ns.DeletedAt),
	}
}
