package bootstrappers

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/internal/contracts"
	"golang.org/x/oauth2"
)

type OidcBootstrapper struct{}

func (D *OidcBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	if !cfg.OidcEnabled {
		return nil
	}

	provider, err := oidc.NewProvider(context.TODO(), cfg.ProviderUrl)
	if err != nil {
		return err
	}
	// Configure an OpenID Connect aware OAuth2 client.
	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectUrl,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "offline"},
	}

	var sessionURLs struct {
		CheckSessionIFrame string `json:"check_session_iframe"`
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := provider.Claims(&sessionURLs); err != nil {
		return err
	}

	app.SetOidc(&contracts.OidcConfig{
		Provider:           provider,
		Config:             oauth2Config,
		EndSessionEndpoint: sessionURLs.EndSessionEndpoint,
	})

	return nil
}
