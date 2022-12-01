package models

import (
    "database/sql"
    "testing"
    "time"

    "github.com/duc-cnzj/mars/internal/contracts"
    "github.com/duc-cnzj/mars/internal/utils/date"

    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"
)

func TestAccessToken_Expired(t *testing.T) {
    at := AccessToken{
        ExpiredAt: time.Now().Add(10 * time.Second),
    }
    assert.False(t, at.Expired())
    at2 := AccessToken{
        ExpiredAt: time.Now().Add(-10 * time.Second),
    }
    assert.True(t, at2.Expired())
}

func TestAccessToken_ProtoTransform(t *testing.T) {
    at := &AccessToken{
        Token:      "xx",
        Usage:      "usage",
        Email:      "a@a.com",
        ExpiredAt:  time.Now(),
        LastUsedAt: sql.NullTime{},
        CreatedAt:  time.Time{},
        UpdatedAt:  time.Time{},
        DeletedAt:  gorm.DeletedAt{},
    }
    p := at.ProtoTransform()
    assert.Equal(t, p.Token, at.Token)
    assert.Equal(t, p.Usage, at.Usage)
    assert.Equal(t, p.Email, at.Email)
    assert.Equal(t, p.LastUsedAt, "")
    assert.Equal(t, p.ExpiredAt, date.ToRFC3339DatetimeString(&at.ExpiredAt))
    assert.Equal(t, p.CreatedAt, date.ToRFC3339DatetimeString(&at.CreatedAt))
    assert.Equal(t, p.UpdatedAt, date.ToRFC3339DatetimeString(&at.UpdatedAt))
    assert.Equal(t, p.DeletedAt, date.ToRFC3339DatetimeString(&at.DeletedAt.Time))
    at.LastUsedAt = sql.NullTime{
        Time:  time.Now(),
        Valid: true,
    }
    pp := at.ProtoTransform()
    assert.Equal(t, date.ToHumanizeDatetimeString(&at.LastUsedAt.Time), pp.LastUsedAt)
}

func TestNewAccessToken(t *testing.T) {
    token := NewAccessToken("my token", time.Now(), &contracts.UserInfo{Email: "xx@x.com"})
    assert.NotEmpty(t, token.Token)
}

func TestAccessToken_GetUserInfo(t *testing.T) {
    uinfo := &contracts.UserInfo{
        ID:        "x",
        Email:     "xx@x.com",
        Name:      "duc",
        Picture:   "ppp",
        Roles:     []string{"aaa"},
        LogoutUrl: "abc",
    }
    assert.Equal(t, *uinfo, NewAccessToken("my token", time.Now(), uinfo).GetUserInfo())
}
