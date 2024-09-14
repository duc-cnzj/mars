package repo

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAccessToken_Expired(t *testing.T) {
	at := &AccessToken{}
	assert.True(t, at.Expired())
	at.ExpiredAt = time.Now().Add(time.Hour)
	assert.False(t, at.Expired())
}

func TestNewAccessTokenRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	repo := NewAccessTokenRepo(timer.NewReal(), mlog.NewForConfig(nil), data.NewMockData(m))
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.(*accessTokenRepo).logger)
	assert.NotNil(t, repo.(*accessTokenRepo).data)
	assert.NotNil(t, repo.(*accessTokenRepo).timer)
}

func TestToAccessToken(t *testing.T) {
	assert.Nil(t, ToAccessToken(nil))

	entToken := &ent.AccessToken{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Token:     "testToken",
		Usage:     "testUsage",
		Email:     "test@example.com",
		ExpiredAt: time.Now().Add(time.Hour),
		UserInfo:  schematype.UserInfo{ID: "1"},
	}

	at := ToAccessToken(entToken)

	assert.NotNil(t, at)
	assert.Equal(t, entToken.ID, at.ID)
	assert.Equal(t, entToken.CreatedAt, at.CreatedAt)
	assert.Equal(t, entToken.UpdatedAt, at.UpdatedAt)
	assert.Equal(t, entToken.Token, at.Token)
	assert.Equal(t, entToken.Usage, at.Usage)
	assert.Equal(t, entToken.Email, at.Email)
	assert.Equal(t, entToken.ExpiredAt, at.ExpiredAt)
	assert.Equal(t, entToken.UserInfo, at.UserInfo)
}

func Test_accessTokenRepo_Grant(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	grant, err := repo.Grant(context.TODO(), &GrantAccessTokenInput{
		ExpireSeconds: 10,
		Usage:         "aa",
		User: &auth.UserInfo{
			ID:    "1",
			Email: "1@q.com",
			Name:  "duc",
		},
	})
	assert.Nil(t, err)

	assert.Equal(t, "aa", grant.Usage)
	assert.Equal(t, "1", grant.UserInfo.ID)
	assert.False(t, grant.Expired())
}

func Test_accessTokenRepo_Lease_Success(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	grant, _ := repo.Grant(context.TODO(), &GrantAccessTokenInput{
		ExpireSeconds: 10,
		Usage:         "aa",
		User: &auth.UserInfo{
			ID:    "1",
			Email: "1@q.com",
			Name:  "duc",
		},
	})

	lease, err := repo.Lease(context.TODO(), grant.Token, 20)
	assert.Nil(t, err)
	assert.Equal(t, grant.Token, lease.Token)
	assert.False(t, lease.Expired())
}

func Test_accessTokenRepo_Lease_TokenNotFound(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)

	_, err := repo.Lease(context.TODO(), "nonexistentToken", 20)
	assert.NotNil(t, err)
}

func Test_accessTokenRepo_Lease_TokenExpired(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	grant, _ := repo.Grant(context.TODO(), &GrantAccessTokenInput{
		ExpireSeconds: -10, // Expired token
		Usage:         "aa",
		User: &auth.UserInfo{
			ID:    "1",
			Email: "1@q.com",
			Name:  "duc",
		},
	})

	_, err := repo.Lease(context.TODO(), grant.Token, 20)
	assert.NotNil(t, err)
}

func Test_accessTokenRepo_List_WithSoftDelete(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	db.AccessToken.Create().
		SetToken("testToken").
		SetUsage("testUsage").
		SetEmail("").
		SetDeletedAt(time.Now()).
		SetExpiredAt(time.Now().Add(time.Hour)).
		SaveX(context.TODO())
	res, _, err := repo.List(context.TODO(), &ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		WithSoftDelete: true,
		Email:          "",
	})
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "testToken", res[0].Token)
}

func Test_accessTokenRepo_List_WithoutSoftDelete(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	db.AccessToken.Create().
		SetToken("testToken").
		SetUsage("testUsage").
		SetEmail("").
		SetDeletedAt(time.Now()).
		SetExpiredAt(time.Now().Add(time.Hour)).
		SaveX(context.TODO())
	res, _, err := repo.List(context.TODO(), &ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		WithSoftDelete: false,
		Email:          "",
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(res))
}

func Test_accessTokenRepo_List_WithEmail(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	db.AccessToken.Create().
		SetToken("testToken").
		SetUsage("testUsage").
		SetEmail("test@example.com").
		SetExpiredAt(time.Now().Add(time.Hour)).
		SaveX(context.TODO())
	db.AccessToken.Create().
		SetToken("texstToken").
		SetUsage("testUsage").
		SetEmail("1xxtest@example.com").
		SetExpiredAt(time.Now().Add(time.Hour)).
		SaveX(context.TODO())
	res, _, err := repo.List(context.TODO(), &ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		WithSoftDelete: false,
		Email:          "test@example.com",
	})
	assert.Nil(t, err)
	assert.Len(t, res, 1)
}

func Test_accessTokenRepo_List_WithoutEmail(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	db.AccessToken.Create().
		SetToken("testToken").
		SetUsage("testUsage").
		SetEmail("test@example.com").
		SetExpiredAt(time.Now().Add(time.Hour)).
		SaveX(context.TODO())
	db.AccessToken.Create().
		SetToken("txestToken").
		SetUsage("testUsage").
		SetEmail("xxtest@example.com").
		SetExpiredAt(time.Now().Add(time.Hour)).
		SaveX(context.TODO())
	res, pag, err := repo.List(context.TODO(), &ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		WithSoftDelete: false,
		Email:          "",
	})
	assert.Nil(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, int32(1), pag.Page)
	assert.Equal(t, int32(10), pag.PageSize)
	assert.Equal(t, int32(2), pag.Count)
}

func Test_accessTokenRepo_Revoke_Success(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)
	grant, _ := repo.Grant(context.TODO(), &GrantAccessTokenInput{
		ExpireSeconds: 10,
		Usage:         "aa",
		User: &auth.UserInfo{
			ID:    "1",
			Email: "1@q.com",
			Name:  "duc",
		},
	})

	err := repo.Revoke(context.TODO(), grant.Token)
	assert.Nil(t, err)

	_, err = repo.Lease(context.TODO(), grant.Token, 20)
	assert.NotNil(t, err)
}

func Test_accessTokenRepo_Revoke_TokenNotFound(t *testing.T) {
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewAccessTokenRepo(
		timer.NewReal(),
		mlog.NewForConfig(nil),
		data.NewDataImpl(&data.NewDataParams{
			DB: db,
		}),
	)

	err := repo.Revoke(context.TODO(), "nonexistentToken")
	assert.Nil(t, err)
}
