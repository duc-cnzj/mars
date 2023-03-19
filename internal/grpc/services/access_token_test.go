package services

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/token"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccessToken_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	db.AutoMigrate(&models.AccessToken{})
	uinfo := &contracts.UserInfo{Email: "admin@admin.com"}
	var tokens = []*models.AccessToken{
		models.NewAccessToken("token 1", time.Now(), uinfo),
		models.NewAccessToken("token 2", time.Now(), uinfo),
		models.NewAccessToken("token 3", time.Now(), &contracts.UserInfo{Email: "user@user.com"}),
	}
	for _, accessToken := range tokens {
		assert.Nil(t, db.Create(accessToken).Error)
	}
	ctx := auth.SetUser(context.TODO(), uinfo)
	list, err := (&AccessToken{}).List(ctx, &token.ListRequest{
		Page:     1,
		PageSize: 10,
	})
	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)
	assert.Equal(t, int64(2), list.Count)
	assert.Equal(t, "token 2", list.Items[0].Usage)
	assert.Equal(t, "token 1", list.Items[1].Usage)
	db.Where("`token`= ?", list.Items[0].Token).Delete(&models.AccessToken{})
	list, _ = (&AccessToken{}).List(ctx, &token.ListRequest{
		Page:     1,
		PageSize: 10,
	})
	assert.Len(t, list.Items, 2)
	assert.Equal(t, int64(2), list.Count)
	assert.Equal(t, "token 2", list.Items[0].Usage)
	assert.True(t, list.Items[0].IsDeleted)
	assert.Equal(t, "token 1", list.Items[1].Usage)
}

var testTime, _ = time.Parse("2006-01-02 15:04:05", "2022-11-30 00:00:00")

func TestAccessToken_Grant(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	db.AutoMigrate(&models.AccessToken{})
	uinfo := &contracts.UserInfo{Email: "admin@admin.com"}

	ctx := auth.SetUser(context.TODO(), uinfo)
	testutil.AssertAuditLogFired(m, app)
	grant, err := (&AccessToken{nowFunc: func() time.Time { return testTime }}).Grant(ctx, &token.GrantRequest{
		ExpireSeconds: 5,
		Usage:         "my token",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, grant.Token.Token)
	assert.Equal(t, uinfo.Email, grant.Token.Email)
	ti := testTime.Add(5 * time.Second)
	assert.Equal(t, date.ToRFC3339DatetimeString(&ti), grant.Token.ExpiredAt)
}

func TestAccessToken_Lease(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	uinfo := &contracts.UserInfo{Email: "admin@admin.com"}
	ctx := auth.SetUser(context.TODO(), uinfo)

	_, err := (&AccessToken{}).Lease(ctx, &token.LeaseRequest{
		Token:         "token-not-exists",
		ExpireSeconds: 100,
	})
	e, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, e.Code())
	db.AutoMigrate(&models.AccessToken{})

	at := models.NewAccessToken("expired", time.Now().Add(-10*time.Second), uinfo)
	db.Create(at)

	_, err = (&AccessToken{nowFunc: func() time.Time { return testTime }}).Lease(ctx, &token.LeaseRequest{
		Token:         at.Token,
		ExpireSeconds: 100,
	})
	fromError, _ := status.FromError(err)

	assert.Equal(t, "token 已经过期", fromError.Message())
	assert.Equal(t, codes.Aborted, fromError.Code())
	n := time.Now().Add(5 * time.Second)
	nt := models.NewAccessToken("my token", n, uinfo)
	db.Create(nt)
	testutil.AssertAuditLogFired(m, app)
	lease, _ := (&AccessToken{}).Lease(ctx, &token.LeaseRequest{
		Token:         nt.Token,
		ExpireSeconds: 100,
	})
	tti := n.Add(100 * time.Second)
	assert.Equal(t, date.ToRFC3339DatetimeString(&tti), lease.Token.ExpiredAt)

	nt2 := models.NewAccessToken("user token", time.Now().Add(10*time.Second), &contracts.UserInfo{Email: "user@user.com"})
	db.Create(nt2)
	_, err = (&AccessToken{}).Lease(ctx, &token.LeaseRequest{
		Token:         nt2.Token,
		ExpireSeconds: 100,
	})
	assert.Error(t, err)
	s, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, s.Code())
	assert.Equal(t, "token not found", s.Message())
}

func TestAccessToken_Revoke(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	db.AutoMigrate(&models.AccessToken{})
	uinfo := &contracts.UserInfo{Email: "admin@admin.com"}

	ctx := auth.SetUser(context.TODO(), uinfo)
	dispatcher := testutil.AssertAuditLogFired(m, app)
	accessToken := models.NewAccessToken("my token", time.Now(), uinfo)
	db.Create(accessToken)
	_, err := (&AccessToken{}).Revoke(ctx, &token.RevokeRequest{
		Token: accessToken.Token,
	})
	assert.Nil(t, err)
	var at models.AccessToken
	db.Unscoped().Where("`email` = ?", uinfo.Email).First(&at)
	assert.True(t, at.DeletedAt.Valid)

	userAccessToken := models.NewAccessToken("my token", time.Now(), &contracts.UserInfo{Email: "user@user.com"})
	db.Create(userAccessToken)
	dispatcher.EXPECT().Dispatch(contracts.Event("audit_log"), gomock.Any()).Times(1)
	(&AccessToken{}).Revoke(ctx, &token.RevokeRequest{
		Token: userAccessToken.Token,
	})

	var at2 models.AccessToken
	db.Unscoped().Where("`token` = ?", at2.Token).First(&at2)
	assert.False(t, at2.DeletedAt.Valid)
}
