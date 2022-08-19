package bootstrappers

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/internal/contracts"
	"golang.org/x/oauth2"
)

type OidcBootstrapper struct{}

func (o *OidcBootstrapper) Tags() []string {
	return []string{}
}

func (o *OidcBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	var oidcConfig contracts.OidcConfig = make(contracts.OidcConfig)
	for _, setting := range cfg.Oidc {
		if !setting.Enabled {
			continue
		}
		provider, err := oidc.NewProvider(context.TODO(), setting.ProviderUrl)
		if err != nil {
			return err
		}

		var extraValues struct {
			CheckSessionIFrame string   `json:"check_session_iframe"`
			ScopesSupported    []string `json:"scopes_supported"`
			EndSessionEndpoint string   `json:"end_session_endpoint"`
		}
		if err := provider.Claims(&extraValues); err != nil {
			return err
		}

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

	app.SetOidc(oidcConfig)

	return nil
}
