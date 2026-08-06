package domainmanager

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
)

// dmApp 是 PluginApp 的最小手写 stub。
type dmApp struct {
	data   data.Data
	logger mlog.Logger
}

func (d dmApp) Logger() mlog.Logger { return d.logger }
func (d dmApp) Data() data.Data     { return d.data }
func (d dmApp) Cache() data.Cache   { return nil }

// genCert 生成一个 DNSNames 匹配给定列表的自签名证书，返回 PEM 格式 crt/key。
func genCert(t *testing.T, dnsNames []string) (crt, key []byte) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: dnsNames[0]},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     dnsNames,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	crt = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	keyDER, err := x509.MarshalECPrivateKey(priv)
	require.NoError(t, err)
	key = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return crt, key
}

// newFakeSecretLister 用 fake clientset + informer 构建一个包含给定 secret 的 SecretLister。
func newFakeSecretLister(t *testing.T, secret *corev1.Secret) corelistersv1.SecretLister {
	t.Helper()
	var objs []runtime.Object
	if secret != nil {
		objs = append(objs, secret)
	}
	client := fake.NewSimpleClientset(objs...)
	factory := informers.NewSharedInformerFactory(client, 0)
	lister := factory.Core().V1().Secrets().Lister()
	ctx, cancel := context.WithCancel(context.Background())
	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())
	cancel()
	return lister
}

// tlsSecret 构造一个 TLS 类型的 k8s secret。
func tlsSecret(name, ns string, crt, key []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": crt,
			"tls.key": key,
		},
	}
}

// ---------------------------------------------------------------------------
// subdomain.go
// ---------------------------------------------------------------------------

func TestSubdomain_HasIndex(t *testing.T) {
	assert.True(t, Subdomain{index: 3}.HasIndex())
	assert.False(t, Subdomain{index: -1}.HasIndex())
}

func TestSubdomain_CompleteSubdomain(t *testing.T) {
	withIndex := Subdomain{projectName: "app", namespace: "devops-prod", index: 1, domainSuffix: "test.com"}
	assert.Equal(t, "app-devops-prod-1.test.com", withIndex.CompleteSubdomain())

	withoutIndex := Subdomain{projectName: "app", namespace: "devops-prod", index: -1, domainSuffix: "test.com"}
	assert.Equal(t, "app-devops-prod.test.com", withoutIndex.CompleteSubdomain())
}

func TestSubdomain_MediumSubdomain(t *testing.T) {
	// ns 前缀（含尾连字符，默认 "devops-"）被剥离。
	withIndex := Subdomain{projectName: "app", namespace: "devops-prod", index: 1, nsPrefix: "devops-", domainSuffix: "test.com"}
	assert.Equal(t, "app-prod-1.test.com", withIndex.MediumSubdomain())

	withoutIndex := Subdomain{projectName: "app", namespace: "devops-prod", index: -1, nsPrefix: "devops-", domainSuffix: "test.com"}
	assert.Equal(t, "app-prod.test.com", withoutIndex.MediumSubdomain())

	// namespace 前缀与 nsPrefix 不匹配时保持原样。
	noPrefix := Subdomain{projectName: "app", namespace: "prod", index: -1, nsPrefix: "devops-", domainSuffix: "test.com"}
	assert.Equal(t, "app-prod.test.com", noPrefix.MediumSubdomain())
}

