package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessToken_Expired(t *testing.T) {
	at := &AccessToken{}
	assert.True(t, at.Expired())
	at.ExpiredAt = time.Now().Add(time.Hour)
	assert.False(t, at.Expired())
}

func TestNewAccessTokenRepo(t *testing.T) {

}

func TestToAccessToken(t *testing.T) {

}

func Test_accessTokenRepo_Grant(t *testing.T) {

}

func Test_accessTokenRepo_Lease(t *testing.T) {

}

func Test_accessTokenRepo_List(t *testing.T) {

}

func Test_accessTokenRepo_Revoke(t *testing.T) {

}
