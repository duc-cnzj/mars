package domain_manager

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSyncSecretDomainManager_Destroy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + (&SyncSecretDomainManager{}).Name() + " plugin Destroy...").Times(1)
	mm := &SyncSecretDomainManager{}
	mm.Destroy()
}

func TestSyncSecretDomainManager_GetCertSecretName(t *testing.T) {
	assert.Equal(t, SyncSecretSecretName, (&SyncSecretDomainManager{}).GetCertSecretName("", 1))
}

func TestSyncSecretDomainManager_GetCerts(t *testing.T) {
	mm := &SyncSecretDomainManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + mm.Name() + " plugin Initialize...").Times(1)
	app := testutil.MockApp(m)
	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kube-public",
			Name:      "my-secret",
		},
		Data: map[string][]byte{
			"tls.key": []byte(tlsKey),
			"tls.crt": []byte(tlsCrt),
		},
		Type: v1.SecretTypeTLS,
	}
	inf := &testInf{}
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:         fake.NewSimpleClientset(sec),
		SecretLister:   testutil.NewSecretLister(sec),
		SecretInformer: inf,
	}).AnyTimes()
	mm.Initialize(map[string]any{
		"ns_prefix":        "pfx",
		"wildcard_domain":  testDomain,
		"secret_name":      "my-secret",
		"secret_namespace": "kube-public",
	})
	name, key, crt := mm.GetCerts()
	assert.Equal(t, SyncSecretSecretName, name)
	assert.Equal(t, tlsKey, key)
	assert.Equal(t, tlsCrt, crt)
	assert.Len(t, inf.handlers, 1)
}

type testInf struct {
	handlers []cache.ResourceEventHandler
	cache.SharedIndexInformer
}

func (i *testInf) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	i.handlers = append(i.handlers, handler)
	return nil, nil
}

func TestSyncSecretDomainManager_GetClusterIssuer(t *testing.T) {
	assert.Equal(t, "", (&SyncSecretDomainManager{}).GetClusterIssuer())
}

func TestSyncSecretDomainManager_GetDomain(t *testing.T) {
	mm := &SyncSecretDomainManager{}
	preOccupiedLen := 0
	projectName := "pj"
	namespace := "ns"
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     mm.nsPrefix,
		domainSuffix: mm.domainSuffix,
	}.SubStr(), mm.GetDomain(projectName, namespace, preOccupiedLen))
}

func TestSyncSecretDomainManager_GetDomainByIndex(t *testing.T) {
	mm := &SyncSecretDomainManager{}
	idx := 0
	preOccupiedLen := 0
	projectName := "pj"
	namespace := "ns"
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        idx,
		nsPrefix:     mm.nsPrefix,
		domainSuffix: mm.domainSuffix,
	}.SubStr(), mm.GetDomainByIndex(projectName, namespace, idx, preOccupiedLen))
}

func TestSyncSecretDomainManager_Initialize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + (&SyncSecretDomainManager{}).Name() + " plugin Initialize...").Times(1)
	mm := &SyncSecretDomainManager{}
	app := testutil.MockApp(m)
	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "my-secret",
		},
		Data: map[string][]byte{
			"tls.key": []byte(tlsKey),
			"tls.crt": []byte(tlsCrt),
		},
		Type: v1.SecretTypeTLS,
	}
	inf := &testInf{}

	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		SecretLister:   testutil.NewSecretLister(sec),
		Client:         fake.NewSimpleClientset(sec),
		SecretInformer: inf,
	}).AnyTimes()
	assert.Nil(t, mm.Initialize(map[string]any{
		"ns_prefix":        "pfx",
		"wildcard_domain":  testDomain,
		"secret_name":      "my-secret",
		"secret_namespace": "test",
	}))
	assert.NotNil(t, mm.GetSecret())
	assert.Equal(t, "my-secret", mm.GetSecret().Name)
	assert.Equal(t, testDomain, mm.wildcardDomain)
	assert.Equal(t, "mars.local", mm.domainSuffix)
	assert.Equal(t, "pfx", mm.nsPrefix)
	assert.Equal(t, "my-secret", mm.secretName)
	assert.Equal(t, "test", mm.secretNamespace)
	assert.Len(t, inf.handlers, 1)
}