func TestSubdomain_SubStr_degrade_chain(t *testing.T) {
	base := Subdomain{projectName: "app", namespace: "devops-prod", index: 1, nsPrefix: "devops-", domainSuffix: "test.com"}

	// maxLen=0 直接返回完整版（该 struct 无 index，故为 app-prod.test.com）。
	assert.Equal(t, "app-prod.test.com", Subdomain{maxLen: 0, projectName: "app", namespace: "prod", index: -1, domainSuffix: "test.com"}.SubStr())

	// 完整版放得下。
	assert.Equal(t, "app-devops-prod-1.test.com", base.SubStr())

	// 完整版超长 → 降级到中等版。完整版 26 字符、中等版 18 字符，maxLen=20 时取中等版。
	medium := Subdomain{projectName: "app", namespace: "devops-prod", index: 1, nsPrefix: "devops-", domainSuffix: "test.com", maxLen: 20}
	assert.Equal(t, "app-prod-1.test.com", medium.SubStr())

	// 中等版也超长（maxLen=15 < 18）→ 降级到简单版（哈希截断）。
	simple := Subdomain{projectName: "app", namespace: "devops-prod", index: 1, nsPrefix: "devops-", domainSuffix: "test.com", maxLen: 15}
	got := simple.SubStr()
	assert.LessOrEqual(t, len(got), 15)
	assert.Equal(t, "test.com", got[len(got)-len("test.com"):])
}

func TestSubdomain_SimpleSubdomain(t *testing.T) {
	withIndex := Subdomain{projectName: "app", namespace: "ns", index: 1, domainSuffix: "test.com", maxLen: 20}
	assert.Equal(t, "test.com", withIndex.SimpleSubdomain()[len(withIndex.SimpleSubdomain())-len("test.com"):])

	withoutIndex := Subdomain{projectName: "app", namespace: "ns", index: -1, domainSuffix: "test.com", maxLen: 20}
	assert.Equal(t, "test.com", withoutIndex.SimpleSubdomain()[len(withoutIndex.SimpleSubdomain())-len("test.com"):])
}

func TestSubdomain_SimpleSubdomain_panics_when_no_room(t *testing.T) {
	// leftLen <= 0 时必须 panic。
	s := Subdomain{projectName: "app", namespace: "ns", index: -1, domainSuffix: "toolongdomain.com", maxLen: 5}
	assert.Panics(t, func() { s.SimpleSubdomain() })
}

func TestSubdomain_substr(t *testing.T) {
	assert.Equal(t, "abc", substr("abc", 5))
	assert.Equal(t, "ab", substr("abc", 2))
}

// ---------------------------------------------------------------------------
// validateTLSWildcardDomain (helper.go)
// ---------------------------------------------------------------------------

func TestValidateTLSWildcardDomain_matching(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	assert.NoError(t, validateTLSWildcardDomain(key, crt, "*.example.com"))
}

func TestValidateTLSWildcardDomain_invalid_keypair(t *testing.T) {
	assert.Error(t, validateTLSWildcardDomain([]byte("bad key"), []byte("bad crt"), "*.example.com"))
}

func TestValidateTLSWildcardDomain_domain_mismatch(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	err := validateTLSWildcardDomain(key, crt, "*.other.com")
	assert.ErrorContains(t, err, "域名和证书不匹配")
}

func TestValidateTLSWildcardDomain_string_types(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	assert.NoError(t, validateTLSWildcardDomain(string(key), string(crt), "*.example.com"))
}

// ---------------------------------------------------------------------------
// certmanager.go
// ---------------------------------------------------------------------------

func TestCertManager_Name(t *testing.T) {
	assert.Equal(t, "cert-manager_domain_manager", (&certManager{}).Name())
}

func TestCertManager_Initialize_valid(t *testing.T) {
	d := &certManager{}
	err := d.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, map[string]any{
		"ns_prefix":       "devops",
		"cluster_issuer":  "letsencrypt",
		"wildcard_domain": "*.example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "devops", d.nsPrefix)
	assert.Equal(t, "letsencrypt", d.clusterIssuer)
	assert.Equal(t, "*.example.com", d.wildcardDomain)
	assert.Equal(t, "example.com", d.domainSuffix)
}

func TestCertManager_Initialize_missing_required(t *testing.T) {
	d := &certManager{}
	err := d.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, map[string]any{})
	assert.ErrorContains(t, err, "cluster_issuer, wildcard_domain required")
}

