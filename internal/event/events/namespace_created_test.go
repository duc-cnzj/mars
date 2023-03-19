package events

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
)

func TestHandleInjectTlsSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	manager := mock.NewMockDomainManager(ctrl)
	manager.EXPECT().Initialize(gomock.Any()).Times(1)
	manager.EXPECT().GetCerts().AnyTimes().Return("name", "key", "crt")
	app.EXPECT().GetPluginByName("test_domain_plugin_driver").AnyTimes().Return(manager)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any())
	app.EXPECT().Config().AnyTimes().Return(&config.Config{DomainManagerPlugin: config.Plugin{
		Name: "test_domain_plugin_driver",
		Args: nil,
	}})
	k8sClient := fake.NewSimpleClientset()
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
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
	// 再次 create 会报错，测试 err 部分
	l := mock.NewMockLoggerInterface(ctrl)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Error(gomock.Any()).Times(1)
	HandleInjectTlsSecret(NamespaceCreatedData{
		NsModel: nil,
		NsK8sObj: &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		},
	}, EventNamespaceCreated)
}