func TestSyncSecretDomainManager_Initialize_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mm := &SyncSecretDomainManager{}
	app := testutil.MockApp(m)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{SecretLister: testutil.NewSecretLister()}).Times(1)

	assert.True(t, apierrors.IsNotFound(mm.Initialize(map[string]any{
		"ns_prefix":        "pfx",
		"wildcard_domain":  testDomain,
		"secret_name":      "my-secret",
		"secret_namespace": "test",
	})))

	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "my-secret",
		},
		Data: map[string][]byte{
			"tls.key": []byte(tlsKey),
			"tls.crt": []byte(tlsCrt),
		},
		Type: v1.SecretTypeOpaque,
	}
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{SecretLister: testutil.NewSecretLister(sec)}).Times(1)
	assert.Equal(t, errors.New("secret not verified"), mm.Initialize(map[string]any{
		"ns_prefix":        "pfx",
		"wildcard_domain":  testDomain,
		"secret_name":      "my-secret",
		"secret_namespace": "test",
	}))

	sec2 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "my-secret",
		},
		Data: map[string][]byte{
			"tls.key": []byte(tlsKey),
			"tls.crt": []byte(tlsCrt),
		},
		Type: v1.SecretTypeTLS,
	}
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{SecretLister: testutil.NewSecretLister(sec2)}).Times(1)

	assert.Equal(t, "域名和证书不匹配, cert dnsNames: [*.mars.local], 域名: errorDomain", mm.Initialize(map[string]any{
		"wildcard_domain":  "errorDomain",
		"ns_prefix":        "pfx",
		"secret_name":      "my-secret",
		"secret_namespace": "test",
	}).Error())

	assert.Equal(t, errors.New("secret_namespace, secret_name, wildcard_domain required"), (&SyncSecretDomainManager{}).Initialize(map[string]any{
		"ns_prefix":        "pfx",
		"secret_name":      "my-secret",
		"secret_namespace": "test",
	}))
	assert.Equal(t, errors.New("secret_namespace, secret_name, wildcard_domain required"), (&SyncSecretDomainManager{}).Initialize(map[string]any{
		"ns_prefix":        "pfx",
		"wildcard_domain":  testDomain,
		"secret_namespace": "test",
	}))
	assert.Equal(t, errors.New("secret_namespace, secret_name, wildcard_domain required"), (&SyncSecretDomainManager{}).Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"wildcard_domain": testDomain,
		"secret_name":     "my-secret",
	}))
}

func TestSyncSecretDomainManager_Name(t *testing.T) {
	assert.Equal(t, "sync_secret_domain_manager", (&SyncSecretDomainManager{}).Name())
}

func TestSyncSecretDomainManager_eventHandler(t *testing.T) {
	tests := []struct {
		name  string
		ns    string
		wants bool
	}{
		{
			name:  "sec",
			ns:    "ns",
			wants: true,
		},
		{
			name:  "sec1",
			ns:    "ns",
			wants: false,
		},
		{
			name:  "sec",
			ns:    "ns1",
			wants: false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			t.Parallel()
			called := 0
			m := (&SyncSecretDomainManager{
				secretName:      "sec",
				secretNamespace: "ns",
			}).eventHandler(func(oldObj, newObj any) {
				called++
			})
			assert.Equal(t, tt.wants, m.FilterFunc(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.name,
					Namespace: tt.ns,
				},
			}))
			m.Handler.(cache.ResourceEventHandlerFuncs).UpdateFunc(nil, nil)
			assert.Equal(t, 1, called)
		})
	}

}

func TestSyncSecretDomainManager_handleSecretChange(t *testing.T) {
	m := &SyncSecretDomainManager{
		secretName:      "sec",
		secretNamespace: "ns",
	}
	m.handleSecretChange(
		&v1.Secret{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "0"}},
		&v1.Secret{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "0"}},
	)

	sec1 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{ResourceVersion: "0"},
		Data: map[string][]byte{
			"tls.key": nil,
			"tls.crt": nil,
		},
	}
	sec2 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"},
		Data: map[string][]byte{
			"tls.key": nil,
			"tls.crt": nil,
		},
	}
	m.handleSecretChange(sec1, sec2)
	assert.Same(t, m.secret, sec2)

	sec3 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"},
		Data: map[string][]byte{
			"tls.key": []byte("aaa"),
			"tls.crt": []byte("bbb"),
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	app := testutil.MockApp(ctrl)
	db, f := testutil.SetGormDB(ctrl, app)
	defer f()
	assert.Nil(t, db.AutoMigrate(&models.Namespace{}))
	called := 0
	m.updateCertTlsFunc = func(name, key, crt string) {
		called++
		assert.Equal(t, SyncSecretSecretName, name)
	}
	m.handleSecretChange(sec1, sec3)
	assert.Equal(t, 1, called)

	sec4 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"},
		Data: map[string][]byte{
			"tls.key": []byte("aaa"),
			"tls.crt": []byte("bbb"),
		},
	}
	m.handleSecretChange(nil, sec4)
	assert.Same(t, m.secret, sec4)
	assert.Equal(t, 2, called)
}
