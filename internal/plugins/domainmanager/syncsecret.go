package domainmanager

import (
	"errors"
	"strings"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	corev1 "k8s.io/api/core/v1"
)

var _ application.DomainManager = (*syncSecretDomainManager)(nil)

// SyncSecretSecretName 和 manual 方式保持名称一致，避免两种方式之间切换时需要手动部署才能生效的问题
const SyncSecretSecretName = ManualCertSecretName

func init() {
	dr := &syncSecretDomainManager{}
	application.RegisterPlugin(dr.Name(), dr)
}

type syncSecretDomainManager struct {
	nsPrefix       string
	wildcardDomain string
	domainSuffix   string

	secretNamespace string
	secretName      string

	data   data.Data
	db     *ent.Client
	logger mlog.Logger
}

const SyncSecretDomainManager = "sync_secret_domain_manager"

func (d *syncSecretDomainManager) Name() string {
	return SyncSecretDomainManager
}

func (d *syncSecretDomainManager) Initialize(app application.App, args map[string]any) error {
	d.data = app.Data()
	d.db = app.DB()
	d.logger = app.Logger()

	if p, ok := args["ns_prefix"]; ok {
		d.nsPrefix = p.(string)
	}

	if p, ok := args["secret_namespace"]; ok {
		d.secretNamespace = p.(string)
	}

	if p, ok := args["secret_name"]; ok {
		d.secretName = p.(string)
	}

	if wd, ok := args["wildcard_domain"]; ok {
		d.wildcardDomain = wd.(string)
		d.domainSuffix = strings.TrimLeft(d.wildcardDomain, "*.")
	}

	if d.secretNamespace == "" || d.secretName == "" || d.wildcardDomain == "" {
		return errors.New("secret_namespace, secret_name, wildcard_domain required")
	}

	secret, err := d.data.K8sClient().SecretLister.Secrets(d.secretNamespace).Get(d.secretName)
	if err != nil {
		return err
	}

	if secret.Type != corev1.SecretTypeTLS {
		return errors.New("secret not verified")
	}

	var (
		tlsKey = secret.Data["tls.key"]
		tlsCrt = secret.Data["tls.crt"]
	)
	err = validateTelsWildcardDomain(tlsKey, tlsCrt, d.wildcardDomain)
	if err != nil {
		return err
	}

	d.logger.Info("[Plugin]: " + d.Name() + " plugin Initialize...")

	return nil
}

func (d *syncSecretDomainManager) Destroy() error {
	d.logger.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

func (d *syncSecretDomainManager) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        index,
		nsPrefix:     d.nsPrefix,
		domainSuffix: d.domainSuffix,
	}.SubStr()
}

func (d *syncSecretDomainManager) GetDomain(projectName, namespace string, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     d.nsPrefix,
		domainSuffix: d.domainSuffix,
	}.SubStr()
}

func (d *syncSecretDomainManager) GetCertSecretName(projectName string, index int) string {
	return SyncSecretSecretName
}

func (d *syncSecretDomainManager) GetClusterIssuer() string {
	return ""
}

func (d *syncSecretDomainManager) GetCerts() (name, key, crt string) {
	sec, err := d.data.K8sClient().SecretLister.Secrets(d.secretNamespace).Get(d.secretName)
	if err != nil {
		d.logger.Error("[TLS]: get secret error: ", err)
		return "", "", ""
	}
	return SyncSecretSecretName, string(sec.Data["tls.key"]), string(sec.Data["tls.crt"])
}
