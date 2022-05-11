package domain_manager

import (
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func Test_Substr(t *testing.T) {
	sd := Subdomain{
		maxLen:       12,
		projectName:  "app",
		namespace:    "devops-test",
		index:        -1,
		nsPrefix:     "devops-",
		domainSuffix: "test.com",
	}
	assert.Equal(t, "app-devops-test.test.com", sd.CompleteSubdomain())
	assert.Equal(t, "app-test.test.com", sd.MediumSubdomain())
	assert.Equal(t, sd.SimpleSubdomain(), sd.SubStr())

	sd.maxLen = 17
	assert.Equal(t, sd.MediumSubdomain(), sd.SubStr())
	assert.True(t, len(sd.SubStr()) <= sd.maxLen)
	sd.maxLen = 9999
	assert.Equal(t, sd.CompleteSubdomain(), sd.SubStr())
	assert.True(t, len(sd.SubStr()) <= sd.maxLen)

	sd2 := Subdomain{
		maxLen:       12,
		projectName:  "app",
		namespace:    "devops-test",
		index:        1,
		nsPrefix:     "devops-",
		domainSuffix: "test.com",
	}

	assert.Equal(t, "app-devops-test-1.test.com", sd2.CompleteSubdomain())
	assert.Equal(t, "app-test-1.test.com", sd2.MediumSubdomain())
	assert.Equal(t, sd2.SimpleSubdomain(), sd2.SubStr())
	assert.Len(t, sd2.SubStr(), sd2.maxLen)

	sd3 := Subdomain{
		maxLen:       1,
		projectName:  "app",
		namespace:    "devops-test",
		index:        1,
		nsPrefix:     "devops-",
		domainSuffix: "test.com",
	}
	assert.Panics(t, func() {
		sd3.SubStr()
	})
}

func TestCertManager_Destroy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	l.EXPECT().Info("[Plugin]: " + (&CertManager{}).Name() + " plugin Destroy...")
	(&CertManager{}).Destroy()
}

func TestCertManager_GetCertSecretName(t *testing.T) {
	assert.Equal(t,
		fmt.Sprintf("mars-tls-%s", utils.Md5(fmt.Sprintf("%s-%d", "", 1))),
		(&CertManager{}).GetCertSecretName("", 1),
	)
}

func TestCertManager_GetCerts(t *testing.T) {
	name, key, crt := (&CertManager{}).GetCerts()
	assert.Empty(t, name)
	assert.Empty(t, key)
	assert.Empty(t, crt)
}

func TestCertManager_GetClusterIssuer(t *testing.T) {
	cm := &CertManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	l.EXPECT().Info(gomock.Any()).Times(1)
	cm.Initialize(map[string]any{"cluster_issuer": "issuer", "wildcard_domain": "*.test.local"})
	assert.Equal(t, "issuer", cm.GetClusterIssuer())
}

func TestCertManager_Initialize(t *testing.T) {
	cm := &CertManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	l.EXPECT().Info(gomock.Any()).Times(1)
	cm.Initialize(map[string]any{"cluster_issuer": "issuer", "ns_prefix": "pre", "wildcard_domain": "*.mars.test"})
	assert.Equal(t, "issuer", cm.clusterIssuer)
	assert.Equal(t, "mars.test", cm.domainSuffix)
	assert.Equal(t, "*.mars.test", cm.wildcardDomain)
	assert.Equal(t, "pre", cm.nsPrefix)
}

func TestCertManager_Name(t *testing.T) {
	assert.Equal(t, name, (&CertManager{}).Name())
}
