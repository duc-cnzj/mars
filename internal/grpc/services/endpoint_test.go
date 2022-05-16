package services

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/endpoint"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEndpointSvc_InNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	_, err := new(EndpointSvc).InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: 123,
	})
	assert.Error(t, err)
	db.AutoMigrate(&models.Namespace{})
	_, err = new(EndpointSvc).InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: 123,
	})
	assert.NotNil(t, err)
	ns := &models.Namespace{
		Name: "duc",
	}
	db.Create(ns)
	fk := fake.NewSimpleClientset(
		&v12.IngressList{
			Items: []v12.Ingress{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "xxx",
						Labels: map[string]string{
							"app": "xxx",
						}},
					Spec: v12.IngressSpec{
						TLS: []v12.IngressTLS{
							{
								Hosts:      []string{"xxx.org"},
								SecretName: "sec2",
							},
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "yyy",
						Labels: map[string]string{
							"app.kubernetes.io/instance": "yyy",
						}},
					Spec: v12.IngressSpec{
						Rules: []v12.IngressRule{
							{
								Host: "yyy.com",
							},
							{
								Host: "zzz.com",
							},
						},
					},
				},
			},
		},
		&corev1.ServiceList{
			Items: []corev1.Service{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "svc1",
						Labels: map[string]string{
							"app.kubernetes.io/instance": "yyy",
						},
					},
					Spec: corev1.ServiceSpec{
						Type: "NodePort",
						Ports: []corev1.ServicePort{
							{
								Name:     "http",
								Protocol: "tcp",
								Port:     80,
								NodePort: 30000,
							},
						},
					},
				},
			},
		})
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client: fk,
	})
	app.EXPECT().Config().Return(&config.Config{ExternalIp: "127.0.0.1"})
	res, _ := new(EndpointSvc).InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: int64(ns.ID),
	})
	assert.Len(t, res.Items, 3)
}

func TestEndpointSvc_InProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	_, err := new(EndpointSvc).InProject(context.TODO(), &endpoint.InProjectRequest{
		ProjectId: 11,
	})
	assert.Error(t, err)
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	_, err = new(EndpointSvc).InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: 123,
	})
	assert.NotNil(t, err)
	p := &models.Project{
		Name:      "duc",
		Namespace: models.Namespace{Name: "duc"},
	}
	db.Create(p)
	fk := fake.NewSimpleClientset(
		&v12.IngressList{
			Items: []v12.Ingress{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "xxx",
						Labels: map[string]string{
							"app": "xxx",
						}},
					Spec: v12.IngressSpec{
						TLS: []v12.IngressTLS{
							{
								Hosts:      []string{"xxx.org"},
								SecretName: "sec2",
							},
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "duc",
						Labels: map[string]string{
							"app.kubernetes.io/instance": "duc",
						}},
					Spec: v12.IngressSpec{
						Rules: []v12.IngressRule{
							{
								Host: "yyy.com",
							},
							{
								Host: "zzz.com",
							},
						},
					},
				},
			},
		},
		&corev1.ServiceList{
			Items: []corev1.Service{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "svc1",
						Labels: map[string]string{
							"app.kubernetes.io/instance": "yyy",
						},
					},
					Spec: corev1.ServiceSpec{
						Type: "NodePort",
						Ports: []corev1.ServicePort{
							{
								Name:     "http",
								Protocol: "tcp",
								Port:     80,
								NodePort: 30000,
							},
						},
					},
				},
			},
		})
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client: fk,
	})
	app.EXPECT().Config().Return(&config.Config{ExternalIp: "127.0.0.1"}).AnyTimes()
	res, _ := new(EndpointSvc).InProject(context.TODO(), &endpoint.InProjectRequest{
		ProjectId: int64(p.ID),
	})
	assert.Len(t, res.Items, 2)
}
