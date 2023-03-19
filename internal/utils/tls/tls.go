package tls

import (
	"context"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddTlsSecret(ns string, name string, key string, crt string) error {
	_, err := app.K8sClientSet().CoreV1().Secrets(ns).Create(context.TODO(), &v1.Secret{
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
		Type: v1.SecretTypeTLS,
	}, metav1.CreateOptions{})
	return err
}

func UpdateCertTls(secretName, tlsKey, tlsCrt string) {
	// 需要更新 tls 证书
	var namespaceList []models.Namespace
	app.DB().Select("ID", "Name").Find(&namespaceList)
	var changed bool
	var changedSecrets []*v1.Secret
	for _, n := range namespaceList {
		secret, err := app.K8sClient().SecretLister.Secrets(n.Name).Get(secretName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				mlog.Infof("[TLS]: Add secret namespace: %s, name %s.", n.Name, secretName)
				AddTlsSecret(n.Name, secretName, tlsKey, tlsCrt)
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
		mlog.Warning("[TLS]: certs changed, updating...")
		for _, secret := range changedSecrets {
			secret.StringData = sdata
			_, err := app.K8sClient().Client.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
			if err == nil {
				mlog.Infof("[TLS]: namespace: %s, name %s updated", secret.Namespace, secret.Name)
			}
		}
	}
}
