package bootstrappers

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt"

	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
)

type AuthBootstrapper struct{}

func (a *AuthBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	pem, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(app.Config().PrivateKey))
	if err != nil {
		return err
	}
	app.SetAuth(auth.NewAuth(pem, pem.Public().(*rsa.PublicKey)))

	return nil
}
