package bootstrappers

import (
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
)

func TestOidcBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Times(1).Return(&config.Config{})
	app.EXPECT().SetOidc(gomock.Any()).Times(1)
	(&OidcBootstrapper{}).Bootstrap(app)
}

func TestOidcBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&OidcBootstrapper{}).Tags())
}

func Test_addOidcCfg(t *testing.T) {
	var (
		provider = &oidc.Provider{}
		setting  = config.OidcSetting{
			Name:         "name",
			Enabled:      true,
			ProviderUrl:  "xxx",
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectUrl:  "redirectUrl",
		}
		ev = extraValues{
			EndSessionEndpoint: "https://xxx",
		}
		oidcConfig = make(contracts.OidcConfig)
	)
	addOidcCfg(provider, ev, setting, oidcConfig)
	assert.Len(t, oidcConfig, 1)
	assert.Equal(t, contracts.OidcConfigItem{
		Provider: provider,
		Config: oauth2.Config{
			ClientID:     setting.ClientID,
			ClientSecret: setting.ClientSecret,
			Endpoint:     oauth2.Endpoint{},
			RedirectURL:  setting.RedirectUrl,
			Scopes:       []string{oidc.ScopeOpenID},
		},
		EndSessionEndpoint: ev.EndSessionEndpoint,
	}, oidcConfig[setting.Name])
}

func Test_addOidcCfg2(t *testing.T) {
	var (
		provider = &oidc.Provider{}
		setting  = config.OidcSetting{
			Name:         "name",
			Enabled:      true,
			ProviderUrl:  "xxx",
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectUrl:  "redirectUrl",
		}
		ev = extraValues{
			ScopesSupported:    []string{"openid", "app", "email"},
			EndSessionEndpoint: "https://xxx",
		}
		oidcConfig = make(contracts.OidcConfig)
	)
	addOidcCfg(provider, ev, setting, oidcConfig)
	assert.Len(t, oidcConfig, 1)
	assert.Equal(t, contracts.OidcConfigItem{
		Provider: provider,
		Config: oauth2.Config{
			ClientID:     setting.ClientID,
			ClientSecret: setting.ClientSecret,
			Endpoint:     oauth2.Endpoint{},
			RedirectURL:  setting.RedirectUrl,
			Scopes:       []string{"openid", "app", "email"},
		},
		EndSessionEndpoint: ev.EndSessionEndpoint,
	}, oidcConfig[setting.Name])
}