func TestCertManager_Initialize_bad_type(t *testing.T) {
	cases := map[string]any{
		"ns_prefix":       123,
		"cluster_issuer":  456,
		"wildcard_domain": 789,
	}
	for key, val := range cases {
		d := &certManager{}
		args := map[string]any{"cluster_issuer": "ci", "wildcard_domain": "*.x.com", key: val}
		err := d.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, args)
		assert.ErrorContains(t, err, "must be string")
	}
}

func TestCertManager_Destroy(t *testing.T) {
	d := &certManager{logger: mlog.NewForConfig(nil)}
	assert.NoError(t, d.Destroy())
}

func TestCertManager_GetCertSecretName(t *testing.T) {
	d := &certManager{}
	name := d.GetCertSecretName("my-app", 2)
	assert.NotEmpty(t, name)
	assert.Contains(t, name, "mars-tls-")
	// 相同输入产出相同结果。
	assert.Equal(t, name, d.GetCertSecretName("my-app", 2))
}

func TestCertManager_GetClusterIssuer(t *testing.T) {
	d := &certManager{clusterIssuer: "letsencrypt"}
	assert.Equal(t, "letsencrypt", d.GetClusterIssuer())
}

func TestCertManager_GetDomain_and_GetDomainByIndex(t *testing.T) {
	d := &certManager{nsPrefix: "devops", domainSuffix: "example.com"}

	domIdx := d.GetDomainByIndex("app", "devops-prod", 2, 0)
	assert.NotEmpty(t, domIdx)
	assert.Contains(t, domIdx, "example.com")
	assert.Contains(t, domIdx, "-2.")
}

func TestCertManager_GetCerts_empty(t *testing.T) {
	d := &certManager{}
	name, key, crt := d.GetCerts()
	assert.Empty(t, name)
	assert.Empty(t, key)
	assert.Empty(t, crt)
}

// ---------------------------------------------------------------------------
// default.go
// ---------------------------------------------------------------------------

func TestDefault_Name(t *testing.T) {
	assert.Equal(t, "default_domain_manager", (&defaultDomainManager{}).Name())
}

func TestDefault_Initialize_and_Destroy(t *testing.T) {
	d := &defaultDomainManager{}
	require.NoError(t, d.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, nil))
	assert.NoError(t, d.Destroy())
}

func TestDefault_NewDefaultDomainManager(t *testing.T) {
	dm := NewDefaultDomainManager()
	assert.Equal(t, "default_domain_manager", dm.Name())
}

func TestDefault_GetDomainByIndex(t *testing.T) {
	d := &defaultDomainManager{}
	assert.Contains(t, d.GetDomainByIndex("app", "devops-prod", 1, 0), "faker-domain.local")
}

func TestDefault_empty_cert_methods(t *testing.T) {
	d := &defaultDomainManager{}
	assert.Empty(t, d.GetCertSecretName("app", 1))
	assert.Empty(t, d.GetClusterIssuer())
	name, key, crt := d.GetCerts()
	assert.Empty(t, name)
	assert.Empty(t, key)
	assert.Empty(t, crt)
}

// ---------------------------------------------------------------------------
// manual.go
// ---------------------------------------------------------------------------

func TestManual_Name(t *testing.T) {
	assert.Equal(t, "manual_domain_manager", (&manualDomainManager{}).Name())
}

func TestManual_Initialize_valid(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	m := &manualDomainManager{}
	err := m.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, map[string]any{
		"ns_prefix":       "devops",
		"tls_crt":         string(crt),
		"tls_key":         string(key),
		"wildcard_domain": "*.example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "example.com", m.domainSuffix)
}

func TestManual_Initialize_missing_required(t *testing.T) {
	m := &manualDomainManager{}
	err := m.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, map[string]any{})
	assert.ErrorContains(t, err, "tls_crt, tls_key, wildcard_domain required")
}

func TestManual_Initialize_cert_mismatch(t *testing.T) {
	crt, key := genCert(t, []string{"*.other.com"})
	m := &manualDomainManager{}
	err := m.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, map[string]any{
		"tls_crt":         string(crt),
		"tls_key":         string(key),
		"wildcard_domain": "*.example.com",
	})
	assert.ErrorContains(t, err, "域名和证书不匹配")
}

