package services

import (
	"context"
	"fmt"
	"sort"

	apiauth "github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/spf13/cast"
)

var _ apiauth.AuthServer = (*authSvc)(nil)

// authSvc 是 apiauth.AuthServer 的 gRPC 实现：提供登录、用户信息、登录设置
// 与 OIDC 授权码换取，审计事件经 eventBiz 落库，登录成功后经 userBiz 同步用户投影，
// 由 NewAuthSvc 构造。
type authSvc struct {
	apiauth.UnimplementedAuthServer

	logger   mlog.Logger
	authBiz  biz.AuthBiz
	eventBiz biz.EventBiz
	userBiz  biz.UserBiz
}

// AuthSvcDeps 收口 NewAuthSvc 的构造依赖，由 wire 按字段注入。
type AuthSvcDeps struct {
	EventBiz biz.EventBiz
	Logger   mlog.Logger
	AuthBiz  biz.AuthBiz
	UserBiz  biz.UserBiz
}

// NewAuthSvc 收口认证服务的构造依赖，由 wire 按字段注入。
func NewAuthSvc(deps AuthSvcDeps) apiauth.AuthServer {
	return &authSvc{
		logger:   deps.Logger.WithModule("services/auth"),
		eventBiz: deps.EventBiz,
		authBiz:  deps.AuthBiz,
		userBiz:  deps.UserBiz,
	}
}

// Login 处理用户名密码登录：校验凭证后签发登录凭证，并落登录审计日志。
func (a *authSvc) Login(ctx context.Context, request *apiauth.LoginRequest) (*apiauth.LoginResponse, error) {
	loginResp, err := a.authBiz.Login(ctx, &biz.LoginInput{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}

	a.eventBiz.AuditLog(
		types.EventActionType_Login,
		loginResp.UserInfo.Name,
		loginResp.UserInfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", loginResp.UserInfo.Name, loginResp.UserInfo.Email),
	)

	return &apiauth.LoginResponse{
		Token:     loginResp.Token,
		ExpiresIn: loginResp.ExpiredIn,
	}, nil
}

// Info 返回当前登录用户信息：用户由鉴权拦截器（middlewares.Login*ServerInterceptor）
// 统一验签后经 biz.SetUser 注入 ctx，本方法只做「取 ctx 用户 → 映射响应」，不再自行验签
// （消除与拦截器的双重验签）。取用户用 biz.MustGetUser（与 AccessBiz/services 全仓惯例
// 一致）：双链路（gRPC 拦截器 + HTTP gateway 经 RegisterAuthHandlerFromEndpoint 回环 dial
// 到同一 gRPC server）必注入用户，ctx 无用户即编程错误（panic，由 grpc_recovery 兜底）。
func (a *authSvc) Info(ctx context.Context, req *apiauth.InfoRequest) (*apiauth.InfoResponse, error) {
	user := biz.MustGetUser(ctx)
	return &apiauth.InfoResponse{
		Id:        cast.ToInt32(user.ID),
		Avatar:    user.Picture,
		Name:      user.Name,
		Email:     user.Email,
		LogoutUrl: user.LogoutUrl,
		Roles:     user.Roles,
	}, nil
}

// Settings 返回可用的 OIDC 登录方式：为每个 provider 生成一次性 state 拼出
// 授权码 URL，按名字排序后返回，供前端渲染登录页。
func (a *authSvc) Settings(ctx context.Context, request *apiauth.SettingsRequest) (*apiauth.SettingsResponse, error) {
	settings, err := a.authBiz.Settings(ctx)
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}
	var items = make([]*apiauth.SettingsResponse_OidcSetting, 0, len(settings))
	for name, setting := range settings {
		state := rand.String(32)

		items = append(items, &apiauth.SettingsResponse_OidcSetting{
			Enabled:            true,
			Name:               name,
			Url:                setting.Config.AuthCodeURL(state),
			EndSessionEndpoint: setting.EndSessionEndpoint,
			State:              state,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return &apiauth.SettingsResponse{Items: items}, nil
}

// Exchange 用 OIDC 授权码换发登录凭证：换发编排（遍历 provider/验签/claims 解码）
// 已下沉 biz.AuthBiz.Exchange，这里只做 transport 份内事——签名、审计与响应映射。
func (a *authSvc) Exchange(ctx context.Context, request *apiauth.ExchangeRequest) (*apiauth.ExchangeResponse, error) {
	userinfo, err := a.authBiz.Exchange(ctx, request.Code)
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}

	data, err := a.authBiz.Sign(ctx, userinfo)
	if err != nil {
		// 与其他 service 一致：直接返回原始错误，不额外包装 codes.Internal（避免丢失原始错误码）。
		return nil, logError(ctx, a.logger, err)
	}
	a.eventBiz.AuditLogWithRequest(
		types.EventActionType_Login,
		userinfo.Name,
		userinfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email),
		request,
	)
	// 登录成功即同步用户投影：不存在则创建、存在则推进最近登录。投影写库失败仅记日志
	// 不阻断登录——OIDC 凭证已校验、登录事件已落库，users 只是管理投影，该用户下次
	// 登录会由 SyncLoginUser 自动补回。
	if err := a.userBiz.SyncLoginUser(ctx, userinfo.Email, userinfo.Name); err != nil {
		a.logger.ErrorCtx(ctx, err)
	}

	return &apiauth.ExchangeResponse{
		Token:     data.Token,
		ExpiresIn: data.ExpiredIn,
	}, nil
}
