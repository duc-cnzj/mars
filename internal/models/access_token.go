package models

import (
	"encoding/json"
	"github.com/duc-cnzj/mars/internal/contracts"
	"time"

	"github.com/google/uuid"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/utils/date"

	"gorm.io/gorm"
)

type AccessToken struct {
	Token string `json:"token" gorm:"size:255;not null;primaryKey;"`

	Usage     string    `json:"usage" gorm:"size:50;"`
	Email     string    `json:"email" gorm:"index;not null;default:'';"`
	ExpiredAt time.Time `json:"expired_at"`

	LastUsedAt time.Time `json:"last_used_at"`

	UserInfo string `json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func NewAccessToken(usage string, expiredAt time.Time, user *contracts.UserInfo) *AccessToken {
	return &AccessToken{Usage: usage, Email: user.Email, ExpiredAt: expiredAt, Token: uuid.New().String(), UserInfo: user.Json()}
}

func (at *AccessToken) Expired() bool {
	return time.Now().After(at.ExpiredAt)
}

func (at *AccessToken) GetUserInfo() contracts.UserInfo {
	var info contracts.UserInfo
	json.Unmarshal([]byte(at.UserInfo), &info)
	return info
}

func (at *AccessToken) ProtoTransform() *types.AccessTokenModel {
	return &types.AccessTokenModel{
		Token:     at.Token,
		Email:     at.Email,
		ExpiredAt: date.ToRFC3339DatetimeString(&at.ExpiredAt),
		Usage:     at.Usage,
		CreatedAt: date.ToRFC3339DatetimeString(&at.CreatedAt),
		UpdatedAt: date.ToRFC3339DatetimeString(&at.UpdatedAt),
		DeletedAt: date.ToRFC3339DatetimeString(&at.DeletedAt.Time),
	}
}
