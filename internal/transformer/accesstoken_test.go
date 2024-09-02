package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/date"
	"github.com/stretchr/testify/assert"
)

func TestFromAccessToken_NilInput(t *testing.T) {
	var at *repo.AccessToken
	result := transformer.FromAccessToken(at)
	assert.Nil(t, result)
}

func TestFromAccessToken_ValidInput(t *testing.T) {
	now := time.Now()
	exp := now.Add(-time.Hour)
	at := &repo.AccessToken{
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
	assert.Equal(t, date.ToRFC3339DatetimeString(&exp), result.ExpiredAt)
	assert.Equal(t, "testUsage", result.Usage)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.CreatedAt)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.UpdatedAt)
	assert.False(t, result.IsDeleted)
	assert.True(t, result.IsExpired)
}

func TestFromAccessToken_DeletedToken(t *testing.T) {
	now := time.Now()
	at := &repo.AccessToken{
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
	at := &repo.AccessToken{
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
