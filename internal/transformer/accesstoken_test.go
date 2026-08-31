package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/stretchr/testify/assert"
)

func TestFromAccessToken_NilInput(t *testing.T) {
	var at *biz.AccessToken
	result := transformer.FromAccessToken(at)
	assert.Nil(t, result)
}

func TestFromAccessToken_ValidInput(t *testing.T) {
	now := time.Now()
	exp := now.Add(-time.Hour)
	at := &biz.AccessToken{
		Token:     "testToken",
		Email:     "testEmail",
		ExpiredAt: exp,
		Usage:     "testUsage",
		CreatedAt: now,
		UpdatedAt: now,
	}
	result := transformer.FromAccessToken(at)
	assert.NotNil(t, result)
	assert.Equal(t, "testToken", result.Token)
	assert.Equal(t, "testEmail", result.Email)
	assert.Equal(t, date.ToRFC3339(&exp), result.ExpiredAt)
	assert.Equal(t, "testUsage", result.Usage)
	assert.Equal(t, date.ToRFC3339(&now), result.CreatedAt)
	assert.Equal(t, date.ToRFC3339(&now), result.UpdatedAt)
	assert.False(t, result.IsDeleted)
	assert.True(t, result.IsExpired)
	assert.Empty(t, result.LastUsedAt)
}

func TestFromAccessToken_DeletedToken(t *testing.T) {
	now := time.Now()
	at := &biz.AccessToken{
		Token:     "testToken",
		Email:     "testEmail",
		ExpiredAt: now,
		Usage:     "testUsage",
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &now,
	}
	result := transformer.FromAccessToken(at)
	assert.NotNil(t, result)
	assert.True(t, result.IsDeleted)
}

func TestFromAccessToken_ExpiredToken(t *testing.T) {
	now := time.Now()
	expiredTime := now.Add(-time.Hour)
	at := &biz.AccessToken{
		Token:     "testToken",
		Email:     "testEmail",
		ExpiredAt: expiredTime,
		Usage:     "testUsage",
		CreatedAt: now,
		UpdatedAt: now,
	}
	result := transformer.FromAccessToken(at)
	assert.NotNil(t, result)
	assert.True(t, result.IsExpired)
}

func TestFromAccessToken_LastUsedAt(t *testing.T) {
	now := time.Now()
	lastUsed := now.Add(-time.Hour)
	at := &biz.AccessToken{
		Token:      "testToken",
		Email:      "testEmail",
		ExpiredAt:  now,
		Usage:      "testUsage",
		LastUsedAt: &lastUsed,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	result := transformer.FromAccessToken(at)
	assert.NotNil(t, result)
	assert.Equal(t, date.ToHumanizeDateTime(&lastUsed), result.LastUsedAt)
}

func TestFromAccessToken_NotExpired(t *testing.T) {
	now := time.Now()
	at := &biz.AccessToken{
		Token:     "testToken",
		Email:     "testEmail",
		ExpiredAt: now.Add(time.Hour),
		Usage:     "testUsage",
		CreatedAt: now,
		UpdatedAt: now,
	}
	result := transformer.FromAccessToken(at)
	assert.NotNil(t, result)
	assert.False(t, result.IsExpired)
}

func TestFromAccessToken_NameFromUserInfo(t *testing.T) {
	now := time.Now()
	at := &biz.AccessToken{
		Token:     "testToken",
		Email:     "testEmail",
		ExpiredAt: now.Add(time.Hour),
		Usage:     "testUsage",
		CreatedAt: now,
		UpdatedAt: now,
		UserInfo:  biz.UserInfo{Name: "Display Name", Email: "testEmail"},
	}
	result := transformer.FromAccessToken(at)
	assert.NotNil(t, result)
	assert.Equal(t, "Display Name", result.Name)
}

func TestFromAccessToken_NameFallbackToEmail(t *testing.T) {
	now := time.Now()
	at := &biz.AccessToken{
		Token:     "testToken",
		Email:     "testEmail",
		ExpiredAt: now.Add(time.Hour),
		Usage:     "testUsage",
		CreatedAt: now,
		UpdatedAt: now,
	}
	result := transformer.FromAccessToken(at)
	assert.NotNil(t, result)
	assert.Equal(t, "testEmail", result.Name)
}
