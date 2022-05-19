package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
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
	fk := fake.NewSimpleClientset(
		&v12.IngressList{
			Items: []v12.Ingress{
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "ing1",
						Labels: map[string]string{
							"app.kubernetes.io/instance": "app1",
						}},
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
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "duc",
						Name:      "ing2",
						Labels: map[string]string{
							"app.kubernetes.io/instance": "app2",
						}},
					Spec: v12.IngressSpec{
						TLS: []v12.IngressTLS{
							{
								Hosts:      []string{"app2.org"},
								SecretName: "sec2",
							},
						},
					},
				},
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
	)
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client: fk,
	})
	mapping := GetIngressMappingByNamespace("duc")
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

func TestGetNodePortMappingByNamespace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	fk := fake.NewSimpleClientset(&corev1.ServiceList{
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
			},
		},
	})
	app.EXPECT().K8sClient().AnyTimes().Return(&contracts.K8sClient{
		Client: fk,
	})
	app.EXPECT().Config().Return(&config.Config{ExternalIp: "127.0.0.1"})
	mapping := GetNodePortMappingByNamespace("duc")
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
