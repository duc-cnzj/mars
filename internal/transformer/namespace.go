package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
)

// FromNamespace 把 biz.Namespace 转换为 proto NamespaceModel。
func FromNamespace(ns *biz.Namespace) *types.NamespaceModel {
	if ns == nil {
		return nil
	}
	return &types.NamespaceModel{
		Id:           int32(ns.ID),
		Name:         ns.Name,
		Projects:     slice.Map(ns.Projects, FromProject),
		Description:  ns.Description,
		Members:      slice.Map(ns.Members, FromMember),
		Private:      ns.Private,
		CreatorEmail: ns.CreatorEmail,
		CreatedAt:    date.ToRFC3339(&ns.CreatedAt),
		UpdatedAt:    date.ToRFC3339(&ns.UpdatedAt),
		DeletedAt:    date.ToRFC3339(ns.DeletedAt),
	}
}
