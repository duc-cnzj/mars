package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent/accesstoken"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	priKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &priKey.PublicKey
)

func TestAuth_Sign(t *testing.T) {
	auth := NewJwtAuth(priKey, publicKey)
	sign, err := auth.Sign(&UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{schematype.MarsAdmin},
		ID:        "1",
		Email:     "1025434218@qq.com",
		Name:      "duc",
	})
	assert.Nil(t, err)
	token, b := auth.VerifyToken(sign.Token)
	assert.True(t, b)
	assert.Equal(t, "mars", token.StandardClaims.Issuer)
	assert.Equal(t, "duc", token.UserInfo.Name)
	assert.Equal(t, "1025434218@qq.com", token.StandardClaims.Subject)
	assert.Equal(t, []string{schematype.MarsAdmin}, token.UserInfo.Roles)
	assert.Equal(t, "xxx", token.UserInfo.LogoutUrl)

	pk := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: big.NewInt(1),
		},
	}
	assert.Less(t, pk.Size(), 11)
	authError := NewJwtAuth(pk, nil)
	_, err = authError.Sign(&UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{schematype.MarsAdmin},
		ID:        "1",
		Email:     "1025434218@qq.com",
		Name:      "duc",
	})
	assert.Error(t, err)
}

func TestAuth_VerifyToken(t *testing.T) {
	auth := NewJwtAuth(priKey, publicKey)
	sign, _ := auth.Sign(&UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{schematype.MarsAdmin},
		ID:        "1",
		Name:      "duc",
	})
	_, b := auth.VerifyToken(sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("Bearer " + sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("bearer " + sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("bearer" + sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("")
	assert.False(t, b)
}

func TestNewAuth(t *testing.T) {
	assert.Implements(t, (*Auth)(nil), NewJwtAuth(nil, nil))
}

func TestNewAccessTokenAuth(t *testing.T) {
	assert.Implements(t, (*Authenticator)(nil), NewAccessTokenAuth(nil))
}

func TestAccessTokenAuth_VerifyToken(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	_, b := NewAccessTokenAuth(nil).VerifyToken("")
	assert.False(t, b)

	db, _ := data.NewSqliteDB()
	defer db.Close()
	at := db.AccessToken.Create().
		SetToken("my token").
		SetUsage("x").
		SetExpiredAt(time.Now().Add(10 * time.Second)).
		SetUserInfo(schematype.UserInfo{
			ID:        "xx",
			Email:     "admin@admin.com",
			Name:      "duc",
			Picture:   "xx",
			Roles:     []string{schematype.MarsAdmin},
			LogoutUrl: "https://xxx",
		}).SaveX(context.TODO())
	assert.Nil(t, at.LastUsedAt)
	dd := data.NewDataImpl(&data.NewDataParams{DB: db})
	u, b := NewAccessTokenAuth(dd).
		VerifyToken(at.Token)
	assert.True(t, b)
	assert.Equal(t, "xx", u.UserInfo.ID)
	assert.Equal(t, "admin@admin.com", u.UserInfo.Email)
	assert.Equal(t, "duc", u.UserInfo.Name)
	assert.Equal(t, "xx", u.UserInfo.Picture)
	assert.Equal(t, []string{schematype.MarsAdmin}, u.UserInfo.Roles)
	assert.Equal(t, "https://xxx", u.UserInfo.LogoutUrl)

	first, _ := db.AccessToken.Query().Where(accesstoken.Token(at.Token)).First(context.Background())

	assert.NotZero(t, first.LastUsedAt)
	_, bb := NewAccessTokenAuth(dd).VerifyToken("bearer " + at.Token)
	assert.True(t, bb)
}

type autha struct{}

func (a *autha) VerifyToken(s string) (*JwtClaims, bool) {
	return nil, false
}

func TestNewAuthn(t *testing.T) {
	authn, _ := NewAuthn(data.NewDataImpl(&data.NewDataParams{Cfg: &config.Config{PrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQCdx5ZBeL3P3lH2fU/8yd4E1L880DjaKCnnnQkya+kOE7kkJNtP
xW4WIKsBgXUPtXUYk/uA5AkklJ/1ssiTbkM/G5J54ThsACarhiNijUznD81c7g0Q
6pbHYGAHU91wQgpcIv39cOKZVpFkEfIwgBMIKUvupBpGyXMU4YALVV23CQIDAQAB
AoGARo+kzeDumlDlvONr6zRoOybd45eHZWEC5JchLtB9qJL/gH+PKQy1X+X6NDEu
JflTxcsgdhMFV7u0EdCDzRNJtPKP/cU8hww0J2l3ZKTGzbbQnLIBFD3In8sEc9xe
3ikEjqs0EgSh3uY5XEq8qzuX3cI+FNlGyOwzM+ZcN7nWfPUCQQDOURX82COQIfAT
RjTshDQ55J/DUPPHyzpTER9OZNXYKp0IBBNzYyhJ6SHQHSuxHfL8W1FVHhmIsIBW
GQWo0y7zAkEAw8ZPJ4QH5otMsIgIfwMuPX0rO+QxwmJ6eg9ADuFr5zv6HizjAVVP
dKXuUU0gnemD4DncgiV2jZ0v2RzHK1aZEwJAR6G7gpgAcPB3jBmaEmwsPdV06rlW
io2y6FhPiEZWQME62CeiITPSLyc0SC94lfwR+zAxYt4ae2zcgggaAO2hpQJAecA5
d7S3iRu2XM6sofijaCAQpBV9EItX6dLUHqz4Av0cxmlZ33ljiYKr3CngD/SqS+cQ
CGwt91H68MXh40TeuwJARxz1VMLq7hKo8J4scAW/YrBTE4N6malYjYoR2HFs+YwL
cSE/4A4yfzTjN2r5GuJr8rTU7gU4Su9C8dLC0htWCA==
-----END RSA PRIVATE KEY-----
`}}))
	assert.Implements(t, (*Auth)(nil), authn)
}
