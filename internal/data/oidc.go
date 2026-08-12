package data

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"golang.org/x/oauth2"
)

// InitOidcProvider 启动期装配 OIDC provider：遍历配置中 enabled 的 OIDC 设置，
// 拉取各 provider 的元数据与额外声明，组装成 biz.OidcConfig 供认证使用。
// 用 oidcOnce 保证幂等；单个 provider 拉取失败只告警并跳过，不影响整体启动。
func (data *dataImpl) InitOidcProvider() {
	data.oidcOnce.Do(func() {
		var (
			oidcConfig = make(biz.OidcConfig)
			cfg        = data.Config()
			logger     = data.logger
		)
		logger.Info("init oidc provider...")
		for _, setting := range cfg.Oidc {
			if !setting.Enabled {
				continue
			}
			// 单个 provider 拉取/声明解析失败只告警并跳过，不中断其余 provider。
			provider, err := oidc.NewProvider(context.TODO(), setting.ProviderUrl)
			if err != nil {
				logger.Warning(setting.ProviderUrl, err)
				continue
			}
			var ev extraValues
			if err = provider.Claims(&ev); err != nil {
				logger.Warning(setting.ProviderUrl, err)
				continue
			}
			addOidcCfg(provider, ev, setting, oidcConfig)
		}
		data.oidc = oidcConfig
	})
}

// extraValues 是 OIDC provider discovery 文档中按名字读取的额外字段。
type extraValues struct {
	ScopesSupported    []string `json:"scopes_supported"`
	EndSessionEndpoint string   `json:"end_session_endpoint"`
}

// addOidcCfg 把单个 provider 组装成 oauth2 配置并写入 oidcConfig 映射：
// scopes 缺省回落 ScopeOpenID，end_session 端点用于登出流程。
func addOidcCfg(provider *oidc.Provider, extraValues extraValues, setting config.OidcSetting, oidcConfig biz.OidcConfig) {
	scopes := extraValues.ScopesSupported
	if len(scopes) < 1 {
		scopes = []string{oidc.ScopeOpenID}
	}

	oauth2Config := oauth2.Config{
		ClientID:     setting.ClientID,
		ClientSecret: setting.ClientSecret,
		RedirectURL:  setting.RedirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}
	oidcConfig[setting.Name] = biz.OidcConfigItem{
		Provider:           provider,
		Config:             oauth2Config,
		EndSessionEndpoint: extraValues.EndSessionEndpoint,
	}
}
