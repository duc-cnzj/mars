package domainmanager

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"go.uber.org/mock/gomock"

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
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + (&certManager{}).Name() + " plugin Destroy...")
	(&certManager{}).Destroy()
}

func TestCertManager_GetCertSecretName(t *testing.T) {
	assert.Equal(t,
		fmt.Sprintf("mars-tls-%s", utils.Hash(fmt.Sprintf("%s-%d", "", 1))),
		(&certManager{}).GetCertSecretName("", 1),
	)
}

func TestCertManager_GetCerts(t *testing.T) {
	name, key, crt := (&certManager{}).GetCerts()
	assert.Empty(t, name)
	assert.Empty(t, key)
	assert.Empty(t, crt)
}

func TestCertManager_GetClusterIssuer(t *testing.T) {
	cm := &certManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info(gomock.Any()).Times(1)
	cm.Initialize(map[string]any{"cluster_issuer": "issuer", "wildcard_domain": "*.test.local"})
	assert.Equal(t, "issuer", cm.GetClusterIssuer())
}

func TestCertManager_Initialize(t *testing.T) {
	cm := &certManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info(gomock.Any()).Times(1)
	cm.Initialize(map[string]any{"cluster_issuer": "issuer", "ns_prefix": "pre", "wildcard_domain": "*.mars.test"})
	assert.Equal(t, "issuer", cm.clusterIssuer)
	assert.Equal(t, "mars.test", cm.domainSuffix)
	assert.Equal(t, "*.mars.test", cm.wildcardDomain)
	assert.Equal(t, "pre", cm.nsPrefix)

	err := cm.Initialize(map[string]any{"cluster_issuer": "", "wildcard_domain": ""})
	assert.Error(t, err)
}

func TestCertManager_Name(t *testing.T) {
	assert.Equal(t, name, (&certManager{}).Name())
}

func TestCertManager_GetDomainByIndex(t *testing.T) {
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - 1,
		projectName:  "pro",
		namespace:    "ns",
		index:        1,
		nsPrefix:     "",
		domainSuffix: "",
	}.SubStr(), (&certManager{}).GetDomainByIndex("pro", "ns", 1, 1))
}

func TestCertManager_GetDomain(t *testing.T) {
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - 1,
		projectName:  "pro",
		namespace:    "ns",
		index:        -1,
		nsPrefix:     "",
		domainSuffix: "",
	}.SubStr(), (&certManager{}).GetDomain("pro", "ns", 1))
}

func Test_SubStr(t *testing.T) {
	assert.Equal(t, "aaa", substr("aaa", 1000))
}

func TestSubdomain_SubStr(t *testing.T) {
	str := (Subdomain{
		maxLen:      0,
		projectName: "pro",
		namespace:   "ns",
		index:       -1,
	}).SubStr()
	assert.Equal(t, fmt.Sprintf("%s-%s.%s", "pro", "ns", ""), str)
}
