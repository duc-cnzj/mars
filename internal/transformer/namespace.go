package transformer

import (
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
	"github.com/duc-cnzj/mars/v4/internal/utils/serialize"
)

func FromNamespace(ns *ent.Namespace) *types.NamespaceModel {
	if ns == nil {
		return nil
	}
	return &types.NamespaceModel{
		Id:               int64(ns.ID),
		Name:             ns.Name,
		ImagePullSecrets: ns.GetImagePullSecrets(),
		Projects:         serialize.Serialize(ns.Edges.Projects, FromProject),
		CreatedAt:        date.ToRFC3339DatetimeString(&ns.CreatedAt),
		UpdatedAt:        date.ToRFC3339DatetimeString(&ns.UpdatedAt),
		DeletedAt:        date.ToRFC3339DatetimeString(ns.DeletedAt),
	}
}
