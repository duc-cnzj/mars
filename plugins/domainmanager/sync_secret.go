package domainmanager

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/namespace"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

var _ application.DomainManager = (*syncSecretDomainManager)(nil)

// SyncSecretSecretName 和 manual 方式保持名称一致，避免两种方式之间切换时需要手动部署才能生效的问题
const SyncSecretSecretName = ManualCertSecretName

func init() {
	dr := &syncSecretDomainManager{
		updateCertTlsFunc: UpdateCertTls,
	}
	application.RegisterPlugin(dr.Name(), dr)
}

type syncSecretDomainManager struct {
	updateCertTlsFunc func(db *ent.Client, k8sCli *data.K8sClient, logger mlog.Logger, secretName, tlsKey, tlsCrt string)
	nsPrefix          string
	wildcardDomain    string
	domainSuffix      string

	secretNamespace string
	secretName      string

	k8sCli *data.K8sClient
	db     *ent.Client
	logger mlog.Logger

	mu     sync.RWMutex
	secret *corev1.Secret
}

func (d *syncSecretDomainManager) SetSecret(s *corev1.Secret) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.secret = s
}

func (d *syncSecretDomainManager) GetSecret() *corev1.Secret {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.secret
}

func (d *syncSecretDomainManager) Name() string {
	return "sync_secret_domain_manager"
}

func (d *syncSecretDomainManager) Initialize(app application.App, args map[string]any) error {
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

	secret, err := d.k8sCli.SecretLister.Secrets(d.secretNamespace).Get(d.secretName)
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
	d.SetSecret(secret)

	d.k8sCli = app.Data().K8sClient()
	d.db = app.DB()
	d.logger = app.Logger()
	d.k8sCli.SecretInformer.AddEventHandler(d.eventHandler(d.handleSecretChange))

	d.logger.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

func (d *syncSecretDomainManager) eventHandler(updateFunc func(oldObj, newObj any)) cache.FilteringResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			sec := obj.(*corev1.Secret)
			return sec.Namespace == d.secretNamespace && sec.Name == d.secretName
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				updateFunc(nil, obj)
			},
			UpdateFunc: updateFunc,
		},
	}
}

func (d *syncSecretDomainManager) handleSecretChange(oldObj any, newObj any) {
	if oldObj == nil && newObj != nil {
		d.SetSecret(newObj.(*corev1.Secret))
		//d.updateCertTlsFunc(d.GetCerts())
		return
	}

	oldSec := oldObj.(*corev1.Secret)
	newSec := newObj.(*corev1.Secret)
	if newSec.ResourceVersion != oldSec.ResourceVersion {
		d.SetSecret(newSec)
		// 更新当前的所有 secret
		var (
			oldKey = string(oldSec.Data["tls.key"])
			newKey = string(newSec.Data["tls.key"])

			oldCrt = string(oldSec.Data["tls.crt"])
			newCrt = string(newSec.Data["tls.crt"])
		)
		if oldKey != newKey || oldCrt != newCrt {
			certs, key, crt := d.GetCerts()
			d.updateCertTlsFunc(d.db, d.k8sCli, d.logger, certs, key, crt)
		}
	}
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
	sec := d.GetSecret()
	return SyncSecretSecretName, string(sec.Data["tls.key"]), string(sec.Data["tls.crt"])
}

func AddTlsSecret(k8sCli *data.K8sClient, ns string, name string, key string, crt string) error {
	_, err := k8sCli.Client.CoreV1().Secrets(ns).Create(context.TODO(), &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				"created-by": "mars",
			},
		},
		StringData: map[string]string{
			"tls.key": key,
			"tls.crt": crt,
		},
		Type: corev1.SecretTypeTLS,
	}, metav1.CreateOptions{})
	return err
}

func UpdateCertTls(db *ent.Client, k8sCli *data.K8sClient, logger mlog.Logger, secretName, tlsKey, tlsCrt string) {
	namespaceList := db.Namespace.Query().Select(namespace.FieldID, namespace.FieldName).AllX(context.TODO())
	var changed bool
	var changedSecrets []*corev1.Secret
	for _, n := range namespaceList {
		secret, err := k8sCli.SecretLister.Secrets(n.Name).Get(secretName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Infof("[TLS]: Register secret namespace: %s, name %s.", n.Name, secretName)
				AddTlsSecret(k8sCli, n.Name, secretName, tlsKey, tlsCrt)
				continue
			}
		}
		if string(secret.Data["tls.crt"]) != tlsCrt || string(secret.Data["tls.key"]) != tlsKey {
			changed = true
			changedSecrets = append(changedSecrets, secret.DeepCopy())
		}
	}
	if changed {
		sdata := map[string]string{
			"tls.key": tlsKey,
			"tls.crt": tlsCrt,
		}
		logger.Warning("[TLS]: certs changed, updating...")
		for _, secret := range changedSecrets {
			secret.StringData = sdata
			_, err := k8sCli.Client.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
			if err == nil {
				logger.Infof("[TLS]: namespace: %s, name %s updated", secret.Namespace, secret.Name)
			}
		}
	}
}
