package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
)

// FromAccessToken transform biz.AccessToken to proto AccessTokenModel.
func FromAccessToken(at *biz.AccessToken) *types.AccessTokenModel {
	if at == nil {
		return nil
	}
	return &types.AccessTokenModel{
		Token:      at.Token,
		Email:      at.Email,
		ExpiredAt:  date.ToRFC3339(&at.ExpiredAt),
		Usage:      at.Usage,
		LastUsedAt: date.ToHumanizeDateTime(at.LastUsedAt),
		IsDeleted:  at.DeletedAt != nil,
		IsExpired:  biz.IsAccessTokenExpired(at),
		CreatedAt:  date.ToRFC3339(&at.CreatedAt),
		UpdatedAt:  date.ToRFC3339(&at.UpdatedAt),
		DeletedAt:  date.ToRFC3339(at.DeletedAt),
	}
}
