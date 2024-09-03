package repo

import (
	"context"

	auth2 "github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginInput struct {
	Username string
	Password string
}

type AuthRepo interface {
	Login(ctx context.Context, input *LoginInput) (*LoginResponse, error)
	VerifyToken(ctx context.Context, token string) (*auth2.UserInfo, error)
	Settings(ctx context.Context) (data.OidcConfig, error)
	Sign(ctx context.Context, input *auth2.UserInfo) (*LoginResponse, error)
}

var _ AuthRepo = (*authRepo)(nil)

type authRepo struct {
	authsvc auth2.Auth
	data    data.Data
	logger  mlog.Logger
}

func NewAuthRepo(authsvc auth2.Auth, logger mlog.Logger, data data.Data) AuthRepo {
	return &authRepo{authsvc: authsvc, logger: logger, data: data}
}

type LoginResponse struct {
	Token     string
	ExpiredIn int64

	UserInfo *auth2.UserInfo
}

var adminUserInfo = &auth2.UserInfo{
	LogoutUrl: "",
	Roles:     []string{schematype.MarsAdmin},
	ID:        "1",
	Name:      "管理员",
	Email:     "1025434218@qq.com",
}

func (a *authRepo) Login(ctx context.Context, input *LoginInput) (*LoginResponse, error) {
	if input.Username != "admin" && a.data.Config().AdminPassword != input.Password {
		return nil, status.Errorf(codes.Unauthenticated, "用户名或密码错误")
	}
	signData, err := a.authsvc.Sign(adminUserInfo)
	if err != nil {
		return nil, ToError(401, err)
	}

	return &LoginResponse{
		Token:     signData.Token,
		ExpiredIn: signData.ExpiredIn,
		UserInfo:  adminUserInfo,
	}, nil
}

func (a *authRepo) VerifyToken(ctx context.Context, token string) (*auth2.UserInfo, error) {
	verifyToken, b := a.authsvc.VerifyToken(token)
	if !b {
		return nil, ToError(401, "token验证失败")
	}
	return verifyToken.UserInfo, nil
}

func (a *authRepo) Settings(ctx context.Context) (data.OidcConfig, error) {
	return a.data.OidcConfig(), nil
}

func (a *authRepo) Sign(ctx context.Context, input *auth2.UserInfo) (*LoginResponse, error) {
	signData, err := a.authsvc.Sign(input)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	return &LoginResponse{
		Token:     signData.Token,
		ExpiredIn: signData.ExpiredIn,
		UserInfo:  input,
	}, nil
}
