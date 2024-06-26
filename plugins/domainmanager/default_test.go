package domainmanager

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDefaultDomainManager_Destroy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + (&defaultDomainManager{}).Name() + " plugin Destroy...").Times(1)
	(&defaultDomainManager{}).Destroy()
}

func TestDefaultDomainManager_GetCertSecretName(t *testing.T) {
	assert.Equal(t, "", (&defaultDomainManager{}).GetCertSecretName("", 0))
}

func TestDefaultDomainManager_GetCerts(t *testing.T) {
	n, key, crt := (&defaultDomainManager{}).GetCerts()
	assert.Equal(t, "", n)
	assert.Equal(t, "", key)
	assert.Equal(t, "", crt)
}

func TestDefaultDomainManager_GetClusterIssuer(t *testing.T) {
	assert.Equal(t, "", (&defaultDomainManager{}).GetClusterIssuer())
}

func TestDefaultDomainManager_GetDomain(t *testing.T) {
	preOccupiedLen := 0
	projectName := "pj"
	namespace := "ns"
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     "devops",
		domainSuffix: "faker-domain.local",
	}.SubStr(), (&defaultDomainManager{}).GetDomain(projectName, namespace, preOccupiedLen))
}

func TestDefaultDomainManager_GetDomainByIndex(t *testing.T) {
	preOccupiedLen := 0
	projectName := "pj"
	namespace := "ns"
	idx := 1
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        idx,
		nsPrefix:     "devops",
		domainSuffix: "faker-domain.local",
	}.SubStr(), (&defaultDomainManager{}).GetDomainByIndex(projectName, namespace, idx, preOccupiedLen))
}

func TestDefaultDomainManager_Initialize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + (&defaultDomainManager{}).Name() + " plugin Initialize...").Times(1)
	(&defaultDomainManager{}).Initialize(map[string]any{})
}

func TestDefaultDomainManager_Name(t *testing.T) {
	assert.Equal(t, "default_domain_manager", (&defaultDomainManager{}).Name())
}

func TestNewDefaultDomainManager(t *testing.T) {
	assert.IsType(t, &defaultDomainManager{}, NewDefaultDomainManager())
}
