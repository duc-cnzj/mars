package domain_manager

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDomainManager_Destroy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	l.EXPECT().Info("[Plugin]: " + (&DefaultDomainManager{}).Name() + " plugin Destroy...")
	(&DefaultDomainManager{}).Destroy()
}

func TestDefaultDomainManager_GetCertSecretName(t *testing.T) {
	assert.Equal(t, "", (&DefaultDomainManager{}).GetCertSecretName("", 0))
}

func TestDefaultDomainManager_GetCerts(t *testing.T) {
	n, key, crt := (&DefaultDomainManager{}).GetCerts()
	assert.Equal(t, "", n)
	assert.Equal(t, "", key)
	assert.Equal(t, "", crt)
}

func TestDefaultDomainManager_GetClusterIssuer(t *testing.T) {
	assert.Equal(t, "", (&DefaultDomainManager{}).GetClusterIssuer())
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
	}.SubStr(), (&DefaultDomainManager{}).GetDomain(projectName, namespace, preOccupiedLen))
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
	}.SubStr(), (&DefaultDomainManager{}).GetDomainByIndex(projectName, namespace, idx, preOccupiedLen))
}

func TestDefaultDomainManager_Initialize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	l.EXPECT().Info("[Plugin]: " + (&DefaultDomainManager{}).Name() + " plugin Initialize...")
	(&DefaultDomainManager{}).Initialize(map[string]any{})
}

func TestDefaultDomainManager_Name(t *testing.T) {
	assert.Equal(t, "default_domain_manager", (&DefaultDomainManager{}).Name())
}
