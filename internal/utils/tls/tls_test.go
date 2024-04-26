package tls

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAddTlsSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	fk := fake.NewSimpleClientset()
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client: fk,
	})
	AddTlsSecret("default", "tls-secret", "key", "crt")
	sec, _ := fk.CoreV1().Secrets("default").Get(context.TODO(), "tls-secret", v1.GetOptions{})
	assert.Equal(t, &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "tls-secret",
			Namespace: "default",
			Annotations: map[string]string{
				"created-by": "mars",
			},
		},
		StringData: map[string]string{
			"tls.key": "key",
			"tls.crt": "crt",
		},
		Type: corev1.SecretTypeTLS,
	}, sec)
}

func TestUpdateCertTls(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)

	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	assert.Nil(t, db.AutoMigrate(&models.Namespace{}))

	assert.Nil(t, db.Create(&models.Namespace{Name: "ns"}).Error)
	assert.Nil(t, db.Create(&models.Namespace{Name: "ns-2"}).Error)
	assert.Nil(t, db.Create(&models.Namespace{Name: "ns-3"}).Error)
	sec := &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "ns-2",
			Name:      "cert",
		},
		StringData: map[string]string{
			"tls.key": "key-2",
			"tls.crt": "crt-2",
		},
	}
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		SecretLister: testutil.NewSecretLister(sec),
		Client:       fake.NewSimpleClientset(sec),
	}).AnyTimes()
	UpdateCertTls("cert", "key", "crt")
	s, _ := app.K8sClient().Client.CoreV1().Secrets("ns").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s.StringData["tls.key"])
	assert.Equal(t, "crt", s.StringData["tls.crt"])
	s2, _ := app.K8sClient().Client.CoreV1().Secrets("ns-2").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s2.StringData["tls.key"])
	assert.Equal(t, "crt", s2.StringData["tls.crt"])
	s3, _ := app.K8sClient().Client.CoreV1().Secrets("ns-3").Get(context.TODO(), "cert", v1.GetOptions{})
	assert.Equal(t, "key", s3.StringData["tls.key"])
	assert.Equal(t, "crt", s3.StringData["tls.crt"])
}
