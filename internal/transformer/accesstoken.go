package transformer

import (
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/date"
)

// FromAccessToken transform to proto model.
func FromAccessToken(at *repo.AccessToken) *types.AccessTokenModel {
	if at == nil {
		return nil
	}
	return &types.AccessTokenModel{
		Token:      at.Token,
		Email:      at.Email,
		ExpiredAt:  date.ToRFC3339DatetimeString(&at.ExpiredAt),
		Usage:      at.Usage,
		LastUsedAt: date.ToHumanizeDatetimeString(at.LastUsedAt),
		IsDeleted:  at.DeletedAt != nil,
		IsExpired:  at.Expired(),
		CreatedAt:  date.ToRFC3339DatetimeString(&at.CreatedAt),
		UpdatedAt:  date.ToRFC3339DatetimeString(&at.UpdatedAt),
		DeletedAt:  date.ToRFC3339DatetimeString(at.DeletedAt),
	}
}
