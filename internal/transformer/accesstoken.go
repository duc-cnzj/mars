package transformer

import (
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
)

// FromAccessToken 把 biz.AccessToken 转换为 proto AccessTokenModel。
func FromAccessToken(at *biz.AccessToken) *types.AccessTokenModel {
	if at == nil {
		return nil
	}
	// name 取签发时快照的显示名（OIDC name）；历史令牌 user_info 缺失/为空时回退 email，
	// 保证创建人字段恒非空展示（前端也可再兜底，但展示语义在边界统一收敛）。
	name := at.Email
	if at.UserInfo.Name != "" {
		name = at.UserInfo.Name
	}
	return &types.AccessTokenModel{
		Token:      at.Token,
		Email:      at.Email,
		Name:       name,
		ExpiredAt:  date.ToRFC3339(&at.ExpiredAt),
		Usage:      at.Usage,
		LastUsedAt: date.ToHumanizeDateTime(at.LastUsedAt),
		IsDeleted:  at.DeletedAt != nil,
		IsExpired:  at.IsExpired(time.Now()),
		CreatedAt:  date.ToRFC3339(&at.CreatedAt),
		UpdatedAt:  date.ToRFC3339(&at.UpdatedAt),
		DeletedAt:  date.ToRFC3339(at.DeletedAt),
	}
}
