package auth

import (
	"crypto/rsa"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/golang-jwt/jwt"
)

type Auth struct {
	priKey *rsa.PrivateKey
	pubKey *rsa.PublicKey
}

func NewAuth(priKey *rsa.PrivateKey, pubKey *rsa.PublicKey) *Auth {
	return &Auth{priKey: priKey, pubKey: pubKey}
}

func (a *Auth) VerifyToken(t string) (*contracts.JwtClaims, bool) {
	var token string = t
	if len(t) > 6 && strings.EqualFold("bearer", t[0:6]) {
		token = strings.TrimSpace(t[6:])
	}
	if token != "" {
		parse, err := jwt.ParseWithClaims(token, &contracts.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			return a.pubKey, nil
		})
		if err == nil && parse.Valid {
			return parse.Claims.(*contracts.JwtClaims), true
		}
	}

	return nil, false
}

func (a *Auth) Sign(info contracts.UserInfo) (*contracts.SignData, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &contracts.JwtClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(contracts.Expired).Unix(),
			Issuer:    "mars",
		},
		UserInfo: info,
	})

	signedString, err := token.SignedString(a.priKey)
	if err != nil {
		return nil, err
	}
	return &contracts.SignData{
		Token:     signedString,
		ExpiredIn: int64(contracts.Expired.Seconds()),
	}, nil
}
