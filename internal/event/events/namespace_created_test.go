package events

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
)

func TestHandleInjectTlsSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	manager := mock.NewMockDomainManager(ctrl)
	manager.EXPECT().Initialize(gomock.Any()).Times(1)
	manager.EXPECT().GetCerts().Return("name", "key", "crt")
	app.EXPECT().GetPluginByName("test_domain_plugin_driver").Return(manager)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any())
	app.EXPECT().Config().Return(&config.Config{DomainManagerPlugin: config.Plugin{
		Name: "test_domain_plugin_driver",
		Args: nil,
	}})
	k8sClient := fake.NewSimpleClientset()
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client: k8sClient,
	})
	HandleInjectTlsSecret(NamespaceCreatedData{
		NsModel: nil,
		NsK8sObj: &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		},
	}, EventNamespaceCreated)
	list, _ := k8sClient.CoreV1().Secrets("test").List(context.TODO(), metav1.ListOptions{})
	assert.Len(t, list.Items, 1)
	assert.Equal(t, "mars", list.Items[0].Annotations["created-by"])
	assert.Equal(t, "crt", list.Items[0].StringData["tls.crt"])
	assert.Equal(t, "key", list.Items[0].StringData["tls.key"])
	assert.Equal(t, "name", list.Items[0].Name)
}
