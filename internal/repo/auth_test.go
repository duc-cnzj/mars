package repo

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthRepo_Login_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data)

	input := &LoginInput{
		Username: "admin",
		Password: "password",
	}

	authsvc.EXPECT().Sign(gomock.Any()).Return(&auth.SignData{Token: "token", ExpiredIn: 3600}, nil).Times(1)

	resp, err := repo.Login(context.Background(), input)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "token", resp.Token)
}

func TestAuthRepo_Login_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data)

	input := &LoginInput{
		Username: "xadmin",
		Password: "wrongpassword",
	}
	data.EXPECT().Config().Return(&config.Config{AdminPassword: "password"}).Times(1)

	resp, err := repo.Login(context.Background(), input)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestAuthRepo_Login_Failure2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data)

	input := &LoginInput{
		Username: "admin",
		Password: "wrongpassword",
	}
	authsvc.EXPECT().Sign(gomock.Any()).Return(nil, errors.New("x"))

	_, err := repo.Login(context.Background(), input)

	assert.Error(t, err)
	s, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, s.Code())
}

func TestAuthRepo_VerifyToken_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data)

	token := "validtoken"

	authsvc.EXPECT().VerifyToken(token).Return(&auth.JwtClaims{UserInfo: &auth.UserInfo{ID: "1"}}, true).Times(1)

	userInfo, err := repo.VerifyToken(context.Background(), token)

	assert.Nil(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, "1", userInfo.ID)
}

func TestAuthRepo_VerifyToken_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data)

	token := "invalidtoken"

	authsvc.EXPECT().VerifyToken(token).Return(nil, false).Times(1)

	userInfo, err := repo.VerifyToken(context.Background(), token)

	assert.NotNil(t, err)
	assert.Nil(t, userInfo)
}
func TestAuthRepo_Settings(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data2 := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data2)

	oidcConfig := data.OidcConfig{}

	data2.EXPECT().OidcConfig().Return(oidcConfig).Times(1)

	config, err := repo.Settings(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, oidcConfig, config)
}

func TestAuthRepo_Sign_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data)

	userInfo := &auth.UserInfo{
		ID:    "1",
		Name:  "name",
		Email: "email@example.com",
	}

	signData := &auth.SignData{
		Token:     "token",
		ExpiredIn: 3600,
	}

	authsvc.EXPECT().Sign(userInfo).Return(signData, nil).Times(1)

	resp, err := repo.Sign(context.Background(), userInfo)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, signData.Token, resp.Token)
	assert.Equal(t, signData.ExpiredIn, resp.ExpiredIn)
	assert.Equal(t, userInfo, resp.UserInfo)
}

func TestAuthRepo_Sign_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authsvc := auth.NewMockAuth(m)
	data := data.NewMockData(m)
	logger := mlog.NewForConfig(nil)

	repo := NewAuthRepo(authsvc, logger, data)

	userInfo := &auth.UserInfo{
		ID:    "1",
		Name:  "name",
		Email: "email@example.com",
	}

	authsvc.EXPECT().Sign(userInfo).Return(nil, errors.New("sign error")).Times(1)

	resp, err := repo.Sign(context.Background(), userInfo)

	assert.NotNil(t, err)
	assert.Nil(t, resp)
}
