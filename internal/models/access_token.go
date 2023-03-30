package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
)

type AccessToken struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Token string `json:"token" gorm:"unique;size:255;not null;"`

	Usage     string    `json:"usage" gorm:"size:50;"`
	Email     string    `json:"email" gorm:"index;not null;default:'';"`
	ExpiredAt time.Time `json:"expired_at"`

	LastUsedAt sql.NullTime `json:"last_used_at"`

	UserInfo string `json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func NewAccessToken(usage string, expiredAt time.Time, user *contracts.UserInfo) *AccessToken {
	return &AccessToken{Usage: usage, Email: user.Email, ExpiredAt: expiredAt, Token: uuid.New().String(), UserInfo: user.Json()}
}

// Expired token is expired.
func (at *AccessToken) Expired() bool {
	return time.Now().After(at.ExpiredAt)
}

// GetUserInfo return token userinfo.
func (at *AccessToken) GetUserInfo() contracts.UserInfo {
	var info contracts.UserInfo
	json.Unmarshal([]byte(at.UserInfo), &info)
	return info
}

func (at *AccessToken) ProtoTransform() *types.AccessTokenModel {
	var lastUsedAt = ""
	if at.LastUsedAt.Valid {
		lastUsedAt = date.ToHumanizeDatetimeString(&at.LastUsedAt.Time)
	}
	return &types.AccessTokenModel{
		Token:      at.Token,
		Email:      at.Email,
		ExpiredAt:  date.ToRFC3339DatetimeString(&at.ExpiredAt),
		Usage:      at.Usage,
		LastUsedAt: lastUsedAt,
		IsDeleted:  at.DeletedAt.Valid,
		IsExpired:  at.Expired(),
		CreatedAt:  date.ToRFC3339DatetimeString(&at.CreatedAt),
		UpdatedAt:  date.ToRFC3339DatetimeString(&at.UpdatedAt),
		DeletedAt:  date.ToRFC3339DatetimeString(&at.DeletedAt.Time),
	}
}
