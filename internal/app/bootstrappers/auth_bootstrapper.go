package bootstrappers

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt"

	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type AuthBootstrapper struct{}

func (a *AuthBootstrapper) Tags() []string {
	return []string{}
}

func (a *AuthBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	pem, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(app.Config().PrivateKey))
	if err != nil {
		return err
	}
	jwtAuth := auth.NewJwtAuth(pem, pem.Public().(*rsa.PublicKey))
	app.SetAuth(auth.NewAuthn(jwtAuth.Sign, jwtAuth, auth.NewAccessTokenAuth(app)))

	return nil
}
