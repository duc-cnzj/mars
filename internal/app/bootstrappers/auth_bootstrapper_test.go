package bootstrappers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
)

type testapp struct {
	contracts.ApplicationInterface
	cfg  *config.Config
	auth contracts.AuthInterface
}

func (t *testapp) Config() *config.Config {
	return t.cfg
}

func (t *testapp) SetAuth(a contracts.AuthInterface) {
	t.auth = a
}

func TestAuthBootstrapper_Bootstrap(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})

	app := &testapp{
		cfg: &config.Config{
			PrivateKey: bf.String(),
		},
	}
	j := auth.NewJwtAuth(key, key.Public().(*rsa.PublicKey))
	(&AuthBootstrapper{}).Bootstrap(app)
	info := contracts.UserInfo{
		ID:        "x",
		Email:     "x",
		Name:      "x",
		Picture:   "x",
		Roles:     nil,
		LogoutUrl: "x",
	}
	sign, _ := j.Sign(info)
	data, _ := app.auth.Sign(info)
	assert.Equal(t, sign, data)
	_, b := app.auth.VerifyToken(sign.Token)
	assert.True(t, b)
	assert.Len(t, app.auth.(*auth.Authn).Authns, 2)
	assert.IsType(t, &auth.JwtAuth{}, app.auth.(*auth.Authn).Authns[0])
	assert.IsType(t, &auth.AccessTokenAuth{}, app.auth.(*auth.Authn).Authns[1])
}

func TestAuthBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&AuthBootstrapper{}).Tags())
}