func TestManual_Initialize_bad_type(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	cases := map[string]any{
		"ns_prefix":       1,
		"tls_crt":         2,
		"tls_key":         3,
		"wildcard_domain": 4,
	}
	for k, v := range cases {
		m := &manualDomainManager{}
		args := map[string]any{"tls_crt": string(crt), "tls_key": string(key), "wildcard_domain": "*.example.com", k: v}
		err := m.Initialize(dmApp{logger: mlog.NewForConfig(nil)}, args)
		assert.ErrorContains(t, err, "must be string")
	}
}

func TestManual_Destroy(t *testing.T) {
	m := &manualDomainManager{logger: mlog.NewForConfig(nil)}
	assert.NoError(t, m.Destroy())
}

func TestManual_GetDomainByIndex(t *testing.T) {
	m := &manualDomainManager{nsPrefix: "devops", domainSuffix: "example.com"}
	assert.Contains(t, m.GetDomainByIndex("app", "devops-prod", 3, 0), "-3.")
}

func TestManual_GetCertSecretName_and_issuer(t *testing.T) {
	m := &manualDomainManager{}
	assert.Equal(t, ManualCertSecretName, m.GetCertSecretName("app", 1))
	assert.Empty(t, m.GetClusterIssuer())
}

func TestManual_GetCerts(t *testing.T) {
	m := &manualDomainManager{tlsKey: "k", tlsCrt: "c"}
	name, key, crt := m.GetCerts()
	assert.Equal(t, ManualCertSecretName, name)
	assert.Equal(t, "k", key)
	assert.Equal(t, "c", crt)
}

// ---------------------------------------------------------------------------
// syncsecret.go
// ---------------------------------------------------------------------------

func TestSyncSecret_Name(t *testing.T) {
	assert.Equal(t, "sync_secret_domain_manager", (&syncSecretDomainManager{}).Name())
}

func syncSecretArgs() map[string]any {
	return map[string]any{
		"ns_prefix":        "devops-",
		"secret_namespace": "default",
		"secret_name":      "tls-secret",
		"wildcard_domain":  "*.example.com",
	}
}

func TestSyncSecret_Initialize_valid(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	secret := tlsSecret("tls-secret", "default", crt, key)
	lister := newFakeSecretLister(t, secret)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().K8sClient().Return(&data.K8sClient{SecretLister: lister})

	d := &syncSecretDomainManager{}
	err := d.Initialize(dmApp{data: md, logger: mlog.NewForConfig(nil)}, syncSecretArgs())
	require.NoError(t, err)
	assert.Equal(t, "devops-", d.nsPrefix)
	assert.Equal(t, "default", d.secretNamespace)
	assert.Equal(t, "tls-secret", d.secretName)
	assert.Equal(t, "example.com", d.domainSuffix)
}

func TestSyncSecret_Initialize_missing_required(t *testing.T) {
	d := &syncSecretDomainManager{}
	err := d.Initialize(dmApp{data: data.NewMockData(gomock.NewController(t)), logger: mlog.NewForConfig(nil)}, map[string]any{})
	assert.ErrorContains(t, err, "secret_namespace, secret_name, wildcard_domain required")
}

func TestSyncSecret_Initialize_secret_not_found(t *testing.T) {
	lister := newFakeSecretLister(t, nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().K8sClient().Return(&data.K8sClient{SecretLister: lister})

	d := &syncSecretDomainManager{}
	err := d.Initialize(dmApp{data: md, logger: mlog.NewForConfig(nil)}, syncSecretArgs())
	assert.Error(t, err)
}

func TestSyncSecret_Initialize_wrong_secret_type(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "tls-secret", Namespace: "default"},
		Type:       corev1.SecretTypeOpaque, // 非 TLS 类型
		Data: map[string][]byte{
			"tls.crt": crt,
			"tls.key": key,
		},
	}
	lister := newFakeSecretLister(t, secret)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().K8sClient().Return(&data.K8sClient{SecretLister: lister})

	d := &syncSecretDomainManager{}
	err := d.Initialize(dmApp{data: md, logger: mlog.NewForConfig(nil)}, syncSecretArgs())
	assert.ErrorContains(t, err, "secret not verified")
}

