package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
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

func TestCreateDockerSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	fk := fake.NewSimpleClientset()
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client: fk,
	})
	secret, _ := CreateDockerSecret("default", "name", "pwd", "1@q.c", "mars.io")
	dockercfgAuth := DockerConfigEntry{
		Username: "name",
		Password: "pwd",
		Email:    "1@q.c",
		Auth:     base64.StdEncoding.EncodeToString([]byte("name:pwd")),
	}

	dockerCfgJSON := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{"mars.io": dockercfgAuth},
	}

	marshal, _ := json.Marshal(dockerCfgJSON)
	assert.Equal(t, &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace:    "default",
			GenerateName: "mars-",
		},
		Data: map[string][]byte{
			".dockerconfigjson": marshal,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}, secret)
}

func TestIsPodRunning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	fk := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "pod1",
				Namespace: "duc",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "pod2",
				Namespace: "duc",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodFailed,
			},
		},
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "pod3",
				Namespace: "duc",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodFailed,
				ContainerStatuses: []corev1.ContainerStatus{
					{
						State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{
								Reason:  "Reason",
								Message: "Message",
							},
						},
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "pod4",
				Namespace: "duc",
			},
			Status: corev1.PodStatus{
				Phase:  corev1.PodFailed,
				Reason: "Evicted",
			},
		})
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client: fk,
	})
	running, _ := IsPodRunning("duc", "pod1")
	assert.True(t, running)
	running, r := IsPodRunning("duc", "pod2")
	assert.False(t, running)
	assert.Equal(t, "pod not running.", r)
	running, r = IsPodRunning("duc", "pod3")
	assert.False(t, running)
	assert.Equal(t, "Reason Message", r)
	running, r = IsPodRunning("duc", "pod4")
	assert.False(t, running)
	assert.Equal(t, "po pod4 already evicted in namespace duc!", r)
	po, err := fk.CoreV1().Pods("duc").Get(context.TODO(), "pod4", v1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, corev1.PodFailed, po.Status.Phase)
}

func TestGetIngressMappingByNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	ing1 := v12.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: "duc",
			Name:      "ing1",
		},
		Spec: v12.IngressSpec{
			TLS: []v12.IngressTLS{
				{
					Hosts:      []string{"app1.com", "app1.io"},
					SecretName: "sec1",
				},
				{
					Hosts:      []string{"app1.org"},
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
			Name:      "ing2",
		},
		Spec: v12.IngressSpec{
			TLS: []v12.IngressTLS{
				{
					Hosts:      []string{"app2.org"},
					SecretName: "sec2",
				},
			},
		},
	}
	ing3 := v12.Ingress{
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
	ing4 := v12.Ingress{
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
	fk := fake.NewSimpleClientset(
		&v12.IngressList{
			Items: []v12.Ingress{
				ing1,
				ing2,
				ing3,
				ing4,
			},
		},
	)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client: fk,
	})
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := models.Namespace{
		Name: "duc",
	}
	db.Create(&ns)
	p1 := &models.Project{
		Name:        "app1",
		Manifest:    strings.Join(encodeToYaml(&ing1), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p1)
	p2 := &models.Project{
		Name:        "app2",
		Manifest:    strings.Join(encodeToYaml(&ing2), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p2)
	p3 := &models.Project{
		Name:        "xxx",
		NamespaceId: ns.ID,
	}
	db.Create(p3)
	p4 := &models.Project{
		Name:        "yyy",
		Manifest:    strings.Join(encodeToYaml(&ing4), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p4)
	db.Preload("Projects").First(&ns)
	mapping := GetIngressMappingByProjects(ns.Name, ns.Projects...)

	assert.Len(t, mapping["app1"], 3)
	assert.Len(t, mapping["app2"], 1)
	for _, endpoint := range mapping["app2"] {
		assert.True(t, strings.HasPrefix(endpoint.Url, "https://"))
	}
	assert.Len(t, mapping["xxx"], 0)
	assert.Len(t, mapping["yyy"], 2)
	for _, endpoint := range mapping["yyy"] {
		assert.True(t, strings.HasPrefix(endpoint.Url, "http://"))
	}
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

func TestGetNodePortMappingByNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
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
				{
					Name:     "ui",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30001,
				},
				{
					Name:     "web",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30002,
				},
				{
					Name:     "api",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30003,
				},
				{
					Name:     "grpc",
					Protocol: "tcp",
					Port:     80,
					NodePort: 30004,
				},
			},
		},
	}
	fk := fake.NewSimpleClientset(&corev1.ServiceList{
		Items: []corev1.Service{
			svc1,
		},
	})
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client: fk,
	})
	app.EXPECT().Config().Return(&config.Config{ExternalIp: "127.0.0.1"})
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.Namespace{}, &models.Project{})
	ns := models.Namespace{
		Name: "duc",
	}
	db.Create(&ns)
	p1 := &models.Project{
		Name:        "svc1",
		Manifest:    strings.Join(encodeToYaml(&svc1), "---"),
		NamespaceId: ns.ID,
	}
	db.Create(p1)
	db.Preload("Projects").First(&ns)
	mapping := GetNodePortMappingByProjects(ns.Name, ns.Projects...)
	httpCount := 0
	grpcCount := 0
	for _, endpoints := range mapping {
		for _, endpoint := range endpoints {
			if strings.HasPrefix(endpoint.Url, "http") {
				httpCount++
			}

			if strings.HasPrefix(endpoint.Url, "grpc://") {
				grpcCount++
			}
		}
	}
	assert.Equal(t, 4, httpCount)
	assert.Equal(t, 1, grpcCount)
}

func TestFilterK8sTypeFromManifest(t *testing.T) {
	data := []string{`apiVersion: v1
kind: Service
metadata:
 name: devops-misc-consul-server
 namespace: devops-aa
 labels:
   app: consul
   chart: consul-helm
   heritage: Helm
   release: devops-misc
   component: server
 annotations:
   service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
spec:
 publishNotReadyAddresses: true
 ports:
 - name: http
   port: 8500
   targetPort: 8500
`, `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: busybox
  name: busybox-56c8cc5468-fd59w
  namespace: default
spec:
  containers:
  - command:
    - sh
    - -c
    - sleep 3600;
    image: busybox:latest
    name: busybox
    resources:
      limits:
        cpu: 10m
        memory: 10Mi
      requests:
        cpu: 10m
        memory: 10Mi
`}
	res := FilterRuntimeObjectFromManifests[*corev1.Service](data)
	assert.Len(t, res, 1)
	res1 := FilterRuntimeObjectFromManifests[*corev1.Pod](data)
	assert.Len(t, res1, 1)
	res2 := FilterRuntimeObjectFromManifests[*corev1.Namespace](data)
	assert.Len(t, res2, 0)
}

func TestRuntimeObjectList_Has(t *testing.T) {
	l := RuntimeObjectList{
		&corev1.Service{
			TypeMeta: v1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "svc1",
			},
		},
		&corev1.Service{
			TypeMeta: v1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "svc2",
			},
		},
	}
	assert.True(t, l.Has(&corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc2",
		},
	}))
	assert.False(t, l.Has(&corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc3",
		},
	}))
	assert.True(t, l.Has(&corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc1",
		},
	}))
	assert.False(t, l.Has(&corev1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "svc2",
		},
	}))
}
