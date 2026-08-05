package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
)

// FromMember transform biz.Member to proto MemberModel.
func FromMember(member *biz.Member) *types.MemberModel {
	if member == nil {
		return nil
	}
	return &types.MemberModel{
		Id:    int32(member.ID),
		Email: member.Email,
	}
}