func TestSyncSecret_Initialize_cert_mismatch(t *testing.T) {
	crt, key := genCert(t, []string{"*.other.com"})
	secret := tlsSecret("tls-secret", "default", crt, key)
	lister := newFakeSecretLister(t, secret)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().K8sClient().Return(&data.K8sClient{SecretLister: lister})

	d := &syncSecretDomainManager{}
	err := d.Initialize(dmApp{data: md, logger: mlog.NewForConfig(nil)}, syncSecretArgs())
	assert.ErrorContains(t, err, "域名和证书不匹配")
}

func TestSyncSecret_Initialize_bad_type(t *testing.T) {
	// 类型校验发生在 K8sClient 调用之前，故无需设置 K8sClient 期望。
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)

	for k, v := range map[string]any{"ns_prefix": 1, "secret_namespace": 2, "secret_name": 3, "wildcard_domain": 4} {
		d := &syncSecretDomainManager{}
		args := map[string]any{"secret_namespace": "default", "secret_name": "tls-secret", "wildcard_domain": "*.example.com", k: v}
		err := d.Initialize(dmApp{data: md, logger: mlog.NewForConfig(nil)}, args)
		assert.ErrorContains(t, err, "must be string")
	}
}

func TestSyncSecret_Destroy(t *testing.T) {
	d := &syncSecretDomainManager{logger: mlog.NewForConfig(nil)}
	assert.NoError(t, d.Destroy())
}

func TestSyncSecret_GetDomainByIndex(t *testing.T) {
	d := &syncSecretDomainManager{nsPrefix: "devops", domainSuffix: "example.com"}
	assert.Contains(t, d.GetDomainByIndex("app", "devops-prod", 1, 0), "example.com")
}

func TestSyncSecret_GetCertSecretName_and_issuer(t *testing.T) {
	d := &syncSecretDomainManager{}
	assert.Equal(t, SyncSecretSecretName, d.GetCertSecretName("app", 1))
	assert.Equal(t, SyncSecretSecretName, ManualCertSecretName)
	assert.Empty(t, d.GetClusterIssuer())
}

func TestSyncSecret_GetCerts_success(t *testing.T) {
	crt, key := genCert(t, []string{"*.example.com"})
	secret := tlsSecret("tls-secret", "default", crt, key)
	lister := newFakeSecretLister(t, secret)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().K8sClient().Return(&data.K8sClient{SecretLister: lister})

	d := &syncSecretDomainManager{
		data:            md,
		secretNamespace: "default",
		secretName:      "tls-secret",
		logger:          mlog.NewForConfig(nil),
	}

	name, k, c := d.GetCerts()
	assert.Equal(t, SyncSecretSecretName, name)
	assert.Equal(t, string(key), k)
	assert.Equal(t, string(crt), c)
}

func TestSyncSecret_GetCerts_read_error(t *testing.T) {
	// 空 lister：Get 失败 → 返回空三元组。
	lister := newFakeSecretLister(t, nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().K8sClient().Return(&data.K8sClient{SecretLister: lister})

	d := &syncSecretDomainManager{
		data:            md,
		secretNamespace: "default",
		secretName:      "missing",
		logger:          mlog.NewForConfig(nil),
	}

	name, k, c := d.GetCerts()
	assert.Empty(t, name)
	assert.Empty(t, k)
	assert.Empty(t, c)
}

// TestRegister_interface ensures implementations satisfy application.DomainManager.
func TestRegister_interface(t *testing.T) {
	var _ application.DomainManager = (*certManager)(nil)
	var _ application.DomainManager = (*defaultDomainManager)(nil)
	var _ application.DomainManager = (*manualDomainManager)(nil)
	var _ application.DomainManager = (*syncSecretDomainManager)(nil)
}
