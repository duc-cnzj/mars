package transformer

import (
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/repo"
)

func FromMember(member *repo.Member) *types.MemberModel {
	if member == nil {
		return nil
	}
	return &types.MemberModel{
		Id:    int32(member.ID),
		Email: member.Email,
	}
}
