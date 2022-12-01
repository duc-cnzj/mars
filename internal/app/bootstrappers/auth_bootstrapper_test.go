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
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestAuthBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	app.EXPECT().Config().Return(&config.Config{
		PrivateKey: bf.String(),
	}).Times(1)
	app.EXPECT().SetAuth(auth.NewJwtAuth(key, key.Public().(*rsa.PublicKey))).Times(1)
	(&AuthBootstrapper{}).Bootstrap(app)
}

func TestAuthBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&AuthBootstrapper{}).Tags())
}
