package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/auth"
	auth2 "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestAuthSvc_AuthFuncOverride(t *testing.T) {
	svc := NewAuthSvc(nil, nil, "")
	_, err := svc.AuthFuncOverride(context.TODO(), "")
	assert.Nil(t, err)
}

func TestAuthSvc_Info(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	authSvc := auth2.NewAuth(key, key.Public().(*rsa.PublicKey))
	sign, _ := authSvc.Sign(contracts.UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{"user"},
		OpenIDClaims: contracts.OpenIDClaims{
			Name:    "duc",
			Sub:     "2022",
			Email:   "1025434218@qq.com",
			Picture: "avatar.png",
		},
	})
	svc := NewAuthSvc(authSvc, nil, "")
	ctx := context.TODO()
	info, err := svc.Info(ctx, &auth.InfoRequest{})
	assert.Nil(t, info)
	assert.Error(t, err)
	md := metadata.New(map[string]string{"Authorization": sign.Token})
	ctx = metadata.NewIncomingContext(ctx, md)
	info, err = svc.Info(ctx, &auth.InfoRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "duc", info.Name)
	assert.Equal(t, "1025434218@qq.com", info.Id)
	assert.Equal(t, "xxx", info.LogoutUrl)
	assert.Equal(t, "avatar.png", info.Avatar)
	assert.Equal(t, []string{"user"}, info.Roles)
}

func TestAuthSvc_Login(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	authSvc := auth2.NewAuth(key, key.Public().(*rsa.PublicKey))
	svc := NewAuthSvc(authSvc, nil, "admin")
	login, err := svc.Login(context.TODO(), &auth.LoginRequest{
		Username: "admin",
		Password: "admin",
	})
	assert.Nil(t, err)
	token, ok := authSvc.VerifyToken(login.Token)
	assert.Equal(t, "admin@mars.com", token.StandardClaims.Subject)
	assert.True(t, ok)
	assert.Equal(t, "管理员", token.UserInfo.Name)
}

func TestNewAuthSvc(t *testing.T) {
	assert.Implements(t, (*auth.AuthServer)(nil), NewAuthSvc(nil, nil, ""))
}
