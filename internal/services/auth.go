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
	"google.golang.org/grpc/metadata"
)

var _ apiauth.AuthServer = (*authSvc)(nil)

// authSvc 是 apiauth.AuthServer 的 gRPC 实现：提供登录、用户信息、登录设置
// 与 OIDC 授权码换取，审计事件经 eventBiz 落库，由 NewAuthSvc 构造。
type authSvc struct {
	apiauth.UnimplementedAuthServer

	logger   mlog.Logger
	authBiz  biz.AuthBiz
	eventBiz biz.EventBiz
}

// AuthSvcDeps 收口 NewAuthSvc 的构造依赖，由 wire 按字段注入。
type AuthSvcDeps struct {
	EventBiz biz.EventBiz
	Logger   mlog.Logger
	AuthBiz  biz.AuthBiz
}

// NewAuthSvc 收口认证服务的构造依赖，由 wire 按字段注入。
func NewAuthSvc(deps AuthSvcDeps) apiauth.AuthServer {
	return &authSvc{
		logger:   deps.Logger.WithModule("services/auth"),
		eventBiz: deps.EventBiz,
		authBiz:  deps.AuthBiz,
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
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", loginResp.UserInfo.Name, loginResp.UserInfo.Email),
	)

	return &apiauth.LoginResponse{
		Token:     loginResp.Token,
		ExpiresIn: loginResp.ExpiredIn,
	}, nil
}

// Info 返回当前登录用户信息：校验请求头 Authorization 中的 token，
// 有效则返回用户资料；未携带 token 或校验失败统一返回 Unauthenticated。
func (a *authSvc) Info(ctx context.Context, req *apiauth.InfoRequest) (*apiauth.InfoResponse, error) {
	incomingContext, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tokenSlice := incomingContext.Get("Authorization")
		if len(tokenSlice) == 1 {
			c, err := a.authBiz.VerifyToken(ctx, tokenSlice[0])
			if err == nil {
				return &apiauth.InfoResponse{
					Id:        cast.ToInt32(c.ID),
					Avatar:    c.Picture,
					Name:      c.Name,
					Email:     c.Email,
					LogoutUrl: c.LogoutUrl,
					Roles:     c.Roles,
				}, nil
			}
			// 带了 token 但校验失败：记录根因（过期/篡改/用户被删），
			// 便于区分"没带 token"和"token 无效"两种未授权场景。
			a.logger.DebugCtx(ctx, "auth info: verify token failed", err)
		}
	}

	a.logger.WarningCtx(ctx, "auth info: unauthorized")
	// 无凭证/凭证无效的兜底返回 Unauthenticated——状态码由 biz 工厂（ToError 401）构造，
	// transport 不再散落 status 构造（协议映射收口 biz）。
	return nil, biz.ToError(401, "Unauthenticated.")
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
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email),
		request,
	)

	return &apiauth.ExchangeResponse{
		Token:     data.Token,
		ExpiresIn: data.ExpiredIn,
	}, nil
}
