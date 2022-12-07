package bootstrappers

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
)

type OidcBootstrapper struct{}

func (o *OidcBootstrapper) Tags() []string {
	return []string{}
}

type extraValues struct {
	CheckSessionIFrame string   `json:"check_session_iframe"`
	ScopesSupported    []string `json:"scopes_supported"`
	EndSessionEndpoint string   `json:"end_session_endpoint"`
}

func (o *OidcBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	var oidcConfig contracts.OidcConfig = make(contracts.OidcConfig)
	for _, setting := range cfg.Oidc {
		if !setting.Enabled {
			continue
		}
		var (
			err      error
			provider *oidc.Provider
		)
		if provider, err = oidc.NewProvider(context.TODO(), setting.ProviderUrl); err != nil {
			return err
		}

		var ev extraValues
		if err = provider.Claims(&ev); err != nil {
			return err
		}
		addOidcCfg(provider, ev, setting, oidcConfig)
	}

	app.SetOidc(oidcConfig)

	return nil
}

func addOidcCfg(provider *oidc.Provider, extraValues extraValues, setting config.OidcSetting, oidcConfig contracts.OidcConfig) {
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
	oidcConfig[setting.Name] = contracts.OidcConfigItem{
		Provider:           provider,
		Config:             oauth2Config,
		EndSessionEndpoint: extraValues.EndSessionEndpoint,
	}
}
