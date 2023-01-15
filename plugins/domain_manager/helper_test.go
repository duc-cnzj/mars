package domain_manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateTelsWildcardDomain(t *testing.T) {
	assert.Nil(t, validateTelsWildcardDomain(tlsKey, tlsCrt, testDomain))
	assert.Error(t, validateTelsWildcardDomain("", tlsCrt, testDomain))
	assert.Error(t, validateTelsWildcardDomain(tlsKey, "", testDomain))
	assert.Equal(t, "域名和证书不匹配, cert dnsNames: [*.mars.local], 域名: *.not-exists.com", validateTelsWildcardDomain(tlsKey, tlsCrt, "*.not-exists.com").Error())
}
