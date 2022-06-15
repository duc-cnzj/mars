package services

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/duc-cnzj/mars-client/v4/endpoint"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEndpointSvc_InNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	_, err := new(EndpointSvc).InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: 123,
	})
	assert.Error(t, err)
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	_, err = new(EndpointSvc).InNamespace(context.TODO(), &endpoint.InNamespaceRequest{
		NamespaceId: 123,
	})
	ing1 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "xxx",
		},
		Spec: v12.IngressSpec{
			TLS: []v12.IngressTLS{
				{
					Hosts:      []string{"xxx.org"},
					SecretName: "sec2",
				},
			},
		},
	}
	ing2 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "yyy",
		},
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
	}
	assert.NotNil(t, err)
	ns := &models.Namespace{
		Name: "duc",
	}
	db.Create(ns)
	p2 := &models.Project{
		Name:        "app2",
		Manifest:    strings.Join(encodeToYaml(&ing2), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p2)
	svc1 := corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "svc1",
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
	}
	p3 := &models.Project{
		Name:        "xxx",
		NamespaceId: ns.ID,
		Manifest:    strings.Join(encodeToYaml(&svc1), "---"),
	}
	db.Create(p3)
	fk := fake.NewSimpleClientset(
		&v12.IngressList{
			Items: []v12.Ingress{
				ing1, ing2,
			},
		},
		&corev1.ServiceList{
			Items: []corev1.Service{
				svc1,
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

func encodeToYaml(objs ...runtime.Object) []string {
	var results []string
	for _, obj := range objs {
		bf := bytes.Buffer{}
		info, _ := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
		info.Serializer.Encode(obj, &bf)
		results = append(results, bf.String())
	}
	return results
}

func TestEndpointSvc_InProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	_, err := new(EndpointSvc).InProject(context.TODO(), &endpoint.InProjectRequest{
		ProjectId: 11,
	})
	assert.Error(t, err)
	db.AutoMigrate(&models.Namespace{}, &models.Project{})

	p1 := &models.Project{Namespace: models.Namespace{Name: "app-ns"}, Name: "app"}
	assert.Nil(t, db.Create(p1).Error)
	assert.Nil(t, db.Where("`id` = ?", p1.Namespace.ID).Delete(&models.Namespace{}).Error)
	_, err = new(EndpointSvc).InProject(context.TODO(), &endpoint.InProjectRequest{
		ProjectId: int64(p1.ID),
	})
	assert.Error(t, err)
	svc1 := corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
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
	}
	ing1 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
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
	}
	ing2 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
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
	}
	p := &models.Project{
		Name:      "duc",
		Manifest:  strings.Join(encodeToYaml(&ing2), "---"),
		Namespace: models.Namespace{Name: "duc"},
	}
	db.Create(p)
	fk := fake.NewSimpleClientset(
		&v12.IngressList{
			Items: []v12.Ingress{
				ing1,
				ing2,
			},
		},
		&corev1.ServiceList{
			Items: []corev1.Service{
				svc1,
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
